package domain

import "errors"

var (
	ErrSizeSequenceNotConfigured = errors.New("size sequence is not configured")
	ErrEmptySizeSequence         = errors.New("size sequence must contain at least one item")
	ErrInvalidSizeName           = errors.New("size name is required")
	ErrInvalidSizePosition       = errors.New("size position must be positive")
	ErrDuplicateSizeName         = errors.New("size names must be unique")
	ErrDuplicateSizePosition     = errors.New("size positions must be unique")

	ErrColorsRequired = errors.New("at least one color is required")
	ErrInvalidColor   = errors.New("color name is required")
	ErrDuplicateColor = errors.New("colors must be unique within an entry")
)
