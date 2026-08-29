package application

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

type ColorBatchResult struct {
	ID       string
	Color    string
	Position int
	Status   string
}

type EntryResult struct {
	ID           string
	ReceivedAt   time.Time
	ColorBatches []ColorBatchResult
}

func NewEntryResult(entry model.Entry) EntryResult {
	batches := entry.ColorBatches()
	result := make([]ColorBatchResult, len(batches))
	for index, batch := range batches {
		result[index] = ColorBatchResult{
			ID:       batch.ID(),
			Color:    batch.Color().Name(),
			Position: batch.Position(),
			Status:   string(batch.Status()),
		}
	}
	return EntryResult{ID: entry.ID(), ReceivedAt: entry.ReceivedAt(), ColorBatches: result}
}
