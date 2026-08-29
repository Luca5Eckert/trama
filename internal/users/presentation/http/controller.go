package http

import (
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"

	"github.com/Luca5Eckert/trama/internal/users/application/command"
	"github.com/Luca5Eckert/trama/internal/users/application/query"
	"github.com/Luca5Eckert/trama/internal/users/domain"
	"github.com/Luca5Eckert/trama/internal/users/presentation/dto"
)

type UserController struct {
	createUser  *command.CreateUser
	getUserByID *query.GetUserByID
	listUsers   *query.ListUsers
}

func NewUserController(createUser *command.CreateUser, getUserByID *query.GetUserByID, listUsers *query.ListUsers) *UserController {
	return &UserController{createUser: createUser, getUserByID: getUserByID, listUsers: listUsers}
}

func (controller *UserController) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("POST /v1/users", controller.create)
	mux.HandleFunc("GET /v1/users/{id}", controller.getByID)
	mux.HandleFunc("GET /v1/users", controller.list)
}

func (controller *UserController) create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var request dto.CreateUserRequest
	if err := decodeJSON(r, &request); err != nil {
		respondError(w, nethttp.StatusBadRequest, "invalid_json", "invalid JSON payload")
		return
	}

	user, err := controller.createUser.Execute(r.Context(), command.CreateUserCommand{Name: request.Name})
	if err != nil {
		respondDomainError(w, err)
		return
	}

	w.Header().Set("Location", "/v1/users/"+user.ID)
	respondJSON(w, nethttp.StatusCreated, dto.FromUser(user))
}

func (controller *UserController) getByID(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, err := controller.getUserByID.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, nethttp.StatusOK, dto.FromUser(user))
}

func (controller *UserController) list(w nethttp.ResponseWriter, r *nethttp.Request) {
	users, err := controller.listUsers.Execute(r.Context())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, nethttp.StatusOK, dto.FromUsers(users))
}

func decodeJSON(r *nethttp.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON value")
	}
	return nil
}

func respondDomainError(w nethttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidName):
		respondError(w, nethttp.StatusUnprocessableEntity, "invalid_name", domain.ErrInvalidName.Error())
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, nethttp.StatusNotFound, "user_not_found", domain.ErrNotFound.Error())
	default:
		respondError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func respondError(w nethttp.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, dto.ErrorResponse{Error: dto.ErrorBody{Code: code, Message: message}})
}

func respondJSON(w nethttp.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
