package main

import (
	"context"
	"encoding/base64"
	"log"
	"net"
	"net/http"
)

var (
	basicAuth = "user:pass"
)

type BasicAuthorizer struct {
}

func NewBasicAuthorizer() *BasicAuthorizer {
	return &BasicAuthorizer{}
}

func (b BasicAuthorizer) Authorize(user string) bool {
	serverAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(basicAuth))
	return user == serverAuth
}

type Authorizer interface {
	Authorize(user string) bool
}

type User struct {
	User string
}

func GetUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(User{}).(User)
	if ok {
		return &user
	}
	return nil
}

func AuthorizerMiddleware(authorizer Authorizer) ProxyMiddleware {
	return func(handler ProxyHandler) ProxyHandler {
		return func(ctx context.Context, con net.Conn, req *http.Request) int64 {
			proxyAuth := req.Header.Get("Proxy-Authorization")

			if !authorizer.Authorize(proxyAuth) {
				defer con.Close()
				log.Printf("[WARN] Authorization failed for request from %s", con.RemoteAddr())
				return httpResponse(con, http.StatusUnauthorized, []byte("Unauthorized"))
			}
			ctx = context.WithValue(ctx, User{}, User{User: proxyAuth})
			return handler(ctx, con, req)
		}
	}
}
