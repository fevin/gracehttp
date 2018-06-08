# gracehttp
## 支持功能
1. 平滑重启（Zero-Downtime restart server）；
2. 平滑关闭；
3. 多 `Server` 添加（包含 `HTTP` 、`HTTPS`）；
4. 自定义日志组件；

## 使用指南
### 添加服务器
```go
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

	gracehttp.Run() // 此方法会阻塞，直到所有的 HTTP 服务退出
```

如上所示，只需创建好 `Server` 对象，调用 `gracehttp.AddServer` 添加即可。

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
