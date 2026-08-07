package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("user not found")

type Repository interface {
	Create(context.Context, User) error
	GetByID(context.Context, string) (User, error)
	GetAll(context.Context) ([]User, error)
}
