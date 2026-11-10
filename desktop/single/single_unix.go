//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

// Status: implemented, not yet tested on real hardware.

package single

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

type platformLock struct {
	lock *os.File
	ln   net.Listener
	sock string
	done chan struct{}
}

// runtimeDir is $XDG_RUNTIME_DIR (per user, Linux) or the temp directory
// (per user on macOS). The file names also carry the user ID, so users who
// share /tmp do not collide.
func runtimeDir() string {
	if d := os.Getenv("XDG_RUNTIME_DIR"); d != "" {
		return d
	}
	return os.TempDir()
}

func acquire(id string, activate func()) (*platformLock, error) {
	base := filepath.Join(runtimeDir(), id+"-"+strconv.Itoa(os.Getuid()))
	sock := base + ".sock"
	f, err := os.OpenFile(base+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			if c, err := net.Dial("unix", sock); err == nil {
				c.Write([]byte{1})
				c.Close()
			}
			return nil, ErrAlreadyRunning
		}
		return nil, err
	}
	// Holding the lock, any socket file left by a crashed instance is stale.
	os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		f.Close()
		return nil, err
	}
	os.Chmod(sock, 0o600)
	p := &platformLock{lock: f, ln: ln, sock: sock, done: make(chan struct{})}
	go p.serve(activate)
	return p, nil
}

func (p *platformLock) serve(activate func()) {
	defer close(p.done)
	for {
		c, err := p.ln.Accept()
		if err != nil {
			return // listener closed by release
		}
		c.Close()
		activate()
	}
}

func (p *platformLock) release() error {
	p.ln.Close()
	<-p.done
	os.Remove(p.sock)
	return p.lock.Close() // closing the file drops the flock
}
