package accounts

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/olehmushka/distributed-social/internal/pg"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return pg.Migrate(ctx, pool, migrationFiles, "migrations")
}
