package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/memory"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
	productionhttp "github.com/Luca5Eckert/trama/internal/production/presentation/http"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func newHandler(now time.Time) http.Handler {
	repository := memory.NewSizeSequenceRepository()
	controller := productionhttp.NewSizeSequenceController(
		command.NewReplaceSizeSequence(repository, fixedClock{now: now}),
		query.NewGetSizeSequence(repository),
	)
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)
	return mux
}

func TestGetSizeSequenceNotConfigured(t *testing.T) {
	handler := newHandler(time.Now())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/production/size-sequence", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
	var body dto.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "size_sequence_not_configured" {
		t.Fatalf("code = %q", body.Error.Code)
	}
}

func TestPutAndGetSizeSequence(t *testing.T) {
	now := time.Date(2026, 8, 29, 22, 0, 0, 0, time.UTC)
	handler := newHandler(now)

	put := httptest.NewRecorder()
	handler.ServeHTTP(put, httptest.NewRequest(http.MethodPut, "/v1/production/size-sequence", bytes.NewBufferString(`{"items":[{"name":"M","position":20},{"name":"P","position":10}]}`)))
	if put.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, body=%s", put.Code, put.Body.String())
	}
	var putBody dto.SizeSequenceResponse
	if err := json.NewDecoder(put.Body).Decode(&putBody); err != nil {
		t.Fatalf("decode PUT: %v", err)
	}
	if putBody.Items[0].Name != "P" || putBody.Items[1].Name != "M" {
		t.Fatalf("unexpected PUT order %#v", putBody.Items)
	}
	if !putBody.UpdatedAt.Equal(now) {
		t.Fatalf("updatedAt = %v, want %v", putBody.UpdatedAt, now)
	}

	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/v1/production/size-sequence", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status = %d, body=%s", get.Code, get.Body.String())
	}
}

func TestPutRejectsMalformedJSON(t *testing.T) {
	handler := newHandler(time.Now())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/v1/production/size-sequence", bytes.NewBufferString(`{"items":`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestPutRejectsInvalidSequence(t *testing.T) {
	handler := newHandler(time.Now())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/v1/production/size-sequence", bytes.NewBufferString(`{"items":[{"name":"P","position":10},{"name":"p","position":20}]}`)))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
	var body dto.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "invalid_size_sequence" {
		t.Fatalf("code = %q", body.Error.Code)
	}
}

func TestRepeatedPutPreservesUpdatedAt(t *testing.T) {
	now := time.Date(2026, 8, 29, 22, 0, 0, 0, time.UTC)
	handler := newHandler(now)
	payload := []byte(`{"items":[{"name":"P","position":10},{"name":"M","position":20}]}`)

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPut, "/v1/production/size-sequence", bytes.NewReader(payload)))
	var firstBody dto.SizeSequenceResponse
	_ = json.NewDecoder(first.Body).Decode(&firstBody)

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPut, "/v1/production/size-sequence", bytes.NewReader(payload)))
	var secondBody dto.SizeSequenceResponse
	_ = json.NewDecoder(second.Body).Decode(&secondBody)

	if !secondBody.UpdatedAt.Equal(firstBody.UpdatedAt) {
		t.Fatalf("updatedAt changed: first=%v second=%v", firstBody.UpdatedAt, secondBody.UpdatedAt)
	}
}
