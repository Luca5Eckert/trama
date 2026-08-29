package dto

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application"
)

type SizeItemRequest struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type ReplaceSizeSequenceRequest struct {
	Items []SizeItemRequest `json:"items"`
}

type SizeItemResponse struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type SizeSequenceResponse struct {
	Items     []SizeItemResponse `json:"items"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func FromResult(result application.SizeSequenceResult) SizeSequenceResponse {
	items := make([]SizeItemResponse, len(result.Items))
	for index, item := range result.Items {
		items[index] = SizeItemResponse{Name: item.Name, Position: item.Position}
	}
	return SizeSequenceResponse{Items: items, UpdatedAt: result.UpdatedAt.UTC()}
}
