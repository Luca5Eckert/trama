package query

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type GetEntry struct{ reader port.EntryReader }

func NewGetEntry(reader port.EntryReader) *GetEntry { return &GetEntry{reader: reader} }

func (useCase *GetEntry) Execute(ctx context.Context, id string) (application.EntryDetailResult, error) {
	value, err := useCase.reader.GetByID(ctx, id)
	if err != nil {
		return application.EntryDetailResult{}, fmt.Errorf("get entry: %w", err)
	}
	return application.NewEntryDetailResult(value), nil
}
