package model

import (
	"strings"

	"github.com/Luca5Eckert/trama/internal/production/domain"
)

type Color struct {
	name string
	key  string
}

func NewColor(raw string) (Color, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return Color{}, domain.ErrInvalidColor
	}
	return Color{name: name, key: strings.ToLower(name)}, nil
}

func (color Color) Name() string {
	return color.name
}

func (color Color) Key() string {
	return color.key
}
