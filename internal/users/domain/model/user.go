package model

import (
	"strings"
	"time"

	"github.com/Luca5Eckert/trama/internal/users/domain"
)

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func NewUser(id, name string, now time.Time) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return User{}, domain.ErrInvalidName
	}

	return User{
		ID:        id,
		Name:      name,
		CreatedAt: now.UTC(),
	}, nil
}
