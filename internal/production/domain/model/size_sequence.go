package model

import (
	"sort"

	"github.com/Luca5Eckert/trama/internal/production/domain"
)

type SizeSequence struct {
	items []SizeDefinition
}

func NewSizeSequence(items []SizeDefinition) (SizeSequence, error) {
	if len(items) == 0 {
		return SizeSequence{}, domain.ErrEmptySizeSequence
	}

	ordered := append([]SizeDefinition(nil), items...)
	seenNames := make(map[string]struct{}, len(ordered))
	seenPositions := make(map[int]struct{}, len(ordered))

	for _, item := range ordered {
		if _, exists := seenNames[item.comparisonKey()]; exists {
			return SizeSequence{}, domain.ErrDuplicateSizeName
		}
		seenNames[item.comparisonKey()] = struct{}{}

		if _, exists := seenPositions[item.Position()]; exists {
			return SizeSequence{}, domain.ErrDuplicateSizePosition
		}
		seenPositions[item.Position()] = struct{}{}
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Position() < ordered[j].Position()
	})

	return SizeSequence{items: ordered}, nil
}

func (sequence SizeSequence) Items() []SizeDefinition {
	return append([]SizeDefinition(nil), sequence.items...)
}

func (sequence SizeSequence) Equal(other SizeSequence) bool {
	if len(sequence.items) != len(other.items) {
		return false
	}
	for index := range sequence.items {
		left := sequence.items[index]
		right := other.items[index]
		if left.Position() != right.Position() || left.Name() != right.Name() {
			return false
		}
	}
	return true
}
