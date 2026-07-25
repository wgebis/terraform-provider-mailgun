package framework

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/mailgun/mailgun-go/v5"
)

func notFoundErr() error {
	return &mailgun.UnexpectedResponseError{Expected: []int{200}, Actual: http.StatusNotFound}
}

func serverErr() error {
	return &mailgun.UnexpectedResponseError{Expected: []int{200}, Actual: http.StatusInternalServerError}
}

// A 404 from the probe means the deletion has propagated: success, no error.
func TestWaitForDeletion_NotFoundSucceeds(t *testing.T) {
	err := waitForDeletion(context.Background(), "domain", "d1", time.Minute, time.Millisecond, func() error {
		return notFoundErr()
	})
	if err != nil {
		t.Fatalf("404 probe should return nil, got %v", err)
	}
}

// A non-404 error must be surfaced, not swallowed as a successful deletion.
// Regression test: the old loop returned on ANY error, so a transient 5xx
// during polling made Terraform drop a resource that still existed.
func TestWaitForDeletion_TransientErrorIsReturned(t *testing.T) {
	err := waitForDeletion(context.Background(), "domain", "d1", time.Minute, time.Millisecond, func() error {
		return serverErr()
	})
	if err == nil {
		t.Fatal("a non-404 error must be returned, not treated as deletion success")
	}
}

// While the resource still exists (probe returns nil), polling continues until
// the timeout elapses and a timeout error is returned.
func TestWaitForDeletion_StillPresentTimesOut(t *testing.T) {
	calls := 0
	err := waitForDeletion(context.Background(), "route", "r1", 5*time.Millisecond, time.Millisecond, func() error {
		calls++
		return nil // resource still present
	})
	if err == nil {
		t.Fatal("expected a timeout error while the resource is still present")
	}
	if calls < 1 {
		t.Fatalf("probe should have been called at least once, got %d", calls)
	}
}

// The resource lingers for a couple of polls, then a 404 arrives: success.
func TestWaitForDeletion_EventuallyGone(t *testing.T) {
	calls := 0
	err := waitForDeletion(context.Background(), "route", "r1", time.Minute, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return nil // still present
		}
		return notFoundErr()
	})
	if err != nil {
		t.Fatalf("expected success once the probe reports 404, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 probes, got %d", calls)
	}
}

// A cancelled context stops polling promptly with the context error.
func TestWaitForDeletion_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := waitForDeletion(ctx, "domain", "d1", time.Minute, time.Hour, func() error {
		return nil // still present, would otherwise poll forever
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
