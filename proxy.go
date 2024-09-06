package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type DataSizer interface {
	ConsumeUsage() chan *UsageReport
	IsLimitReached(user string) bool
	GetUsage(user string) int64
}

type Proxy struct {
	proxyOptions
}

func NewProxy(optsFn ...ProxyOptionsFn) *Proxy {
	opts := DefaultOptions
	for _, fn := range optsFn {
		fn(&opts)
	}

	return &Proxy{opts}
}

func (r *Proxy) proxyRequest(con net.Conn) {
	con.SetDeadline(time.Now().Add(2 * time.Hour))
	br := bufio.NewReader(con)
	req, err := http.ReadRequest(br)
	if err != nil {
		defer con.Close()
		log.Print("[ERROR] Failed to read http request, ", err)
		httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))

		return
	}

	method := chainMiddlewares(handlePlainHttp, r.middlewares...)
	ctx := context.Background()
	method(ctx, con, req)
}

func chainMiddlewares(h ProxyHandler, rest ...ProxyMiddleware) ProxyHandler {
	if len(rest) == 0 {
		return h
	}

	return rest[0](chainMiddlewares(h, rest[1:cap(rest)]...))
}

func handlePlainHttp(ctx context.Context, con net.Conn, req *http.Request) (total int64) {
	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL.String(), req.Body)
	if err != nil {
		log.Print("[ERROR] Failed to create proxy request, ", err)
		httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))

		return 0
	}

	proxyReq.Header = req.Header
	proxyClient := &http.Client{}
	proxyRes, err := proxyClient.Do(proxyReq)
	if err != nil {
		log.Print("[ERROR] Failed to send request to proxy, ", err)
		httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))

		return 0
	}
	defer proxyRes.Body.Close()
	total = total + req.ContentLength

	res := http.Response{
		StatusCode: proxyRes.StatusCode,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	res.Header = proxyRes.Header.Clone()
	res.Write(con)
	n, _ := io.Copy(con, proxyRes.Body)
	return total + n
}

func (r *Proxy) Serve(addr *net.TCPAddr) error {
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	defer l.Close()

	for {
		con, err := l.Accept()
		if err != nil {
			log.Print("[ERROR] Failed open tcp listener, ", err)
			continue
		}
		go r.proxyRequest(con)
	}
}

func httpResponse(con net.Conn, status int, body []byte) int64 {
	res := http.Response{
		StatusCode:    status,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(body)),
	}
	res.Write(con)
	n, _ := con.Write(body)
	return int64(n)
}

func transfer(destination io.WriteCloser, source io.ReadCloser, usageReport chan int64) {
	defer destination.Close()
	defer source.Close()
	n, _ := io.Copy(destination, source)
	usageReport <- n
}

func HandleSecureHttpMiddleware(next ProxyHandler) ProxyHandler {
	return func(ctx context.Context, con net.Conn, req *http.Request) int64 {
		if req.Method != http.MethodConnect {
			return next(ctx, con, req)
		}
		httpResponse(con, http.StatusOK, []byte{})

		proxy, err := net.DialTimeout("tcp", req.URL.Host, 2*time.Hour)
		if err != nil {
			defer con.Close()
			log.Print("[ERROR] Failed to create proxy connection, ", err)
			return httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))
		}
		resultChan := make(chan int64)

		go transfer(proxy, con, resultChan)
		go transfer(con, proxy, resultChan)

		n := <-resultChan
		return n + <-resultChan
	}
}
