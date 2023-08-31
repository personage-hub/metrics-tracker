package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/personage-hub/metrics-tracker/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

func main() {

	config := parseFlag()
	log, err := logger.Initialize(config.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	log.Info("Running agent", zap.String("Server address", config.ServerAddress))

	mc := NewMonitoringClient(http.DefaultClient, log, config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mc.StartMonitoring(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("Received signal", zap.String("signal", sig.String()))
		cancel()
	}
}
