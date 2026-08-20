package admins

import (
	"context"
	"errors"

	logging "github.com/ipfs/go-log"

	"github.com/olehmushka/distributed-social/internal/eventbus"
	"github.com/olehmushka/distributed-social/internal/eventsapi"
)

var (
	ErrMissingActor  = errors.New("actor id is required")
	ErrMissingTarget = errors.New("target id is required")
	ErrMissingReason = errors.New("reason is required")
)

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

func (s *Service) SuspendUser(ctx context.Context, actorID, userID, reason string) (ModerationAction, error) {
	action, err := s.recordAndPublish(ctx, actorID, TargetUser, userID, ActionSuspendUser, reason,
		eventsapi.SubjectUserSuspended, eventsapi.UserSuspendedPayload{UserID: userID})
	return action, err
}

func (s *Service) RestoreUser(ctx context.Context, actorID, userID, reason string) (ModerationAction, error) {
	return s.recordAndPublish(ctx, actorID, TargetUser, userID, ActionRestoreUser, reason,
		eventsapi.SubjectUserRestored, eventsapi.UserRestoredPayload{UserID: userID})
}

func (s *Service) RemovePost(ctx context.Context, actorID, postID, reason string) (ModerationAction, error) {
	return s.recordAndPublish(ctx, actorID, TargetPost, postID, ActionRemovePost, reason,
		eventsapi.SubjectPostRemoved, eventsapi.PostRemovedPayload{PostID: postID})
}

func (s *Service) ListActions(ctx context.Context, limit, offset int) ([]ModerationAction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListActions(ctx, limit, offset)
}

func (s *Service) recordAndPublish(ctx context.Context, actorID string, target TargetType, targetID string, action ActionType, reason, subject string, payload any) (ModerationAction, error) {
	if actorID == "" {
		return ModerationAction{}, ErrMissingActor
	}
	if targetID == "" {
		return ModerationAction{}, ErrMissingTarget
	}
	if reason == "" {
		return ModerationAction{}, ErrMissingReason
	}

	recorded, err := s.repo.RecordAction(ctx, ModerationAction{
		ActorID:  actorID,
		Target:   target,
		TargetID: targetID,
		Action:   action,
		Reason:   reason,
	})
	if err != nil {
		return ModerationAction{}, err
	}

	env, err := eventbus.NewEnvelope(subject, payload)
	if err != nil {
		s.logger.Errorw("failed to build moderation envelope", "action", action, "targetId", targetID, "err", err)
		return recorded, nil
	}

	if err := s.publisher.Publish(ctx, subject, env); err != nil {
		s.logger.Errorw("failed to publish moderation event", "action", action, "targetId", targetID, "err", err)
	}

	return recorded, nil
}
