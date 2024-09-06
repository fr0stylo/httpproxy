package main

import (
	"bufio"
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

type Authorizer interface {
	Authorize(user string) bool
}

type Proxy struct {
	proxyOptions
	usageReportChan chan *UsageReport
}

func NewProxy(optsFn ...ProxyOptionsFn) *Proxy {
	opts := DefaultOptions
	for _, fn := range optsFn {
		fn(&opts)
	}
	return &Proxy{opts, opts.dataLimiter.ConsumeUsage()}
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

	proxyAuth := req.Header.Get("Proxy-Authorization")

	if !r.authorizer.Authorize(proxyAuth) {
		defer con.Close()
		log.Printf("[WARN] Authorization failed for request from %s", con.RemoteAddr())
		httpResponse(con, http.StatusUnauthorized, []byte("Unauthorized"))

		return
	}

	if r.dataLimiter.IsLimitReached(proxyAuth) {
		defer con.Close()
		log.Printf("[WARN] Usage limit reached %s", con.RemoteAddr())
		httpResponse(con, http.StatusTooManyRequests, []byte("Unauthorized"))

		return
	}

	r.requestLogger.Log(req)

	if req.Method == http.MethodConnect {
		r.handleSecureHttp(con, req, func(sent int64) {
			r.usageReportChan <- NewUsageReport(proxyAuth, sent)
		})
	} else {
		r.handlePlainHttp(con, req, func(sent int64) {
			r.usageReportChan <- NewUsageReport(proxyAuth, sent)
		})
	}
}

func (r *Proxy) handleSecureHttp(con net.Conn, req *http.Request, reporter TransferReporter) {
	httpResponse(con, http.StatusOK, []byte{})

	proxy, err := net.DialTimeout("tcp", req.URL.Host, 2*time.Hour)
	if err != nil {
		defer con.Close()
		log.Print("[ERROR] Failed to create proxy connection, ", err)
		httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))

		return
	}

	go transfer(proxy, con, reporter)
	go transfer(con, proxy, reporter)
}

func (r *Proxy) handlePlainHttp(con net.Conn, req *http.Request, reporter TransferReporter) {
	proxyReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		log.Print("[ERROR] Failed to create proxy request, ", err)
		httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))

		return
	}

	proxyReq.Header = req.Header
	proxyClient := &http.Client{}
	proxyRes, err := proxyClient.Do(proxyReq)
	if err != nil {
		log.Print("[ERROR] Failed to send request to proxy, ", err)
		httpResponse(con, http.StatusInternalServerError, []byte("Something went wrong"))

		return
	}
	defer proxyRes.Body.Close()

	res := http.Response{
		StatusCode: proxyRes.StatusCode,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	res.Header = proxyRes.Header.Clone()
	res.Write(con)
	n, _ := io.Copy(con, proxyRes.Body)
	reporter(n)
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

type TransferReporter = func(sent int64)

func transfer(destination io.WriteCloser, source io.ReadCloser, usageReport TransferReporter) {
	defer destination.Close()
	defer source.Close()
	n, _ := io.Copy(destination, source)
	usageReport(n)
}

var (
	basicAuth = "user:pass"
)

func httpResponse(con net.Conn, status int, body []byte) {
	res := http.Response{
		StatusCode:    status,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(body)),
	}
	res.Write(con)
	con.Write(body)
}
