package gracehttp

import (
	"errors"
	"net/http"
)

type ServerOption struct {
	HTTPServer          *http.Server
	IsTLS               bool
	CertFile            string
	KeyFile             string
	MaxListenConnection int // 限制 server 连接数
}

func (this *ServerOption) init() error {
	if this == nil {
		return errors.New("option cannot be nil")
	}

	if this.MaxListenConnection <= 0 {
		this.MaxListenConnection = 10000
	}
	return nil
}
