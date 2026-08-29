package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

type UserRepository struct {
	mutex sync.RWMutex
	users map[string]model.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{users: make(map[string]model.User)}
}

func (repo *UserRepository) Create(_ context.Context, user model.User) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.users[user.ID] = user
	return nil
}

func (repo *UserRepository) GetByID(_ context.Context, id string) (model.User, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	user, ok := repo.users[id]
	if !ok {
		return model.User{}, domain.ErrNotFound
	}

	return user, nil
}

func (repo *UserRepository) List(_ context.Context) ([]model.User, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	users := make([]model.User, 0, len(repo.users))
	for _, user := range repo.users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		if users[i].CreatedAt.Equal(users[j].CreatedAt) {
			return users[i].ID < users[j].ID
		}
		return users[i].CreatedAt.Before(users[j].CreatedAt)
	})

	return users, nil
}
