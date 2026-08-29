package command

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type SizeInput struct {
	Name     string
	Position int
}

type ReplaceSizeSequenceCommand struct {
	Items []SizeInput
}

type ReplaceSizeSequence struct {
	repository port.SizeSequenceRepository
	clock      port.Clock
}

func NewReplaceSizeSequence(repository port.SizeSequenceRepository, clock port.Clock) *ReplaceSizeSequence {
	return &ReplaceSizeSequence{repository: repository, clock: clock}
}

func (useCase *ReplaceSizeSequence) Execute(ctx context.Context, cmd ReplaceSizeSequenceCommand) (application.SizeSequenceResult, error) {
	definitions := make([]model.SizeDefinition, len(cmd.Items))
	for index, input := range cmd.Items {
		definition, err := model.NewSizeDefinition(input.Name, input.Position)
		if err != nil {
			return application.SizeSequenceResult{}, err
		}
		definitions[index] = definition
	}

	sequence, err := model.NewSizeSequence(definitions)
	if err != nil {
		return application.SizeSequenceResult{}, err
	}

	stored, err := useCase.repository.Replace(ctx, sequence, useCase.clock.Now().UTC())
	if err != nil {
		return application.SizeSequenceResult{}, fmt.Errorf("replace size sequence: %w", err)
	}
	return application.NewSizeSequenceResult(stored), nil
}
