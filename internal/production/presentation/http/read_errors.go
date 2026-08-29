package http

import (
	nethttp "net/http"

	"github.com/Luca5Eckert/trama/internal/production/application"
)

func respondReadError(w nethttp.ResponseWriter, err error) {
	switch {
	case application.IsEntryNotFound(err):
		respondError(w, nethttp.StatusNotFound, "entry_not_found", "entry not found")
	case application.IsColorBatchNotFound(err):
		respondError(w, nethttp.StatusNotFound, "color_batch_not_found", "color batch not found")
	case application.IsInvalidPagination(err):
		respondError(w, nethttp.StatusUnprocessableEntity, "invalid_pagination", "limit must be between 1 and 100 and offset must be non-negative")
	case application.IsInvalidColorBatchStatus(err):
		respondError(w, nethttp.StatusUnprocessableEntity, "invalid_status", "invalid color batch status")
	default:
		respondError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error")
	}
}
