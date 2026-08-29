package receipt

import (
	"fmt"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

func ParseColors(raw []string) ([]model.Color, error) {
	if len(raw) == 0 {
		return nil, domain.ErrColorsRequired
	}

	colors := make([]model.Color, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for index, value := range raw {
		color, err := model.NewColor(value)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[color.Key()]; exists {
			return nil, domain.ErrDuplicateColor
		}
		seen[color.Key()] = struct{}{}
		colors[index] = color
	}
	return colors, nil
}

func Receive(receivedAt time.Time, colors []model.Color, sequence model.SizeSequence, ids port.IDGenerator) (model.Entry, error) {
	if len(colors) == 0 {
		return model.Entry{}, domain.ErrColorsRequired
	}

	entryID, err := ids.NewID()
	if err != nil {
		return model.Entry{}, fmt.Errorf("generate entry id: %w", err)
	}

	now := receivedAt.UTC()
	sizes := sequence.Items()
	batches := make([]model.ColorBatch, len(colors))
	for colorIndex, color := range colors {
		batchID, err := ids.NewID()
		if err != nil {
			return model.Entry{}, fmt.Errorf("generate color batch id: %w", err)
		}

		runs := make([]model.SizeRun, len(sizes))
		for sizeIndex, size := range sizes {
			runID, err := ids.NewID()
			if err != nil {
				return model.Entry{}, fmt.Errorf("generate size run id: %w", err)
			}
			runs[sizeIndex] = model.NewPendingSizeRun(runID, batchID, size)
		}

		batches[colorIndex] = model.NewWaitingColorBatch(batchID, entryID, color, colorIndex+1, now, runs)
	}

	return model.NewEntry(entryID, now, batches), nil
}
