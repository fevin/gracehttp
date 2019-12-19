package gracehttp

import (
	"flag"
	"io/ioutil"
	"net/http"
	"testing"
)

var (
	runChan   chan struct{}
	httpPort1 string
	httpPort2 string
)

func init() {
	flag.StringVar(&httpPort1, "http_port_1", "9090", "the port of http server 1")
	flag.StringVar(&httpPort2, "http_port_2", "9091", "the port of http server 2")

	runChan = make(chan struct{}, 1)
}

type Controller struct {
}

func (this *Controller) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/ping" {
		resp.Write([]byte("pong"))
	} else {
		resp.Write([]byte("unknown"))
	}
}

func runServer(t *testing.T) {
	sc := &Controller{}
	srv1 := &http.Server{
		Addr:    ":" + httpPort1,
		Handler: sc,
	}
	AddServer(srv1, false, "", "")
	srv2 := &http.Server{
		Addr:    ":" + httpPort2,
		Handler: sc,
	}
	AddServer(srv2, false, "", "")
	runChan <- struct{}{}
	if err := Run(); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPServer(t *testing.T) {
	go runServer(t)
	<-runChan

	{
		t.Log("======== test http server 1 ========")
		resp, err := http.Get("http://localhost:" + httpPort1 + "/ping")
		if err != nil {
			t.Fatal("http server 1 error:", err)
		} else {
			defer resp.Body.Close()
			data, respErr := ioutil.ReadAll(resp.Body)
			if respErr != nil {
				t.Fatal("http server 1 error:", respErr)
			}
			t.Log("http server 1 success, response:", string(data))
		}
	}

	{
		t.Log("======== test http server 2 ========")
		resp, err := http.Get("http://localhost:" + httpPort2 + "/ping")
		if err != nil {
			t.Fatal("http server 2 error:", err)
		} else {
			defer resp.Body.Close()
			data, respErr := ioutil.ReadAll(resp.Body)
			if respErr != nil {
				t.Fatal("http server 2 error:", respErr)
			}
			t.Log("http server 2 success, response:", string(data))
		}
	}
}
