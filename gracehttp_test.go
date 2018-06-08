package gracehttp

import (
	"code.aliyun.com/nextdata/gracehttp"
	"fmt"
	"net/http"
	"time"
)

func main() {
	sc := &Controller{}
	srv1 := &http.Server{
		Addr:    ":9090",
		Handler: sc,
	}
	gracehttp.AddServer(srv1, false, "", "")
	srv2 := &http.Server{
		Addr:    ":9091",
		Handler: sc,
	}
	gracehttp.AddServer(srv2, false, "", "")
	gracehttp.Run()

	fmt.Print("main over")
}

type Controller struct {
}

func (this *Controller) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("hello01"))
	time.Sleep(20 * time.Second)
	resp.Write([]byte("hello02"))
}
