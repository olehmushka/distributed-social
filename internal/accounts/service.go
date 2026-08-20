package accounts

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	logging "github.com/ipfs/go-log"

	"github.com/olehmushka/distributed-social/internal/eventbus"
	"github.com/olehmushka/distributed-social/internal/eventsapi"
)

var (
	ErrInvalidUsername = errors.New("username must be 3-32 characters of letters, digits, or underscores")
	ErrEmptyPost       = errors.New("post content must not be empty")
	ErrPostTooLong     = errors.New("post content must be 500 characters or fewer")
	ErrAuthorSuspended = errors.New("author is suspended")
)

const maxPostLength = 500

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type Publisher interface {
	Publish(ctx context.Context, subject string, env eventbus.Envelope) error
}

type Service struct {
	repo      Repository
	publisher Publisher
	logger    *logging.ZapEventLogger
}

func NewService(repo Repository, publisher Publisher, logger *logging.ZapEventLogger) *Service {
	return &Service{repo: repo, publisher: publisher, logger: logger}
}

func (s *Service) CreateUser(ctx context.Context, username, displayName string) (User, error) {
	if !usernamePattern.MatchString(username) {
		return User{}, ErrInvalidUsername
	}

	if _, err := s.repo.GetUserByUsername(ctx, username); err == nil {
		return User{}, ErrUsernameTaken
	} else if !errors.Is(err, ErrNotFound) {
		return User{}, fmt.Errorf("check existing username: %w", err)
	}

	return s.repo.CreateUser(ctx, User{
		Username:    username,
		DisplayName: displayName,
		Status:      UserStatusActive,
	})
}

func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *Service) CreatePost(ctx context.Context, authorID, content string) (Post, error) {
	if content == "" {
		return Post{}, ErrEmptyPost
	}
	if len(content) > maxPostLength {
		return Post{}, ErrPostTooLong
	}

	author, err := s.repo.GetUser(ctx, authorID)
	if err != nil {
		return Post{}, fmt.Errorf("look up post author: %w", err)
	}
	if author.Status == UserStatusSuspended {
		return Post{}, ErrAuthorSuspended
	}

	post, err := s.repo.CreatePost(ctx, Post{
		AuthorID: authorID,
		Content:  content,
		Status:   PostStatusActive,
	})
	if err != nil {
		return Post{}, err
	}

	env, err := eventbus.NewEnvelope(eventsapi.SubjectPostCreated, eventsapi.PostCreatedPayload{
		PostID:         post.ID,
		AuthorID:       post.AuthorID,
		AuthorUsername: author.Username,
		Content:        post.Content,
		CreatedAt:      post.CreatedAt,
	})
	if err != nil {
		s.logger.Errorw("failed to build post.created envelope", "postId", post.ID, "err", err)
		return post, nil
	}

	// The post is already durably persisted; a publish failure here means
	// the search index falls behind until a future event catches it up
	// rather than the write failing outright. A transactional outbox would
	// close this gap in a production deployment.
	if err := s.publisher.Publish(ctx, eventsapi.SubjectPostCreated, env); err != nil {
		s.logger.Errorw("failed to publish post.created event", "postId", post.ID, "err", err)
	}

	return post, nil
}

func (s *Service) GetPost(ctx context.Context, id string) (Post, error) {
	return s.repo.GetPost(ctx, id)
}

func (s *Service) ListUserPosts(ctx context.Context, authorID string, limit, offset int) ([]Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListPostsByAuthor(ctx, authorID, limit, offset)
}

// HandlePostRemoved reacts to a moderation decision made by the admins
// service. accounts never writes to admins' data directly, and vice versa
// -- this is the only channel that keeps the two consistent.
func (s *Service) HandlePostRemoved(ctx context.Context, payload eventsapi.PostRemovedPayload) error {
	return s.repo.UpdatePostStatus(ctx, payload.PostID, PostStatusRemoved)
}

func (s *Service) HandleUserSuspended(ctx context.Context, payload eventsapi.UserSuspendedPayload) error {
	return s.repo.UpdateUserStatus(ctx, payload.UserID, UserStatusSuspended)
}

func (s *Service) HandleUserRestored(ctx context.Context, payload eventsapi.UserRestoredPayload) error {
	return s.repo.UpdateUserStatus(ctx, payload.UserID, UserStatusActive)
}
