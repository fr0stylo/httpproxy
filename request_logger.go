package main

import (
	"log"
	"net/http"
)

type ConsoleRequestLogger struct {
}

func (c ConsoleRequestLogger) Log(r *http.Request) {
	log.Print(r.Host)
}

func NewConsoleRequestLogger() *ConsoleRequestLogger {
	return &ConsoleRequestLogger{}
}
