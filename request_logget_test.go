package main

import (
	"context"
	"net"
	"net/http"
	"testing"
)

type MockRequestLogger struct {
	loggedRequests []*http.Request
}

func (m *MockRequestLogger) Log(req *http.Request) {
	m.loggedRequests = append(m.loggedRequests, req)
}

func MockProxyHandler(ctx context.Context, con net.Conn, req *http.Request) int64 {
	return 12345
}

func TestRequestLoggerMiddleware(t *testing.T) {
	logger := &MockRequestLogger{}

	middleware := RequestLoggerMiddleware(logger)
	wrappedHandler := middleware(MockProxyHandler)

	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)

	var mockConn net.Conn

	responseSize := wrappedHandler(context.Background(), mockConn, req)

	if responseSize != 12345 {
		t.Errorf("expected response size 12345, got %v", responseSize)
	}

	if len(logger.loggedRequests) != 1 {
		t.Fatalf("expected 1 logged request, got %v", len(logger.loggedRequests))
	}

	if logger.loggedRequests[0].URL.Path != "/foo" {
		t.Errorf("expected logged request to be GET /foo, got %v %v", logger.loggedRequests[0].Method, logger.loggedRequests[0].URL.Path)
	}
}
