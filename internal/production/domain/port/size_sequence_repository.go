package port

import (
	"context"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

type StoredSizeSequence struct {
	Sequence  model.SizeSequence
	UpdatedAt time.Time
}

type SizeSequenceRepository interface {
	Get(context.Context) (StoredSizeSequence, error)
	Replace(context.Context, model.SizeSequence, time.Time) (StoredSizeSequence, error)
}
