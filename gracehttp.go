package gracehttp

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

var (
	nextSrvId     int
	notifySignals []os.Signal
)

func init() {
	nextSrvId = 1
	notifySignals = append(notifySignals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		notifySignals = append(notifySignals, syscall.SIGTSTP, syscall.SIGUSR1)
	}
}

func dispatchSrvId() int {
	id := nextSrvId
	nextSrvId++
	return id
}

func NewGraceHTTP() *GraceHTTP {
	grace := new(GraceHTTP)
	grace.server = newGracefulServer()
	grace.sig = make(chan os.Signal)
	signal.Notify(grace.sig, notifySignals...)

	go grace.exitHandler()
	return grace
}

type GraceHTTP struct {
	serverWg sync.WaitGroup
	server   *gracefulServer
	sig      chan os.Signal
}

func (this *GraceHTTP) AddServer(option *ServerOption) (*Server, error) {
	if err := option.init(); err != nil {
		return nil, err
	}
	srv := &Server{
		id:           dispatchSrvId(),
		ServerOption: option,
	}
	this.server.AddServer(srv)
	return srv, nil
}

func (this *GraceHTTP) Run() (retErr error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	return this.server.RunAllServer()
}

func (this *GraceHTTP) exitHandler() {
	capturedSig := <-this.sig
	srvLog.Info(fmt.Sprintf("Received SIG. [PID:%d, SIG:%v]", syscall.Getpid(), capturedSig))
	switch capturedSig {
	case syscall.SIGHUP:
		if err := this.startNewProcess(); err != nil {
			srvLog.Error(fmt.Sprintf("Received SIG. [PID:%d, SIG:%v]", syscall.Getpid(), capturedSig))
		}
		this.shutdown()
	case syscall.SIGINT:
		fallthrough
	case syscall.SIGTERM:
		fallthrough
	case syscall.SIGTSTP:
		fallthrough
	case syscall.SIGQUIT:
		this.shutdown()
	}
}

// 启动子进程执行新程序
func (this *GraceHTTP) startNewProcess() error {
	// 获取 args
	var args []string
	for _, arg := range os.Args {
		args = append(args, arg)
	}

	// 获取 fds
	fds := []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()}
	for _, srv := range this.server.srvList {
		setRestartEnv(srv.getAddr())
		fds = append(fds, srv.fd)
	}

	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: fds,
	}

	forkId, err := syscall.ForkExec(os.Args[0], args, execSpec)
	if err != nil {
		srvLog.Error(fmt.Sprintf("failed to forkexec: %v", err))
	}
	srvLog.Info(fmt.Sprintf("start new process success, pid %d.", forkId))
	return nil
}

func (this *GraceHTTP) shutdown() {
	this.server.shutdownAllServer()
}
