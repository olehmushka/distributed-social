package search_test

import (
	"context"
	"errors"
	"testing"

	logging "github.com/ipfs/go-log"

	"github.com/olehmushka/distributed-social/internal/eventsapi"
	"github.com/olehmushka/distributed-social/internal/search"
)

type fakeRepo struct {
	docs            map[string]search.Document
	removed         map[string]bool
	authorSuspended map[string]bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		docs:            map[string]search.Document{},
		removed:         map[string]bool{},
		authorSuspended: map[string]bool{},
	}
}

func (f *fakeRepo) UpsertDocument(_ context.Context, doc search.Document) error {
	f.docs[doc.PostID] = doc
	return nil
}

func (f *fakeRepo) SetPostRemoved(_ context.Context, postID string, removed bool) error {
	f.removed[postID] = removed
	return nil
}

func (f *fakeRepo) SetAuthorSuspended(_ context.Context, authorID string, suspended bool) error {
	f.authorSuspended[authorID] = suspended
	return nil
}

func (f *fakeRepo) Search(_ context.Context, query string, _, _ int) ([]search.Result, error) {
	var results []search.Result
	for _, doc := range f.docs {
		if f.removed[doc.PostID] || f.authorSuspended[doc.AuthorID] {
			continue
		}
		results = append(results, search.Result{Document: doc, Rank: 1})
	}
	return results, nil
}

func TestSearch_RejectsEmptyQuery(t *testing.T) {
	svc := search.NewService(newFakeRepo(), logging.Logger("test"))

	_, err := svc.Search(context.Background(), "  ", 20, 0)
	if !errors.Is(err, search.ErrEmptyQuery) {
		t.Fatalf("expected ErrEmptyQuery, got %v", err)
	}
}

func TestHandlePostCreated_ThenSearch_FindsDocument(t *testing.T) {
	repo := newFakeRepo()
	svc := search.NewService(repo, logging.Logger("test"))
	ctx := context.Background()

	if err := svc.HandlePostCreated(ctx, eventsapi.PostCreatedPayload{
		PostID: "post-1", AuthorID: "author-1", AuthorUsername: "alice", Content: "hello world",
	}); err != nil {
		t.Fatalf("handle post created: %v", err)
	}

	results, err := svc.Search(ctx, "hello", 20, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].PostID != "post-1" {
		t.Fatalf("expected to find post-1, got %+v", results)
	}
}

func TestHandlePostRemoved_ExcludesFromSearch(t *testing.T) {
	repo := newFakeRepo()
	svc := search.NewService(repo, logging.Logger("test"))
	ctx := context.Background()

	_ = svc.HandlePostCreated(ctx, eventsapi.PostCreatedPayload{
		PostID: "post-1", AuthorID: "author-1", AuthorUsername: "alice", Content: "hello world",
	})
	if err := svc.HandlePostRemoved(ctx, eventsapi.PostRemovedPayload{PostID: "post-1"}); err != nil {
		t.Fatalf("handle post removed: %v", err)
	}

	results, err := svc.Search(ctx, "hello", 20, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected removed post to be excluded, got %+v", results)
	}
}

func TestHandleUserSuspendedThenRestored_TogglesVisibility(t *testing.T) {
	repo := newFakeRepo()
	svc := search.NewService(repo, logging.Logger("test"))
	ctx := context.Background()

	_ = svc.HandlePostCreated(ctx, eventsapi.PostCreatedPayload{
		PostID: "post-1", AuthorID: "author-1", AuthorUsername: "alice", Content: "hello world",
	})
	if err := svc.HandleUserSuspended(ctx, eventsapi.UserSuspendedPayload{UserID: "author-1"}); err != nil {
		t.Fatalf("handle user suspended: %v", err)
	}
	if results, _ := svc.Search(ctx, "hello", 20, 0); len(results) != 0 {
		t.Fatalf("expected suspended author's posts to be excluded, got %+v", results)
	}

	if err := svc.HandleUserRestored(ctx, eventsapi.UserRestoredPayload{UserID: "author-1"}); err != nil {
		t.Fatalf("handle user restored: %v", err)
	}
	if results, _ := svc.Search(ctx, "hello", 20, 0); len(results) != 1 {
		t.Fatalf("expected restored author's posts to reappear, got %+v", results)
	}
}
