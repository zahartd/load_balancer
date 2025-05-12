package balancer

import (
	"context"
	"log"
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

	healthCheckInterval := time.Duration(config.HealthCheckInterval) * time.Second
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

	httpTimeout := 2 * time.Second // TODO: move to config
	client := http.Client{Timeout: httpTimeout}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, b := range lb.backends {
				go func(be *backend.Backend) {
					healthURL := be.URL.ResolveReference(&url.URL{Path: "/ping"}).String()
					resp, err := client.Get(healthURL)
					if err != nil {
						log.Printf("Backend %s health chek failed: %s", be.URL.String(), err.Error())
					}
					alive := err == nil && resp.StatusCode == http.StatusOK
					if resp != nil {
						resp.Body.Close()
					}
					lb.MarkBackendStatus(be.URL.String(), alive)
				}(b)
			}
		}
	}
}

func (lb *LoadBalancer) NextBackend() (*backend.Backend, error) {
	alives := lb.getAlive()

	nextBackend, err := lb.balancer.Next(alives)
	if err != nil {
		log.Printf("Failed to get next backend: %s\n", err.Error())
		return nil, err
	}

	nextBackend.Mutex.Lock()
	nextBackend.ActiveConns++
	nextBackend.Mutex.Unlock()

	return nextBackend, nil
}

func (lb *LoadBalancer) MarkBackendStatus(url string, alive bool) {
	for _, b := range lb.backends {
		if b.URL.String() == url {
			b.Mutex.Lock()
			b.Alive = alive
			b.Mutex.Unlock()
			log.Printf(
				"backend health update: url=%s alive=%t",
				url, alive,
			)
			return
		}
	}
}
