package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/memory"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
	productionhttp "github.com/Luca5Eckert/trama/internal/production/presentation/http"
)

type entryIDs struct {
	values []string
	index int
}

func (ids *entryIDs) NewID() (string, error) {
	if ids.index >= len(ids.values) {
		return "", errors.New("no id available")
	}
	value := ids.values[ids.index]
	ids.index++
	return value, nil
}

func entryHandler(t *testing.T, configured bool) http.Handler {
	t.Helper()
	sequences := memory.NewSizeSequenceRepository()
	if configured {
		definition, _ := model.NewSizeDefinition("2", 10)
		sequence, _ := model.NewSizeSequence([]model.SizeDefinition{definition})
		if _, err := sequences.Replace(context.Background(), sequence, time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("configure sequence: %v", err)
		}
	}
	entries := memory.NewEntryRepository()
	ids := &entryIDs{values: []string{"entry-1", "batch-1", "run-1", "batch-2", "run-2", "batch-3", "run-3"}}
	controller := productionhttp.NewEntryController(command.NewCreateEntry(sequences, entries, ids, fixedClock{now: time.Date(2026, 8, 29, 21, 0, 0, 0, time.UTC)}))
	mux := http.NewServeMux()
	controller.RegisterRoutes(mux)
	return mux
}

func TestCreateEntryReturns201LocationAndPreservesColorOrder(t *testing.T) {
	handler := entryHandler(t, true)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/entries", bytes.NewBufferString(`{"colors":["Preto","Azul","Bege"]}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
	if location := response.Header().Get("Location"); location != "/v1/entries/entry-1" {
		t.Fatalf("Location = %q", location)
	}
	var body dto.EntryResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.ColorBatches) != 3 || body.ColorBatches[0].Color != "Preto" || body.ColorBatches[1].Color != "Azul" || body.ColorBatches[2].Color != "Bege" {
		t.Fatalf("unexpected batches: %#v", body.ColorBatches)
	}
	if body.ColorBatches[0].Position != 1 || body.ColorBatches[2].Position != 3 {
		t.Fatalf("unexpected positions: %#v", body.ColorBatches)
	}
}

func TestCreateEntryRejectsMalformedJSON(t *testing.T) {
	response := httptest.NewRecorder()
	entryHandler(t, true).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/entries", bytes.NewBufferString(`{"colors":`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestCreateEntryValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		payload string
		code string
	}{
		{name: "colors required", payload: `{"colors":[]}`, code: "colors_required"},
		{name: "invalid color", payload: `{"colors":[" "]}`, code: "invalid_color"},
		{name: "duplicate", payload: `{"colors":["Preto"," preto "]}`, code: "duplicate_color"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			entryHandler(t, true).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/entries", bytes.NewBufferString(test.payload)))
			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
			}
			var body dto.ErrorResponse
			_ = json.NewDecoder(response.Body).Decode(&body)
			if body.Error.Code != test.code {
				t.Fatalf("code = %q, want %q", body.Error.Code, test.code)
			}
		})
	}
}

func TestCreateEntryRequiresSizeSequence(t *testing.T) {
	response := httptest.NewRecorder()
	entryHandler(t, false).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/entries", bytes.NewBufferString(`{"colors":["Preto"]}`)))
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
	var body dto.ErrorResponse
	_ = json.NewDecoder(response.Body).Decode(&body)
	if body.Error.Code != "size_sequence_not_configured" {
		t.Fatalf("code = %q", body.Error.Code)
	}
}
