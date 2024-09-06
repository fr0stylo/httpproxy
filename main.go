package main

import (
	"flag"
	"log"
	"net"
)

var (
	port      = flag.Int("port", 8080, "HTTP proxy port")
	dataLimit = flag.Int64("limit", 1024*1024*1024, "Limit of data per user")
)

func main() {
	flag.Parse()

	proxy := NewProxy(
		WithMiddlewares(
			RequestLoggerMiddleware(NewConsoleRequestLogger()),
			AuthorizerMiddleware(NewBasicAuthorizer()),
			DataSizerMiddleware(NewDataLimiter(*dataLimit)),
			HandleSecureHttpMiddleware,
		))

	log.Fatal("Something went really wrong: ", proxy.Serve(&net.TCPAddr{Port: *port}))
}
