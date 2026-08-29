package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
	"github.com/Luca5Eckert/trama/internal/production/domain/snapshot"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
	productionhttp "github.com/Luca5Eckert/trama/internal/production/presentation/http"
)

type httpEntryReader struct { entry snapshot.Entry; items []snapshot.EntrySummary; missing bool }
func (reader *httpEntryReader) GetByID(context.Context, string) (snapshot.Entry, error) { if reader.missing { return snapshot.Entry{}, domain.ErrEntryNotFound }; return reader.entry, nil }
func (reader *httpEntryReader) List(context.Context, port.PageCriteria) ([]snapshot.EntrySummary, error) { return reader.items, nil }

type httpBatchReader struct { batch snapshot.ColorBatch; items []snapshot.ColorBatch; missing bool }
func (reader *httpBatchReader) GetByID(context.Context, string) (snapshot.ColorBatch, error) { if reader.missing { return snapshot.ColorBatch{}, domain.ErrColorBatchNotFound }; return reader.batch, nil }
func (reader *httpBatchReader) List(context.Context, port.ColorBatchListCriteria) ([]snapshot.ColorBatch, error) { return reader.items, nil }

func TestEntryReadEndpoints(t *testing.T) {
	now := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	next := "2"
	reader := &httpEntryReader{entry: snapshot.Entry{ID: "entry", ReceivedAt: now, ColorBatches: []snapshot.ColorBatch{{ID: "batch", EntryID: "entry", Color: "Preto", Position: 1, Status: model.ColorBatchWaiting, NextSize: &next, CreatedAt: now}}}}
	controller := productionhttp.NewEntryQueryController(query.NewGetEntry(reader), query.NewListEntries(reader))
	mux := http.NewServeMux(); controller.RegisterRoutes(mux)

	response := httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/entries/entry", nil))
	if response.Code != http.StatusOK { t.Fatalf("detail status=%d body=%s", response.Code, response.Body.String()) }
	var detail dto.EntryDetailResponse; _ = json.NewDecoder(response.Body).Decode(&detail)
	if len(detail.ColorBatches) != 1 || detail.ColorBatches[0].NextSize == nil || *detail.ColorBatches[0].NextSize != "2" { t.Fatalf("detail=%#v", detail) }

	response = httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/entries?limit=101", nil))
	if response.Code != http.StatusUnprocessableEntity { t.Fatalf("pagination status=%d", response.Code) }
}

func TestEntryDetailNotFound(t *testing.T) {
	reader := &httpEntryReader{missing: true}
	controller := productionhttp.NewEntryQueryController(query.NewGetEntry(reader), query.NewListEntries(reader))
	mux := http.NewServeMux(); controller.RegisterRoutes(mux)
	response := httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/entries/missing", nil))
	if response.Code != http.StatusNotFound { t.Fatalf("status=%d", response.Code) }
}

func TestColorBatchReadEndpointsAndValidation(t *testing.T) {
	now := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	current, next := "4", "6"
	reader := &httpBatchReader{batch: snapshot.ColorBatch{ID: "batch", EntryID: "entry", Color: "Azul", Position: 1, Status: model.ColorBatchInProduction, CurrentSize: &current, NextSize: &next, CreatedAt: now, SizeRuns: []snapshot.SizeRun{{ID: "run", SizeName: "4", Position: 20, Status: model.SizeRunInProgress}}}}
	controller := productionhttp.NewColorBatchController(query.NewGetColorBatch(reader), query.NewListColorBatches(reader))
	mux := http.NewServeMux(); controller.RegisterRoutes(mux)

	response := httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/color-batches/batch", nil))
	if response.Code != http.StatusOK { t.Fatalf("detail status=%d body=%s", response.Code, response.Body.String()) }
	var detail dto.ColorBatchDetailResponse; _ = json.NewDecoder(response.Body).Decode(&detail)
	if detail.CurrentSize == nil || *detail.CurrentSize != "4" || len(detail.SizeRuns) != 1 { t.Fatalf("detail=%#v", detail) }

	response = httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/color-batches?status=PAUSED", nil))
	if response.Code != http.StatusUnprocessableEntity { t.Fatalf("status filter status=%d", response.Code) }
}

func TestReadCollectionsReturnEmpty200(t *testing.T) {
	entryController := productionhttp.NewEntryQueryController(query.NewGetEntry(&httpEntryReader{}), query.NewListEntries(&httpEntryReader{}))
	batchController := productionhttp.NewColorBatchController(query.NewGetColorBatch(&httpBatchReader{}), query.NewListColorBatches(&httpBatchReader{}))
	mux := http.NewServeMux(); entryController.RegisterRoutes(mux); batchController.RegisterRoutes(mux)
	response := httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/entries", nil))
	if response.Code != http.StatusOK { t.Fatalf("entries status=%d", response.Code) }
	var entries dto.EntryListResponse; _ = json.NewDecoder(response.Body).Decode(&entries)
	if len(entries.Items) != 0 || entries.Page.Count != 0 { t.Fatalf("entries=%#v", entries) }

	response = httptest.NewRecorder(); mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/color-batches", nil))
	if response.Code != http.StatusOK { t.Fatalf("batches status=%d", response.Code) }
}
