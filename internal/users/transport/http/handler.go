package http

import (
	"encoding/json"
	"errors"
	"github.com/Luca5Eckert/trama/internal/users/application"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"net/http"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service} }

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/users", h.create)
	mux.HandleFunc("GET /v1/users/{id}", h.getByID)
	mux.HandleFunc("GET /v1/users", h.getAll)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input application.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	user, err := h.service.Create(r.Context(), input)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, user)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetAll(r.Context())

	if err != nil {
		respondDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}


func respondDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidName):
		respondError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
