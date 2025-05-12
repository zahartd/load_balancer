package balancer

import (
	"context"
	"net/http"
	"net/url"
	"time"

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

	healthCheckInterval := time.Duration(int64(config.HealthCheckInterval))
	go lb.healthCheckingLoop(ctx, healthCheckInterval)

	return &lb
}

func (lb *LoadBalancer) getAlive() []*backend.Backend {
	var alive []*backend.Backend
	for _, b := range lb.backends {
		b.Mutex.RLock()
		if b.Alive {
			alive = append(alive, b)
		}
		b.Mutex.RUnlock()
	}
	return alive
}

func (lb *LoadBalancer) healthCheckingLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	client := http.Client{Timeout: interval}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, b := range lb.backends {
				go func(b *backend.Backend) {
					healthURL := b.URL.ResolveReference(&url.URL{Path: "/ping"}).String()
					resp, err := client.Get(healthURL)
					alive := err == nil && resp.StatusCode == http.StatusOK
					if resp != nil {
						resp.Body.Close()
					}
					lb.MarkBackendStatus(b.URL.String(), alive)
				}(b)
			}
		}
	}
}

func (lb *LoadBalancer) NextBackend() (*backend.Backend, error) {
	alives := lb.getAlive()

	nextBackend, err := lb.balancer.Next(alives)
	if err != nil {
		return nil, err
	}

	nextBackend.Mutex.Lock()
	nextBackend.ActiveConns++
	nextBackend.Mutex.Unlock()

	return nextBackend, nil
}

func (lb *LoadBalancer) MarkBackendStatus(url string, alive bool) {
	for _, b := range lb.backends {
		if b.URL.Hostname() == url {
			b.Mutex.Lock()
			b.Alive = alive
			b.Mutex.Unlock()
			return
		}
	}
}
