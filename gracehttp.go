package gracehttp

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
)

var (
	nextSrvId int
)

func init() {
	nextSrvId = 1
}

func dispatchSrvId() int {
	id := nextSrvId
	nextSrvId++
	return id
}

func NewGraceHTTP() *GraceHTTP {
	grace := new(GraceHTTP)
	grace.server = new(gracefulServer)
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
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		} else {
			this.server.WaitAllServer()
		}
	}()

	if !flag.Parsed() {
		flag.Parse()
	}

	return this.server.AsyncRunAllServer()
}

func (this *GraceHTTP) exitHandler() {
	exitForSignal(this)
}

// 启动子进程执行新程序
func (this *GraceHTTP) startNewProcess() error {
	// 获取 args
	var args []string
	for _, arg := range os.Args {
		args = append(args, arg)
	}

	// 获取 fds
	fds := make([]*os.File, 0, len(this.server.srvList))
	for _, srv := range this.server.srvList {
		srvFd, err := srv.listener.(*Listener).TCPListener.File()
		if err != nil {
			srvLog.Error(fmt.Sprintf("failed to forkexec: %v", err))
			return err
		}
		setRestartEnv(srv.getAddr())

		fds = append(fds, srvFd)
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ())
	cmd.ExtraFiles = fds
	if err := cmd.Run(); err != nil {
		srvLog.Error(fmt.Sprintf("failed to forkexec: %v", err))
		return err
	} else {
		srvLog.Info("start new process success")
		return nil
	}
}

func (this *GraceHTTP) shutdown() {
	this.server.shutdownAllServer()
}

func (this *GraceHTTP) restart() error {
	if err := this.startNewProcess(); err != nil {
		srvLog.Error(fmt.Sprintf("restart error:%v", err))
		return err
	}
	this.shutdown()
	return nil
}
