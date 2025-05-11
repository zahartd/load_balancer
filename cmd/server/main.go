package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/zahartd/load_balancer/internal/balancer"
	"github.com/zahartd/load_balancer/internal/config"
	httpGateway "github.com/zahartd/load_balancer/internal/gateways/http"
)

const gracefulShutdownTime = 7 * time.Second // TODD: move it to env

func main() {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configs: %s\n", err.Error())
	}

	lb := balancer.New(cfg.Backends, cfg.LoadBalancer.Algorithm)

	r := httpGateway.NewServer(
		lb,
		httpGateway.WithHost(cfg.Server.Host),
		httpGateway.WithPort(cfg.Server.Port),
	)

	go func() {
		if err := r.Run(appCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %s\n", err)
		}
	}()

	<-appCtx.Done()

	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, shutdown := context.WithTimeout(context.Background(), gracefulShutdownTime)
	defer shutdown()
	if err := r.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
