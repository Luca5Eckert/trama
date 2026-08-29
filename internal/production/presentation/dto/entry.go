package dto

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application"
)

type CreateEntryRequest struct {
	Colors []string `json:"colors"`
}

type ColorBatchResponse struct {
	ID       string `json:"id"`
	Color    string `json:"color"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type EntryResponse struct {
	ID           string               `json:"id"`
	ReceivedAt   time.Time            `json:"receivedAt"`
	ColorBatches []ColorBatchResponse `json:"colorBatches"`
}

func FromEntryResult(result application.EntryResult) EntryResponse {
	batches := make([]ColorBatchResponse, len(result.ColorBatches))
	for index, batch := range result.ColorBatches {
		batches[index] = ColorBatchResponse{
			ID: batch.ID, Color: batch.Color, Position: batch.Position, Status: batch.Status,
		}
	}
	return EntryResponse{ID: result.ID, ReceivedAt: result.ReceivedAt.UTC(), ColorBatches: batches}
}
