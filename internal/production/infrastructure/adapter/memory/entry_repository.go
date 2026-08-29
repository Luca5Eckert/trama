package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/Luca5Eckert/trama/internal/production/domain/model"
)

type EntryRepository struct {
	mu      sync.RWMutex
	entries map[string]model.Entry
}

func NewEntryRepository() *EntryRepository {
	return &EntryRepository{entries: make(map[string]model.Entry)}
}

func (repository *EntryRepository) Create(_ context.Context, entry model.Entry) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, exists := repository.entries[entry.ID()]; exists {
		return fmt.Errorf("entry %s already exists", entry.ID())
	}
	repository.entries[entry.ID()] = entry
	return nil
}

func (repository *EntryRepository) Entries() []model.Entry {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entries := make([]model.Entry, 0, len(repository.entries))
	for _, entry := range repository.entries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ReceivedAt().Equal(entries[j].ReceivedAt()) {
			return entries[i].ID() < entries[j].ID()
		}
		return entries[i].ReceivedAt().Before(entries[j].ReceivedAt())
	})
	return entries
}
