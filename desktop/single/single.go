// Package single keeps one instance of an application running per user.
//
// The first instance holds a lock. A later instance fails with
// ErrAlreadyRunning after asking the first one to come forward, which then
// receives a value on Activated, typically to show its window:
//
//	inst, err := single.Acquire("my-app")
//	if errors.Is(err, single.ErrAlreadyRunning) {
//		return // the running instance has been asked to show itself
//	}
//	defer inst.Release()
//	go func() {
//		for range inst.Activated() {
//			showWindow()
//		}
//	}()
//
// Platforms: Windows uses a named mutex and event in the session namespace;
// macOS, Linux and the BSDs use flock on a lock file and a Unix socket for
// the activation request.
package single

import (
	"errors"
	"fmt"
	"sync"
)

// ErrAlreadyRunning is returned by Acquire when another instance holds the
// lock. That instance has been asked to activate, when the platform allows.
var ErrAlreadyRunning = errors.New("single: another instance is already running")

// ErrInvalidID is returned for an ID that is empty, too long or contains
// characters other than ASCII letters, digits, '.', '_' and '-'.
var ErrInvalidID = errors.New("single: invalid application ID")

const maxIDLen = 64

func checkID(id string) error {
	if id == "" || len(id) > maxIDLen {
		return ErrInvalidID
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
		default:
			return ErrInvalidID
		}
	}
	return nil
}

// Instance is the lock held by the running instance.
type Instance struct {
	activated chan struct{}
	p         *platformLock
	once      sync.Once
	err       error
}

// Acquire takes the single-instance lock for id, which should be the
// application's stable identifier. See the package documentation.
func Acquire(id string) (*Instance, error) {
	if err := checkID(id); err != nil {
		return nil, err
	}
	in := &Instance{activated: make(chan struct{}, 1)}
	p, err := acquire(id, in.notify)
	if err != nil {
		if errors.Is(err, ErrAlreadyRunning) {
			return nil, err
		}
		return nil, fmt.Errorf("single: %w", err)
	}
	in.p = p
	return in, nil
}

// notify queues one activation; requests arriving before the previous one
// is received are merged.
func (in *Instance) notify() {
	select {
	case in.activated <- struct{}{}:
	default:
	}
}

// Activated receives a value each time another instance asks this one to
// come forward. The channel is never closed.
func (in *Instance) Activated() <-chan struct{} { return in.activated }

// Release gives up the lock; later instances can then start. It is safe to
// call more than once. The lock is also released when the process exits.
func (in *Instance) Release() error {
	in.once.Do(func() { in.err = in.p.release() })
	return in.err
}
