package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (s *MySQLStore) Records(ctx context.Context, query RecordQuery) (RecordPage, error) {
	where, args := recordWhere(query)
	const source = `
		FROM commit_records c
		JOIN repositories r ON r.id = c.repository_id`

	var page RecordPage
	page.Records = make([]DashboardRecord, 0)
	page.Page = query.Page
	page.PageSize = query.PageSize

	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
			COALESCE(SUM(c.is_ai_commit), 0),
			COALESCE(SUM(c.ai_lines), 0),
			COALESCE(SUM(c.total_lines), 0),
			COUNT(DISTINCT c.repository_id),
			COUNT(DISTINCT c.author)`+source+where, args...).Scan(
		&page.TotalCommits,
		&page.AICommits,
		&page.AILines,
		&page.TotalLines,
		&page.Repositories,
		&page.Contributors,
	); err != nil {
		return RecordPage{}, err
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*)`+source+where, args...).Scan(&page.TotalRecords); err != nil {
		return RecordPage{}, err
	}
	if page.TotalRecords > 0 {
		page.TotalPages = (page.TotalRecords + query.PageSize - 1) / query.PageSize
	}

	pageArgs := append(append([]any{}, args...), query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.origin, c.commit_id, c.author, c.ai_lines, c.total_lines, c.is_ai_commit,
			DATE_FORMAT(c.committed_at, '%Y-%m-%d %H:%i:%s'), c.message`+source+where+`
		ORDER BY c.committed_at DESC, c.commit_id DESC
		LIMIT ? OFFSET ?`, pageArgs...)
	if err != nil {
		return RecordPage{}, err
	}
	defer rows.Close()

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
			return RecordPage{}, err
		}
		page.Records = append(page.Records, record)
	}
	if err := rows.Err(); err != nil {
		return RecordPage{}, err
	}
	return page, nil
}

func recordWhere(query RecordQuery) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if query.Author != "" {
		clauses = append(clauses, `c.author LIKE ? ESCAPE '\\'`)
		args = append(args, "%"+escapeLike(query.Author)+"%")
	}
	if query.Repository != "" {
		clauses = append(clauses, `r.origin LIKE ? ESCAPE '\\'`)
		args = append(args, "%"+escapeLike(query.Repository)+"%")
	}
	if !query.Start.IsZero() {
		clauses = append(clauses, "c.committed_at >= ?")
		args = append(args, query.Start)
	}
	if !query.End.IsZero() {
		clauses = append(clauses, "c.committed_at < ?")
		args = append(args, query.End)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	return strings.ReplaceAll(value, "_", `\_`)
}
