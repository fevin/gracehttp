# gracehttp
优雅的使用 `HTTP Server`

   * [gracehttp](#gracehttp)
      * [环境](#环境)
      * [支持功能](#支持功能)
      * [使用此 pkg 会给你带来什么](#使用此-pkg-会给你带来什么)
      * [使用指南](#使用指南)
         * [添加服务](#添加服务)
            * [退出或者重启服务](#退出或者重启服务)
         * [添加自定义日志组件](#添加自定义日志组件)
      * [关于链接数限制](#关于链接数限制)
      * [测试](#测试)
         * [HTTP Server 常规测试](#http-server-常规测试)
         * [平滑测试](#平滑测试)

## 环境
* Go 1.14+

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
```
go get github.com/fevin/gracehttp
```

### 添加服务
@see `main_test/gracehttp_main.go`

#### 退出或者重启服务
* 重启：`kill -HUP pid` 或 `kill -USR1 pid`
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


## 测试
### HTTP Server 常规测试
```
✗ go test -v github.com/fevin/gracehttp
=== RUN   TestHTTPServer
--- PASS: TestHTTPServer (0.00s)
    gracehttp_test.go:106: ******* test multi server  *******
    gracehttp_test.go:77: [test http server 1]
    gracehttp_test.go:87: http server 1 success, response: pong by pid:5385
    gracehttp_test.go:92: [test http server 2]
    gracehttp_test.go:102: http server 2 success, response: pong by pid:5385
PASS
ok  	github.com/fevin/gracehttp	0.016s
```

### 平滑重启功能测试
```
// 以下操作需要开启 go mod
✗ git clone https://github.com/fevin/gracehttp.git
✗ go build main_test/gracehttp_main.go
✗ nohup ./gracehttp_main 2>&1 > gracehttp.log &

✗ curl http://localhost:9090/ping
pong by pid:86703
✗ kill -USR1 86703
[1]  + 86703 done       nohup ./bin/gracehttp_main 2>&1 > gracehttp.log
✗ cat gracehttp.log
2019/12/20 12:07:38 [gracehtto-log][Info] [Received SIG. [PID:86703, SIG:user defined signal 1]]
2019/12/20 12:07:38 [gracehtto-log][Info] [start new process success, pid 86818.]

# 再次执行 curl，发现 pid 已经变化
✗ curl http://localhost:9090/ping
pong by pid:86818

✗ kill -QUIT 86818
✗ cat gracehttp.log
2019/12/20 12:07:38 [gracehtto-log][Info] [Received SIG. [PID:86703, SIG:user defined signal 1]]
2019/12/20 12:07:38 [gracehtto-log][Info] [start new process success, pid 86818.]
2019/12/20 12:08:56 [gracehtto-log][Info] [Received SIG. [PID:86818, SIG:quit]]
```

---
@see test `./gracehttp_test.go`    
@see demo `./main_test/gracehttp_main.go`
