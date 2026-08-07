package domain

import "context"

type Repository interface {
	Create(context.Context, User) error
	GetByID(context.Context, string) (User, error)
	GetAll(context.Context) ([]User, error)
}
