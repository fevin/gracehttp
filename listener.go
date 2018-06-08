package gracehttp

import (
	"net"
	"time"
)

type Listener struct {
	*net.TCPListener
}

func newListener(tl *net.TCPListener) net.Listener {
	return &Listener{
		TCPListener: tl,
	}
}

func (l *Listener) Fd() (uintptr, error) {
	file, err := l.TCPListener.File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}

func (l *Listener) Accept() (net.Conn, error) {

	tc, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(time.Minute)

	conn := &Connection{
		Conn:     tc,
		listener: l,
	}
	return conn, nil
}
