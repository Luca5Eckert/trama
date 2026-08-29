package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Luca5Eckert/trama/internal/users/application/command"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

type repositoryStub struct {
	created []model.User
	err     error
}

func (repo *repositoryStub) Create(_ context.Context, user model.User) error {
	repo.created = append(repo.created, user)
	return repo.err
}
func (*repositoryStub) GetByID(context.Context, string) (model.User, error) { return model.User{}, nil }
func (*repositoryStub) List(context.Context) ([]model.User, error)          { return nil, nil }

type idStub struct {
	id  string
	err error
}

func (stub idStub) NewID() (string, error) { return stub.id, stub.err }

type clockStub struct{ now time.Time }

func (stub clockStub) Now() time.Time { return stub.now }

func TestCreateUserUsesInjectedPorts(t *testing.T) {
	now := time.Date(2026, 8, 29, 22, 0, 0, 0, time.UTC)
	repository := &repositoryStub{}
	useCase := command.NewCreateUser(repository, idStub{id: "user-1"}, clockStub{now: now})

	user, err := useCase.Execute(context.Background(), command.CreateUserCommand{Name: " Ada "})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if user.ID != "user-1" || user.Name != "Ada" || !user.CreatedAt.Equal(now) {
		t.Fatalf("unexpected user: %#v", user)
	}
	if len(repository.created) != 1 || repository.created[0] != user {
		t.Fatalf("repository got %#v", repository.created)
	}
}

func TestCreateUserPropagatesIDFailure(t *testing.T) {
	idErr := errors.New("entropy unavailable")
	useCase := command.NewCreateUser(&repositoryStub{}, idStub{err: idErr}, clockStub{})

	_, err := useCase.Execute(context.Background(), command.CreateUserCommand{Name: "Ada"})
	if !errors.Is(err, idErr) {
		t.Fatalf("got %v, want wrapped id error", err)
	}
}
