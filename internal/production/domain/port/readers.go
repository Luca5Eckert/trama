package port

import (
	"context"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/snapshot"
)

type PageCriteria struct {
	Limit  int
	Offset int
}

type ColorBatchListCriteria struct {
	Status  *model.ColorBatchStatus
	EntryID *string
	Page    PageCriteria
}

type EntryReader interface {
	GetByID(context.Context, string) (snapshot.Entry, error)
	List(context.Context, PageCriteria) ([]snapshot.EntrySummary, error)
}

type ColorBatchReader interface {
	GetByID(context.Context, string) (snapshot.ColorBatch, error)
	List(context.Context, ColorBatchListCriteria) ([]snapshot.ColorBatch, error)
}
