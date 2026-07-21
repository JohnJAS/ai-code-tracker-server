package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-code-tracker-server/internal/domain"
	_ "github.com/go-sql-driver/mysql"
)

const commitDateLayout = "2006-01-02 15:04:05"

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) UpsertBatch(ctx context.Context, origin string, records []domain.Record) (InsertResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return InsertResult{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO repositories (origin) VALUES (?)
		ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)`, origin)
	if err != nil {
		return InsertResult{}, err
	}
	repositoryID, err := result.LastInsertId()
	if err != nil {
		return InsertResult{}, err
	}

	statement, err := tx.PrepareContext(ctx, `
		INSERT IGNORE INTO commit_records
		(repository_id, commit_id, author, ai_lines, total_lines, is_ai_commit, committed_at, message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return InsertResult{}, err
	}
	defer statement.Close()

	inserted := 0
	for _, record := range records {
		committedAt, err := time.Parse(commitDateLayout, record.Date)
		if err != nil {
			return InsertResult{}, fmt.Errorf("parse commit date: %w", err)
		}
		result, err := statement.ExecContext(
			ctx,
			repositoryID,
			record.CommitID,
			record.Author,
			record.AiLines,
			record.TotalLines,
			record.IsAICommit,
			committedAt,
			record.Message,
		)
		if err != nil {
			return InsertResult{}, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return InsertResult{}, err
		}
		inserted += int(rows)
	}

	if err := tx.Commit(); err != nil {
		return InsertResult{}, err
	}
	return InsertResult{Inserted: inserted, Duplicates: len(records) - inserted}, nil
}

func (s *MySQLStore) Dashboard(ctx context.Context) (Dashboard, error) {
	var dashboard Dashboard
	if err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(is_ai_commit), 0),
			COALESCE(SUM(ai_lines), 0),
			COALESCE(SUM(total_lines), 0),
			(SELECT COUNT(*) FROM repositories)
		FROM commit_records`).Scan(
		&dashboard.TotalCommits,
		&dashboard.AICommits,
		&dashboard.AILines,
		&dashboard.TotalLines,
		&dashboard.Repositories,
	); err != nil {
		return Dashboard{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT r.origin, c.commit_id, c.author, c.ai_lines, c.total_lines, c.is_ai_commit,
			DATE_FORMAT(c.committed_at, '%Y-%m-%d %H:%i:%s'), c.message
		FROM commit_records c
		JOIN repositories r ON r.id = c.repository_id
		ORDER BY c.committed_at DESC, c.commit_id DESC
		LIMIT 20`)
	if err != nil {
		return Dashboard{}, err
	}
	defer rows.Close()

	dashboard.RecentRecords = make([]DashboardRecord, 0)
	for rows.Next() {
		var record DashboardRecord
		if err := rows.Scan(
			&record.RepositoryURL,
			&record.CommitID,
			&record.Author,
			&record.AILines,
			&record.TotalLines,
			&record.IsAICommit,
			&record.Date,
			&record.Message,
		); err != nil {
			return Dashboard{}, err
		}
		dashboard.RecentRecords = append(dashboard.RecentRecords, record)
	}
	if err := rows.Err(); err != nil {
		return Dashboard{}, err
	}
	return dashboard, nil
}
