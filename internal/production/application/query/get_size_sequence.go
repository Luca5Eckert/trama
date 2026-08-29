package query

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type GetSizeSequence struct {
	repository port.SizeSequenceRepository
}

func NewGetSizeSequence(repository port.SizeSequenceRepository) *GetSizeSequence {
	return &GetSizeSequence{repository: repository}
}

func (useCase *GetSizeSequence) Execute(ctx context.Context) (application.SizeSequenceResult, error) {
	stored, err := useCase.repository.Get(ctx)
	if err != nil {
		return application.SizeSequenceResult{}, fmt.Errorf("get size sequence: %w", err)
	}
	return application.NewSizeSequenceResult(stored), nil
}
