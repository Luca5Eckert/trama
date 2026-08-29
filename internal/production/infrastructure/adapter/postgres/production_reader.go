package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
	"github.com/Luca5Eckert/trama/internal/production/domain/snapshot"
)

type ProductionReader struct{ pool *pgxpool.Pool }

func NewProductionReader(pool *pgxpool.Pool) *ProductionReader { return &ProductionReader{pool: pool} }

func (reader *ProductionReader) GetByID(ctx context.Context, id string) (snapshot.Entry, error) {
	var entry snapshot.Entry
	if err := reader.pool.QueryRow(ctx, `SELECT id, received_at FROM production_entries WHERE id = $1`, id).Scan(&entry.ID, &entry.ReceivedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return snapshot.Entry{}, domain.ErrEntryNotFound }
		return snapshot.Entry{}, fmt.Errorf("select entry: %w", err)
	}
	batches, err := reader.listBatchesForEntry(ctx, id)
	if err != nil { return snapshot.Entry{}, err }
	entry.ReceivedAt = entry.ReceivedAt.UTC()
	entry.ColorBatches = batches
	return entry, nil
}

func (reader *ProductionReader) List(ctx context.Context, page port.PageCriteria) ([]snapshot.EntrySummary, error) {
	rows, err := reader.pool.Query(ctx, `
		SELECT e.id, e.received_at, count(b.id)::int
		FROM production_entries e
		LEFT JOIN production_color_batches b ON b.entry_id = e.id
		GROUP BY e.id, e.received_at
		ORDER BY e.received_at DESC, e.id DESC
		LIMIT $1 OFFSET $2
	`, page.Limit, page.Offset)
	if err != nil { return nil, fmt.Errorf("list entries: %w", err) }
	defer rows.Close()
	result := make([]snapshot.EntrySummary, 0)
	for rows.Next() {
		var item snapshot.EntrySummary
		if err := rows.Scan(&item.ID, &item.ReceivedAt, &item.ColorBatchCount); err != nil { return nil, fmt.Errorf("scan entry summary: %w", err) }
		item.ReceivedAt = item.ReceivedAt.UTC()
		result = append(result, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate entries: %w", err) }
	return result, nil
}

func (reader *ProductionReader) GetColorBatchByID(ctx context.Context, id string) (snapshot.ColorBatch, error) {
	batch, err := scanBatch(reader.pool.QueryRow(ctx, batchSelect+` WHERE b.id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return snapshot.ColorBatch{}, domain.ErrColorBatchNotFound }
		return snapshot.ColorBatch{}, fmt.Errorf("select color batch: %w", err)
	}
	runs, err := reader.listSizeRuns(ctx, id)
	if err != nil { return snapshot.ColorBatch{}, err }
	batch.SizeRuns = runs
	return batch, nil
}

func (reader *ProductionReader) ListColorBatches(ctx context.Context, criteria port.ColorBatchListCriteria) ([]snapshot.ColorBatch, error) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if criteria.Status != nil {
		args = append(args, string(*criteria.Status))
		clauses = append(clauses, fmt.Sprintf("b.status = $%d", len(args)))
	}
	if criteria.EntryID != nil {
		args = append(args, *criteria.EntryID)
		clauses = append(clauses, fmt.Sprintf("b.entry_id = $%d", len(args)))
	}
	queryText := batchSelect
	if len(clauses) > 0 { queryText += " WHERE " + strings.Join(clauses, " AND ") }
	args = append(args, criteria.Page.Limit)
	queryText += fmt.Sprintf(" ORDER BY b.created_at ASC, b.entry_id ASC, b.position ASC, b.id ASC LIMIT $%d", len(args))
	args = append(args, criteria.Page.Offset)
	queryText += fmt.Sprintf(" OFFSET $%d", len(args))

	rows, err := reader.pool.Query(ctx, queryText, args...)
	if err != nil { return nil, fmt.Errorf("list color batches: %w", err) }
	defer rows.Close()
	result := make([]snapshot.ColorBatch, 0)
	for rows.Next() {
		batch, err := scanBatch(rows)
		if err != nil { return nil, fmt.Errorf("scan color batch: %w", err) }
		result = append(result, batch)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate color batches: %w", err) }
	return result, nil
}

func (reader *ProductionReader) listBatchesForEntry(ctx context.Context, entryID string) ([]snapshot.ColorBatch, error) {
	rows, err := reader.pool.Query(ctx, batchSelect+` WHERE b.entry_id = $1 ORDER BY b.position ASC, b.id ASC`, entryID)
	if err != nil { return nil, fmt.Errorf("select entry batches: %w", err) }
	defer rows.Close()
	result := make([]snapshot.ColorBatch, 0)
	for rows.Next() {
		batch, err := scanBatch(rows)
		if err != nil { return nil, fmt.Errorf("scan entry batch: %w", err) }
		result = append(result, batch)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate entry batches: %w", err) }
	return result, nil
}

func (reader *ProductionReader) listSizeRuns(ctx context.Context, batchID string) ([]snapshot.SizeRun, error) {
	rows, err := reader.pool.Query(ctx, `
		SELECT id, size_name, position, status, quantity, started_at, completed_at
		FROM production_size_runs
		WHERE color_batch_id = $1
		ORDER BY position ASC, id ASC
	`, batchID)
	if err != nil { return nil, fmt.Errorf("select size runs: %w", err) }
	defer rows.Close()
	result := make([]snapshot.SizeRun, 0)
	for rows.Next() {
		var item snapshot.SizeRun
		var status string
		var quantity pgtype.Int4
		var startedAt, completedAt pgtype.Timestamptz
		if err := rows.Scan(&item.ID, &item.SizeName, &item.Position, &status, &quantity, &startedAt, &completedAt); err != nil {
			return nil, fmt.Errorf("scan size run: %w", err)
		}
		item.Status = model.SizeRunStatus(status)
		if quantity.Valid { value := int(quantity.Int32); item.Quantity = &value }
		if startedAt.Valid { value := startedAt.Time.UTC(); item.StartedAt = &value }
		if completedAt.Valid { value := completedAt.Time.UTC(); item.CompletedAt = &value }
		result = append(result, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate size runs: %w", err) }
	return result, nil
}

type scanner interface { Scan(...any) error }

func scanBatch(row scanner) (snapshot.ColorBatch, error) {
	var batch snapshot.ColorBatch
	var status string
	var startedAt, completedAt pgtype.Timestamptz
	var currentSize, nextSize pgtype.Text
	if err := row.Scan(
		&batch.ID, &batch.EntryID, &batch.Color, &batch.Position, &status, &batch.CreatedAt,
		&startedAt, &completedAt, &currentSize, &nextSize,
	); err != nil { return snapshot.ColorBatch{}, err }
	batch.Status = model.ColorBatchStatus(status)
	batch.CreatedAt = batch.CreatedAt.UTC()
	if startedAt.Valid { value := startedAt.Time.UTC(); batch.StartedAt = &value }
	if completedAt.Valid { value := completedAt.Time.UTC(); batch.CompletedAt = &value }
	if currentSize.Valid { value := currentSize.String; batch.CurrentSize = &value }
	if nextSize.Valid { value := nextSize.String; batch.NextSize = &value }
	return batch, nil
}

const batchSelect = `
	SELECT b.id, b.entry_id, b.color_name, b.position, b.status, b.created_at, b.started_at, b.completed_at,
		CASE
			WHEN b.status = 'IN_PRODUCTION' THEN (
				SELECT sr.size_name FROM production_size_runs sr
				WHERE sr.color_batch_id = b.id AND sr.status = 'IN_PROGRESS'
				ORDER BY sr.position ASC, sr.id ASC LIMIT 1
			)
			ELSE NULL
		END AS current_size,
		CASE
			WHEN b.status = 'COMPLETED' THEN NULL
			ELSE (
				SELECT sr.size_name FROM production_size_runs sr
				WHERE sr.color_batch_id = b.id AND sr.status = 'PENDING'
				ORDER BY sr.position ASC, sr.id ASC LIMIT 1
			)
		END AS next_size
	FROM production_color_batches b`

var _ port.EntryReader = (*ProductionReader)(nil)
