package backend

import (
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL         *url.URL
	alive       atomic.Bool
	activeConns atomic.Int64
}

func (b *Backend) IsAlive() bool {
	return b.alive.Load()
}

func (b *Backend) SetAlive(up bool) {
	b.alive.Store(up)
}

func (b *Backend) IncConns() {
	b.activeConns.Add(1)
}

func (b *Backend) DecConns() {
	b.activeConns.Add(-1)
}

func (b *Backend) ActiveConns() int64 {
	return b.activeConns.Load()
}
