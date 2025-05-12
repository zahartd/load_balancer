package balancer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zahartd/load_balancer/internal/config"
	"github.com/zahartd/load_balancer/internal/models"
)

type LoadBalancer struct {
	// Note: Must provide threadsafe access to itself
	balancer Algorithm

	// Note: it's field in current impl should be constant for thread safe
	// Currently, the backend list is statistical and sets at the start of the application through the config
	// TODO: Make it dynamicly and added synchronization
	backends []*models.Backend
}

func New(ctx context.Context, backendsConfigs []config.BackendConfig, config config.LoadBalancerConfig) *LoadBalancer {
	// Set backends list on startup (this list is constant in all time of app working)
	// Therefore, we consider access to backends from different flows safe
	backends := make([]*models.Backend, 0, len(backendsConfigs))
	for _, bc := range backendsConfigs {
		backends = append(backends, &models.Backend{
			URL: bc.URL,
		})
	}

	// Create new balancer
	lb := &LoadBalancer{
		balancer: CreateAlgorithm(config.Algorithm),
		backends: backends,
	}

	// Start in separate goroutine periodical task with healthchecking
	healthCheckInterval := config.HealthCheckIntervalMS.AsDuration()
	go lb.healthCheckingRoutine(ctx, healthCheckInterval)

	return lb
}

func (lb *LoadBalancer) AliveBackends() int {
	return len(lb.getAlive())
}

func (lb *LoadBalancer) getAlive() []*models.Backend {
	// Fixed current states of backends
	// Non-blocking for other goroutines
	var alive []*models.Backend
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
	// Pass through the backends and ping each
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

var ErrNoAvailableBackends = errors.New("no available backends")

func (lb *LoadBalancer) NextBackend() (*models.Backend, error) {
	alives := lb.getAlive()

	if len(alives) == 0 {
		log.Println("There is no available backend")
		return nil, ErrNoAvailableBackends
	}

	nextBackend := lb.balancer.Next(alives)
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
