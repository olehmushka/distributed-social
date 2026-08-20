package search

import (
	"context"
	"errors"
	"strings"

	logging "github.com/ipfs/go-log"

	"github.com/olehmushka/distributed-social/internal/eventsapi"
)

var ErrEmptyQuery = errors.New("query must not be empty")

type Service struct {
	repo   Repository
	logger *logging.ZapEventLogger
}

func NewService(repo Repository, logger *logging.ZapEventLogger) *Service {
	return &Service{repo: repo, logger: logger}
}

// HandlePostCreated indexes a post published by the accounts service. It is
// safe to run more than once for the same post -- the repository upserts by
// post id -- so at-least-once JetStream delivery doesn't produce duplicates.
func (s *Service) HandlePostCreated(ctx context.Context, payload eventsapi.PostCreatedPayload) error {
	return s.repo.UpsertDocument(ctx, Document{
		PostID:         payload.PostID,
		AuthorID:       payload.AuthorID,
		AuthorUsername: payload.AuthorUsername,
		Content:        payload.Content,
		CreatedAt:      payload.CreatedAt,
	})
}

func (s *Service) HandlePostRemoved(ctx context.Context, payload eventsapi.PostRemovedPayload) error {
	return s.repo.SetPostRemoved(ctx, payload.PostID, true)
}

func (s *Service) HandleUserSuspended(ctx context.Context, payload eventsapi.UserSuspendedPayload) error {
	return s.repo.SetAuthorSuspended(ctx, payload.UserID, true)
}

func (s *Service) HandleUserRestored(ctx context.Context, payload eventsapi.UserRestoredPayload) error {
	return s.repo.SetAuthorSuspended(ctx, payload.UserID, false)
}

func (s *Service) Search(ctx context.Context, query string, limit, offset int) ([]Result, error) {
	if strings.TrimSpace(query) == "" {
		return nil, ErrEmptyQuery
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.Search(ctx, query, limit, offset)
}
