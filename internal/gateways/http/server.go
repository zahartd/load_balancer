package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Server struct {
	host       string
	port       uint16
	httpServer *http.Server
}

func NewServer(options ...func(*Server)) *Server {
	s := &Server{}
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

func (s *Server) Run(_ context.Context) error {
	s.httpServer = &http.Server{
		Addr: fmt.Sprintf("%s:%d", s.host, s.port),
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
