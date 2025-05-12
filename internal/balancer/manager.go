package balancer

import (
	"github.com/zahartd/load_balancer/internal/backend"
	"github.com/zahartd/load_balancer/internal/config"
)

type LoadBalancer struct {
	balancer Algorithm
	backends []*backend.Backend
}

func New(backendsConfigs []config.BackendConfig, algorithm string) *LoadBalancer {
	backends := make([]*backend.Backend, 0, len(backendsConfigs))
	for _, bc := range backendsConfigs {
		backends = append(backends, &backend.Backend{
			URL: bc.URL,
		})
	}
	return &LoadBalancer{
		balancer: CreateAlgorithm(algorithm),
		backends: backends,
	}
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
