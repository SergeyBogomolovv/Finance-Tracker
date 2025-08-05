package limiter

import (
	"FinanceTracker/gateway/pkg/utils"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
	_ "golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func New(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}
	return limiter
}

func (rl *RateLimiter) Wrap(next http.Handler) http.Handler {
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
