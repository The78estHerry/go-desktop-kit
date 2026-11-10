package single

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCheckID(t *testing.T) {
	good := []string{"my-app", "com.example.app", "a_b-1"}
	bad := []string{"", "a/b", `a\b`, "a b", "名前", string(make([]byte, maxIDLen+1))}
	for _, id := range good {
		if err := checkID(id); err != nil {
			t.Errorf("checkID(%q) = %v, want nil", id, err)
		}
	}
	for _, id := range bad {
		if err := checkID(id); !errors.Is(err, ErrInvalidID) {
			t.Errorf("checkID(%q) = %v, want ErrInvalidID", id, err)
		}
	}
}

func TestNotifyMerges(t *testing.T) {
	in := &Instance{activated: make(chan struct{}, 1)}
	in.notify()
	in.notify() // must not block
	<-in.Activated()
	select {
	case <-in.Activated():
		t.Fatal("two requests were not merged")
	default:
	}
}

// TestAcquire runs against the real OS objects with a unique ID.
func TestAcquire(t *testing.T) {
	id := fmt.Sprintf("go-desktop-kit-test-%d", os.Getpid())
	first, err := Acquire(id)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	defer first.Release()

	if _, err := Acquire(id); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("second Acquire = %v, want ErrAlreadyRunning", err)
	}
	select {
	case <-first.Activated():
	case <-time.After(5 * time.Second):
		t.Fatal("first instance was not activated")
	}

	if err := first.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if err := first.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
	again, err := Acquire(id)
	if err != nil {
		t.Fatalf("Acquire after Release: %v", err)
	}
	again.Release()
}
