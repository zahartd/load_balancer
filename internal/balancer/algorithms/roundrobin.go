package balancer_algorithms

import (
	"errors"
	"sync/atomic"

	"github.com/zahartd/load_balancer/internal/backend"
)

type RoundRobin struct {
	current uint32
}

var ErrNoAvailableBackends = errors.New("no available backends")

func NewRoundRobinAlghoritm() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) Next(backends []*backend.Backend) (*backend.Backend, error) {
	backendCount := len(backends)
	if backendCount == 0 {
		return nil, ErrNoAvailableBackends
	}
	next := atomic.AddUint32(&rr.current, 1)
	return backends[(int(next)-1)%backendCount], nil
}
