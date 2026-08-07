package memory

import (
	"context"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"sync"
)

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
		return domain.User{}, domain.ErrNotFound
	}
	
	return user, nil
}

func (repo *UserRepository) GetAll(_ context.Context) ([]domain.User, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	users := make([]domain.User, 0, len(repo.users))
	for _, user := range repo.users {
		users = append(users, user)
	}
	return users, nil
}	
