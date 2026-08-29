package query

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type GetColorBatch struct{ reader port.ColorBatchReader }

func NewGetColorBatch(reader port.ColorBatchReader) *GetColorBatch { return &GetColorBatch{reader: reader} }

func (useCase *GetColorBatch) Execute(ctx context.Context, id string) (application.ReadColorBatchResult, error) {
	value, err := useCase.reader.GetByID(ctx, id)
	if err != nil {
		return application.ReadColorBatchResult{}, fmt.Errorf("get color batch: %w", err)
	}
	return application.NewColorBatchDetailResult(value), nil
}
