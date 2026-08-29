package http

import (
	nethttp "net/http"

	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
)

type EntryQueryController struct {
	getEntry   *query.GetEntry
	listEntries *query.ListEntries
}

func NewEntryQueryController(getEntry *query.GetEntry, listEntries *query.ListEntries) *EntryQueryController {
	return &EntryQueryController{getEntry: getEntry, listEntries: listEntries}
}

func (controller *EntryQueryController) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /v1/entries/{id}", controller.get)
	mux.HandleFunc("GET /v1/entries", controller.list)
}

func (controller *EntryQueryController) get(w nethttp.ResponseWriter, r *nethttp.Request) {
	result, err := controller.getEntry.Execute(r.Context(), r.PathValue("id"))
	if err != nil { respondReadError(w, err); return }
	respondJSON(w, nethttp.StatusOK, dto.FromEntryDetailResult(result))
}

func (controller *EntryQueryController) list(w nethttp.ResponseWriter, r *nethttp.Request) {
	values := r.URL.Query()
	limit, err := parseOptionalInt(values, "limit")
	if err != nil { respondError(w, nethttp.StatusUnprocessableEntity, "invalid_pagination", "limit must be an integer"); return }
	offset, err := parseOptionalInt(values, "offset")
	if err != nil { respondError(w, nethttp.StatusUnprocessableEntity, "invalid_pagination", "offset must be an integer"); return }
	result, err := controller.listEntries.Execute(r.Context(), query.ListEntriesQuery{Limit: limit, Offset: offset})
	if err != nil { respondReadError(w, err); return }
	respondJSON(w, nethttp.StatusOK, dto.FromEntryListResult(result))
}
