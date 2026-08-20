package accounts

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrUsernameTaken = errors.New("username already taken")
)

type Repository interface {
	CreateUser(ctx context.Context, u User) (User, error)
	GetUser(ctx context.Context, id string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	UpdateUserStatus(ctx context.Context, id string, status UserStatus) error

	CreatePost(ctx context.Context, p Post) (Post, error)
	GetPost(ctx context.Context, id string) (Post, error)
	ListPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]Post, error)
	UpdatePostStatus(ctx context.Context, id string, status PostStatus) error
}
