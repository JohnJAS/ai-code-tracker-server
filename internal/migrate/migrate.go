package migrate

import (
	"context"
	"database/sql"

	"ai-code-tracker-server/internal/migrations"
)

func Apply(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, migrations.InitialSQL())
	return err
}
