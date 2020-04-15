package gracehttp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

type gracefulServer struct {
	srvList       []*Server
	srvSucessList []*Server
	srvWg         sync.WaitGroup
	signalChan    chan os.Signal
}

func (this *gracefulServer) AddServer(srv *Server) {
	this.srvList = append(this.srvList, srv)
}

func (this *gracefulServer) WaitAllServer() {
	this.srvWg.Wait()
}

func (this *gracefulServer) AsyncRunAllServer() error {
	for _, srv := range this.srvList {
		if err := srv.Run(); err != nil {
			this.shutdownAllServer()
			return err
		}
		this.addSucessfulServer(srv)
	}

	return nil
}

func (this *gracefulServer) addSucessfulServer(srv *Server) {
	this.srvSucessList = append(this.srvSucessList, srv)
	this.srvWg.Add(1)
}

func (this *gracefulServer) shutdownAllServer() {
	for _, srv := range this.srvSucessList {
		if err := srv.Shutdown(context.Background()); err != nil {
			srvLog.Error(fmt.Sprintf("srv  closed fail, [pid:%d, srvrd:%d]", os.Getpid(), srv.id))
		}
		this.srvWg.Done()
	}
	this.srvSucessList = this.srvSucessList[:0]
}

// 支持优雅重启的http服务
type Server struct {
	*ServerOption
	id       int
	fd       uintptr
	listener net.Listener
}

func (this *Server) getAddr() string {
	return this.HTTPServer.Addr
}

func (this *Server) Run() error {
	if this.IsTLS {
		return this.ListenAndServeTLS()
	} else {
		return this.ListenAndServe()
	}
}

func (this *Server) ListenAndServe() error {
	if this.getAddr() == "" {
		return errors.New("http port must be set!")
	}

	ln, err := this.getNetTCPListener()
	if err != nil {
		return err
	}

	wrapListener := newListener(ln, this.MaxListenConnection)
	lnFd, fdErr := wrapListener.Fd()
	if fdErr != nil {
		log.Panicf("get listener fd error:%v", fdErr)
	}
	this.fd = lnFd
	this.listener = wrapListener
	go this.Serve()
	return nil
}

func (this *Server) ListenAndServeTLS() error {
	if this.getAddr() == "" {
		return errors.New("https port must be set!")
	}

	config := &tls.Config{}
	if this.HTTPServer.TLSConfig != nil {
		*config = *this.HTTPServer.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(this.CertFile, this.KeyFile)
	if err != nil {
		return err
	}

	ln, err := this.getNetTCPListener()
	if err != nil {
		return err
	}

	wrapListener := newListener(ln, this.MaxListenConnection)
	lnFd, fdErr := wrapListener.Fd()
	if fdErr != nil {
		log.Panicf("get listener fd error:%v", fdErr)
	}
	this.fd = lnFd
	this.listener = tls.NewListener(wrapListener, config)
	go this.Serve()
	return nil
}

func (this *Server) Serve() {
	if err := this.HTTPServer.Serve(this.listener); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("srv err:%v, [pid:%d, srvid:%d, srvAddr:%v]", err, os.Getpid(), this.id, this.getAddr()))
	}
}

func (this *Server) Shutdown(ctx context.Context) error {
	return this.HTTPServer.Shutdown(ctx)
}

func (this *Server) getNetTCPListener() (*net.TCPListener, error) {

	var ln net.Listener
	var err error

	if isRestartEnv(this.getAddr()) {
		file := os.NewFile(uintptr(this.id+2), "") // 此处加 2，因为 0/1/2 分别对应标准输入/输出/错误
		ln, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("net.FileListener error: %v", err)
			return nil, err
		}
	} else {
		ln, err = net.Listen("tcp", this.getAddr())
		if err != nil {
			err = fmt.Errorf("net.Listen error: %v", err)
			return nil, err
		}
	}
	return ln.(*net.TCPListener), nil
}
