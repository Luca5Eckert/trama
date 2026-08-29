package http

import (
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
)

type SizeSequenceController struct {
	replaceSizeSequence *command.ReplaceSizeSequence
	getSizeSequence     *query.GetSizeSequence
}

func NewSizeSequenceController(replaceSizeSequence *command.ReplaceSizeSequence, getSizeSequence *query.GetSizeSequence) *SizeSequenceController {
	return &SizeSequenceController{replaceSizeSequence: replaceSizeSequence, getSizeSequence: getSizeSequence}
}

func (controller *SizeSequenceController) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("PUT /v1/production/size-sequence", controller.replace)
	mux.HandleFunc("GET /v1/production/size-sequence", controller.get)
}

func (controller *SizeSequenceController) replace(w nethttp.ResponseWriter, r *nethttp.Request) {
	var request dto.ReplaceSizeSequenceRequest
	if err := decodeJSON(r, &request); err != nil {
		respondError(w, nethttp.StatusBadRequest, "invalid_json", "invalid JSON payload")
		return
	}

	items := make([]command.SizeInput, len(request.Items))
	for index, item := range request.Items {
		items[index] = command.SizeInput{Name: item.Name, Position: item.Position}
	}

	result, err := controller.replaceSizeSequence.Execute(r.Context(), command.ReplaceSizeSequenceCommand{Items: items})
	if err != nil {
		respondSizeSequenceError(w, err)
		return
	}
	respondJSON(w, nethttp.StatusOK, dto.FromResult(result))
}

func (controller *SizeSequenceController) get(w nethttp.ResponseWriter, r *nethttp.Request) {
	result, err := controller.getSizeSequence.Execute(r.Context())
	if err != nil {
		respondSizeSequenceError(w, err)
		return
	}
	respondJSON(w, nethttp.StatusOK, dto.FromResult(result))
}

func decodeJSON(r *nethttp.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON value")
	}
	return nil
}

func respondSizeSequenceError(w nethttp.ResponseWriter, err error) {
	switch {
	case application.IsInvalidSizeSequence(err):
		respondError(w, nethttp.StatusUnprocessableEntity, "invalid_size_sequence", "invalid size sequence")
	case application.IsSizeSequenceNotConfigured(err):
		respondError(w, nethttp.StatusNotFound, "size_sequence_not_configured", "size sequence is not configured")
	default:
		respondError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func respondError(w nethttp.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, dto.ErrorResponse{Error: dto.ErrorBody{Code: code, Message: message}})
}

func respondJSON(w nethttp.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
