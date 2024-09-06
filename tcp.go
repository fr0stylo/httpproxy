package main

import (
	"flag"
	"log"
	"net"
)

var (
	port      = flag.Int64("port", 8080, "HTTP proxy port")
	dataLimit = flag.Int64("limit", 1024*1024*1024, "Limit of data per user")
)

func main() {
	flag.Parse()

	dl := NewDataLimiter(*dataLimit)
	ba := NewBasicAuthorizer()
	proxy := NewProxy(WithPort(*port), WithAuthorizer(ba), WithDataLimiter(dl))

	log.Fatal("Something went really wrong: ", proxy.Serve(&net.TCPAddr{Port: 8080}))
}
