package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
)

type DataLimiter struct {
	usage map[string]int64
	lock  sync.Mutex
	limit int64
}

type UsageReport struct {
	user string
	sent int64
}

func NewUsageReport(user string, sent int64) *UsageReport {
	return &UsageReport{user: user, sent: sent}
}

func (r *DataLimiter) ConsumeUsage() chan *UsageReport {
	reportChan := make(chan *UsageReport)
	go func() {
		for ur := range reportChan {
			r.AddUsage(ur.user, ur.sent)
		}
	}()

	return reportChan
}

func (r *DataLimiter) AddUsage(user string, sent int64) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.usage[user] = r.usage[user] + sent
}

func (r *DataLimiter) IsLimitReached(user string) bool {
	if usage, ok := r.usage[user]; ok {
		return usage >= r.limit
	}

	return false
}

func (r *DataLimiter) GetUsage(user string) int64 {
	return r.usage[user]
}

func NewDataLimiter(limit int64) *DataLimiter {
	return &DataLimiter{usage: make(map[string]int64), limit: limit, lock: sync.Mutex{}}
}

func DataSizerMiddleware(sizer DataSizer) ProxyMiddleware {
	reportChan := sizer.ConsumeUsage()
	return func(handler ProxyHandler) ProxyHandler {
		return func(ctx context.Context, con net.Conn, req *http.Request) int64 {
			user := GetUserFromContext(ctx)

			if sizer.IsLimitReached(user.User) {
				defer con.Close()
				log.Printf("[WARN] Usage limit reached %s", con.RemoteAddr())
				httpResponse(con, http.StatusTooManyRequests, []byte("Unauthorized"))

				return 0
			}
			sent := handler(ctx, con, req)

			reportChan <- NewUsageReport(user.User, sent)

			return sent
		}
	}
}
