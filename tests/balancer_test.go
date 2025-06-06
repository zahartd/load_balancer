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
	"github.com/zahartd/load_balancer/tests/utils"
)

type BalancerTestSuite struct {
	suite.Suite

	// backends
	healthyServer *httptest.Server
	switchServer  *utils.SwitchServer

	// componets
	lb *balancer.LoadBalancer
	// Can use withous rate limiter

	// main api server
	apiServer *httptest.Server
}

func TestBalancerSuite(t *testing.T) {
	// discard non-fatal logs  if it runs without -v
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}

	suite.Run(t, new(BalancerTestSuite))
}

func (s *BalancerTestSuite) SetupSuite() {
	s.healthyServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("always healthy"))
	}))

	s.switchServer = utils.NewSwitchServer()

	var backends []config.BackendConfig
	for _, u := range []string{s.healthyServer.URL, s.switchServer.URL()} {
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

	srvImpl := httpGateway.NewServer(
		context.Background(),
		s.lb,
		nil,
		httpGateway.WithHost(""),
		httpGateway.WithPort(0),
	)
	s.apiServer = httptest.NewServer(srvImpl.Handler())
}

func (s *BalancerTestSuite) TearDownSuite() {
	s.healthyServer.Close()
	s.switchServer.Close()
	s.apiServer.Close()
}

func (s *BalancerTestSuite) waitAlive(want int) {
	require.Eventually(
		s.T(),
		func() bool { return s.lb.AliveBackends() == want },
		200*time.Millisecond,
		50*time.Millisecond,
	)
}

func (s *BalancerTestSuite) doRequest() string {
	resp, err := http.Get(s.apiServer.URL + "/")
	s.Require().NoError(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return string(body)
}

func (s *BalancerTestSuite) TestRoundRobin_Common() {
	s.waitAlive(2)

	results := map[string]struct{}{}
	for range 4 {
		results[s.doRequest()] = struct{}{}
	}

	s.Require().Contains(results, "always healthy")
	s.Require().Contains(results, "ok")
	s.Require().Len(results, 2)
}

func (s *BalancerTestSuite) TestRoundRobin_DownAndRecovery() {
	s.switchServer.SetDown(true)
	s.waitAlive(1)

	results := map[string]struct{}{}
	for range 4 {
		results[s.doRequest()] = struct{}{}
	}
	s.Require().Equal(map[string]struct{}{"always healthy": {}}, results)

	s.switchServer.SetDown(false)
	s.waitAlive(2)

	results = map[string]struct{}{}
	for range 4 {
		results[s.doRequest()] = struct{}{}
	}
	s.Require().Contains(results, "always healthy")
	s.Require().Contains(results, "ok")
	s.Require().Len(results, 2)
}
