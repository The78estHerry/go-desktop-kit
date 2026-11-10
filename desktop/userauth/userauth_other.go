//go:build !windows && !(darwin && cgo)

package userauth

import "context"

// Linux has no common way for an application to ask for the user's own
// presence (polkit needs an installed action and usually asks for an
// administrator); applications use an App PIN instead. macOS needs CGO.

func available(context.Context) error { return &UnavailableError{Unsupported} }

func verify(context.Context, Request) error { return &UnavailableError{Unsupported} }
