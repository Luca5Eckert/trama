package production

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Luca5Eckert/trama/internal/production/application/command"
	"github.com/Luca5Eckert/trama/internal/production/application/query"
	"github.com/Luca5Eckert/trama/internal/production/domain/port"
	"github.com/Luca5Eckert/trama/internal/production/domain/snapshot"
	"github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/postgres"
	"github.com/Luca5Eckert/trama/internal/production/infrastructure/adapter/system"
	presentationhttp "github.com/Luca5Eckert/trama/internal/production/presentation/http"
)

type Module struct {
	sizeSequenceController *presentationhttp.SizeSequenceController
	entryController        *presentationhttp.EntryController
	entryQueryController   *presentationhttp.EntryQueryController
	colorBatchController   *presentationhttp.ColorBatchController
}

func NewModule(pool *pgxpool.Pool) Module {
	sizeSequences := postgres.NewSizeSequenceRepository(pool)
	entries := postgres.NewEntryRepository(pool)
	reads := postgres.NewProductionReader(pool)
	batchReads := colorBatchReaderAdapter{reader: reads}
	clock := system.NewUTCClock()
	ids := system.NewRandomIDGenerator()

	replaceSizeSequence := command.NewReplaceSizeSequence(sizeSequences, clock)
	getSizeSequence := query.NewGetSizeSequence(sizeSequences)
	createEntry := command.NewCreateEntry(sizeSequences, entries, ids, clock)
	getEntry := query.NewGetEntry(reads)
	listEntries := query.NewListEntries(reads)
	getBatch := query.NewGetColorBatch(batchReads)
	listBatches := query.NewListColorBatches(batchReads)

	return Module{
		sizeSequenceController: presentationhttp.NewSizeSequenceController(replaceSizeSequence, getSizeSequence),
		entryController:        presentationhttp.NewEntryController(createEntry),
		entryQueryController:   presentationhttp.NewEntryQueryController(getEntry, listEntries),
		colorBatchController:   presentationhttp.NewColorBatchController(getBatch, listBatches),
	}
}

func (module Module) RegisterRoutes(mux *http.ServeMux) {
	module.sizeSequenceController.RegisterRoutes(mux)
	module.entryController.RegisterRoutes(mux)
	module.entryQueryController.RegisterRoutes(mux)
	module.colorBatchController.RegisterRoutes(mux)
}

type colorBatchReaderAdapter struct{ reader *postgres.ProductionReader }

func (adapter colorBatchReaderAdapter) GetByID(ctx context.Context, id string) (snapshot.ColorBatch, error) {
	return adapter.reader.GetColorBatchByID(ctx, id)
}

func (adapter colorBatchReaderAdapter) List(ctx context.Context, criteria port.ColorBatchListCriteria) ([]snapshot.ColorBatch, error) {
	return adapter.reader.ListColorBatches(ctx, criteria)
}
