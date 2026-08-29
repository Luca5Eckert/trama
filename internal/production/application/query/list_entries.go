package query

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type ListEntriesQuery struct {
	Limit  *int
	Offset *int
}

type ListEntries struct{ reader port.EntryReader }

func NewListEntries(reader port.EntryReader) *ListEntries { return &ListEntries{reader: reader} }

func (useCase *ListEntries) Execute(ctx context.Context, input ListEntriesQuery) (application.EntryListResult, error) {
	page, err := normalizePage(input.Limit, input.Offset)
	if err != nil {
		return application.EntryListResult{}, err
	}
	values, err := useCase.reader.List(ctx, page)
	if err != nil {
		return application.EntryListResult{}, fmt.Errorf("list entries: %w", err)
	}
	return application.NewEntryListResult(values, page.Limit, page.Offset), nil
}
