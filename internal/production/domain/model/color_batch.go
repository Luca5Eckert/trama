package model

import "time"

type ColorBatchStatus string

const ColorBatchWaiting ColorBatchStatus = "WAITING"

type ColorBatch struct {
	id          string
	entryID     string
	color       Color
	position    int
	status      ColorBatchStatus
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	sizeRuns    []SizeRun
}

func NewWaitingColorBatch(id, entryID string, color Color, position int, createdAt time.Time, sizeRuns []SizeRun) ColorBatch {
	return ColorBatch{
		id:        id,
		entryID:   entryID,
		color:     color,
		position:  position,
		status:    ColorBatchWaiting,
		createdAt: createdAt.UTC(),
		sizeRuns:  append([]SizeRun(nil), sizeRuns...),
	}
}

func (batch ColorBatch) ID() string { return batch.id }
func (batch ColorBatch) EntryID() string { return batch.entryID }
func (batch ColorBatch) Color() Color { return batch.color }
func (batch ColorBatch) Position() int { return batch.position }
func (batch ColorBatch) Status() ColorBatchStatus { return batch.status }
func (batch ColorBatch) CreatedAt() time.Time { return batch.createdAt.UTC() }
func (batch ColorBatch) StartedAt() *time.Time { return copyTime(batch.startedAt) }
func (batch ColorBatch) CompletedAt() *time.Time { return copyTime(batch.completedAt) }
func (batch ColorBatch) SizeRuns() []SizeRun { return append([]SizeRun(nil), batch.sizeRuns...) }
