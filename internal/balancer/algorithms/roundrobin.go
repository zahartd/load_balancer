package balancer_algorithms

import (
	"sync/atomic"

	"github.com/zahartd/load_balancer/internal/models"
)

type RoundRobin struct {
	current uint32
}

func NewRoundRobinAlghoritm() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) Next(backends []*models.Backend) *models.Backend {
	backendCount := len(backends)
	next := atomic.AddUint32(&rr.current, 1)
	return backends[(int(next)-1)%backendCount]
}
