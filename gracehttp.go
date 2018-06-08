package gracehttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	srvM  *sync.Mutex
	srvWg sync.WaitGroup
)

func init() {
	srvM = new(sync.Mutex)
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
		srvWg.Wait()
	}()
	for _, srv := range gracefulSrv.srvList {
		srvWg.Add(1)
		srv.Run()
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
