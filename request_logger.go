package main

import (
	"context"
	"log"
	"net"
	"net/http"
)

type ConsoleRequestLogger struct {
}

func (c ConsoleRequestLogger) Log(r *http.Request) {
	host, _, _ := net.SplitHostPort(r.URL.Host)
	log.Print(host)
}

func NewConsoleRequestLogger() *ConsoleRequestLogger {
	return &ConsoleRequestLogger{}
}

func RequestLoggerMiddleware(logger RequestLogger) ProxyMiddleware {
	return func(handler ProxyHandler) ProxyHandler {
		return func(ctx context.Context, con net.Conn, req *http.Request) int64 {
			logger.Log(req)
			return handler(ctx, con, req)
		}
	}
}
