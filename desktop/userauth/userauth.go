// Package userauth asks the operating system to confirm that the signed-in
// user is present: Windows Hello (face, fingerprint or PIN) on Windows,
// Touch ID or the login password on macOS.
//
// It is a presence check, not encryption: it stops someone (or an AI
// agent) driving the GUI, not a program already running as the user.
// Where it is unavailable (Linux, Hello not set up, disabled by policy),
// fall back to an application PIN such as desktop/apppin.
//
//	err := userauth.Verify(ctx, userauth.Request{
//		Message: "Show the token for gitea-home",
//		Window:  owner, // from the GUI toolkit, e.g. go-gio-kit's giowin
//	})
//	switch {
//	case err == nil: // verified
//	case errors.Is(err, userauth.ErrUnavailable): // use the App PIN instead
//	case errors.Is(err, userauth.ErrCanceled):
//	}
//
// Verify blocks until the user answers. Call it from a goroutine, never
// from a GUI event loop: the prompt needs that loop to keep running.
package userauth

import (
	"context"
	"errors"
)

// Errors returned by Verify and Available.
var (
	// ErrUnavailable: the OS cannot verify the user here. The error is an
	// *UnavailableError carrying the reason.
	ErrUnavailable = errors.New("userauth: user verification is not available")
	// ErrCanceled: the user dismissed the prompt, or ctx ended.
	ErrCanceled = errors.New("userauth: verification canceled")
	// ErrRetriesExhausted: too many failed attempts; the OS refuses for now.
	ErrRetriesExhausted = errors.New("userauth: too many failed attempts")
)

// Reason says why verification is unavailable.
type Reason int

const (
	Unsupported      Reason = iota // the platform has no implementation
	NotConfigured                  // e.g. Windows Hello is not set up
	DisabledByPolicy               // turned off by an administrator
	DeviceNotPresent               // no biometric device or PIN provider
	DeviceBusy                     // the device is in use; try again later
)

var reasonNames = [...]string{"unsupported", "not configured", "disabled by policy", "device not present", "device busy"}

func (r Reason) String() string {
	if r >= 0 && int(r) < len(reasonNames) {
		return reasonNames[r]
	}
	return "unknown"
}

// UnavailableError is returned when verification cannot be done.
// errors.Is(err, ErrUnavailable) is true for it.
type UnavailableError struct{ Reason Reason }

func (e *UnavailableError) Error() string        { return ErrUnavailable.Error() + ": " + e.Reason.String() }
func (e *UnavailableError) Is(target error) bool { return target == ErrUnavailable }

// Window is the window the prompt belongs to; the OS shows the prompt in
// front of it. The zero Window lets the OS choose (the foreground window
// on Windows). GUI toolkits create it, for example go-gio-kit's giowin,
// so application code never handles native handles.
type Window struct{ native uintptr }

// NativeWindow wraps a native window handle: an HWND on Windows, an
// NSWindow pointer on macOS. It is meant for GUI toolkit adapters.
func NativeWindow(handle uintptr) Window { return Window{native: handle} }

// Request describes one verification.
type Request struct {
	// Message tells the user what they are approving, e.g. "Show the
	// token for gitea-home". Keep it short; the OS may truncate it.
	Message string
	Window  Window
}

// Available reports whether Verify can work now: nil, or an
// *UnavailableError.
func Available(ctx context.Context) error { return available(ctx) }

// Verify shows the OS prompt and waits for the user. It returns nil when
// the user is verified, ErrCanceled, ErrRetriesExhausted, or an
// *UnavailableError.
func Verify(ctx context.Context, r Request) error {
	if err := ctx.Err(); err != nil {
		return ErrCanceled
	}
	return verify(ctx, r)
}
