package ratelimit_algorithms

import (
	"flag"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zahartd/load_balancer/internal/config"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}
	os.Exit(m.Run())
}

func TestTokenBucketLimiter_Exhaustion(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	opts := config.TokenBucketLimiterOptions{
		DefaultCapacity:         3,
		DefaultRefillIntervalMS: config.DurationMs(300),
	}
	tbl := NewTokenBucketLimiter(ctx, opts)

	for i := range opts.DefaultCapacity {
		require.True(t, tbl.Allow(), "token %d should be allowed", i+1)
	}

	require.False(t, tbl.Allow(), "4th token should be denied when capacity exhausted")
}

func TestTokenBucketLimiter_Refill(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	opts := config.TokenBucketLimiterOptions{
		DefaultCapacity:         1,
		DefaultRefillIntervalMS: config.DurationMs(100),
	}
	tbl := NewTokenBucketLimiter(ctx, opts)

	require.True(t, tbl.Allow(), "initial token should be allowed")
	require.False(t, tbl.Allow(), "no tokens left immediately after consumption")

	time.Sleep(150 * time.Millisecond)

	require.True(t, tbl.Allow(), "token should be refilled after interval")
	require.False(t, tbl.Allow(), "only one token refilled, next should be denied")
}

func TestTokenBucketLimiter_NoOverflow(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	opts := config.TokenBucketLimiterOptions{
		DefaultCapacity:         2,
		DefaultRefillIntervalMS: config.DurationMs(50),
	}
	tbl := NewTokenBucketLimiter(ctx, opts)

	require.True(t, tbl.Allow())
	require.True(t, tbl.Allow())
	require.False(t, tbl.Allow())

	time.Sleep(250 * time.Millisecond)

	require.True(t, tbl.Allow(), "first refill token")
	require.True(t, tbl.Allow(), "second refill token")
	require.False(t, tbl.Allow(), "no overflow beyond capacity")
}

func TestTokenBucketLimiter_Concurrent(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	opts := config.TokenBucketLimiterOptions{
		DefaultCapacity:         5,
		DefaultRefillIntervalMS: config.DurationMs(10),
	}
	tbl := NewTokenBucketLimiter(ctx, opts)

	results := make(chan bool, 10)
	for range 10 {
		go func() {
			results <- tbl.Allow()
		}()
	}

	allowed := 0
	for range 10 {
		if <-results {
			allowed++
		}
	}

	require.LessOrEqual(t, allowed, 5, "at most capacity requests should pass concurrently")
}
