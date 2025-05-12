package balancer_algorithms

import (
	"errors"
	"sync/atomic"

	"github.com/zahartd/load_balancer/internal/backend"
)

type RoundRobin struct {
	current uint32
}

func NewRoundRobinAlghoritm() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) Next(backends []*backend.Backend) (*backend.Backend, error) {
	backendCount := len(backends)
	if backendCount == 0 {
		return nil, errors.New("no available backends")
	}
	next := atomic.AddUint32(&rr.current, 1)
	return backends[int(next)%backendCount], nil
}
