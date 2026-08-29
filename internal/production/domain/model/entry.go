package model

import "time"

type Entry struct {
	id           string
	receivedAt   time.Time
	colorBatches []ColorBatch
}

func NewEntry(id string, receivedAt time.Time, colorBatches []ColorBatch) Entry {
	return Entry{
		id:           id,
		receivedAt:   receivedAt.UTC(),
		colorBatches: append([]ColorBatch(nil), colorBatches...),
	}
}

func (entry Entry) ID() string { return entry.id }
func (entry Entry) ReceivedAt() time.Time { return entry.receivedAt.UTC() }
func (entry Entry) ColorBatches() []ColorBatch { return append([]ColorBatch(nil), entry.colorBatches...) }
