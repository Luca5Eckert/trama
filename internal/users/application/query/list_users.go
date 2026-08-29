package query

import (
	"context"

	"github.com/Luca5Eckert/trama/internal/users/domain/model"
	"github.com/Luca5Eckert/trama/internal/users/domain/port"
)

type ListUsers struct {
	users port.UserRepository
}

func NewListUsers(users port.UserRepository) *ListUsers {
	return &ListUsers{users: users}
}

func (uc *ListUsers) Execute(ctx context.Context) ([]model.User, error) {
	return uc.users.List(ctx)
}
