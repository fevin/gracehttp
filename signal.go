package gracehttp

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var sig chan os.Signal
var notifySignals []os.Signal

func init() {
	sig = make(chan os.Signal)
	notifySignals = append(notifySignals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGQUIT, syscall.SIGUSR1)
	signal.Notify(sig, notifySignals...)
}

// 捕获系统信号
func handleSignals() {
	capturedSig := <-sig
	srvLog.Info(fmt.Sprintf("Received SIG. [PID:%d, SIG:%v]", syscall.Getpid(), capturedSig))
	switch capturedSig {
	case syscall.SIGHUP:
	case syscall.SIGUSR1:
		startNewProcess()
		shutdown()
	case syscall.SIGINT:
		fallthrough
	case syscall.SIGTERM:
		fallthrough
	case syscall.SIGTSTP:
		fallthrough
	case syscall.SIGQUIT:
		shutdown()
	}
}
