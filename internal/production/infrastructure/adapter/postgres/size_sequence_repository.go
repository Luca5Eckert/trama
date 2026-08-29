package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

const singletonID int16 = 1

type SizeSequenceRepository struct {
	pool *pgxpool.Pool
}

func NewSizeSequenceRepository(pool *pgxpool.Pool) *SizeSequenceRepository {
	return &SizeSequenceRepository{pool: pool}
}

func (repository *SizeSequenceRepository) Get(ctx context.Context) (port.StoredSizeSequence, error) {
	var updatedAt time.Time
	if err := repository.pool.QueryRow(ctx, `SELECT updated_at FROM production_size_sequence WHERE singleton_id = $1`, singletonID).Scan(&updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return port.StoredSizeSequence{}, domain.ErrSizeSequenceNotConfigured
		}
		return port.StoredSizeSequence{}, fmt.Errorf("select size sequence header: %w", err)
	}

	sequence, err := loadSequence(ctx, repository.pool, singletonID)
	if err != nil {
		return port.StoredSizeSequence{}, err
	}
	return port.StoredSizeSequence{Sequence: sequence, UpdatedAt: updatedAt.UTC()}, nil
}

func (repository *SizeSequenceRepository) Replace(ctx context.Context, sequence model.SizeSequence, updatedAt time.Time) (stored port.StoredSizeSequence, err error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return port.StoredSizeSequence{}, fmt.Errorf("begin size sequence replace: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); err == nil && rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = fmt.Errorf("rollback size sequence replace: %w", rollbackErr)
		}
	}()

	if _, err = tx.Exec(ctx, `
		INSERT INTO production_size_sequence (singleton_id, updated_at)
		VALUES ($1, $2)
		ON CONFLICT (singleton_id) DO NOTHING
	`, singletonID, updatedAt.UTC()); err != nil {
		return port.StoredSizeSequence{}, fmt.Errorf("ensure size sequence header: %w", err)
	}

	var persistedUpdatedAt time.Time
	if err = tx.QueryRow(ctx, `SELECT updated_at FROM production_size_sequence WHERE singleton_id = $1 FOR UPDATE`, singletonID).Scan(&persistedUpdatedAt); err != nil {
		return port.StoredSizeSequence{}, fmt.Errorf("lock size sequence header: %w", err)
	}

	existing, hasItems, loadErr := loadSequenceIfPresent(ctx, tx, singletonID)
	if loadErr != nil {
		return port.StoredSizeSequence{}, loadErr
	}
	if hasItems && existing.Equal(sequence) {
		if err = tx.Commit(ctx); err != nil {
			return port.StoredSizeSequence{}, fmt.Errorf("commit unchanged size sequence: %w", err)
		}
		return port.StoredSizeSequence{Sequence: existing, UpdatedAt: persistedUpdatedAt.UTC()}, nil
	}

	if _, err = tx.Exec(ctx, `UPDATE production_size_sequence SET updated_at = $2 WHERE singleton_id = $1`, singletonID, updatedAt.UTC()); err != nil {
		return port.StoredSizeSequence{}, fmt.Errorf("update size sequence header: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM production_size_sequence_items WHERE sequence_id = $1`, singletonID); err != nil {
		return port.StoredSizeSequence{}, fmt.Errorf("delete old size sequence items: %w", err)
	}
	for _, item := range sequence.Items() {
		if _, err = tx.Exec(ctx, `
			INSERT INTO production_size_sequence_items (sequence_id, name, position)
			VALUES ($1, $2, $3)
		`, singletonID, item.Name(), item.Position()); err != nil {
			return port.StoredSizeSequence{}, fmt.Errorf("insert size sequence item: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return port.StoredSizeSequence{}, fmt.Errorf("commit size sequence replace: %w", err)
	}
	return port.StoredSizeSequence{Sequence: sequence, UpdatedAt: updatedAt.UTC()}, nil
}

type rowQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func loadSequence(ctx context.Context, querier rowQuerier, sequenceID int16) (model.SizeSequence, error) {
	sequence, hasItems, err := loadSequenceIfPresent(ctx, querier, sequenceID)
	if err != nil {
		return model.SizeSequence{}, err
	}
	if !hasItems {
		return model.SizeSequence{}, errors.New("persisted size sequence has no items")
	}
	return sequence, nil
}

func loadSequenceIfPresent(ctx context.Context, querier rowQuerier, sequenceID int16) (model.SizeSequence, bool, error) {
	rows, err := querier.Query(ctx, `
		SELECT name, position
		FROM production_size_sequence_items
		WHERE sequence_id = $1
		ORDER BY position ASC
	`, sequenceID)
	if err != nil {
		return model.SizeSequence{}, false, fmt.Errorf("select size sequence items: %w", err)
	}
	defer rows.Close()

	definitions := make([]model.SizeDefinition, 0)
	for rows.Next() {
		var name string
		var position int
		if err := rows.Scan(&name, &position); err != nil {
			return model.SizeSequence{}, false, fmt.Errorf("scan size sequence item: %w", err)
		}
		definition, err := model.NewSizeDefinition(name, position)
		if err != nil {
			return model.SizeSequence{}, false, fmt.Errorf("invalid persisted size definition: %v", err)
		}
		definitions = append(definitions, definition)
	}
	if err := rows.Err(); err != nil {
		return model.SizeSequence{}, false, fmt.Errorf("iterate size sequence items: %w", err)
	}
	if len(definitions) == 0 {
		return model.SizeSequence{}, false, nil
	}
	sequence, err := model.NewSizeSequence(definitions)
	if err != nil {
		return model.SizeSequence{}, false, fmt.Errorf("invalid persisted size sequence: %v", err)
	}
	return sequence, true, nil
}
