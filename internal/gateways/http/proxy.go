package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"syscall"

	"github.com/zahartd/load_balancer/internal/balancer"
)

func NewProxy(lb *balancer.LoadBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get backend for request
		backend, err := lb.NextBackend()
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		defer backend.DecConns()

		proxy := httputil.NewSingleHostReverseProxy(backend.URL)

		// Processing next backend errors:
		// reaching the backend or errors from ModifyResponse.
		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
			log.Printf("Backend %s return error: %s\n", backend.URL.String(), e.Error())

			if errors.Is(e, context.Canceled) {
				log.Printf("client canceled: %v\n", err)
				return
			}

			// handle hetwork error
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				// Transport/connection error
				if opErr.Op == "dial" || opErr.Timeout() || errors.Is(opErr.Err, syscall.ECONNREFUSED) {
					lb.MarkBackendStatus(backend.URL.String(), false)
					http.Error(rw, "Bad gateway", http.StatusBadGateway)
					return
				}

				// Read timeot
				if opErr.Op == "read" && opErr.Timeout() {
					lb.MarkBackendStatus(backend.URL.String(), false)
					http.Error(rw, "Upstream timeout", http.StatusGatewayTimeout)
					return
				}
			}

			lb.MarkBackendStatus(backend.URL.String(), false)
			http.Error(rw, fmt.Sprintf("Backend error: %v", e), http.StatusBadGateway)
		}

		proxy.ServeHTTP(w, r)
	})
}
