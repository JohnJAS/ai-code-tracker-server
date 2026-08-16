package store

import (
	"context"
	"time"

	"ai-code-tracker-server/internal/domain"
)

type Store interface {
	UpsertBatch(ctx context.Context, origin string, records []domain.Record) (InsertResult, error)
	Dashboard(ctx context.Context) (Dashboard, error)
	Records(ctx context.Context, query RecordQuery) (RecordPage, error)
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

type RecordQuery struct {
	Author     string
	Repository string
	Start      time.Time
	End        time.Time
	Page       int
	PageSize   int
}

type RecordPage struct {
	TotalCommits int               `json:"total_commits"`
	AICommits    int               `json:"ai_commits"`
	AILines      int               `json:"ai_lines"`
	TotalLines   int               `json:"total_lines"`
	Repositories int               `json:"repositories"`
	Contributors int               `json:"contributors"`
	Records      []DashboardRecord `json:"records"`
	TotalRecords int               `json:"total_records"`
	Page         int               `json:"page"`
	PageSize     int               `json:"page_size"`
	TotalPages   int               `json:"total_pages"`
}
