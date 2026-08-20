package search

import "context"

type Repository interface {
	UpsertDocument(ctx context.Context, doc Document) error
	SetPostRemoved(ctx context.Context, postID string, removed bool) error
	SetAuthorSuspended(ctx context.Context, authorID string, suspended bool) error
	Search(ctx context.Context, query string, limit, offset int) ([]Result, error)
}
