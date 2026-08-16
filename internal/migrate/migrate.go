package migrate

import (
	"context"
	"database/sql"
	"strings"

	"ai-code-tracker-server/internal/migrations"
)

func Apply(ctx context.Context, db *sql.DB) error {
	for _, statement := range strings.Split(migrations.InitialSQL(), ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	var columnCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
			AND table_name = 'commit_records'
			AND column_name = 'ai_tool'`).Scan(&columnCount); err != nil {
		return err
	}
	if columnCount == 0 {
		if _, err := db.ExecContext(ctx, `
			ALTER TABLE commit_records
			ADD COLUMN ai_tool VARCHAR(255) NOT NULL DEFAULT '' AFTER is_ai_commit`); err != nil {
			return err
		}
	}
	return nil
}
