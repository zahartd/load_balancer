package ratelimit_algorithms

import (
	"context"
	"log"
	"time"

	"github.com/zahartd/load_balancer/internal/config"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
}

func NewTokenBucketLimiter(ctx context.Context, options config.TokenBucketLimiterOptions) *TokenBucketLimiter {
	tbl := &TokenBucketLimiter{
		tokens: make(chan struct{}, options.DefaultCapacity),
	}

	for range options.DefaultCapacity {
		tbl.tokens <- struct{}{}
	}

	refilInterval := options.DefaultRefillIntervalMS.AsDuration()
	go tbl.refillRoutine(ctx, refilInterval)

	return tbl
}

func (tbl *TokenBucketLimiter) refillRoutine(ctx context.Context, interval time.Duration) {
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
		log.Printf("In total, tokens are left: %d", len(tbl.tokens))
		return true
	default:
		return false
	}
}
