package gracehttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	srvM                *sync.Mutex
	srvWg               sync.WaitGroup
	maxListenConnection int
)

func init() {
	srvM = new(sync.Mutex)
	maxListenConnection = 100000 // concurrent connections for single server, default c100k
}

// 设置单个 server 最大并发链接数
func SetMaxConcurrentForOneServer(max int) {
	maxListenConnection = max
}

func AddServer(srv *http.Server, isTLS bool, certFile, keyFile string) *Server {
	srvM.Lock()
	defer srvM.Unlock()
	pSrv := &Server{
		id:         dispatchSrvId(),
		httpServer: srv,
		isTLS:      isTLS,
		certFile:   certFile,
		keyFile:    keyFile,
	}
	gracefulSrv.srvList = append(gracefulSrv.srvList, pSrv)
	return pSrv
}

func dispatchSrvId() int {
	id := nextSrvId
	nextSrvId++
	return id
}

func Run() {
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		} else {
			srvWg.Wait()
		}
	}()

	for _, srv := range gracefulSrv.srvList {
		srv.Run()
		srvWg.Add(1)
	}
}

func shutdown() {
	for _, srv := range gracefulSrv.srvList {
		if err := srv.httpServer.Shutdown(context.Background()); err != nil {
			srvLog.Error(fmt.Sprintf("srv  closed fail, [pid:%d, srvrd:%d]", os.Getpid(), srv.id))
		}
		srvWg.Done()
	}
}
