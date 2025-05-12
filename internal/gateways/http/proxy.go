package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/zahartd/load_balancer/internal/balancer"
)

func NewProxy(lb *balancer.LoadBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend, err := lb.NextBackend()
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(backend.URL)

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
			lb.MarkBackendStatus(backend.URL.String(), false)
			http.Error(rw, fmt.Sprintf("Backend error: %v", e), http.StatusBadGateway)
		}

		proxy.ServeHTTP(w, r)

		backend.DecConns()
	})
}
