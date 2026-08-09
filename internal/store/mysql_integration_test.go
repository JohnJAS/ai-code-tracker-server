package store_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

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

func TestMySQLStoreRecordsFiltersSummarizeAndPaginate(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}

	repository := store.NewMySQLStore(newMigratedDatabase(t, dsn))
	seedRecordPage(t, repository)

	page, err := repository.Records(context.Background(), store.RecordQuery{
		Author:     "dev",
		Repository: "acme",
		Start:      time.Date(2026, time.August, 2, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC),
		Page:       1,
		PageSize:   1,
	})
	if err != nil {
		t.Fatalf("Records() error = %v", err)
	}
	if page.TotalCommits != 2 || page.AICommits != 1 || page.AILines != 7 ||
		page.TotalLines != 10 || page.Repositories != 1 {
		t.Fatalf("statistics = %#v", page)
	}
	if page.TotalRecords != 2 || page.Page != 1 || page.PageSize != 1 || page.TotalPages != 2 {
		t.Fatalf("pagination = %#v", page)
	}
	if len(page.Records) != 1 || page.Records[0].CommitID != strings.Repeat("c", 40) {
		t.Fatalf("records = %#v", page.Records)
	}
}

func TestMySQLStoreRecordsReturnsEmptyOutOfRangePage(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}

	repository := store.NewMySQLStore(newMigratedDatabase(t, dsn))
	seedRecordPage(t, repository)

	page, err := repository.Records(context.Background(), store.RecordQuery{
		Author:   "dev",
		Page:     3,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("Records() error = %v", err)
	}
	if page.TotalRecords != 2 || page.TotalPages != 2 || page.Page != 3 {
		t.Fatalf("pagination = %#v", page)
	}
	if page.Records == nil || len(page.Records) != 0 {
		t.Fatalf("records = %#v, want empty non-nil slice", page.Records)
	}
}

func TestMySQLStoreRecordsEscapesLikeWildcards(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}

	repository := store.NewMySQLStore(newMigratedDatabase(t, dsn))
	seedRecordPage(t, repository)

	page, err := repository.Records(context.Background(), store.RecordQuery{
		Author:   "%",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("Records() error = %v", err)
	}
	if page.TotalRecords != 0 || page.TotalCommits != 0 || len(page.Records) != 0 {
		t.Fatalf("page = %#v, want no matches", page)
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

func seedRecordPage(t *testing.T, repository *store.MySQLStore) {
	t.Helper()
	records := []domain.Record{
		{
			Author:     "dev",
			CommitID:   strings.Repeat("a", 40),
			AiLines:    3,
			TotalLines: 4,
			IsAICommit: true,
			Date:       "2026-08-01 10:00:00",
			Message:    "outside start date",
		},
		{
			Author:     "dev",
			CommitID:   strings.Repeat("b", 40),
			AiLines:    7,
			TotalLines: 10,
			IsAICommit: true,
			Date:       "2026-08-02 10:00:00",
			Message:    "first match",
		},
		{
			Author:     "dev",
			CommitID:   strings.Repeat("c", 40),
			AiLines:    0,
			TotalLines: 0,
			IsAICommit: false,
			Date:       "2026-08-04 10:00:00",
			Message:    "end date match",
		},
	}
	if _, err := repository.UpsertBatch(context.Background(), "github.com/acme/demo", records); err != nil {
		t.Fatalf("seed first repository: %v", err)
	}
	if _, err := repository.UpsertBatch(context.Background(), "github.com/other/demo", []domain.Record{{
		Author:     "reviewer",
		CommitID:   strings.Repeat("d", 40),
		AiLines:    2,
		TotalLines: 5,
		IsAICommit: true,
		Date:       "2026-08-03 10:00:00",
		Message:    "other repository",
	}}); err != nil {
		t.Fatalf("seed second repository: %v", err)
	}
}
