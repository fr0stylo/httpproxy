package main

import "encoding/base64"

type BasicAuthorizer struct {
}

func NewBasicAuthorizer() *BasicAuthorizer {
	return &BasicAuthorizer{}
}

func (b BasicAuthorizer) Authorize(user string) bool {
	serverAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(basicAuth))
	return user == serverAuth
}
