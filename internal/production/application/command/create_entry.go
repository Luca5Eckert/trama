package command

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
	"github.com/Luca5Eckert/trama/internal/production/domain/receipt"
)

type CreateEntryCommand struct {
	Colors []string
}

type CreateEntry struct {
	sequences port.SizeSequenceRepository
	entries   port.EntryRepository
	ids       port.IDGenerator
	clock     port.Clock
}

func NewCreateEntry(sequences port.SizeSequenceRepository, entries port.EntryRepository, ids port.IDGenerator, clock port.Clock) *CreateEntry {
	return &CreateEntry{sequences: sequences, entries: entries, ids: ids, clock: clock}
}

func (useCase *CreateEntry) Execute(ctx context.Context, cmd CreateEntryCommand) (application.EntryResult, error) {
	colors, err := receipt.ParseColors(cmd.Colors)
	if err != nil {
		return application.EntryResult{}, err
	}

	stored, err := useCase.sequences.Get(ctx)
	if err != nil {
		return application.EntryResult{}, fmt.Errorf("get size sequence for entry: %w", err)
	}

	entry, err := receipt.Receive(useCase.clock.Now().UTC(), colors, stored.Sequence, useCase.ids)
	if err != nil {
		return application.EntryResult{}, err
	}
	if err := useCase.entries.Create(ctx, entry); err != nil {
		return application.EntryResult{}, fmt.Errorf("persist entry receipt: %w", err)
	}
	return application.NewEntryResult(entry), nil
}
