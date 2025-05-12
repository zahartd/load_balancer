package integration_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zahartd/load_balancer/internal/balancer"
	"github.com/zahartd/load_balancer/internal/config"
	httpGateway "github.com/zahartd/load_balancer/internal/gateways/http"
	"github.com/zahartd/load_balancer/internal/ratelimit"
	"github.com/zahartd/load_balancer/tests/utils"

	"net/url"
)

func TestBalancerFailsOverWhenBackendDown(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))
	defer healthy.Close()

	ss := utils.NewSwitchServer()
	defer ss.Close()

	bURLs := []string{healthy.URL, ss.URL()}
	var backends []config.BackendConfig
	for _, u := range bURLs {
		parsed, err := url.Parse(u)
		require.NoError(t, err)
		backends = append(backends, config.BackendConfig{URL: parsed})
	}

	lb := balancer.New(
		context.Background(),
		backends,
		config.LoadBalancerConfig{
			Algorithm:           "round_robin",
			HealthCheckInterval: 50, // ms
		},
	)

	rateLimitConfig := config.RateLimitConfig{
		Algorithm: "token_buckets",
		Options: config.TokenBucketLimiterOptions{
			DefaultCapacity:     10,
			DefaultRefillPeriod: 50,
		},
	}

	rl := ratelimit.New(rateLimitConfig.Algorithm, rateLimitConfig)

	srvImpl := httpGateway.NewServer(
		context.Background(),
		lb,
		rl,
		httpGateway.WithHost(""),
		httpGateway.WithPort(0),
	)

	h := srvImpl.Handler()
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, "healthy", string(b))

	resp, err = http.Get(srv.URL + "/")
	require.NoError(t, err)
	b, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, "ok", string(b))

	ss.SetDown(true)
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 5; i++ {
		resp, err := http.Get(srv.URL + "/")
		require.NoError(t, err)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.Equal(t, "healthy", string(b))
	}

	ss.SetDown(false)
	time.Sleep(100 * time.Millisecond)

	results := map[string]bool{}
	for i := 0; i < 4; i++ {
		resp, err := http.Get(srv.URL + "/")
		require.NoError(t, err)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		results[string(b)] = true
	}
	require.Contains(t, results, "healthy")
	require.Contains(t, results, "ok")
}
