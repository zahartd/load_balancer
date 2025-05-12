package ratelimit

import (
	"context"
	"log"

	ratelimit_algorithms "github.com/zahartd/load_balancer/internal/ratelimit/algorithms"
)

func CreateAlgorithm(ctx context.Context, algorithmType string, options any) Algorithm {
	var algorithm Algorithm
	switch algorithmType {
	case "token_bucket":
		tokenBucketOptions, ok := options.(ratelimit_algorithms.TokenBucketLimiterOptions)
		if !ok {
			log.Fatal("Invalid algorithm options, expected TokenBucketLimiterOptions\n")
		}
		algorithm = ratelimit_algorithms.NewTokenBucketLimiter(ctx, tokenBucketOptions)
	default:
		log.Fatalf("Uknown algorithm type type: %s\n", algorithmType)
	}
	return algorithm
}
