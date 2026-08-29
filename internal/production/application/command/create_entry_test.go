package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type fakeEntryRepository struct {
	createCalls int
	entry       model.Entry
	err         error
}

func (repository *fakeEntryRepository) Create(_ context.Context, entry model.Entry) error {
	repository.createCalls++
	repository.entry = entry
	return repository.err
}

type fakeSequenceRepository struct {
	getCalls int
	stored   port.StoredSizeSequence
	err      error
}

func (repository *fakeSequenceRepository) Get(context.Context) (port.StoredSizeSequence, error) {
	repository.getCalls++
	return repository.stored, repository.err
}

func (repository *fakeSequenceRepository) Replace(context.Context, model.SizeSequence, time.Time) (port.StoredSizeSequence, error) {
	return port.StoredSizeSequence{}, errors.New("unexpected replace")
}

type fakeIDs struct {
	values []string
	index  int
	err    error
}

func (ids *fakeIDs) NewID() (string, error) {
	if ids.err != nil {
		return "", ids.err
	}
	value := ids.values[ids.index]
	ids.index++
	return value, nil
}

type entryClock struct{ now time.Time }
func (clock entryClock) Now() time.Time { return clock.now }

func commandSequence(t *testing.T) model.SizeSequence {
	t.Helper()
	definition, _ := model.NewSizeDefinition("2", 10)
	sequence, err := model.NewSizeSequence([]model.SizeDefinition{definition})
	if err != nil {
		t.Fatalf("sequence: %v", err)
	}
	return sequence
}

func TestCreateEntryValidatesColorsBeforeReadingSequence(t *testing.T) {
	sequences := &fakeSequenceRepository{}
	entries := &fakeEntryRepository{}
	useCase := command.NewCreateEntry(sequences, entries, &fakeIDs{}, entryClock{})

	_, err := useCase.Execute(context.Background(), command.CreateEntryCommand{})
	if !errors.Is(err, domain.ErrColorsRequired) {
		t.Fatalf("error = %v", err)
	}
	if sequences.getCalls != 0 || entries.createCalls != 0 {
		t.Fatalf("unexpected side effects: sequence=%d create=%d", sequences.getCalls, entries.createCalls)
	}
}

func TestCreateEntryRequiresConfiguredSizeSequence(t *testing.T) {
	sequences := &fakeSequenceRepository{err: domain.ErrSizeSequenceNotConfigured}
	entries := &fakeEntryRepository{}
	useCase := command.NewCreateEntry(sequences, entries, &fakeIDs{}, entryClock{})

	_, err := useCase.Execute(context.Background(), command.CreateEntryCommand{Colors: []string{"Preto"}})
	if !application.IsSizeSequenceNotConfigured(err) {
		t.Fatalf("error = %v", err)
	}
	if entries.createCalls != 0 {
		t.Fatalf("create calls = %d", entries.createCalls)
	}
}

func TestCreateEntryPersistsAggregateOnce(t *testing.T) {
	now := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
	sequences := &fakeSequenceRepository{stored: port.StoredSizeSequence{Sequence: commandSequence(t), UpdatedAt: now}}
	entries := &fakeEntryRepository{}
	ids := &fakeIDs{values: []string{"entry", "batch-preto", "run-preto", "batch-azul", "run-azul"}}
	useCase := command.NewCreateEntry(sequences, entries, ids, entryClock{now: now})

	result, err := useCase.Execute(context.Background(), command.CreateEntryCommand{Colors: []string{"Preto", "Azul"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if entries.createCalls != 1 {
		t.Fatalf("create calls = %d, want 1", entries.createCalls)
	}
	if result.ID != "entry" || len(result.ColorBatches) != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.ColorBatches[0].Color != "Preto" || result.ColorBatches[1].Color != "Azul" {
		t.Fatalf("unexpected color order: %#v", result.ColorBatches)
	}
	if len(entries.entry.ColorBatches()[0].SizeRuns()) != 1 {
		t.Fatal("aggregate snapshot was not persisted through the port")
	}
}

func TestCreateEntryDoesNotPersistWhenIDGenerationFails(t *testing.T) {
	sequences := &fakeSequenceRepository{stored: port.StoredSizeSequence{Sequence: commandSequence(t)}}
	entries := &fakeEntryRepository{}
	useCase := command.NewCreateEntry(sequences, entries, &fakeIDs{err: errors.New("entropy unavailable")}, entryClock{now: time.Now()})

	_, err := useCase.Execute(context.Background(), command.CreateEntryCommand{Colors: []string{"Preto"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if entries.createCalls != 0 {
		t.Fatalf("create calls = %d", entries.createCalls)
	}
}

func TestCreateEntryWrapsPersistenceFailure(t *testing.T) {
	sequences := &fakeSequenceRepository{stored: port.StoredSizeSequence{Sequence: commandSequence(t)}}
	entries := &fakeEntryRepository{err: errors.New("database unavailable")}
	ids := &fakeIDs{values: []string{"entry", "batch", "run"}}
	useCase := command.NewCreateEntry(sequences, entries, ids, entryClock{now: time.Now()})

	_, err := useCase.Execute(context.Background(), command.CreateEntryCommand{Colors: []string{"Preto"}})
	if err == nil {
		t.Fatal("expected persistence error")
	}
	if entries.createCalls != 1 {
		t.Fatalf("create calls = %d", entries.createCalls)
	}
}
