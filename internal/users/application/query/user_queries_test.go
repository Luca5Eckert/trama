package query_test

import (
	"context"
	"testing"

	"github.com/Luca5Eckert/trama/internal/users/application/query"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

type repositoryStub struct {
	byID  model.User
	users []model.User
}

func (*repositoryStub) Create(context.Context, model.User) error { return nil }
func (repo *repositoryStub) GetByID(context.Context, string) (model.User, error) {
	return repo.byID, nil
}
func (repo *repositoryStub) List(context.Context) ([]model.User, error) { return repo.users, nil }

func TestQueriesDelegateToRepository(t *testing.T) {
	repository := &repositoryStub{
		byID:  model.User{ID: "user-1", Name: "Ada"},
		users: []model.User{{ID: "user-1", Name: "Ada"}},
	}

	found, err := query.NewGetUserByID(repository).Execute(context.Background(), "user-1")
	if err != nil || found.ID != "user-1" {
		t.Fatalf("get by id: user=%#v err=%v", found, err)
	}

	users, err := query.NewListUsers(repository).Execute(context.Background())
	if err != nil || len(users) != 1 || users[0].ID != "user-1" {
		t.Fatalf("list: users=%#v err=%v", users, err)
	}
}
