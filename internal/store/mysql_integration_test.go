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
