package port

import (
	"context"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

type EntryRepository interface {
	Create(context.Context, model.Entry) error
}
