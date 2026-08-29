package model_test

import (
	"errors"
	"testing"

	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

func definition(t *testing.T, name string, position int) model.SizeDefinition {
	t.Helper()
	item, err := model.NewSizeDefinition(name, position)
	if err != nil {
		t.Fatalf("new size definition: %v", err)
	}
	return item
}

func TestSizeSequenceOrdersByExplicitPosition(t *testing.T) {
	sequence, err := model.NewSizeSequence([]model.SizeDefinition{
		definition(t, "G", 30),
		definition(t, "P", 10),
		definition(t, "M", 20),
	})
	if err != nil {
		t.Fatalf("new sequence: %v", err)
	}

	items := sequence.Items()
	for index, expected := range []string{"P", "M", "G"} {
		if items[index].Name() != expected {
			t.Fatalf("item %d = %q, want %q", index, items[index].Name(), expected)
		}
	}
}

func TestSizeSequenceValidation(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T) error
		want  error
	}{
		{
			name: "empty sequence",
			build: func(t *testing.T) error {
				_, err := model.NewSizeSequence(nil)
				return err
			},
			want: domain.ErrEmptySizeSequence,
		},
		{
			name: "empty name",
			build: func(t *testing.T) error {
				_, err := model.NewSizeDefinition("   ", 10)
				return err
			},
			want: domain.ErrInvalidSizeName,
		},
		{
			name: "invalid position",
			build: func(t *testing.T) error {
				_, err := model.NewSizeDefinition("P", 0)
				return err
			},
			want: domain.ErrInvalidSizePosition,
		},
		{
			name: "duplicate name after normalization",
			build: func(t *testing.T) error {
				_, err := model.NewSizeSequence([]model.SizeDefinition{
					definition(t, " P ", 10),
					definition(t, "p", 20),
				})
				return err
			},
			want: domain.ErrDuplicateSizeName,
		},
		{
			name: "duplicate position",
			build: func(t *testing.T) error {
				_, err := model.NewSizeSequence([]model.SizeDefinition{
					definition(t, "P", 10),
					definition(t, "M", 10),
				})
				return err
			},
			want: domain.ErrDuplicateSizePosition,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.build(t); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestSizeSequenceAcceptsPositionGaps(t *testing.T) {
	_, err := model.NewSizeSequence([]model.SizeDefinition{
		definition(t, "P", 10),
		definition(t, "M", 40),
	})
	if err != nil {
		t.Fatalf("new sequence: %v", err)
	}
}

func TestItemsReturnsCopy(t *testing.T) {
	sequence, err := model.NewSizeSequence([]model.SizeDefinition{
		definition(t, "P", 10),
		definition(t, "M", 20),
	})
	if err != nil {
		t.Fatalf("new sequence: %v", err)
	}

	items := sequence.Items()
	items[0] = definition(t, "X", 999)

	if sequence.Items()[0].Name() != "P" {
		t.Fatal("Items exposed mutable internal state")
	}
}
