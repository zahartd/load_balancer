package balancer_algorithms

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zahartd/load_balancer/internal/models"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}
	os.Exit(m.Run())
}

func mustURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

func BenchmarkRoundRobin(b *testing.B) {
	log.SetOutput(io.Discard)

	rr := NewRoundRobinAlghoritm()
	b1 := &models.Backend{URL: mustURL("http://a")}
	b2 := &models.Backend{URL: mustURL("http://b")}
	b3 := &models.Backend{URL: mustURL("http://c")}
	backends := []*models.Backend{b1, b2, b3}

	for b.Loop() {
		rr.Next(backends)
	}
}

func TestRoundRobin_Next(t *testing.T) {
	t.Parallel()
	rr := NewRoundRobinAlghoritm()
	b1 := &models.Backend{URL: mustURL("http://a")}
	b2 := &models.Backend{URL: mustURL("http://b")}
	b3 := &models.Backend{URL: mustURL("http://c")}
	backends := []*models.Backend{b1, b2, b3}

	expected := []*models.Backend{b1, b2, b3, b1, b2, b3}

	for i, want := range expected {
		got := rr.Next(backends)
		require.Equal(t, want, got, "iteration %d: expected %v, got %v", i, want, got)
	}
}

func TestRoundRobin_Concurrent(t *testing.T) {
	t.Parallel()
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
			b := rr.Next(backends)
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
