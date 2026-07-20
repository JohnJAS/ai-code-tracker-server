package store

import (
	"context"

	"ai-code-tracker-server/internal/domain"
)

type Store interface {
	UpsertBatch(ctx context.Context, origin string, records []domain.Record) (InsertResult, error)
}

type InsertResult struct {
	Inserted   int `json:"inserted"`
	Duplicates int `json:"duplicates"`
}
