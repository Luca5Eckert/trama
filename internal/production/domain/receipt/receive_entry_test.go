package receipt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/receipt"
)

type sequenceIDs struct {
	values []string
	index  int
	errAt  int
}

func (generator *sequenceIDs) NewID() (string, error) {
	if generator.errAt > 0 && generator.index+1 == generator.errAt {
		return "", errors.New("id failure")
	}
	if generator.index >= len(generator.values) {
		return "", errors.New("no id available")
	}
	value := generator.values[generator.index]
	generator.index++
	return value, nil
}

func sizeSequence(t *testing.T) model.SizeSequence {
	t.Helper()
	first, _ := model.NewSizeDefinition("2", 10)
	second, _ := model.NewSizeDefinition("4", 20)
	sequence, err := model.NewSizeSequence([]model.SizeDefinition{second, first})
	if err != nil {
		t.Fatalf("sequence: %v", err)
	}
	return sequence
}

func TestParseColors(t *testing.T) {
	tests := []struct {
		name string
		input []string
		want error
	}{
		{name: "empty", input: nil, want: domain.ErrColorsRequired},
		{name: "blank", input: []string{"  "}, want: domain.ErrInvalidColor},
		{name: "duplicate normalized", input: []string{"Preto", " preto "}, want: domain.ErrDuplicateColor},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := receipt.ParseColors(test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestReceiveBuildsOneBatchPerColorWithSizeSnapshot(t *testing.T) {
	colors, err := receipt.ParseColors([]string{" Preto ", "Azul"})
	if err != nil {
		t.Fatalf("colors: %v", err)
	}
	ids := &sequenceIDs{values: []string{
		"entry-1",
		"batch-preto", "run-preto-2", "run-preto-4",
		"batch-azul", "run-azul-2", "run-azul-4",
	}}
	now := time.Date(2026, 8, 29, 20, 0, 0, 0, time.FixedZone("BRT", -3*60*60))

	entry, err := receipt.Receive(now, colors, sizeSequence(t), ids)
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if entry.ID() != "entry-1" {
		t.Fatalf("entry id = %q", entry.ID())
	}
	if entry.ReceivedAt().Location() != time.UTC {
		t.Fatalf("receivedAt location = %v", entry.ReceivedAt().Location())
	}

	batches := entry.ColorBatches()
	if len(batches) != 2 {
		t.Fatalf("batches = %d, want 2", len(batches))
	}
	if batches[0].Color().Name() != "Preto" || batches[0].Position() != 1 || batches[0].Status() != model.ColorBatchWaiting {
		t.Fatalf("unexpected first batch: color=%q position=%d status=%s", batches[0].Color().Name(), batches[0].Position(), batches[0].Status())
	}
	if batches[1].Color().Name() != "Azul" || batches[1].Position() != 2 {
		t.Fatalf("unexpected second batch")
	}

	for _, batch := range batches {
		runs := batch.SizeRuns()
		if len(runs) != 2 {
			t.Fatalf("runs = %d, want 2", len(runs))
		}
		if runs[0].SizeName() != "2" || runs[0].Position() != 10 || runs[1].SizeName() != "4" || runs[1].Position() != 20 {
			t.Fatalf("unexpected snapshot for %s", batch.Color().Name())
		}
		for _, run := range runs {
			if run.Status() != model.SizeRunPending {
				t.Fatalf("run status = %s", run.Status())
			}
			if run.Quantity() != nil {
				t.Fatalf("quantity should be unknown")
			}
		}
	}
}

func TestReceivePropagatesIDFailure(t *testing.T) {
	colors, _ := receipt.ParseColors([]string{"Preto"})
	ids := &sequenceIDs{values: []string{"entry-1"}, errAt: 2}
	_, err := receipt.Receive(time.Now(), colors, sizeSequence(t), ids)
	if err == nil {
		t.Fatal("expected id error")
	}
}
