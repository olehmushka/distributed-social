package admins_test

import (
	"context"
	"errors"
	"testing"
	"time"

	logging "github.com/ipfs/go-log"

	"github.com/olehmushka/distributed-social/internal/admins"
	"github.com/olehmushka/distributed-social/internal/eventbus"
	"github.com/olehmushka/distributed-social/internal/eventsapi"
)

type fakeRepo struct {
	actions []admins.ModerationAction
	nextID  int
}

func (f *fakeRepo) RecordAction(_ context.Context, a admins.ModerationAction) (admins.ModerationAction, error) {
	f.nextID++
	a.ID = string(rune('a' + f.nextID))
	a.CreatedAt = time.Now()
	f.actions = append(f.actions, a)
	return a, nil
}

func (f *fakeRepo) ListActions(_ context.Context, limit, offset int) ([]admins.ModerationAction, error) {
	if offset >= len(f.actions) {
		return nil, nil
	}
	end := offset + limit
	if end > len(f.actions) {
		end = len(f.actions)
	}
	return f.actions[offset:end], nil
}

type fakePublisher struct {
	published []eventbus.Envelope
}

func (f *fakePublisher) Publish(_ context.Context, _ string, env eventbus.Envelope) error {
	f.published = append(f.published, env)
	return nil
}

func newTestService() (*admins.Service, *fakeRepo, *fakePublisher) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	return admins.NewService(repo, pub, logging.Logger("test")), repo, pub
}

func TestSuspendUser_RequiresReason(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.SuspendUser(context.Background(), "admin-1", "user-1", "")
	if !errors.Is(err, admins.ErrMissingReason) {
		t.Fatalf("expected ErrMissingReason, got %v", err)
	}
}

func TestSuspendUser_RecordsAndPublishes(t *testing.T) {
	svc, repo, pub := newTestService()

	action, err := svc.SuspendUser(context.Background(), "admin-1", "user-1", "spam")
	if err != nil {
		t.Fatalf("suspend user: %v", err)
	}
	if action.Action != admins.ActionSuspendUser {
		t.Fatalf("expected action %q, got %q", admins.ActionSuspendUser, action.Action)
	}
	if len(repo.actions) != 1 {
		t.Fatalf("expected 1 recorded action, got %d", len(repo.actions))
	}
	if len(pub.published) != 1 || pub.published[0].Type != eventsapi.SubjectUserSuspended {
		t.Fatalf("expected user.suspended event to be published, got %+v", pub.published)
	}
}

func TestRemovePost_RequiresTargetID(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.RemovePost(context.Background(), "admin-1", "", "abusive content")
	if !errors.Is(err, admins.ErrMissingTarget) {
		t.Fatalf("expected ErrMissingTarget, got %v", err)
	}
}
