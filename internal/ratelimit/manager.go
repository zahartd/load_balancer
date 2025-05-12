package ratelimit

import (
	"context"
	"sync"
)

type RateLimiter struct {
	limiters    map[string]Algorithm
	limiterType string
	options     any
	mu          sync.Mutex
}

func New(limiterType string, limiterOptions any) *RateLimiter {
	return &RateLimiter{
		limiterType: limiterType,
		options:     limiterOptions,
		limiters:    make(map[string]Algorithm),
	}
}

func (rl *RateLimiter) getLimiter(ctx context.Context, key string) Algorithm {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if l, exists := rl.limiters[key]; exists {
		return l
	}
	l := CreateAlgorithm(ctx, rl.limiterType, rl.options)
	rl.limiters[key] = l
	return l
}

func (rl *RateLimiter) AllowRequest(ctx context.Context, key string) (bool, error) {
	limiter := rl.getLimiter(ctx, key)
	return limiter.Allow(), nil
}
