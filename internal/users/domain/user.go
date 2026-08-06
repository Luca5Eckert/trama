package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidName = errors.New("name is required")

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewUser(id, name string, now time.Time) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return User{}, ErrInvalidName
	}
	return User{ID: id, Name: name, CreatedAt: now.UTC()}, nil
}
