package query

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type ListColorBatchesQuery struct {
	Status  *string
	EntryID *string
	Limit   *int
	Offset  *int
}

type ListColorBatches struct{ reader port.ColorBatchReader }

func NewListColorBatches(reader port.ColorBatchReader) *ListColorBatches { return &ListColorBatches{reader: reader} }

func (useCase *ListColorBatches) Execute(ctx context.Context, input ListColorBatchesQuery) (application.ColorBatchListResult, error) {
	page, err := normalizePage(input.Limit, input.Offset)
	if err != nil {
		return application.ColorBatchListResult{}, err
	}
	status, err := parseColorBatchStatus(input.Status)
	if err != nil {
		return application.ColorBatchListResult{}, err
	}
	criteria := port.ColorBatchListCriteria{Status: status, EntryID: input.EntryID, Page: page}
	values, err := useCase.reader.List(ctx, criteria)
	if err != nil {
		return application.ColorBatchListResult{}, fmt.Errorf("list color batches: %w", err)
	}
	return application.NewColorBatchListResult(values, page.Limit, page.Offset), nil
}

func parseColorBatchStatus(raw *string) (*model.ColorBatchStatus, error) {
	if raw == nil { return nil, nil }
	var status model.ColorBatchStatus
	switch *raw {
	case string(model.ColorBatchWaiting):
		status = model.ColorBatchWaiting
	case string(model.ColorBatchInProduction):
		status = model.ColorBatchInProduction
	case string(model.ColorBatchCompleted):
		status = model.ColorBatchCompleted
	default:
		return nil, domain.ErrInvalidColorBatchStatus
	}
	return &status, nil
}
