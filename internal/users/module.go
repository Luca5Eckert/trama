package users

import (
	"net/http"

	"github.com/Luca5Eckert/trama/internal/users/application"
	"github.com/Luca5Eckert/trama/internal/users/infrastructure/memory"
	transport "github.com/Luca5Eckert/trama/internal/users/transport/http"
)

type Module struct{ handler *transport.Handler }

func NewModule() Module {
	repository := memory.NewUserRepository()
	service := application.NewService(repository)
	return Module{handler: transport.NewHandler(service)}
}

func (m Module) RegisterRoutes(mux *http.ServeMux) { m.handler.RegisterRoutes(mux) }
