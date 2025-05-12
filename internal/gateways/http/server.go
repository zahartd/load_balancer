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
	handler      http.Handler
	httpServer   *http.Server
	loadBalancer *balancer.LoadBalancer
	rateLimiter  *ratelimit.RateLimiter
}

func NewServer(ctx context.Context, lb *balancer.LoadBalancer, rl *ratelimit.RateLimiter, options ...func(*Server)) *Server {
	s := &Server{
		loadBalancer: lb,
		rateLimiter:  rl,
	}
	for _, o := range options {
		o(s)
	}

	proxyHandler := NewProxy(s.loadBalancer)
	s.handler = RateLimitMiddleware(ctx, s.rateLimiter)(proxyHandler)
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

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) Run(_ context.Context) error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.handler,
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
