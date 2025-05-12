package integration_test

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zahartd/load_balancer/internal/balancer"
	"github.com/zahartd/load_balancer/internal/config"
	httpGateway "github.com/zahartd/load_balancer/internal/gateways/http"
	"github.com/zahartd/load_balancer/internal/ratelimit"
)

type RateLimiterSuite struct {
	suite.Suite

	// backends
	healthyServer *httptest.Server

	// componets
	lb *balancer.LoadBalancer
	rl *ratelimit.RateLimiter

	// main api server
	apiServer *httptest.Server
}

func TestRateLimiterSuite(t *testing.T) {
	// discard non-fatal logs  if it runs without -v
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}

	suite.Run(t, new(RateLimiterSuite))
}

func (s *RateLimiterSuite) SetupSuite() {
	s.healthyServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	var backends []config.BackendConfig
	for _, u := range []string{s.healthyServer.URL} {
		parsed, err := url.Parse(u)
		s.Require().NoError(err)
		backends = append(backends, config.BackendConfig{
			URL: parsed,
		})
	}

	s.lb = balancer.New(
		context.Background(),
		backends,
		config.LoadBalancerConfig{
			Algorithm:             "round_robin",
			HealthCheckIntervalMS: 50,
		},
	)

	s.rl = ratelimit.New("token_bucket", config.TokenBucketLimiterOptions{
		DefaultCapacity:         3,
		DefaultRefillIntervalMS: 200,
	})

	srvImpl := httpGateway.NewServer(
		context.Background(),
		s.lb,
		s.rl,
		httpGateway.WithHost(""),
		httpGateway.WithPort(0),
	)
	s.apiServer = httptest.NewServer(srvImpl.Handler())
}

func (s *RateLimiterSuite) TearDownSuite() {
	s.healthyServer.Close()
	s.apiServer.Close()
}

func (s *RateLimiterSuite) waitAlive(want int) {
	require.Eventually(
		s.T(),
		func() bool { return s.lb.AliveBackends() == want },
		200*time.Millisecond,
		50*time.Millisecond,
	)
}

func (s *RateLimiterSuite) doRequest(key string) (int, string) {
	req, _ := http.NewRequest("GET", s.apiServer.URL+"/", nil)
	req.Header.Set("X-API-Key", key)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func (s *RateLimiterSuite) TestTokenBucket_LimitAndRefill() {
	s.waitAlive(1)

	const key = "test-client"
	for i := range 3 {
		code, body := s.doRequest(key)
		s.Equal(http.StatusOK, code, "request %d should pass", i+1)
		s.Equal("ok", body)
	}

	code, _ := s.doRequest(key)
	s.Equal(http.StatusTooManyRequests, code, "4th request should be rate-limited")

	time.Sleep(300 * time.Millisecond)

	code, body := s.doRequest(key)
	s.Equal(http.StatusOK, code, "after refill one token, request should pass")
	s.Equal("ok", body)

	code, _ = s.doRequest(key)
	s.Equal(http.StatusTooManyRequests, code, "next request after using refill should be limited")
}
