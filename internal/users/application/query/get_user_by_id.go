package query

import (
	"context"

	"github.com/Luca5Eckert/trama/internal/users/domain/model"
	"github.com/Luca5Eckert/trama/internal/users/domain/port"
)

type GetUserByID struct {
	users port.UserRepository
}

func NewGetUserByID(users port.UserRepository) *GetUserByID {
	return &GetUserByID{users: users}
}

func (uc *GetUserByID) Execute(ctx context.Context, id string) (model.User, error) {
	return uc.users.GetByID(ctx, id)
}
