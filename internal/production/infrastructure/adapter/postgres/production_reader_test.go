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
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
	productionpostgres "github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/postgres"
)

func TestProductionReaderIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil { t.Fatalf("create pool: %v", err) }
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil { t.Fatalf("migrate: %v", err) }

	cleanup := func() { _, _ = pool.Exec(context.Background(), `DELETE FROM production_entries WHERE id LIKE 'read-%'`) }
	cleanup()
	defer cleanup()

	oldTime := time.Date(2026, 8, 29, 18, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	if _, err := pool.Exec(ctx, `
		INSERT INTO production_entries (id, received_at) VALUES
		('read-entry-a', $1),
		('read-entry-b', $2)
	`, oldTime, newTime); err != nil { t.Fatalf("insert entries: %v", err) }

	if _, err := pool.Exec(ctx, `
		INSERT INTO production_color_batches
		(id, entry_id, color_name, color_key, position, status, created_at, started_at, completed_at)
		VALUES
		('read-batch-a1', 'read-entry-a', 'Preto', 'preto', 1, 'IN_PRODUCTION', $1, $2, NULL),
		('read-batch-a2', 'read-entry-a', 'Azul', 'azul', 2, 'WAITING', $1, NULL, NULL),
		('read-batch-b1', 'read-entry-b', 'Branco', 'branco', 1, 'COMPLETED', $3, $3, $4)
	`, oldTime.Add(time.Minute), oldTime.Add(2*time.Minute), newTime.Add(time.Minute), newTime.Add(20*time.Minute)); err != nil {
		t.Fatalf("insert batches: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO production_size_runs
		(id, color_batch_id, size_name, position, status, quantity, started_at, completed_at)
		VALUES
		('read-run-a1-p', 'read-batch-a1', 'P', 10, 'COMPLETED', NULL, $1, $2),
		('read-run-a1-m', 'read-batch-a1', 'M', 20, 'IN_PROGRESS', NULL, $3, NULL),
		('read-run-a1-g', 'read-batch-a1', 'G', 30, 'PENDING', NULL, NULL, NULL),
		('read-run-a2-2', 'read-batch-a2', '2', 10, 'PENDING', NULL, NULL, NULL),
		('read-run-a2-4', 'read-batch-a2', '4', 20, 'PENDING', NULL, NULL, NULL),
		('read-run-b1-p', 'read-batch-b1', 'P', 10, 'COMPLETED', 12, $4, $5)
	`, oldTime.Add(2*time.Minute), oldTime.Add(10*time.Minute), oldTime.Add(11*time.Minute), newTime.Add(time.Minute), newTime.Add(20*time.Minute)); err != nil {
		t.Fatalf("insert runs: %v", err)
	}

	reader := productionpostgres.NewProductionReader(pool)

	entry, err := reader.GetByID(ctx, "read-entry-a")
	if err != nil { t.Fatalf("get entry: %v", err) }
	if len(entry.ColorBatches) != 2 || entry.ColorBatches[0].ID != "read-batch-a1" || entry.ColorBatches[1].ID != "read-batch-a2" {
		t.Fatalf("entry batches order = %#v", entry.ColorBatches)
	}
	if entry.ColorBatches[0].CurrentSize == nil || *entry.ColorBatches[0].CurrentSize != "M" || entry.ColorBatches[0].NextSize == nil || *entry.ColorBatches[0].NextSize != "G" {
		t.Fatalf("in-production projection = %#v", entry.ColorBatches[0])
	}
	if entry.ColorBatches[1].CurrentSize != nil || entry.ColorBatches[1].NextSize == nil || *entry.ColorBatches[1].NextSize != "2" {
		t.Fatalf("waiting projection = %#v", entry.ColorBatches[1])
	}

	entries, err := reader.List(ctx, port.PageCriteria{Limit: 1, Offset: 0})
	if err != nil { t.Fatalf("list entries: %v", err) }
	if len(entries) != 1 || entries[0].ID != "read-entry-b" || entries[0].ColorBatchCount != 1 {
		t.Fatalf("entry page 1 = %#v", entries)
	}
	entries, err = reader.List(ctx, port.PageCriteria{Limit: 1, Offset: 1})
	if err != nil { t.Fatalf("list entries page 2: %v", err) }
	if len(entries) != 1 || entries[0].ID != "read-entry-a" || entries[0].ColorBatchCount != 2 {
		t.Fatalf("entry page 2 = %#v", entries)
	}

	batch, err := reader.GetColorBatchByID(ctx, "read-batch-a1")
	if err != nil { t.Fatalf("get batch: %v", err) }
	if batch.CurrentSize == nil || *batch.CurrentSize != "M" || batch.NextSize == nil || *batch.NextSize != "G" {
		t.Fatalf("batch sizes = %#v", batch)
	}
	if len(batch.SizeRuns) != 3 || batch.SizeRuns[0].SizeName != "P" || batch.SizeRuns[1].SizeName != "M" || batch.SizeRuns[2].SizeName != "G" {
		t.Fatalf("runs order = %#v", batch.SizeRuns)
	}
	if batch.SizeRuns[1].Quantity != nil { t.Fatalf("quantity should stay nil: %#v", batch.SizeRuns[1].Quantity) }

	waiting := model.ColorBatchWaiting
	batches, err := reader.ListColorBatches(ctx, port.ColorBatchListCriteria{Status: &waiting, Page: port.PageCriteria{Limit: 50}})
	if err != nil { t.Fatalf("filter status: %v", err) }
	if len(batches) != 1 || batches[0].ID != "read-batch-a2" { t.Fatalf("status filter = %#v", batches) }

	entryID := "read-entry-a"
	batches, err = reader.ListColorBatches(ctx, port.ColorBatchListCriteria{EntryID: &entryID, Page: port.PageCriteria{Limit: 50}})
	if err != nil { t.Fatalf("filter entry: %v", err) }
	if len(batches) != 2 || batches[0].ID != "read-batch-a1" || batches[1].ID != "read-batch-a2" { t.Fatalf("entry filter/order = %#v", batches) }

	batches, err = reader.ListColorBatches(ctx, port.ColorBatchListCriteria{Page: port.PageCriteria{Limit: 1, Offset: 1}})
	if err != nil { t.Fatalf("batch pagination: %v", err) }
	if len(batches) != 1 || batches[0].ID != "read-batch-a2" { t.Fatalf("batch page = %#v", batches) }

	completed := model.ColorBatchCompleted
	batches, err = reader.ListColorBatches(ctx, port.ColorBatchListCriteria{Status: &completed, Page: port.PageCriteria{Limit: 50}})
	if err != nil { t.Fatalf("completed filter: %v", err) }
	if len(batches) != 1 || batches[0].CurrentSize != nil || batches[0].NextSize != nil { t.Fatalf("completed projection = %#v", batches) }

	if _, err := reader.GetByID(ctx, "read-missing"); !errors.Is(err, domain.ErrEntryNotFound) { t.Fatalf("missing entry error = %v", err) }
	if _, err := reader.GetColorBatchByID(ctx, "read-missing"); !errors.Is(err, domain.ErrColorBatchNotFound) { t.Fatalf("missing batch error = %v", err) }
}
