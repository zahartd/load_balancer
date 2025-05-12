package balancer_algorithms

import (
	"errors"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zahartd/load_balancer/internal/models"
)

func mustURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

func TestRoundRobin_Next(t *testing.T) {
	rr := NewRoundRobinAlghoritm()
	b1 := &models.Backend{URL: mustURL("http://a")}
	b2 := &models.Backend{URL: mustURL("http://b")}
	b3 := &models.Backend{URL: mustURL("http://c")}
	backends := []*models.Backend{b1, b2, b3}

	expected := []*models.Backend{b1, b2, b3, b1, b2, b3}

	for i, want := range expected {
		got, err := rr.Next(backends)
		require.NoError(t, err, "iteration %d: unexpected error", i)
		require.Equal(t, want, got, "iteration %d: expected %v, got %v", i, want, got)
	}
}

func TestRoundRobin_Empty(t *testing.T) {
	rr := NewRoundRobinAlghoritm()

	_, err := rr.Next(nil)

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoAvailableBackends), "expected ErrNoAvailableBackends, got %v", err)
}

func TestRoundRobin_Concurrent(t *testing.T) {
	rr := NewRoundRobinAlghoritm()

	b1 := &models.Backend{URL: mustURL("http://1")}
	b2 := &models.Backend{URL: mustURL("http://2")}
	backends := []*models.Backend{b1, b2}

	var wg sync.WaitGroup
	n := 100
	results := make(chan *models.Backend, n)

	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			b, err := rr.Next(backends)
			require.NoError(t, err)
			results <- b
		}()
	}
	wg.Wait()
	close(results)

	count := map[string]int{}
	for b := range results {
		count[b.URL.String()]++
	}

	require.Equal(t, n/2, count["http://1"], "unexpected distribution for b1")
	require.Equal(t, n/2, count["http://2"], "unexpected distribution for b2")
}
