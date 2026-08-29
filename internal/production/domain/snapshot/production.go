package snapshot

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

type SizeRun struct {
	ID          string
	SizeName    string
	Position    int
	Status      model.SizeRunStatus
	Quantity    *int
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type ColorBatch struct {
	ID          string
	EntryID     string
	Color       string
	Position    int
	Status      model.ColorBatchStatus
	CurrentSize *string
	NextSize    *string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	SizeRuns    []SizeRun
}

type Entry struct {
	ID           string
	ReceivedAt   time.Time
	ColorBatches []ColorBatch
}

type EntrySummary struct {
	ID              string
	ReceivedAt      time.Time
	ColorBatchCount int
}
