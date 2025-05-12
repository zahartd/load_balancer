package ratelimit

import (
	"context"
	"log"
	"sync"
)

type RateLimiter struct {
	limiters    map[string]Algorithm
	limiterType string
	options     any
	mu          sync.RWMutex
}

func New(limiterType string, limiterOptions any) *RateLimiter {
	return &RateLimiter{
		limiterType: limiterType,
		options:     limiterOptions,
		limiters:    make(map[string]Algorithm),
	}
}

func (rl *RateLimiter) getLimiter(ctx context.Context, key string) Algorithm {
	// Check if it already created limiter for this client
	rl.mu.RLock()
	if l, exists := rl.limiters[key]; exists {
		rl.mu.RUnlock()
		log.Printf("Use rate limiter (type=%s) for %s", rl.limiterType, key)
		return l
	}
	rl.mu.RUnlock()

	// Create new limmiter
	l := CreateAlgorithm(ctx, rl.limiterType, rl.options)
	rl.mu.Lock()
	rl.limiters[key] = l
	rl.mu.Unlock()

	log.Printf("Created rate limiter (type=%s) for %s", rl.limiterType, key)
	return l
}

func (rl *RateLimiter) AllowRequest(ctx context.Context, key string) (bool, error) {
	limiter := rl.getLimiter(ctx, key)
	return limiter.Allow(), nil
}
