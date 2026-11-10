//go:build windows

package userauth

import (
	"context"
	"errors"
	"fmt"
	"unsafe"
)

// Windows Hello decides which methods (face, fingerprint, PIN) to offer
// from the user's sign-in options; an application cannot restrict them.

// UserConsentVerifierAvailability values.
const (
	availAvailable = iota
	availDeviceNotPresent
	availNotConfigured
	availDisabledByPolicy
	availDeviceBusy
)

// UserConsentVerificationResult values.
const (
	resVerified = iota
	resDeviceNotPresent
	resNotConfigured
	resDisabledByPolicy
	resDeviceBusy
	resRetriesExhausted
	resCanceled
)

func available(ctx context.Context) error {
	var v int32
	err := onWinRTThread(func() error {
		statics, err := activationFactory(classUserConsentVerifier, &iidUserConsentVerifierStatics)
		if err != nil {
			return fmt.Errorf("userauth: activation factory: %w", err)
		}
		defer release(statics)
		var op unsafe.Pointer
		if err := check(comCall(statics, slotStaticsCheckAvailabilityAsync, uintptr(unsafe.Pointer(&op)))); err != nil {
			return fmt.Errorf("userauth: CheckAvailabilityAsync: %w", err)
		}
		defer release(op)
		v, err = waitInt32(ctx, op)
		return err
	})
	if errors.Is(err, errAsyncCanceled) {
		return ErrCanceled
	}
	if err != nil {
		return err
	}
	switch v {
	case availAvailable:
		return nil
	case availDeviceNotPresent:
		return &UnavailableError{DeviceNotPresent}
	case availNotConfigured:
		return &UnavailableError{NotConfigured}
	case availDisabledByPolicy:
		return &UnavailableError{DisabledByPolicy}
	case availDeviceBusy:
		return &UnavailableError{DeviceBusy}
	}
	return fmt.Errorf("userauth: unknown availability %d", v)
}

func verify(ctx context.Context, r Request) error {
	hwnd := r.Window.native
	if hwnd == 0 {
		hwnd, _, _ = procGetForegroundWindow.Call()
	}
	var v int32
	err := onWinRTThread(func() error {
		interop, err := activationFactory(classUserConsentVerifier, &iidIUserConsentVerifierInterop)
		if err != nil {
			return fmt.Errorf("userauth: interop factory: %w", err)
		}
		defer release(interop)
		msg, err := newHString(r.Message)
		if err != nil {
			return err
		}
		defer deleteHString(msg)
		var op unsafe.Pointer
		hr := comCall(interop, slotInteropRequestVerificationForHWND, hwnd, msg,
			uintptr(unsafe.Pointer(&iidAsyncOpVerificationResult)), uintptr(unsafe.Pointer(&op)))
		if err := check(hr); err != nil {
			return fmt.Errorf("userauth: RequestVerificationForWindowAsync: %w", err)
		}
		defer release(op)
		v, err = waitInt32(ctx, op)
		return err
	})
	if errors.Is(err, errAsyncCanceled) {
		return ErrCanceled
	}
	if err != nil {
		return err
	}
	switch v {
	case resVerified:
		return nil
	case resCanceled:
		return ErrCanceled
	case resRetriesExhausted:
		return ErrRetriesExhausted
	case resDeviceNotPresent:
		return &UnavailableError{DeviceNotPresent}
	case resNotConfigured:
		return &UnavailableError{NotConfigured}
	case resDisabledByPolicy:
		return &UnavailableError{DisabledByPolicy}
	case resDeviceBusy:
		return &UnavailableError{DeviceBusy}
	}
	return fmt.Errorf("userauth: unknown verification result %d", v)
}
