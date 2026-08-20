package httputils_test

import (
	"context"
	"errors"
	"testing"

	"github.com/olehmushka/distributed-social/utils/httputils"
)

// CreateFail and CreateError both call GetMetadata, which returns its own
// (nil, in the common case) error using the same `err` identifier as the
// domain error being reported. A `metadata, err := GetMetadata(ctx)` inside
// either function silently overwrites the caller's error before it's
// formatted, and calling .Error() on the now-nil error panics. This test
// exists because that exact regression shipped once already and only
// surfaced under live HTTP traffic, never under mocked unit tests.
func TestCreateFail_PreservesOriginalError(t *testing.T) {
	domainErr := errors.New("boom")

	resp, err := httputils.CreateFail[any](context.Background(), domainErr, nil)
	if err != nil {
		t.Fatalf("CreateFail returned unexpected error: %v", err)
	}
	if resp.Message != domainErr.Error() {
		t.Fatalf("expected message %q, got %q", domainErr.Error(), resp.Message)
	}
}

func TestCreateError_PreservesOriginalError(t *testing.T) {
	domainErr := errors.New("boom")

	resp, err := httputils.CreateError(context.Background(), domainErr)
	if err != nil {
		t.Fatalf("CreateError returned unexpected error: %v", err)
	}
	if resp.Message != domainErr.Error() {
		t.Fatalf("expected message %q, got %q", domainErr.Error(), resp.Message)
	}
}
