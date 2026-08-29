package model

import (
	"strings"

	"github.com/Luca5Eckert/trama/internal/production/domain"
)

type SizeDefinition struct {
	name     string
	position int
}

func NewSizeDefinition(name string, position int) (SizeDefinition, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return SizeDefinition{}, domain.ErrInvalidSizeName
	}
	if position <= 0 {
		return SizeDefinition{}, domain.ErrInvalidSizePosition
	}
	return SizeDefinition{name: trimmedName, position: position}, nil
}

func (definition SizeDefinition) Name() string { return definition.name }

func (definition SizeDefinition) Position() int { return definition.position }

func (definition SizeDefinition) comparisonKey() string {
	return strings.ToLower(definition.name)
}
