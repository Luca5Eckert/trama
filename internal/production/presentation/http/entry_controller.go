package http

import (
	nethttp "net/http"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
)

type EntryController struct {
	createEntry *command.CreateEntry
}

func NewEntryController(createEntry *command.CreateEntry) *EntryController {
	return &EntryController{createEntry: createEntry}
}

func (controller *EntryController) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("POST /v1/entries", controller.create)
}

func (controller *EntryController) create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var request dto.CreateEntryRequest
	if err := decodeJSON(r, &request); err != nil {
		respondError(w, nethttp.StatusBadRequest, "invalid_json", "invalid JSON payload")
		return
	}

	result, err := controller.createEntry.Execute(r.Context(), command.CreateEntryCommand{Colors: request.Colors})
	if err != nil {
		respondCreateEntryError(w, err)
		return
	}

	w.Header().Set("Location", "/v1/entries/"+result.ID)
	respondJSON(w, nethttp.StatusCreated, dto.FromEntryResult(result))
}

func respondCreateEntryError(w nethttp.ResponseWriter, err error) {
	switch {
	case application.IsColorsRequired(err):
		respondError(w, nethttp.StatusUnprocessableEntity, "colors_required", "at least one color is required")
	case application.IsInvalidColor(err):
		respondError(w, nethttp.StatusUnprocessableEntity, "invalid_color", "color name is required")
	case application.IsDuplicateColor(err):
		respondError(w, nethttp.StatusUnprocessableEntity, "duplicate_color", "colors must be unique within an entry")
	case application.IsSizeSequenceNotConfigured(err):
		respondError(w, nethttp.StatusConflict, "size_sequence_not_configured", "size sequence is not configured")
	default:
		respondError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error")
	}
}
