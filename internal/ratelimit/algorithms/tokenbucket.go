package ratelimit_algorithms

import (
	"context"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
}

type TokenBucketLimiterOptions struct {
	Capacity int
	Period   time.Duration
}

func NewTokenBucketLimiter(ctx context.Context, options TokenBucketLimiterOptions) *TokenBucketLimiter {
	tbl := &TokenBucketLimiter{
		tokens: make(chan struct{}, options.Capacity),
	}

	for range options.Capacity {
		tbl.tokens <- struct{}{}
	}

	refilInterval := options.Period.Nanoseconds() / int64(options.Capacity)

	go tbl.refillLoop(ctx, time.Duration(refilInterval))
	return tbl
}

func (tbl *TokenBucketLimiter) refillLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case tbl.tokens <- struct{}{}:
			default:
			}
		}
	}
}

func (tbl *TokenBucketLimiter) Allow() bool {
	select {
	case <-tbl.tokens:
		return true
	default:
		return false
	}
}
