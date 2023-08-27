package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		go mc.StartMonitoring(ctx)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Println("Received signal:", sig)
		cancel()
	}()
	wg.Wait()
	<-sigCh
}
