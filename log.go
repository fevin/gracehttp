package gracehttp

import (
	"log"
)

type Log struct {
	Info   func(args ...interface{})
	Notice func(args ...interface{})
	Error  func(args ...interface{})
}

var srvLog *Log

func init() {
	srvLog = new(Log)
	srvLog.Info = func(args ...interface{}) {
		log.Printf("[gracehtto-log][Info] %v \n", args)
	}
	srvLog.Notice = func(args ...interface{}) {
		log.Printf("[gracehtto-log][Notice] %v \n", args)
	}
	srvLog.Error = func(args ...interface{}) {
		log.Printf("[gracehtto-log][Error] %v \n", args)
	}
}

func SetInfoLogCallback(infoFunc func(args ...interface{})) {
	srvLog.Info = infoFunc
}

func SetNoticeLogCallback(noticeFunc func(args ...interface{})) {
	srvLog.Notice = noticeFunc
}

func SetErrorLogCallback(errFunc func(args ...interface{})) {
	srvLog.Error = errFunc
}
