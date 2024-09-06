package main

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

type MockAuthorizer struct {
	users map[string]bool
}

func (a MockAuthorizer) Authorize(user string) bool {
	_, ok := a.users[user]
	return ok
}

func newMockAuthorize() *MockAuthorizer {
	return &MockAuthorizer{
		users: map[string]bool{
			"valid user": true,
		},
	}
}

func TestAuthorizerMiddleware(t *testing.T) {
	mockAuthorizer := newMockAuthorize()

	proxyMiddleware := AuthorizerMiddleware(mockAuthorizer)

	testCases := []struct {
		name           string
		proxyAuth      string
		expectedStatus int
	}{
		{
			name:           "Valid Authorization",
			proxyAuth:      "valid user",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Authorization",
			proxyAuth:      "invalid user",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			request := http.Request{Header: http.Header{"Proxy-Authorization": {testcase.proxyAuth}}}

			mockConn := newMockConn()
			mockHandler := func(ctx context.Context, con net.Conn, req *http.Request) int64 {
				res := http.Response{
					StatusCode: testcase.expectedStatus,
					ProtoMinor: 1,
					ProtoMajor: 1,
				}
				res.Write(con)
				return 0
			}

			proxyHandler := proxyMiddleware(mockHandler)
			proxyHandler(context.Background(), mockConn, &request)

			res, _ := http.ReadResponse(bufio.NewReader(mockConn.record), &request)
			if res.StatusCode != testcase.expectedStatus {
				t.Errorf("Unexpected status. Expected %d but got %d", testcase.expectedStatus, res.StatusCode)
			}
		})
	}
}

type MockConn struct {
	net.Conn
	record *bytes.Buffer
}

func (MockConn) Close() error                        { return nil }
func (MockConn) Read([]byte) (n int, err error)      { return 0, nil }
func (r MockConn) Write(b []byte) (n int, err error) { return r.record.Write(b) }
func (MockConn) LocalAddr() net.Addr                 { return nil }
func (MockConn) RemoteAddr() net.Addr                { return nil }
func (MockConn) SetDeadline(t time.Time) error       { return nil }
func (MockConn) SetReadDeadline(t time.Time) error   { return nil }
func (MockConn) SetWriteDeadline(t time.Time) error  { return nil }

func newMockConn() *MockConn {
	return &MockConn{
		record: bytes.NewBuffer([]byte{}),
	}
}
