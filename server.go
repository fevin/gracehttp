package gracehttp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func newGracefulServer() *gracefulServer {
	server := new(gracefulServer)
	server.srvList = make([]*Server, 0, 2)
	server.srvErrChan = make(chan error)
	return server
}

type gracefulServer struct {
	srvList    []*Server
	srvWg      sync.WaitGroup
	srvErrChan chan error
	signalChan chan os.Signal
}

func (this *gracefulServer) AddServer(srv *Server) {
	this.srvList = append(this.srvList, srv)
}

func (this *gracefulServer) RunAllServer() error {
	for _, srv := range this.srvList {
		this.srvWg.Add(1)
		server := srv
		go func() {
			if err := server.Run(); err != nil {
				select {
				case this.srvErrChan <- err:
				default:
				}
			}
		}()
	}

	this.srvWg.Wait()
	select {
	case err := <-this.srvErrChan:
		return err
	default:
		return nil
	}
}

func (this *gracefulServer) shutdownAllServer() {
	for index, srv := range this.srvList {
		t := time.NewTimer(time.Duration(10 * time.Second))
		done := make(chan struct{})
		server := srv
		go func() {
			defer func() {
				done <- struct{}{}
			}()
			if err := server.Shutdown(); err != nil {
				srvLog.Error(fmt.Sprintf("srv closed fail, [pid:%d, srvrd:%d]", os.Getpid(), server.id))
			}
		}()

		select {
		case <-t.C:
			srvLog.Error(fmt.Sprintf("server %d shutdown fail, exit directly!", index))
		case <-done:
			srvLog.Info(fmt.Sprintf("server %d shutdown successfully!", index))
		}
		this.srvWg.Done()
	}
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
		return fmt.Errorf("get listener fd error:%v", fdErr)
	}
	this.fd = lnFd
	this.listener = wrapListener
	return this.Serve()
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
		return fmt.Errorf("get listener fd error:%v", fdErr)
	}
	this.fd = lnFd
	this.listener = tls.NewListener(wrapListener, config)
	return this.Serve()
}

func (this *Server) Serve() error {
	return this.HTTPServer.Serve(this.listener)
}

func (this *Server) Shutdown() error {
	return this.HTTPServer.Shutdown(context.Background())
}

func (this *Server) getNetTCPListener() (*net.TCPListener, error) {

	var ln net.Listener
	var err error

	addr := this.getAddr()
	if isRestartEnv(addr) {
		resetRestartEnv(addr)
		file := os.NewFile(uintptr(this.id+2), "") // 此处加 2，因为 0/1/2 分别对应标准输入/输出/错误
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
