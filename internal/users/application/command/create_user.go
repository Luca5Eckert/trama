package command

import (
	"context"
	"fmt"

	"github.com/Luca5Eckert/trama/internal/users/domain/model"
	"github.com/Luca5Eckert/trama/internal/users/domain/port"
)

type CreateUserCommand struct {
	Name string
}

type CreateUser struct {
	users port.UserRepository
	ids   port.IDGenerator
	clock port.Clock
}

func NewCreateUser(users port.UserRepository, ids port.IDGenerator, clock port.Clock) *CreateUser {
	return &CreateUser{users: users, ids: ids, clock: clock}
}

func (uc *CreateUser) Execute(ctx context.Context, cmd CreateUserCommand) (model.User, error) {
	id, err := uc.ids.NewID()
	if err != nil {
		return model.User{}, fmt.Errorf("generate user id: %w", err)
	}

	user, err := model.NewUser(id, cmd.Name, uc.clock.Now())
	if err != nil {
		return model.User{}, err
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return model.User{}, fmt.Errorf("persist user: %w", err)
	}

	return user, nil
}
