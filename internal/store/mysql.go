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
