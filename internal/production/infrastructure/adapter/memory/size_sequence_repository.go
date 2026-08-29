package memory

import (
	"context"
	"sync"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type SizeSequenceRepository struct {
	mutex      sync.RWMutex
	configured bool
	stored     port.StoredSizeSequence
}

func NewSizeSequenceRepository() *SizeSequenceRepository {
	return &SizeSequenceRepository{}
}

func (repository *SizeSequenceRepository) Get(context.Context) (port.StoredSizeSequence, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	if !repository.configured {
		return port.StoredSizeSequence{}, domain.ErrSizeSequenceNotConfigured
	}
	return repository.stored, nil
}

func (repository *SizeSequenceRepository) Replace(_ context.Context, sequence model.SizeSequence, updatedAt time.Time) (port.StoredSizeSequence, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.configured && repository.stored.Sequence.Equal(sequence) {
		return repository.stored, nil
	}

	repository.stored = port.StoredSizeSequence{Sequence: sequence, UpdatedAt: updatedAt.UTC()}
	repository.configured = true
	return repository.stored, nil
}
