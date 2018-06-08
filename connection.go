package gracehttp

import (
	"net"
)

type Connection struct {
	net.Conn
	listener *Listener

	closed bool
}
