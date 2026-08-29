package model

import "time"

type SizeRunStatus string

const SizeRunPending SizeRunStatus = "PENDING"

type SizeRun struct {
	id             string
	colorBatchID   string
	sizeName       string
	position       int
	status         SizeRunStatus
	quantity       *int
	startedAt      *time.Time
	completedAt    *time.Time
}

func NewPendingSizeRun(id, colorBatchID string, size SizeDefinition) SizeRun {
	return SizeRun{
		id:           id,
		colorBatchID: colorBatchID,
		sizeName:     size.Name(),
		position:     size.Position(),
		status:       SizeRunPending,
	}
}

func (run SizeRun) ID() string { return run.id }
func (run SizeRun) ColorBatchID() string { return run.colorBatchID }
func (run SizeRun) SizeName() string { return run.sizeName }
func (run SizeRun) Position() int { return run.position }
func (run SizeRun) Status() SizeRunStatus { return run.status }

func (run SizeRun) Quantity() *int {
	if run.quantity == nil {
		return nil
	}
	value := *run.quantity
	return &value
}

func (run SizeRun) StartedAt() *time.Time {
	return copyTime(run.startedAt)
}

func (run SizeRun) CompletedAt() *time.Time {
	return copyTime(run.completedAt)
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := value.UTC()
	return &copy
}
