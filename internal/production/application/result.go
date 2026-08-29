package application

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type SizeItemResult struct {
	Name     string
	Position int
}

type SizeSequenceResult struct {
	Items     []SizeItemResult
	UpdatedAt time.Time
}

func NewSizeSequenceResult(stored port.StoredSizeSequence) SizeSequenceResult {
	items := stored.Sequence.Items()
	result := SizeSequenceResult{
		Items:     make([]SizeItemResult, len(items)),
		UpdatedAt: stored.UpdatedAt.UTC(),
	}
	for index, item := range items {
		result.Items[index] = SizeItemResult{Name: item.Name(), Position: item.Position()}
	}
	return result
}
