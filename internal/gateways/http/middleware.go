package http

import (
	"context"
	"net/http"

	"github.com/zahartd/load_balancer/internal/ratelimit"
)

func RateLimitMiddleware(ctx context.Context, rl *ratelimit.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := r.Header.Get("X-API-Key")
			if clientID == "" {
				clientID = r.RemoteAddr
			}

			allowed, err := rl.AllowRequest(ctx, clientID)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"code":429,"message":"Rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
