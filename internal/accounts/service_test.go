package accounts_test

import (
	"context"
	"errors"
	"testing"

	logging "github.com/ipfs/go-log"

	"github.com/olehmushka/distributed-social/internal/accounts"
	"github.com/olehmushka/distributed-social/internal/eventbus"
	"github.com/olehmushka/distributed-social/internal/eventsapi"
)

type fakeRepo struct {
	usersByID       map[string]accounts.User
	usersByUsername map[string]accounts.User
	posts           map[string]accounts.Post
	nextID          int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		usersByID:       map[string]accounts.User{},
		usersByUsername: map[string]accounts.User{},
		posts:           map[string]accounts.Post{},
	}
}

func (f *fakeRepo) genID() string {
	f.nextID++
	return string(rune('a' + f.nextID))
}

func (f *fakeRepo) CreateUser(_ context.Context, u accounts.User) (accounts.User, error) {
	if _, ok := f.usersByUsername[u.Username]; ok {
		return accounts.User{}, accounts.ErrUsernameTaken
	}
	u.ID = f.genID()
	f.usersByID[u.ID] = u
	f.usersByUsername[u.Username] = u
	return u, nil
}

func (f *fakeRepo) GetUser(_ context.Context, id string) (accounts.User, error) {
	u, ok := f.usersByID[id]
	if !ok {
		return accounts.User{}, accounts.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) GetUserByUsername(_ context.Context, username string) (accounts.User, error) {
	u, ok := f.usersByUsername[username]
	if !ok {
		return accounts.User{}, accounts.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) UpdateUserStatus(_ context.Context, id string, status accounts.UserStatus) error {
	u, ok := f.usersByID[id]
	if !ok {
		return accounts.ErrNotFound
	}
	u.Status = status
	f.usersByID[id] = u
	f.usersByUsername[u.Username] = u
	return nil
}

func (f *fakeRepo) CreatePost(_ context.Context, p accounts.Post) (accounts.Post, error) {
	p.ID = f.genID()
	f.posts[p.ID] = p
	return p, nil
}

func (f *fakeRepo) GetPost(_ context.Context, id string) (accounts.Post, error) {
	p, ok := f.posts[id]
	if !ok {
		return accounts.Post{}, accounts.ErrNotFound
	}
	return p, nil
}

func (f *fakeRepo) ListPostsByAuthor(_ context.Context, authorID string, _, _ int) ([]accounts.Post, error) {
	var out []accounts.Post
	for _, p := range f.posts {
		if p.AuthorID == authorID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakeRepo) UpdatePostStatus(_ context.Context, id string, status accounts.PostStatus) error {
	p, ok := f.posts[id]
	if !ok {
		return accounts.ErrNotFound
	}
	p.Status = status
	f.posts[id] = p
	return nil
}

type fakePublisher struct {
	published []eventbus.Envelope
}

func (f *fakePublisher) Publish(_ context.Context, _ string, env eventbus.Envelope) error {
	f.published = append(f.published, env)
	return nil
}

func newTestService() (*accounts.Service, *fakeRepo, *fakePublisher) {
	repo := newFakeRepo()
	pub := &fakePublisher{}
	svc := accounts.NewService(repo, pub, logging.Logger("test"))
	return svc, repo, pub
}

func TestCreateUser_RejectsInvalidUsername(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.CreateUser(context.Background(), "a", "A")
	if !errors.Is(err, accounts.ErrInvalidUsername) {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestCreateUser_RejectsDuplicateUsername(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	if _, err := svc.CreateUser(ctx, "alice", "Alice"); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := svc.CreateUser(ctx, "alice", "Someone Else")
	if !errors.Is(err, accounts.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestCreatePost_PublishesPostCreatedEvent(t *testing.T) {
	svc, _, pub := newTestService()
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "alice", "Alice")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	post, err := svc.CreatePost(ctx, user.ID, "hello distributed social")
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.published))
	}
	if pub.published[0].Type != eventsapi.SubjectPostCreated {
		t.Fatalf("expected event type %q, got %q", eventsapi.SubjectPostCreated, pub.published[0].Type)
	}
	if post.Status != accounts.PostStatusActive {
		t.Fatalf("expected new post to be active, got %q", post.Status)
	}
}

func TestCreatePost_RejectsSuspendedAuthor(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "alice", "Alice")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := repo.UpdateUserStatus(ctx, user.ID, accounts.UserStatusSuspended); err != nil {
		t.Fatalf("suspend user: %v", err)
	}

	_, err = svc.CreatePost(ctx, user.ID, "should not be allowed")
	if !errors.Is(err, accounts.ErrAuthorSuspended) {
		t.Fatalf("expected ErrAuthorSuspended, got %v", err)
	}
}

func TestCreatePost_RejectsOversizedContent(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "alice", "Alice")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	huge := make([]byte, 501)
	for i := range huge {
		huge[i] = 'a'
	}

	_, err = svc.CreatePost(ctx, user.ID, string(huge))
	if !errors.Is(err, accounts.ErrPostTooLong) {
		t.Fatalf("expected ErrPostTooLong, got %v", err)
	}
}

func TestHandlePostRemoved_IsIdempotent(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	user, _ := svc.CreateUser(ctx, "alice", "Alice")
	post, _ := svc.CreatePost(ctx, user.ID, "will be removed")

	for i := 0; i < 2; i++ {
		if err := svc.HandlePostRemoved(ctx, eventsapi.PostRemovedPayload{PostID: post.ID}); err != nil {
			t.Fatalf("handle post removed (iteration %d): %v", i, err)
		}
	}

	got, err := repo.GetPost(ctx, post.ID)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if got.Status != accounts.PostStatusRemoved {
		t.Fatalf("expected post status removed, got %q", got.Status)
	}
}
