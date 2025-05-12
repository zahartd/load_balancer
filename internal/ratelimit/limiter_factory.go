package ratelimit

import (
	"context"
	"log"

	"github.com/zahartd/load_balancer/internal/config"
	ratelimit_algorithms "github.com/zahartd/load_balancer/internal/ratelimit/algorithms"
)

func CreateAlgorithm(ctx context.Context, algorithmType string, options any) Algorithm {
	var algorithm Algorithm
	switch algorithmType {
	case "token_bucket":
		tokenBucketOptions, ok := options.(config.TokenBucketLimiterOptions)
		if !ok {
			log.Fatalf(
				"Invalid algorithm options: expected TokenBucketLimiterOptions, but got %T\n",
				options,
			)
		}
		algorithm = ratelimit_algorithms.NewTokenBucketLimiter(ctx, tokenBucketOptions)
	default:
		log.Fatalf("Uknown algorithm type type: %s\n", algorithmType)
	}
	return algorithm
}
