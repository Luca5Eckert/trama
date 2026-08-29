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
	sizeSequenceController *presentationhttp.SizeSequenceController
	entryController        *presentationhttp.EntryController
}

func NewModule(pool *pgxpool.Pool) Module {
	sizeSequences := postgres.NewSizeSequenceRepository(pool)
	entries := postgres.NewEntryRepository(pool)
	clock := system.NewUTCClock()
	ids := system.NewRandomIDGenerator()

	replaceSizeSequence := command.NewReplaceSizeSequence(sizeSequences, clock)
	getSizeSequence := query.NewGetSizeSequence(sizeSequences)
	createEntry := command.NewCreateEntry(sizeSequences, entries, ids, clock)

	return Module{
		sizeSequenceController: presentationhttp.NewSizeSequenceController(replaceSizeSequence, getSizeSequence),
		entryController:        presentationhttp.NewEntryController(createEntry),
	}
}

func (module Module) RegisterRoutes(mux *http.ServeMux) {
	module.sizeSequenceController.RegisterRoutes(mux)
	module.entryController.RegisterRoutes(mux)
}
