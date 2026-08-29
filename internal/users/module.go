package users

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/users/application/command"
	"github.com/Luca5Eckert/trama/internal/users/application/query"
	"github.com/Luca5Eckert/trama/internal/users/infrastructure/adapter/postgres"
	"github.com/Luca5Eckert/trama/internal/users/infrastructure/adapter/system"
	presentationhttp "github.com/Luca5Eckert/trama/internal/users/presentation/http"
)

type Module struct { controller *presentationhttp.UserController }

func NewModule(pool *pgxpool.Pool) Module {
	repository := postgres.NewUserRepository(pool)
	createUser := command.NewCreateUser(repository, system.NewRandomIDGenerator(), system.NewUTCClock())
	return Module{controller: presentationhttp.NewUserController(createUser, query.NewGetUserByID(repository), query.NewListUsers(repository))}
}

func (module Module) RegisterRoutes(mux *http.ServeMux) { module.controller.RegisterRoutes(mux) }
