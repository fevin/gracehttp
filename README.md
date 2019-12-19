# gracehttp
优雅的使用 `HTTP Server`

## 环境
* Go 1.9+

## 支持功能
1. 平滑重启（`Zero-Downtime`）；
2. 平滑关闭；
3. 多 `Server` 添加（包含 `HTTP` 、`HTTPS`）；
4. 自定义日志组件；
5. 支持单个端口 server 链接数限流，默认值为：C100K。超过该限制之后，请求阻塞，并且不会消耗文件句柄，避免发生雪崩，压坏服务。

## 使用此 pkg 会给你带来什么
* 平滑升级服务，不影响用户体验；
* 方便的多 `server` 添加；
* 对 `server` 进行过载保护，通过设置合适的阀值，避免 `too many open files` 错误；

## 使用指南
`go get github.com/fevin/gracehttp`

**单元测试**
```
go test -v github.com/fevin/gracehttp
=== RUN   TestHTTPServer
--- PASS: TestHTTPServer (0.00s)
    gracehttp_test.go:57: ======== test http server 1 ========
    gracehttp_test.go:67: http server 1 success, response: pong
    gracehttp_test.go:72: ======== test http server 2 ========
    gracehttp_test.go:82: http server 2 success, response: pong
PASS
ok  	github.com/fevin/gracehttp	0.016s
```

### 添加服务
```go
    import (
        "flag"
        "github.com/fevin/gracehttp"
        "http"
    )

    // ...

    // http
    srv1 := &http.Server{
        Addr:    ":80",
        Handler: sc,
    }
    gracehttp.AddServer(srv1, false, "", "")

    // https
    srv2 := &http.Server{
        Addr:    ":443",
        Handler: sc,
    }
    gracehttp.AddServer(srv2, true, "../config/https.crt", "../config/https.key")

    gracehttp.SetMaxConcurrentForOneServer(1) // 限制同时只能处理一个链接

    gracehttp.Run() // 此方法会阻塞，直到所有的 HTTP 服务退出
```

如上所示，只需创建好 `Server` 对象，调用 `gracehttp.AddServer` 添加即可。

#### 退出或者重启服务
* 重启：`kill -HUP pid`
* 退出：`kill -QUIT pid`

### 添加自定义日志组件
```go
    gracehttp.SetErrorLogCallback(logger.LogConfigLoadError)
```

此处提供了三个 `Set*` 方法，分别对应不同的日志等级：
* SetInfoLogCallback
* SetNoticeLogCallback
* SetErrorLogCallback

## 关于链接数限制
参考实现：[golang/net/netutil](https://github.com/golang/net/blob/master/netutil/listen.go)
实现原理：
1. 通过 `channel-buffer` 来控制并发量：每个请求都会获取一个缓冲单元，直到缓冲区满；
2. 只有获取 `buffer` 的请求才能进行 `accept`；
3. 如果缓冲区满了，后来的请求会阻塞，直到 `conn.close` 或者 缓冲区有空闲

---
@see `gracehttp_test.go`

## DEMO
`vim main.go`
```go
package main

import (
    "flag"
    "fmt"
    "github.com/fevin/gracehttp"
    "net/http"
)

func main() {
    flag.Parse()
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
    gracehttp.SetMaxConcurrentForOneServer(1)
    gracehttp.Run()

    fmt.Print("main over")
}

type Controller struct {
}

func (this *Controller) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
    resp.Write([]byte("hello01"))
}
```
