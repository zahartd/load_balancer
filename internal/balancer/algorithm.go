package balancer

import "github.com/zahartd/load_balancer/internal/backend"

type Algorithm interface {
	Next(backends []*backend.Backend) (*backend.Backend, error)
}
