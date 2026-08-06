package application_test

import (
	"context"
	"github.com/Luca5Eckert/trama/internal/users/application"
	"github.com/Luca5Eckert/trama/internal/users/infrastructure/memory"
	"testing"
)

func TestCreateAndGetUser(t *testing.T) {
	service := application.NewService(memory.NewUserRepository())
	created, err := service.Create(context.Background(), application.CreateInput{Name: "Ada Lovelace"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	found, err := service.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if found != created {
		t.Fatalf("got %#v, want %#v", found, created)
	}
}
