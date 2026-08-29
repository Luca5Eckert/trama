package application

import (
	"errors"

	"github.com/Luca5Eckert/trama/internal/production/domain"
)

func IsEntryNotFound(err error) bool { return errors.Is(err, domain.ErrEntryNotFound) }
func IsColorBatchNotFound(err error) bool { return errors.Is(err, domain.ErrColorBatchNotFound) }
func IsInvalidPagination(err error) bool { return errors.Is(err, domain.ErrInvalidPagination) }
func IsInvalidColorBatchStatus(err error) bool { return errors.Is(err, domain.ErrInvalidColorBatchStatus) }
