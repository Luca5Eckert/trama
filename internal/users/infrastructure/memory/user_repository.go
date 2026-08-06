package memory

import (
	"context"
	"errors"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"sync"
)

var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewUserRepository() *UserRepository { return &UserRepository{users: make(map[string]domain.User)} }
func (r *UserRepository) Create(_ context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}
func (r *UserRepository) GetByID(_ context.Context, id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	return user, nil
}
