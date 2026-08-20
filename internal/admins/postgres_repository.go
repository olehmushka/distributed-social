package admins

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) RecordAction(ctx context.Context, a ModerationAction) (ModerationAction, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO moderation_actions (actor_id, target_type, target_id, action, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, actor_id, target_type, target_id, action, reason, created_at
	`, a.ActorID, a.Target, a.TargetID, a.Action, a.Reason)

	return scanAction(row)
}

func (r *postgresRepository) ListActions(ctx context.Context, limit, offset int) ([]ModerationAction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, actor_id, target_type, target_id, action, reason, created_at
		FROM moderation_actions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list moderation actions: %w", err)
	}
	defer rows.Close()

	var actions []ModerationAction
	for rows.Next() {
		a, err := scanAction(rows)
		if err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}

	return actions, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAction(row rowScanner) (ModerationAction, error) {
	var a ModerationAction
	err := row.Scan(&a.ID, &a.ActorID, &a.Target, &a.TargetID, &a.Action, &a.Reason, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ModerationAction{}, ErrNotFound
	}
	if err != nil {
		return ModerationAction{}, fmt.Errorf("scan moderation action: %w", err)
	}
	return a, nil
}
