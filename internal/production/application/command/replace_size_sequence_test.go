package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type repositorySpy struct {
	calls  int
	stored port.StoredSizeSequence
	err    error
}

func (repository *repositorySpy) Get(context.Context) (port.StoredSizeSequence, error) {
	return repository.stored, repository.err
}

func (repository *repositorySpy) Replace(_ context.Context, sequence model.SizeSequence, updatedAt time.Time) (port.StoredSizeSequence, error) {
	repository.calls++
	if repository.err != nil {
		return port.StoredSizeSequence{}, repository.err
	}
	repository.stored = port.StoredSizeSequence{Sequence: sequence, UpdatedAt: updatedAt}
	return repository.stored, nil
}

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func TestReplaceSizeSequenceValidatesBeforePersistence(t *testing.T) {
	repository := &repositorySpy{}
	useCase := command.NewReplaceSizeSequence(repository, fixedClock{now: time.Now()})

	_, err := useCase.Execute(context.Background(), command.ReplaceSizeSequenceCommand{Items: []command.SizeInput{
		{Name: "P", Position: 10},
		{Name: "p", Position: 20},
	}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repository.calls != 0 {
		t.Fatalf("repository called %d times, want 0", repository.calls)
	}
}

func TestReplaceSizeSequenceUsesInjectedClockAndOrdersItems(t *testing.T) {
	repository := &repositorySpy{}
	now := time.Date(2026, 8, 29, 22, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	useCase := command.NewReplaceSizeSequence(repository, fixedClock{now: now})

	result, err := useCase.Execute(context.Background(), command.ReplaceSizeSequenceCommand{Items: []command.SizeInput{
		{Name: "M", Position: 20},
		{Name: "P", Position: 10},
	}})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	if repository.calls != 1 {
		t.Fatalf("repository called %d times, want 1", repository.calls)
	}
	if !result.UpdatedAt.Equal(now.UTC()) {
		t.Fatalf("updatedAt = %v, want %v", result.UpdatedAt, now.UTC())
	}
	if result.Items[0].Name != "P" || result.Items[1].Name != "M" {
		t.Fatalf("unexpected order %#v", result.Items)
	}
}

func TestReplaceSizeSequencePropagatesRepositoryError(t *testing.T) {
	want := errors.New("storage unavailable")
	repository := &repositorySpy{err: want}
	useCase := command.NewReplaceSizeSequence(repository, fixedClock{now: time.Now()})

	_, err := useCase.Execute(context.Background(), command.ReplaceSizeSequenceCommand{Items: []command.SizeInput{{Name: "P", Position: 10}}})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}
