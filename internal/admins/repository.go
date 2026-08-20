package admins

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	RecordAction(ctx context.Context, a ModerationAction) (ModerationAction, error)
	ListActions(ctx context.Context, limit, offset int) ([]ModerationAction, error)
}
