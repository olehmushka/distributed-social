// Package pg provides the Postgres connection pool and a minimal embedded
// SQL migration runner shared by every service's storage layer.
package pg

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

// Migrate applies every *.sql file under dir (in an embedded fs.FS, sorted
// by filename) that hasn't been recorded in schema_migrations yet, each in
// its own transaction.
func Migrate(ctx context.Context, pool *pgxpool.Pool, files fs.FS, dir string) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     text PRIMARY KEY,
			applied_at  timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(files, dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var alreadyApplied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&alreadyApplied); err != nil {
			return fmt.Errorf("check migration %q: %w", name, err)
		}
		if alreadyApplied {
			continue
		}

		script, err := fs.ReadFile(files, path.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %q: %w", name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %q: %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(script)); err != nil {
			_ = tx.Rollback(ctx) //nolint:errcheck
			return fmt.Errorf("apply migration %q: %w", name, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx) //nolint:errcheck
			return fmt.Errorf("record migration %q: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %q: %w", name, err)
		}
	}

	return nil
}
