package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/platform/database"
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	productionpostgres "github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/postgres"
)

func sequence(t *testing.T, items ...struct {
	name     string
	position int
}) model.SizeSequence {
	t.Helper()
	definitions := make([]model.SizeDefinition, len(items))
	for index, item := range items {
		definition, err := model.NewSizeDefinition(item.name, item.position)
		if err != nil {
			t.Fatalf("definition: %v", err)
		}
		definitions[index] = definition
	}
	result, err := model.NewSizeSequence(definitions)
	if err != nil {
		t.Fatalf("sequence: %v", err)
	}
	return result
}

func TestSizeSequenceRepositoryIntegration(t *testing.T) {
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
	if _, err := pool.Exec(ctx, `DELETE FROM production_size_sequence`); err != nil {
		t.Fatalf("clean sequence: %v", err)
	}
	defer func() { _, _ = pool.Exec(context.Background(), `DELETE FROM production_size_sequence`) }()

	repository := productionpostgres.NewSizeSequenceRepository(pool)
	if _, err := repository.Get(ctx); !errors.Is(err, domain.ErrSizeSequenceNotConfigured) {
		t.Fatalf("initial Get error = %v", err)
	}

	firstTime := time.Date(2026, 8, 29, 22, 0, 0, 0, time.UTC)
	firstSequence := sequence(t,
		struct {
			name     string
			position int
		}{"M", 20},
		struct {
			name     string
			position int
		}{"P", 10},
	)
	first, err := repository.Replace(ctx, firstSequence, firstTime)
	if err != nil {
		t.Fatalf("first replace: %v", err)
	}
	if !first.UpdatedAt.Equal(firstTime) {
		t.Fatalf("first updatedAt = %v", first.UpdatedAt)
	}

	secondTime := firstTime.Add(time.Hour)
	unchanged, err := repository.Replace(ctx, firstSequence, secondTime)
	if err != nil {
		t.Fatalf("idempotent replace: %v", err)
	}
	if !unchanged.UpdatedAt.Equal(firstTime) {
		t.Fatalf("idempotent replace changed updatedAt to %v", unchanged.UpdatedAt)
	}
	var itemCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM production_size_sequence_items`).Scan(&itemCount); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if itemCount != 2 {
		t.Fatalf("item count = %d, want 2", itemCount)
	}

	replacement := sequence(t,
		struct {
			name     string
			position int
		}{"GG", 40},
		struct {
			name     string
			position int
		}{"P", 10},
	)
	replaced, err := repository.Replace(ctx, replacement, secondTime)
	if err != nil {
		t.Fatalf("replacement: %v", err)
	}
	if !replaced.UpdatedAt.Equal(secondTime) {
		t.Fatalf("replacement updatedAt = %v", replaced.UpdatedAt)
	}

	found, err := repository.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	items := found.Sequence.Items()
	if items[0].Name() != "P" || items[1].Name() != "GG" {
		t.Fatalf("unexpected persisted order: %q, %q", items[0].Name(), items[1].Name())
	}
}
