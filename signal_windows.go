// +build windows

package gracehttp

import (
	"fmt"
	"syscall"
)

func init() {
	notifySignals = append(notifySignals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func exitForSignal(grace *GraceHTTP) {
	capturedSig := <-grace.sig
	srvLog.Info(fmt.Sprintf("Received SIG. [PID:%d, SIG:%v]", syscall.Getpid(), capturedSig))
	switch capturedSig {
	case syscall.SIGHUP:
		grace.restart()
	case syscall.SIGINT:
		fallthrough
	case syscall.SIGTERM:
		fallthrough
	case syscall.SIGQUIT:
		grace.shutdown()
	}
}
