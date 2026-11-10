//go:build windows

package single

import (
	"errors"

	"golang.org/x/sys/windows"
)

type platformLock struct {
	mutex, event, stop windows.Handle
	done               chan struct{}
}

// acquire uses objects in the Local\ namespace, so the lock is per logon
// session: another user on the same machine runs their own instance.
func acquire(id string, activate func()) (*platformLock, error) {
	mname, err := windows.UTF16PtrFromString(`Local\` + id + `-single`)
	if err != nil {
		return nil, err
	}
	ename, err := windows.UTF16PtrFromString(`Local\` + id + `-activate`)
	if err != nil {
		return nil, err
	}
	mutex, err := windows.CreateMutex(nil, false, mname)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		windows.CloseHandle(mutex)
		if ev, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, ename); err == nil {
			windows.SetEvent(ev)
			windows.CloseHandle(ev)
		}
		return nil, ErrAlreadyRunning
	}
	if err != nil {
		return nil, err
	}
	event, err := windows.CreateEvent(nil, 0, 0, ename) // auto-reset
	if err != nil {
		windows.CloseHandle(mutex)
		return nil, err
	}
	stop, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		windows.CloseHandle(event)
		windows.CloseHandle(mutex)
		return nil, err
	}
	p := &platformLock{mutex: mutex, event: event, stop: stop, done: make(chan struct{})}
	go p.wait(activate)
	return p, nil
}

func (p *platformLock) wait(activate func()) {
	defer close(p.done)
	for {
		i, err := windows.WaitForMultipleObjects([]windows.Handle{p.event, p.stop}, false, windows.INFINITE)
		if err != nil || i != windows.WAIT_OBJECT_0 {
			return
		}
		activate()
	}
}

func (p *platformLock) release() error {
	windows.SetEvent(p.stop)
	<-p.done
	windows.CloseHandle(p.stop)
	windows.CloseHandle(p.event)
	return windows.CloseHandle(p.mutex)
}
