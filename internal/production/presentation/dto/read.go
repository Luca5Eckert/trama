package dto

import (
	"time"

	"github.com/Luca5Eckert/trama/internal/production/application"
)

type PageResponse struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type ReadSizeRunResponse struct {
	ID          string     `json:"id"`
	Size        string     `json:"size"`
	Position    int        `json:"position"`
	Status      string     `json:"status"`
	Quantity    *int       `json:"quantity"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

type ReadColorBatchResponse struct {
	ID          string     `json:"id"`
	EntryID     string     `json:"entryId"`
	Color       string     `json:"color"`
	Position    int        `json:"position"`
	Status      string     `json:"status"`
	CurrentSize *string    `json:"currentSize"`
	NextSize    *string    `json:"nextSize"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

type ColorBatchDetailResponse struct {
	ReadColorBatchResponse
	SizeRuns []ReadSizeRunResponse `json:"sizeRuns"`
}

type EntryDetailResponse struct {
	ID           string                   `json:"id"`
	ReceivedAt   time.Time                `json:"receivedAt"`
	ColorBatches []ReadColorBatchResponse `json:"colorBatches"`
}

type EntrySummaryResponse struct {
	ID              string    `json:"id"`
	ReceivedAt      time.Time `json:"receivedAt"`
	ColorBatchCount int       `json:"colorBatchCount"`
}

type EntryListResponse struct {
	Items []EntrySummaryResponse `json:"items"`
	Page  PageResponse           `json:"page"`
}

type ColorBatchListResponse struct {
	Items []ReadColorBatchResponse `json:"items"`
	Page  PageResponse             `json:"page"`
}

func FromEntryDetailResult(result application.EntryDetailResult) EntryDetailResponse {
	batches := make([]ReadColorBatchResponse, len(result.ColorBatches))
	for index, batch := range result.ColorBatches { batches[index] = fromReadColorBatch(batch) }
	return EntryDetailResponse{ID: result.ID, ReceivedAt: result.ReceivedAt.UTC(), ColorBatches: batches}
}

func FromEntryListResult(result application.EntryListResult) EntryListResponse {
	items := make([]EntrySummaryResponse, len(result.Items))
	for index, item := range result.Items {
		items[index] = EntrySummaryResponse{ID: item.ID, ReceivedAt: item.ReceivedAt.UTC(), ColorBatchCount: item.ColorBatchCount}
	}
	return EntryListResponse{Items: items, Page: PageResponse{Limit: result.Page.Limit, Offset: result.Page.Offset, Count: result.Page.Count}}
}

func FromColorBatchDetailResult(result application.ReadColorBatchResult) ColorBatchDetailResponse {
	runs := make([]ReadSizeRunResponse, len(result.SizeRuns))
	for index, run := range result.SizeRuns {
		runs[index] = ReadSizeRunResponse{ID: run.ID, Size: run.Size, Position: run.Position, Status: run.Status, Quantity: run.Quantity, StartedAt: run.StartedAt, CompletedAt: run.CompletedAt}
	}
	return ColorBatchDetailResponse{ReadColorBatchResponse: fromReadColorBatch(result), SizeRuns: runs}
}

func FromColorBatchListResult(result application.ColorBatchListResult) ColorBatchListResponse {
	items := make([]ReadColorBatchResponse, len(result.Items))
	for index, item := range result.Items { items[index] = fromReadColorBatch(item) }
	return ColorBatchListResponse{Items: items, Page: PageResponse{Limit: result.Page.Limit, Offset: result.Page.Offset, Count: result.Page.Count}}
}

func fromReadColorBatch(result application.ReadColorBatchResult) ReadColorBatchResponse {
	return ReadColorBatchResponse{
		ID: result.ID, EntryID: result.EntryID, Color: result.Color, Position: result.Position, Status: result.Status,
		CurrentSize: result.CurrentSize, NextSize: result.NextSize, CreatedAt: result.CreatedAt.UTC(), StartedAt: result.StartedAt, CompletedAt: result.CompletedAt,
	}
}
