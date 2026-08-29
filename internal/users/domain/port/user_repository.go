package port

import (
	"context"

	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

type UserRepository interface {
	Create(context.Context, model.User) error
	GetByID(context.Context, string) (model.User, error)
	List(context.Context) ([]model.User, error)
}
