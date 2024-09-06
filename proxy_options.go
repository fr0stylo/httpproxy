package main

type proxyOptions struct {
	port        int64
	dataLimiter DataSizer
	authorizer  Authorizer
}

var (
	DefaultAuthorizer Authorizer = NewBasicAuthorizer()
	DefaultSizer      DataSizer  = NewDataLimiter(0)
)

var DefaultOptions = proxyOptions{
	port:        8080,
	dataLimiter: DefaultSizer,
	authorizer:  DefaultAuthorizer,
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
