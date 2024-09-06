package main

import (
	"context"
	"net"
	"net/http"
)

type RequestLogger interface {
	Log(r *http.Request)
}

type ProxyHandler func(ctx context.Context, conn net.Conn, r *http.Request) int64
type ProxyMiddleware func(handler ProxyHandler) ProxyHandler

type proxyOptions struct {
	port        int64
	middlewares []ProxyMiddleware
}

var DefaultOptions = proxyOptions{
	port: 8080,
}

type ProxyOptionsFn func(options *proxyOptions)

func WithPort(port int64) ProxyOptionsFn {
	return func(options *proxyOptions) {
		options.port = port
	}
}

func WithMiddlewares(middleware ...ProxyMiddleware) ProxyOptionsFn {
	return func(options *proxyOptions) {
		options.middlewares = middleware
	}
}
