package application

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain/snapshot"
)

type PageInfoResult struct {
	Limit  int
	Offset int
	Count  int
}

type ReadSizeRunResult struct {
	ID          string
	Size        string
	Position    int
	Status      string
	Quantity    *int
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type ReadColorBatchResult struct {
	ID          string
	EntryID     string
	Color       string
	Position    int
	Status      string
	CurrentSize *string
	NextSize    *string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	SizeRuns    []ReadSizeRunResult
}

type EntryDetailResult struct {
	ID           string
	ReceivedAt   time.Time
	ColorBatches []ReadColorBatchResult
}

type EntrySummaryResult struct {
	ID              string
	ReceivedAt      time.Time
	ColorBatchCount int
}

type EntryListResult struct {
	Items []EntrySummaryResult
	Page  PageInfoResult
}

type ColorBatchListResult struct {
	Items []ReadColorBatchResult
	Page  PageInfoResult
}

func NewEntryDetailResult(value snapshot.Entry) EntryDetailResult {
	batches := make([]ReadColorBatchResult, len(value.ColorBatches))
	for index, batch := range value.ColorBatches {
		batches[index] = newReadColorBatchResult(batch)
	}
	return EntryDetailResult{ID: value.ID, ReceivedAt: value.ReceivedAt.UTC(), ColorBatches: batches}
}

func NewEntryListResult(values []snapshot.EntrySummary, limit, offset int) EntryListResult {
	items := make([]EntrySummaryResult, len(values))
	for index, value := range values {
		items[index] = EntrySummaryResult{ID: value.ID, ReceivedAt: value.ReceivedAt.UTC(), ColorBatchCount: value.ColorBatchCount}
	}
	return EntryListResult{Items: items, Page: PageInfoResult{Limit: limit, Offset: offset, Count: len(items)}}
}

func NewColorBatchDetailResult(value snapshot.ColorBatch) ReadColorBatchResult {
	return newReadColorBatchResult(value)
}

func NewColorBatchListResult(values []snapshot.ColorBatch, limit, offset int) ColorBatchListResult {
	items := make([]ReadColorBatchResult, len(values))
	for index, value := range values {
		items[index] = newReadColorBatchResult(value)
	}
	return ColorBatchListResult{Items: items, Page: PageInfoResult{Limit: limit, Offset: offset, Count: len(items)}}
}

func newReadColorBatchResult(value snapshot.ColorBatch) ReadColorBatchResult {
	runs := make([]ReadSizeRunResult, len(value.SizeRuns))
	for index, run := range value.SizeRuns {
		runs[index] = ReadSizeRunResult{
			ID: run.ID, Size: run.SizeName, Position: run.Position, Status: string(run.Status), Quantity: copyInt(run.Quantity),
			StartedAt: copyTimeResult(run.StartedAt), CompletedAt: copyTimeResult(run.CompletedAt),
		}
	}
	return ReadColorBatchResult{
		ID: value.ID, EntryID: value.EntryID, Color: value.Color, Position: value.Position, Status: string(value.Status),
		CurrentSize: copyString(value.CurrentSize), NextSize: copyString(value.NextSize), CreatedAt: value.CreatedAt.UTC(),
		StartedAt: copyTimeResult(value.StartedAt), CompletedAt: copyTimeResult(value.CompletedAt), SizeRuns: runs,
	}
}

func copyString(value *string) *string {
	if value == nil { return nil }
	copy := *value
	return &copy
}

func copyInt(value *int) *int {
	if value == nil { return nil }
	copy := *value
	return &copy
}

func copyTimeResult(value *time.Time) *time.Time {
	if value == nil { return nil }
	copy := value.UTC()
	return &copy
}
