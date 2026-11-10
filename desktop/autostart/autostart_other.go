//go:build !windows && !darwin && !linux && !freebsd && !netbsd && !openbsd && !dragonfly

package autostart

import "errors"

func enable(Entry) error           { return errors.ErrUnsupported }
func disable(string) error         { return errors.ErrUnsupported }
func enabled(string) (bool, error) { return false, errors.ErrUnsupported }
