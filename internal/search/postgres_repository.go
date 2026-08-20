package search

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) UpsertDocument(ctx context.Context, doc Document) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO documents (post_id, author_id, author_username, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (post_id) DO UPDATE SET
			author_username = EXCLUDED.author_username,
			content         = EXCLUDED.content,
			updated_at      = now()
	`, doc.PostID, doc.AuthorID, doc.AuthorUsername, doc.Content, doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert search document: %w", err)
	}
	return nil
}

func (r *postgresRepository) SetPostRemoved(ctx context.Context, postID string, removed bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE documents SET removed = $2, updated_at = now() WHERE post_id = $1`, postID, removed)
	if err != nil {
		return fmt.Errorf("set post removed: %w", err)
	}
	return nil
}

func (r *postgresRepository) SetAuthorSuspended(ctx context.Context, authorID string, suspended bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE documents SET author_suspended = $2, updated_at = now() WHERE author_id = $1`, authorID, suspended)
	if err != nil {
		return fmt.Errorf("set author suspended: %w", err)
	}
	return nil
}

func (r *postgresRepository) Search(ctx context.Context, query string, limit, offset int) ([]Result, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT post_id, author_id, author_username, content, created_at,
		       ts_rank(search_vector, websearch_to_tsquery('english', $1)) AS rank
		FROM documents
		WHERE removed = false
		  AND author_suspended = false
		  AND search_vector @@ websearch_to_tsquery('english', $1)
		ORDER BY rank DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search documents: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var res Result
		if err := rows.Scan(&res.PostID, &res.AuthorID, &res.AuthorUsername, &res.Content, &res.CreatedAt, &res.Rank); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		results = append(results, res)
	}

	return results, rows.Err()
}
