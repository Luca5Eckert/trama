package dto

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/users/domain/model"
)

type CreateUserRequest struct {
	Name string `json:"name"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func FromUser(user model.User) UserResponse {
	return UserResponse{ID: user.ID, Name: user.Name, CreatedAt: user.CreatedAt}
}

func FromUsers(users []model.User) []UserResponse {
	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, FromUser(user))
	}
	return response
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
