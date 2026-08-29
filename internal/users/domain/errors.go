package domain

import "errors"

var (
	ErrInvalidName = errors.New("name is required")
	ErrNotFound    = errors.New("user not found")
)
