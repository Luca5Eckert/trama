package query

import (
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

const (
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

func normalizePage(limit, offset *int) (port.PageCriteria, error) {
	resolvedLimit := DefaultPageLimit
	if limit != nil {
		resolvedLimit = *limit
	}
	resolvedOffset := 0
	if offset != nil {
		resolvedOffset = *offset
	}
	if resolvedLimit <= 0 || resolvedLimit > MaxPageLimit || resolvedOffset < 0 {
		return port.PageCriteria{}, domain.ErrInvalidPagination
	}
	return port.PageCriteria{Limit: resolvedLimit, Offset: resolvedOffset}, nil
}
