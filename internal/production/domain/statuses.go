package domain

import "errors"

var (
	ErrEntryNotFound          = errors.New("entry not found")
	ErrColorBatchNotFound     = errors.New("color batch not found")
	ErrInvalidPagination      = errors.New("invalid pagination")
	ErrInvalidColorBatchStatus = errors.New("invalid color batch status")
)
