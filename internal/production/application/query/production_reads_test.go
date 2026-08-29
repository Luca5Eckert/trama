package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
	"github.com/Luca5Eckert/trama/internal/production/domain/snapshot"
)

type fakeEntryReader struct {
	entry snapshot.Entry
	items []snapshot.EntrySummary
	err error
	lastPage port.PageCriteria
}
func (reader *fakeEntryReader) GetByID(context.Context, string) (snapshot.Entry, error) { return reader.entry, reader.err }
func (reader *fakeEntryReader) List(_ context.Context, page port.PageCriteria) ([]snapshot.EntrySummary, error) { reader.lastPage = page; return reader.items, reader.err }

type fakeBatchReader struct {
	batch snapshot.ColorBatch
	items []snapshot.ColorBatch
	err error
	criteria port.ColorBatchListCriteria
}
func (reader *fakeBatchReader) GetByID(context.Context, string) (snapshot.ColorBatch, error) { return reader.batch, reader.err }
func (reader *fakeBatchReader) List(_ context.Context, criteria port.ColorBatchListCriteria) ([]snapshot.ColorBatch, error) { reader.criteria = criteria; return reader.items, reader.err }

func TestListEntriesUsesDefaultPagination(t *testing.T) {
	reader := &fakeEntryReader{}
	result, err := query.NewListEntries(reader).Execute(context.Background(), query.ListEntriesQuery{})
	if err != nil { t.Fatalf("execute: %v", err) }
	if reader.lastPage.Limit != 50 || reader.lastPage.Offset != 0 { t.Fatalf("page = %#v", reader.lastPage) }
	if result.Page.Count != 0 { t.Fatalf("count = %d", result.Page.Count) }
}

func TestListEntriesRejectsInvalidPagination(t *testing.T) {
	limit := 101
	_, err := query.NewListEntries(&fakeEntryReader{}).Execute(context.Background(), query.ListEntriesQuery{Limit: &limit})
	if !application.IsInvalidPagination(err) { t.Fatalf("error = %v", err) }
}

func TestListColorBatchesMapsFilters(t *testing.T) {
	status := "IN_PRODUCTION"
	entryID := "entry-1"
	limit, offset := 20, 5
	reader := &fakeBatchReader{}
	_, err := query.NewListColorBatches(reader).Execute(context.Background(), query.ListColorBatchesQuery{Status: &status, EntryID: &entryID, Limit: &limit, Offset: &offset})
	if err != nil { t.Fatalf("execute: %v", err) }
	if reader.criteria.Status == nil || *reader.criteria.Status != model.ColorBatchInProduction { t.Fatalf("status = %#v", reader.criteria.Status) }
	if reader.criteria.EntryID == nil || *reader.criteria.EntryID != entryID { t.Fatalf("entry = %#v", reader.criteria.EntryID) }
	if reader.criteria.Page.Limit != limit || reader.criteria.Page.Offset != offset { t.Fatalf("page = %#v", reader.criteria.Page) }
}

func TestListColorBatchesRejectsUnknownStatus(t *testing.T) {
	status := "PAUSED"
	_, err := query.NewListColorBatches(&fakeBatchReader{}).Execute(context.Background(), query.ListColorBatchesQuery{Status: &status})
	if !application.IsInvalidColorBatchStatus(err) { t.Fatalf("error = %v", err) }
}

func TestGetQueriesPreserveNotFound(t *testing.T) {
	_, err := query.NewGetEntry(&fakeEntryReader{err: domain.ErrEntryNotFound}).Execute(context.Background(), "missing")
	if !application.IsEntryNotFound(err) { t.Fatalf("entry error = %v", err) }
	_, err = query.NewGetColorBatch(&fakeBatchReader{err: domain.ErrColorBatchNotFound}).Execute(context.Background(), "missing")
	if !application.IsColorBatchNotFound(err) { t.Fatalf("batch error = %v", err) }
}

func TestGetColorBatchMapsCurrentAndNextSize(t *testing.T) {
	current, next := "M", "G"
	now := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	reader := &fakeBatchReader{batch: snapshot.ColorBatch{ID: "batch", EntryID: "entry", Color: "Preto", Status: model.ColorBatchInProduction, CurrentSize: &current, NextSize: &next, CreatedAt: now}}
	result, err := query.NewGetColorBatch(reader).Execute(context.Background(), "batch")
	if err != nil { t.Fatalf("execute: %v", err) }
	if result.CurrentSize == nil || *result.CurrentSize != "M" || result.NextSize == nil || *result.NextSize != "G" { t.Fatalf("sizes = %#v / %#v", result.CurrentSize, result.NextSize) }
}
