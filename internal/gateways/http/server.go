package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zahartd/load_balancer/internal/balancer"
	"github.com/zahartd/load_balancer/internal/ratelimit"
)

type Server struct {
	host         string
	port         uint16
	httpServer   *http.Server
	loadBalancer *balancer.LoadBalancer
	rateLimiter  *ratelimit.RateLimiter
}

func NewServer(lb *balancer.LoadBalancer, rl *ratelimit.RateLimiter, options ...func(*Server)) *Server {
	s := &Server{
		loadBalancer: lb,
		rateLimiter:  rl,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	proxyHandler := NewProxy(s.loadBalancer)

	// Added middlewares
	rateLimitedHandler := RateLimitMiddleware(ctx, s.rateLimiter)(proxyHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: rateLimitedHandler,
	}
	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutdownError error
	if err := s.httpServer.Shutdown(ctx); err != nil {
		shutdownError = errors.Join(err)
	}
	return shutdownError
}
