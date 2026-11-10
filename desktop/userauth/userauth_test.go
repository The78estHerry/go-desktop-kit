package userauth

import (
	"context"
	"errors"
	"testing"
)

func TestUnavailableError(t *testing.T) {
	var err error = &UnavailableError{NotConfigured}
	if !errors.Is(err, ErrUnavailable) {
		t.Error("UnavailableError is not ErrUnavailable")
	}
	if errors.Is(err, ErrCanceled) {
		t.Error("UnavailableError matched ErrCanceled")
	}
	var u *UnavailableError
	if !errors.As(err, &u) || u.Reason != NotConfigured {
		t.Errorf("errors.As gave %v", u)
	}
	if got := err.Error(); got != "userauth: user verification is not available: not configured" {
		t.Errorf("Error() = %q", got)
	}
}

func TestVerifyCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Verify(ctx, Request{Message: "test"}); !errors.Is(err, ErrCanceled) {
		t.Errorf("Verify with a canceled context = %v, want ErrCanceled", err)
	}
}

// TestAvailable calls the real API; it shows no prompt.
func TestAvailable(t *testing.T) {
	err := Available(context.Background())
	var u *UnavailableError
	if err != nil && !errors.As(err, &u) {
		t.Fatalf("Available returned an unexpected error: %v", err)
	}
	t.Logf("Available: %v", err)
}
