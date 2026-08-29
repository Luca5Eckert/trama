package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application"
	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/domain"
	"github.com/Luca5Eckert/trama/internal/production/domain/model"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
)

type repositoryStub struct {
	stored port.StoredSizeSequence
	err    error
}

func (repository repositoryStub) Get(context.Context) (port.StoredSizeSequence, error) {
	return repository.stored, repository.err
}

func (repository repositoryStub) Replace(context.Context, model.SizeSequence, time.Time) (port.StoredSizeSequence, error) {
	panic("unexpected Replace")
}

func TestGetSizeSequenceNotConfigured(t *testing.T) {
	useCase := query.NewGetSizeSequence(repositoryStub{err: domain.ErrSizeSequenceNotConfigured})
	_, err := useCase.Execute(context.Background())
	if !application.IsSizeSequenceNotConfigured(err) {
		t.Fatalf("got %v, want not configured", err)
	}
}

func TestGetSizeSequencePropagatesUnexpectedError(t *testing.T) {
	want := errors.New("database unavailable")
	useCase := query.NewGetSizeSequence(repositoryStub{err: want})
	_, err := useCase.Execute(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}
