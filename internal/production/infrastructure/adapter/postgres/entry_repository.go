package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

type EntryRepository struct {
	pool *pgxpool.Pool
}

func NewEntryRepository(pool *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{pool: pool}
}

func (repository *EntryRepository) Create(ctx context.Context, entry model.Entry) (err error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin entry receipt: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); err == nil && rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = fmt.Errorf("rollback entry receipt: %w", rollbackErr)
		}
	}()

	if _, err = tx.Exec(ctx, `
		INSERT INTO production_entries (id, received_at)
		VALUES ($1, $2)
	`, entry.ID(), entry.ReceivedAt()); err != nil {
		return fmt.Errorf("insert production entry: %w", err)
	}

	for _, batch := range entry.ColorBatches() {
		if _, err = tx.Exec(ctx, `
			INSERT INTO production_color_batches (
				id, entry_id, color_name, color_key, position, status, created_at, started_at, completed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, batch.ID(), batch.EntryID(), batch.Color().Name(), batch.Color().Key(), batch.Position(), string(batch.Status()), batch.CreatedAt(), batch.StartedAt(), batch.CompletedAt()); err != nil {
			return fmt.Errorf("insert color batch: %w", err)
		}

		for _, run := range batch.SizeRuns() {
			if _, err = tx.Exec(ctx, `
				INSERT INTO production_size_runs (
					id, color_batch_id, size_name, position, status, quantity, started_at, completed_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, run.ID(), run.ColorBatchID(), run.SizeName(), run.Position(), string(run.Status()), run.Quantity(), run.StartedAt(), run.CompletedAt()); err != nil {
				return fmt.Errorf("insert size run: %w", err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit entry receipt: %w", err)
	}
	return nil
}
