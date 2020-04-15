package gracehttp

// about limit @see: "golang.org/x/net/netutil"

import (
	"net"
	"sync"
	"time"
)

func newListener(tl *net.TCPListener, n int) *Listener {
	return &Listener{
		TCPListener: tl,
		sem:         make(chan struct{}, n),
		done:        make(chan struct{}),
	}
}

type Listener struct {
	*net.TCPListener
	sem       chan struct{}
	closeOnce sync.Once     // ensures the done chan is only closed once
	done      chan struct{} // no values sent; closed when Close is called
}

func (l *Listener) Fd() (uintptr, error) {
	file, err := l.TCPListener.File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}

// override
func (l *Listener) Accept() (net.Conn, error) {
	acquired := l.acquire()
	tc, err := l.AcceptTCP()
	if err != nil {
		if acquired {
			l.release()
		}
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(time.Minute)

	return &ListenerConn{Conn: tc, release: l.release}, nil
}

// override
func (l *Listener) Close() error {
	err := l.TCPListener.Close()
	l.closeOnce.Do(func() { close(l.done) })
	return err
}

// acquire acquires the limiting semaphore. Returns true if successfully
// accquired, false if the listener is closed and the semaphore is not
// acquired.
func (l *Listener) acquire() bool {
	select {
	case <-l.done:
		return false
	case l.sem <- struct{}{}:
		return true
	}
}

func (l *Listener) release() { <-l.sem }

type ListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (l *ListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}
