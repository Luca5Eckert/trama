package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/platform/database"
	"github.com/Luca5Eckert/trama/internal/production/domain/receipt"
	productionpostgres "github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/postgres"
)

type repositoryIDs struct {
	values []string
	index int
}

func (ids *repositoryIDs) NewID() (string, error) {
	if ids.index >= len(ids.values) {
		return "", errors.New("no id available")
	}
	value := ids.values[ids.index]
	ids.index++
	return value, nil
}

func TestEntryRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM production_entries`); err != nil {
		t.Fatalf("clean entries: %v", err)
	}
	defer func() { _, _ = pool.Exec(context.Background(), `DELETE FROM production_entries`) }()

	repository := productionpostgres.NewEntryRepository(pool)
	colors, _ := receipt.ParseColors([]string{"Preto", "Azul"})
	ids := &repositoryIDs{values: []string{
		"entry-integration",
		"batch-preto", "run-preto-2", "run-preto-4",
		"batch-azul", "run-azul-2", "run-azul-4",
	}}
	entry, err := receipt.Receive(time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC), colors, sequence(t,
		struct{name string; position int}{"4", 20},
		struct{name string; position int}{"2", 10},
	), ids)
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if err := repository.Create(ctx, entry); err != nil {
		t.Fatalf("create: %v", err)
	}

	var entryCount, batchCount, runCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM production_entries WHERE id = $1`, entry.ID()).Scan(&entryCount); err != nil {
		t.Fatalf("count entry: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM production_color_batches WHERE entry_id = $1`, entry.ID()).Scan(&batchCount); err != nil {
		t.Fatalf("count batches: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM production_size_runs r
		JOIN production_color_batches b ON b.id = r.color_batch_id
		WHERE b.entry_id = $1
	`, entry.ID()).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if entryCount != 1 || batchCount != 2 || runCount != 4 {
		t.Fatalf("counts entry=%d batches=%d runs=%d", entryCount, batchCount, runCount)
	}

	rows, err := pool.Query(ctx, `
		SELECT r.size_name, r.position, r.quantity
		FROM production_size_runs r
		JOIN production_color_batches b ON b.id = r.color_batch_id
		WHERE b.id = $1
		ORDER BY r.position
	`, "batch-preto")
	if err != nil {
		t.Fatalf("query snapshot: %v", err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		var position int
		var quantity *int
		if err := rows.Scan(&name, &position, &quantity); err != nil {
			t.Fatalf("scan snapshot: %v", err)
		}
		if quantity != nil {
			t.Fatalf("quantity = %v, want nil", *quantity)
		}
		names = append(names, name)
	}
	if len(names) != 2 || names[0] != "2" || names[1] != "4" {
		t.Fatalf("snapshot = %#v", names)
	}

	// Changing the global configuration cannot mutate the persisted snapshot.
	sequenceRepository := productionpostgres.NewSizeSequenceRepository(pool)
	_, err = sequenceRepository.Replace(ctx, sequence(t, struct{name string; position int}{"6", 10}), time.Now().UTC())
	if err != nil {
		t.Fatalf("replace global sequence: %v", err)
	}
	var persistedName string
	if err := pool.QueryRow(ctx, `
		SELECT r.size_name
		FROM production_size_runs r
		WHERE r.color_batch_id = $1
		ORDER BY r.position LIMIT 1
	`, "batch-preto").Scan(&persistedName); err != nil {
		t.Fatalf("read persisted snapshot: %v", err)
	}
	if persistedName != "2" {
		t.Fatalf("snapshot changed to %q", persistedName)
	}

	// Duplicate batch IDs force an intermediate insert failure; the whole receipt must roll back.
	rollbackColors, _ := receipt.ParseColors([]string{"Rosa", "Verde"})
	rollbackIDs := &repositoryIDs{values: []string{
		"entry-rollback",
		"batch-duplicate", "run-rollback-1",
		"batch-duplicate", "run-rollback-2",
	}}
	rollbackEntry, err := receipt.Receive(time.Now(), rollbackColors, sequence(t, struct{name string; position int}{"2", 10}), rollbackIDs)
	if err != nil {
		t.Fatalf("build rollback entry: %v", err)
	}
	if err := repository.Create(ctx, rollbackEntry); err == nil {
		t.Fatal("expected intermediate insert failure")
	}
	var rollbackCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM production_entries WHERE id = $1`, "entry-rollback").Scan(&rollbackCount); err != nil {
		t.Fatalf("count rollback entry: %v", err)
	}
	if rollbackCount != 0 {
		t.Fatalf("rollback entry count = %d, want 0", rollbackCount)
	}
}
