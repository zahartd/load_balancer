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
				http.Error(w, "Header X-API-Key is required", http.StatusBadRequest)
			}

			allowed, err := rl.AllowRequest(ctx, clientID)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				log.Printf("Request from %s did not alloweded\n", clientID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, err := w.Write([]byte(`{"code":429,"message":"Rate limit exceeded"}`))
				if err != nil {
					log.Printf("Failed to return rate limit exceed answer: %s\n", err.Error())
				}
				return
			}
			log.Printf("Request from %s alloweded\n", clientID)
			next.ServeHTTP(w, r)
		})
	}
}
