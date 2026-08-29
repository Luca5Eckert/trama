package http

import (
	nethttp "net/http"

	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/presentation/dto"
)

type ColorBatchController struct {
	getBatch   *query.GetColorBatch
	listBatches *query.ListColorBatches
}

func NewColorBatchController(getBatch *query.GetColorBatch, listBatches *query.ListColorBatches) *ColorBatchController {
	return &ColorBatchController{getBatch: getBatch, listBatches: listBatches}
}

func (controller *ColorBatchController) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /v1/color-batches/{id}", controller.get)
	mux.HandleFunc("GET /v1/color-batches", controller.list)
}

func (controller *ColorBatchController) get(w nethttp.ResponseWriter, r *nethttp.Request) {
	result, err := controller.getBatch.Execute(r.Context(), r.PathValue("id"))
	if err != nil { respondReadError(w, err); return }
	respondJSON(w, nethttp.StatusOK, dto.FromColorBatchDetailResult(result))
}

func (controller *ColorBatchController) list(w nethttp.ResponseWriter, r *nethttp.Request) {
	values := r.URL.Query()
	limit, err := parseOptionalInt(values, "limit")
	if err != nil { respondError(w, nethttp.StatusUnprocessableEntity, "invalid_pagination", "limit must be an integer"); return }
	offset, err := parseOptionalInt(values, "offset")
	if err != nil { respondError(w, nethttp.StatusUnprocessableEntity, "invalid_pagination", "offset must be an integer"); return }
	result, err := controller.listBatches.Execute(r.Context(), query.ListColorBatchesQuery{
		Status: optionalString(values, "status"), EntryID: optionalString(values, "entryId"), Limit: limit, Offset: offset,
	})
	if err != nil { respondReadError(w, err); return }
	respondJSON(w, nethttp.StatusOK, dto.FromColorBatchListResult(result))
}
