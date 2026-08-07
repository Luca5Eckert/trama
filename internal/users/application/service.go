package application

import (
	"context"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/infrastructure/utils"
	"time"
)

type Service struct{ 
	repository domain.Repository 
}

type CreateInput struct{ Name string }

func NewService(repository domain.Repository) *Service {
	return &Service{repository} 
}

func (s *Service) Create(ctx context.Context, input CreateInput) (domain.User, error) {
	user, err := domain.NewUser(utils.GenerateUUIDId(), input.Name, time.Now())

	if err != nil {
		return domain.User{}, err
	}
	if err := s.repository.Create(ctx, user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.User, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]domain.User, error) {
	return s.repository.GetAll(ctx)
}

