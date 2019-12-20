package main

import (
	"flag"
	"fmt"
	"github.com/fevin/gracehttp"
	"net/http"
	"syscall"
)

var (
	httpPort string
)

func init() {
	flag.StringVar(&httpPort, "http_port", "9090", "the port of http server")
}

func main() {
	hd := &Controller{}
	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: hd,
	}
	gracehttp.AddServer(srv, false, "", "")
	if err := gracehttp.Run(); err != nil {
		fmt.Println(err)
	}
}

type Controller struct {
}

func (this *Controller) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/ping" {
		resp.Write([]byte(fmt.Sprintf("pong by pid:%d", syscall.Getpid())))
	} else {
		resp.Write([]byte("unknown"))
	}
}
