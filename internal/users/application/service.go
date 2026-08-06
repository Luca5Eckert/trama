package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"time"
)

type Service struct{ repository domain.Repository }
type CreateInput struct{ Name string }

func NewService(repository domain.Repository) Service { return Service{repository} }
func (s Service) Create(ctx context.Context, input CreateInput) (domain.User, error) {
	user, err := domain.NewUser(newID(), input.Name, time.Now())
	if err != nil {
		return domain.User{}, err
	}
	if err := s.repository.Create(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}
func (s Service) GetByID(ctx context.Context, id string) (domain.User, error) {
	return s.repository.GetByID(ctx, id)
}
func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
