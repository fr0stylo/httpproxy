package main

import "net/http"

type RequestLogger interface {
	Log(r *http.Request)
}

type proxyOptions struct {
	port          int64
	dataLimiter   DataSizer
	requestLogger RequestLogger
	authorizer    Authorizer
}

var DefaultOptions = proxyOptions{
	port:          8080,
	dataLimiter:   NewDataLimiter(0),
	authorizer:    NewBasicAuthorizer(),
	requestLogger: NewConsoleRequestLogger(),
}

type ProxyOptionsFn func(options *proxyOptions)

func WithPort(port int64) ProxyOptionsFn {
	return func(options *proxyOptions) {
		options.port = port
	}
}

func WithAuthorizer(authorizer Authorizer) ProxyOptionsFn {
	return func(options *proxyOptions) {
		options.authorizer = authorizer
	}
}

func WithDataLimiter(limiter DataSizer) ProxyOptionsFn {
	return func(options *proxyOptions) {
		options.dataLimiter = limiter
	}
}

func WithRequestLogger(logger RequestLogger) ProxyOptionsFn {
	return func(options *proxyOptions) {
		options.requestLogger = logger
	}
}
