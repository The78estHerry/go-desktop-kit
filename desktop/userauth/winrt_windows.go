//go:build windows

package userauth

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Minimal WinRT plumbing for Windows.Security.Credentials.UI.UserConsentVerifier,
// called directly so no CGO or generated bindings are needed.

var (
	combase                    = windows.NewLazySystemDLL("combase.dll")
	procRoInitialize           = combase.NewProc("RoInitialize")
	procRoGetActivationFactory = combase.NewProc("RoGetActivationFactory")
	procWindowsCreateString    = combase.NewProc("WindowsCreateString")
	procWindowsDeleteString    = combase.NewProc("WindowsDeleteString")
	procGetForegroundWindow    = windows.NewLazySystemDLL("user32.dll").NewProc("GetForegroundWindow")
)

var (
	iidIAsyncInfo                  = mustGUID("{00000036-0000-0000-C000-000000000046}")
	iidUserConsentVerifierStatics  = mustGUID("{AF4F3F91-564C-4DDC-B8B5-973447627C65}")
	iidIUserConsentVerifierInterop = mustGUID("{39E050C3-4E74-441A-8DC0-B81104DF949C}")

	// IAsyncOperation<UserConsentVerificationResult>, derived with the WinRT
	// parameterized-interface algorithm.
	iidAsyncOpVerificationResult = pinterfaceIID(
		"pinterface({9fc2b0bb-e446-44e2-aa61-9cab8f636af2};enum(Windows.Security.Credentials.UI.UserConsentVerificationResult;i4))")
)

const classUserConsentVerifier = "Windows.Security.Credentials.UI.UserConsentVerifier"

// Vtable slots. Every WinRT interface starts with IInspectable (6 slots).
const (
	slotQueryInterface                    = 0
	slotRelease                           = 2
	slotStaticsCheckAvailabilityAsync     = 6
	slotInteropRequestVerificationForHWND = 6
	slotAsyncInfoStatus                   = 7
	slotAsyncInfoErrorCode                = 8
	slotAsyncInfoCancel                   = 9
	slotAsyncOperationGetResults          = 8
)

func mustGUID(s string) windows.GUID {
	g, err := windows.GUIDFromString(s)
	if err != nil {
		panic(err)
	}
	return g
}

// pinterfaceIID computes the IID of a parameterized WinRT interface
// instance from its type signature (UUID v5 over the WinRT namespace).
func pinterfaceIID(signature string) windows.GUID {
	ns := []byte{0x11, 0xf4, 0x7a, 0xd5, 0x7b, 0x73, 0x42, 0xc0, 0xab, 0xae, 0x87, 0x8b, 0x1e, 0x16, 0xad, 0xee}
	sum := sha1.Sum(append(ns, signature...))
	b := sum[:16]
	b[6] = b[6]&0x0f | 0x50
	b[8] = b[8]&0x3f | 0x80
	var g windows.GUID
	g.Data1 = binary.BigEndian.Uint32(b[0:4])
	g.Data2 = binary.BigEndian.Uint16(b[4:6])
	g.Data3 = binary.BigEndian.Uint16(b[6:8])
	copy(g.Data4[:], b[8:16])
	return g
}

type hresultError uint32

func (h hresultError) Error() string { return fmt.Sprintf("HRESULT 0x%08X", uint32(h)) }

func check(hr uintptr) error {
	if int32(hr) < 0 {
		return hresultError(hr)
	}
	return nil
}

// comCall invokes vtable slot idx of obj with obj as the implicit this.
func comCall(obj unsafe.Pointer, idx int, args ...uintptr) uintptr {
	vtbl := *(*unsafe.Pointer)(obj)
	fn := *(*uintptr)(unsafe.Add(vtbl, idx*int(unsafe.Sizeof(uintptr(0)))))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{uintptr(obj)}, args...)...)
	return r
}

func release(obj unsafe.Pointer) {
	if obj != nil {
		comCall(obj, slotRelease)
	}
}

func newHString(s string) (uintptr, error) {
	u, err := windows.UTF16FromString(s)
	if err != nil {
		return 0, err
	}
	var h uintptr
	hr, _, _ := procWindowsCreateString.Call(uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(&h)))
	return h, check(hr)
}

func deleteHString(h uintptr) { procWindowsDeleteString.Call(h) }

// onWinRTThread runs f on a fresh goroutine locked to an OS thread that has
// joined the multithreaded apartment, and waits for it.
func onWinRTThread(f func() error) error {
	done := make(chan error, 1)
	go func() {
		runtime.LockOSThread()               // never unlocked: the thread ends with the goroutine
		hr, _, _ := procRoInitialize.Call(1) // RO_INIT_MULTITHREADED
		const rpcEChangedMode = 0x80010106
		if err := check(hr); err != nil && uint32(hr) != rpcEChangedMode {
			done <- fmt.Errorf("RoInitialize: %w", err)
			return
		}
		done <- f()
	}()
	return <-done
}

func activationFactory(class string, iid *windows.GUID) (unsafe.Pointer, error) {
	h, err := newHString(class)
	if err != nil {
		return nil, err
	}
	defer deleteHString(h)
	var f unsafe.Pointer
	hr, _, _ := procRoGetActivationFactory.Call(h, uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&f)))
	return f, check(hr)
}

var errAsyncCanceled = errors.New("async operation canceled")

// waitInt32 polls an IAsyncOperation<T> whose T is a 32-bit enum and
// returns its result. When ctx ends it cancels the operation.
func waitInt32(ctx context.Context, op unsafe.Pointer) (int32, error) {
	var info unsafe.Pointer
	if err := check(comCall(op, slotQueryInterface, uintptr(unsafe.Pointer(&iidIAsyncInfo)), uintptr(unsafe.Pointer(&info)))); err != nil {
		return 0, fmt.Errorf("query IAsyncInfo: %w", err)
	}
	defer release(info)
	canceled := false
	for {
		var status int32
		if err := check(comCall(info, slotAsyncInfoStatus, uintptr(unsafe.Pointer(&status)))); err != nil {
			return 0, err
		}
		switch status {
		case 0: // Started
			if ctx.Err() != nil && !canceled {
				comCall(info, slotAsyncInfoCancel)
				canceled = true
			}
			time.Sleep(50 * time.Millisecond)
		case 1: // Completed
			var result int32
			err := check(comCall(op, slotAsyncOperationGetResults, uintptr(unsafe.Pointer(&result))))
			return result, err
		case 2: // Canceled
			return 0, errAsyncCanceled
		default: // Error
			var code uint32
			comCall(info, slotAsyncInfoErrorCode, uintptr(unsafe.Pointer(&code)))
			return 0, hresultError(code)
		}
	}
}
