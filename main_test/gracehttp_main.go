package main

import (
	"flag"
	"fmt"
	"github.com/fevin/gracehttp"
	"net/http"
	"syscall"
	"time"
)

var (
	httpPort string
)

func init() {
	flag.StringVar(&httpPort, "http_port", "9090", "the port of http server")
}

func main() {
	hd := &Controller{}
	grace := gracehttp.NewGraceHTTP()
	srv := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      hd,
		ReadTimeout:  time.Duration(time.Second),
		WriteTimeout: time.Duration(time.Second),
	}
	option := &gracehttp.ServerOption{
		HTTPServer:          srv,
		MaxListenConnection: 1000, // 限制 server 连接数
	}
	grace.AddServer(option)
	if err := grace.Run(); err != nil {
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
