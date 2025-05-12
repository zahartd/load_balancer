package http

import (
	"context"
	"log"
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
				_, err := w.Write([]byte(`{"code":429,"message":"Rate limit exceeded"}`))
				if err != nil {
					log.Printf("Failed to return rate limit exceed answer: %s\n", err.Error())
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
