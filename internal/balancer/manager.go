package balancer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zahartd/load_balancer/internal/backend"
	"github.com/zahartd/load_balancer/internal/config"
)

type LoadBalancer struct {
	balancer Algorithm
	backends []*backend.Backend
}

func New(ctx context.Context, backendsConfigs []config.BackendConfig, config config.LoadBalancerConfig) *LoadBalancer {
	backends := make([]*backend.Backend, 0, len(backendsConfigs))
	for _, bc := range backendsConfigs {
		backends = append(backends, &backend.Backend{
			URL: bc.URL,
		})
	}

	lb := LoadBalancer{
		balancer: CreateAlgorithm(config.Algorithm),
		backends: backends,
	}

	healthCheckInterval := config.HealthCheckIntervalMS.AsDuration()
	go lb.healthCheckingRoutine(ctx, healthCheckInterval)

	return &lb
}

func (lb *LoadBalancer) getAlive() []*backend.Backend {
	var alive []*backend.Backend
	for _, b := range lb.backends {
		if b.IsAlive() {
			alive = append(alive, b)
		}
	}
	return alive
}

func (lb *LoadBalancer) healthCheckingRoutine(ctx context.Context, interval time.Duration) {
	httpTimeout := 2 * time.Second // TODO: move to config
	client := http.Client{Timeout: httpTimeout}

	// First healthcheck
	lb.healthCheck(ctx, &client)

	// Health checking loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb.healthCheck(ctx, &client)
		}
	}
}

func (lb *LoadBalancer) healthCheck(ctx context.Context, client *http.Client) {
	eg, egCtx := errgroup.WithContext(ctx)
	for _, b := range lb.backends {
		eg.Go(func() error {
			healthURL := b.URL.ResolveReference(&url.URL{Path: "/ping"}).String()
			req, err := http.NewRequestWithContext(egCtx, http.MethodGet, healthURL, nil)
			if err != nil {
				return fmt.Errorf("bad request for %s with %v", healthURL, err)
			}

			resp, err := client.Do(req)
			alive := err == nil && resp.StatusCode == http.StatusOK
			if resp != nil {
				resp.Body.Close()
			}

			b.SetAlive(alive)

			log.Printf("backend health update: url=%s alive=%t", b.URL, alive)
			return nil
		})
	}
	if err := eg.Wait(); err != nil && err != context.Canceled {
		log.Printf("health checks terminated early: %v", err)
	}
}

func (lb *LoadBalancer) NextBackend() (*backend.Backend, error) {
	alives := lb.getAlive()

	nextBackend, err := lb.balancer.Next(alives)
	if err != nil {
		log.Printf("Failed to get next backend: %s\n", err.Error())
		return nil, err
	}

	nextBackend.IncConns()

	return nextBackend, nil
}

func (lb *LoadBalancer) MarkBackendStatus(url string, alive bool) {
	for _, b := range lb.backends {
		if b.URL.String() == url {
			b.SetAlive(alive)
			log.Printf(
				"backend health update: url=%s alive=%t",
				url, alive,
			)
			return
		}
	}
}
