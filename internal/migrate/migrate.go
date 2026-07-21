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
	return nil
}
