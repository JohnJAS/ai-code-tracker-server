package store_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"ai-code-tracker-server/internal/domain"
	"ai-code-tracker-server/internal/migrate"
	"ai-code-tracker-server/internal/store"
)

func TestMySQLStoreUpsertIsIdempotent(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}

	db := newMigratedDatabase(t, dsn)
	repository := store.NewMySQLStore(db)
	records := []domain.Record{validRecord()}

	first, err := repository.UpsertBatch(context.Background(), "github.com/acme/demo", records)
	if err != nil {
		t.Fatalf("first UpsertBatch() error = %v", err)
	}
	second, err := repository.UpsertBatch(context.Background(), "github.com/acme/demo", records)
	if err != nil {
		t.Fatalf("second UpsertBatch() error = %v", err)
	}
	if first.Inserted != 1 || first.Duplicates != 0 {
		t.Fatalf("first result = %#v", first)
	}
	if second.Inserted != 0 || second.Duplicates != 1 {
		t.Fatalf("second result = %#v", second)
	}
}

func TestMySQLStoreDashboardSummarizesAndOrdersRecords(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}

	db := newMigratedDatabase(t, dsn)
	repository := store.NewMySQLStore(db)
	first := validRecord()
	second := domain.Record{
		Author:     "reviewer",
		CommitID:   strings.Repeat("b", 40),
		AiLines:    0,
		TotalLines: 3,
		IsAICommit: false,
		Date:       "2026-07-20 10:01:00",
		Message:    "fix: dashboard",
	}
	if _, err := repository.UpsertBatch(context.Background(), "github.com/acme/demo", []domain.Record{first, second}); err != nil {
		t.Fatalf("UpsertBatch() error = %v", err)
	}

	dashboard, err := repository.Dashboard(context.Background())
	if err != nil {
		t.Fatalf("Dashboard() error = %v", err)
	}
	if dashboard.TotalCommits != 2 || dashboard.AICommits != 1 || dashboard.AILines != 1 || dashboard.TotalLines != 5 || dashboard.Repositories != 1 {
		t.Fatalf("dashboard summary = %#v", dashboard)
	}
	if len(dashboard.RecentRecords) != 2 {
		t.Fatalf("recent records = %#v", dashboard.RecentRecords)
	}
	if dashboard.RecentRecords[0].CommitID != second.CommitID || dashboard.RecentRecords[0].Message != second.Message {
		t.Fatalf("newest record = %#v", dashboard.RecentRecords[0])
	}
}

func newMigratedDatabase(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("DROP TABLE IF EXISTS commit_records"); err != nil {
		t.Fatalf("drop commit_records: %v", err)
	}
	if _, err := db.Exec("DROP TABLE IF EXISTS repositories"); err != nil {
		t.Fatalf("drop repositories: %v", err)
	}
	if err := migrate.Apply(context.Background(), db); err != nil {
		t.Fatalf("migrate.Apply() error = %v", err)
	}
	return db
}

func validRecord() domain.Record {
	return domain.Record{
		Author:     "dev",
		CommitID:   strings.Repeat("a", 40),
		AiLines:    1,
		TotalLines: 2,
		IsAICommit: true,
		Date:       "2026-07-20 10:00:00",
		Message:    "feat: demo",
	}
}
