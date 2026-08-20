package accounts

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolation = "23505"

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) CreateUser(ctx context.Context, u User) (User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, display_name, status)
		VALUES ($1, $2, $3)
		RETURNING id, username, display_name, status, created_at
	`, u.Username, u.DisplayName, u.Status)

	user, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return User{}, ErrUsernameTaken
		}
		return User{}, err
	}

	return user, nil
}

func (r *postgresRepository) GetUser(ctx context.Context, id string) (User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, display_name, status, created_at FROM users WHERE id = $1
	`, id)

	return scanUser(row)
}

func (r *postgresRepository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, display_name, status, created_at FROM users WHERE username = $1
	`, username)

	return scanUser(row)
}

func (r *postgresRepository) UpdateUserStatus(ctx context.Context, id string, status UserStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	return nil
}

func (r *postgresRepository) CreatePost(ctx context.Context, p Post) (Post, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO posts (author_id, content, status)
		VALUES ($1, $2, $3)
		RETURNING id, author_id, content, status, created_at
	`, p.AuthorID, p.Content, p.Status)

	return scanPost(row)
}

func (r *postgresRepository) GetPost(ctx context.Context, id string) (Post, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, author_id, content, status, created_at FROM posts WHERE id = $1
	`, id)

	return scanPost(row)
}

func (r *postgresRepository) ListPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]Post, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, author_id, content, status, created_at
		FROM posts
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts by author: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (r *postgresRepository) UpdatePostStatus(ctx context.Context, id string, status PostStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update post status: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Status, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}

func scanPost(row rowScanner) (Post, error) {
	var p Post
	err := row.Scan(&p.ID, &p.AuthorID, &p.Content, &p.Status, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, fmt.Errorf("scan post: %w", err)
	}
	return p, nil
}
