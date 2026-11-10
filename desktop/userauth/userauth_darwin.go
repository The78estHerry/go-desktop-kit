//go:build darwin && cgo

// Status: implemented, not yet tested on real hardware.

package userauth

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework LocalAuthentication -framework Foundation
#include <stdlib.h>
#import <LocalAuthentication/LocalAuthentication.h>

// Returns 0 when the device owner can be authenticated, otherwise the
// LAError code.
static int la_available(void) {
	LAContext *ctx = [[LAContext alloc] init];
	NSError *err = nil;
	if ([ctx canEvaluatePolicy:LAPolicyDeviceOwnerAuthentication error:&err]) {
		return 0;
	}
	return err ? (int)err.code : -1;
}

// Shows Touch ID or the login password prompt and blocks until the user
// answers. Returns 0 on success, otherwise the LAError code.
static int la_verify(const char *msg) {
	LAContext *ctx = [[LAContext alloc] init];
	NSString *reason = [NSString stringWithUTF8String:msg];
	dispatch_semaphore_t done = dispatch_semaphore_create(0);
	__block int result = -1;
	[ctx evaluatePolicy:LAPolicyDeviceOwnerAuthentication
	    localizedReason:reason
	              reply:^(BOOL ok, NSError *err) {
		result = ok ? 0 : (err ? (int)err.code : -1);
		dispatch_semaphore_signal(done);
	}];
	dispatch_semaphore_wait(done, DISPATCH_TIME_FOREVER);
	return result;
}
*/
import "C"

import (
	"context"
	"fmt"
	"unsafe"
)

// LAError codes (LocalAuthentication/LAError.h).
const (
	laAuthenticationFailed = -1
	laUserCancel           = -2
	laUserFallback         = -3
	laSystemCancel         = -4
	laPasscodeNotSet       = -5
	laBiometryNotAvailable = -6
	laBiometryNotEnrolled  = -7
	laBiometryLockout      = -8
	laAppCancel            = -9
	laNotInteractive       = -1004
)

func laError(code int) error {
	switch code {
	case 0:
		return nil
	case laUserCancel, laUserFallback, laSystemCancel, laAppCancel:
		return ErrCanceled
	case laAuthenticationFailed, laBiometryLockout:
		return ErrRetriesExhausted
	case laPasscodeNotSet, laBiometryNotEnrolled:
		return &UnavailableError{NotConfigured}
	case laBiometryNotAvailable:
		return &UnavailableError{DeviceNotPresent}
	case laNotInteractive:
		return &UnavailableError{Unsupported}
	}
	return fmt.Errorf("userauth: LocalAuthentication error %d", code)
}

func available(context.Context) error { return laError(int(C.la_available())) }

// The prompt is a system sheet; macOS places it, so r.Window is not used.
// Cancelling ctx after the prompt is shown does not close it.
func verify(_ context.Context, r Request) error {
	msg := C.CString(r.Message)
	defer C.free(unsafe.Pointer(msg))
	return laError(int(C.la_verify(msg)))
}
