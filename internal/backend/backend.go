package backend

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL         *url.URL
	Alive       bool
	Mutex       sync.RWMutex
	ActiveConns int
}
