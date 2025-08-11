package middleware

import (
	"FinanceTracker/gateway/pkg/utils"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type limiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func newLimiter(r rate.Limit, burst int) *limiter {
	return &limiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (rl *limiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}
	return limiter
}

func NewBodyLimiter(limit rate.Limit, burst int, key string) func(next http.Handler) http.Handler {
	rl := newLimiter(limit, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				utils.WriteError(w, "cannot parse IP", http.StatusInternalServerError)
				return
			}

			rawBody, err := io.ReadAll(r.Body)
			if err != nil {
				utils.WriteError(w, "cannot read body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			body := make(map[string]string)
			if err := json.Unmarshal(rawBody, &body); err != nil {
				utils.WriteError(w, "cannot parse body", http.StatusBadRequest)
				return
			}

			val, exists := body[key]
			if !exists || val == "" {
				utils.WriteError(w, "invalid body", http.StatusBadRequest)
				return
			}

			limiter := rl.getLimiter(ip + "|" + val)
			if !limiter.Allow() {
				utils.WriteError(w, "Too many requests. Please wait.", http.StatusTooManyRequests)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(rawBody))

			next.ServeHTTP(w, r)
		})
	}
}

func NewIPLimiter(limit rate.Limit, burst int) func(next http.Handler) http.Handler {
	rl := newLimiter(limit, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				utils.WriteError(w, "cannot parse IP", http.StatusInternalServerError)
				return
			}

			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				utils.WriteError(w, "Too many requests. Please wait.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
