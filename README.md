# gracehttp
优雅的使用 `HTTP Server`

## 支持功能
1. 平滑重启（`Zero-Downtime`）；
2. 平滑关闭；
3. 多 `Server` 添加（包含 `HTTP` 、`HTTPS`）；
4. 自定义日志组件；
5. 支持单个端口 server 链接数限流，默认值为：C100K。超过该限制之后，拒绝服务，避免发生雪崩，压坏服务。

## 使用指南
`go get github.com/fevin/gracehttp`

### 添加服务
```go
    import (
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

---
@see `gracehttp_test.go`
