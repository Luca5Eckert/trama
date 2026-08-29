package application

import (
	"errors"

	"github.com/Luca5Eckert/trama/internal/production/domain"
)

func IsInvalidSizeSequence(err error) bool {
	return errors.Is(err, domain.ErrEmptySizeSequence) ||
		errors.Is(err, domain.ErrInvalidSizeName) ||
		errors.Is(err, domain.ErrInvalidSizePosition) ||
		errors.Is(err, domain.ErrDuplicateSizeName) ||
		errors.Is(err, domain.ErrDuplicateSizePosition)
}

func IsSizeSequenceNotConfigured(err error) bool {
	return errors.Is(err, domain.ErrSizeSequenceNotConfigured)
}

func IsColorsRequired(err error) bool {
	return errors.Is(err, domain.ErrColorsRequired)
}

func IsInvalidColor(err error) bool {
	return errors.Is(err, domain.ErrInvalidColor)
}

func IsDuplicateColor(err error) bool {
	return errors.Is(err, domain.ErrDuplicateColor)
}
