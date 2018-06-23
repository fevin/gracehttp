package gracehttp

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
)

type GracefulServer struct {
	srvList    []*Server
	signalChan chan os.Signal
}

// 支持优雅重启的http服务
type Server struct {
	id         int
	httpServer *http.Server
	listener   net.Listener
	isTLS      bool
	certFile   string
	keyFile    string
}

var (
	gracefulSrv *GracefulServer
	isChild     bool
	nextSrvId   int
)

func init() {
	flag.BoolVar(&isChild, "continue", false, "listen on open fd (after forking)")
	gracefulSrv = new(GracefulServer)
	nextSrvId = 1

	// 处理信号
	go handleSignals()
}

func (srv *Server) Run() {
	if srv.isTLS {
		srv.ListenAndServeTLS()
	} else {
		srv.ListenAndServe()
	}
}

func (srv *Server) ListenAndServe() {
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":http"
	}

	ln, err := srv.getNetTCPListener(addr, srv.id)
	if err != nil {
		panic(fmt.Sprintf("Get listener fail. err:", err))
	}
	srv.listener = newListener(ln, maxListenConnection)
	go srv.Serve()
}

func (srv *Server) ListenAndServeTLS() {
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":https"
	}

	config := &tls.Config{}
	if srv.httpServer.TLSConfig != nil {
		*config = *srv.httpServer.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(srv.certFile, srv.keyFile)
	if err != nil {
		panic(fmt.Sprintf("tls load fail. err:", err))
	}

	ln, err := srv.getNetTCPListener(addr, srv.id)
	if err != nil {
		panic(fmt.Sprintf("Get listener fail. err:", err))
	}

	srv.listener = tls.NewListener(newListener(ln, maxListenConnection), config)
	go srv.Serve()
}

func (srv *Server) Serve() {
	if err := srv.httpServer.Serve(srv.listener); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("srv err:%v, [pid:%d, srvid:%d, srvAddr:%v]", err, os.Getpid(), srv.id, srv.httpServer.Addr))
	}
}

func (srv *Server) getNetTCPListener(addr string, connOrder int) (*net.TCPListener, error) {

	var ln net.Listener
	var err error

	if isChild {
		file := os.NewFile(uintptr(connOrder+2), "") // 此处加 2，因为 0/1/2 分别对应标准输入/输出/错误
		ln, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("net.FileListener error: %v", err)
			return nil, err
		}
	} else {
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			err = fmt.Errorf("net.Listen error: %v", err)
			return nil, err
		}
	}
	return ln.(*net.TCPListener), nil
}

// 启动子进程执行新程序
func startNewProcess() {
	// 获取 args
	var args []string
	for _, arg := range os.Args {
		if arg == "-continue" {
			break
		}
		args = append(args, arg)
	}
	args = append(args, "-continue")

	// 获取 fds
	fds := []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()}
	for _, srv := range gracefulSrv.srvList {
		if srv.listener == nil {
			panic(fmt.Sprintf("srv.listener cannot be nil!id:%v, isTLS:%v", srv.id, srv.isTLS))
		}

		srvFd, err := srv.listener.(*Listener).Fd()
		if err != nil {
			panic(fmt.Sprintf("when start new pro, get listener fd fail. err:", err))
		}

		fds = append(fds, srvFd)
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
}
