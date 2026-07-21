package store

import (
	"context"

	"ai-code-tracker-server/internal/domain"
)

type Store interface {
	UpsertBatch(ctx context.Context, origin string, records []domain.Record) (InsertResult, error)
	Dashboard(ctx context.Context) (Dashboard, error)
}

type InsertResult struct {
	Inserted   int `json:"inserted"`
	Duplicates int `json:"duplicates"`
}

type Dashboard struct {
	TotalCommits  int               `json:"total_commits"`
	AICommits     int               `json:"ai_commits"`
	AILines       int               `json:"ai_lines"`
	TotalLines    int               `json:"total_lines"`
	Repositories  int               `json:"repositories"`
	RecentRecords []DashboardRecord `json:"recent_records"`
}

type DashboardRecord struct {
	RepositoryURL string `json:"repository_url"`
	CommitID      string `json:"commit_id"`
	Author        string `json:"author"`
	AILines       int    `json:"ai_lines"`
	TotalLines    int    `json:"total_lines"`
	IsAICommit    bool   `json:"is_ai_commit"`
	Date          string `json:"date"`
	Message       string `json:"message"`
}
