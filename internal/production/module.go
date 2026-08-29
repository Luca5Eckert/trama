package production

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/postgres"
	"github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/system"
	presentationhttp "github.com/Luca5Eckert/trama/internal/production/presentation/http"
)

type Module struct {
	controller *presentationhttp.SizeSequenceController
}

func NewModule(pool *pgxpool.Pool) Module {
	repository := postgres.NewSizeSequenceRepository(pool)
	replaceSizeSequence := command.NewReplaceSizeSequence(repository, system.NewUTCClock())
	getSizeSequence := query.NewGetSizeSequence(repository)
	return Module{controller: presentationhttp.NewSizeSequenceController(replaceSizeSequence, getSizeSequence)}
}

func (module Module) RegisterRoutes(mux *http.ServeMux) {
	module.controller.RegisterRoutes(mux)
}
