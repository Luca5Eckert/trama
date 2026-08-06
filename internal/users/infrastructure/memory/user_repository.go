package memory

import (
	"context"
	"errors"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"sync"
)

var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	mutex    sync.RWMutex
	users map[string]domain.User
}

func NewUserRepository() *UserRepository { 
	return &UserRepository{users: make(map[string]domain.User)} 
}

func (repo *UserRepository) Create(_ context.Context, user domain.User) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.users[user.ID] = user

	return nil
}
func (repo *UserRepository) GetByID(_ context.Context, id string) (domain.User, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	user, ok := repo.users[id]

	if !ok {
		return domain.User{}, ErrNotFound
	}
	
	return user, nil
}
