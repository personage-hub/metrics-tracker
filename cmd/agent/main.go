package main

import (
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

func main() {

	config := parseFlag()
	if err := logger.Initialize(config.FlagLogLevel); err != nil {
		panic(err)
	}

	logger.Log.Info("Running agent", zap.String("Server address", config.ServerAddress))

	mc := NewMonitoringClient(http.DefaultClient, config)

	go func() {
		err := mc.StartMonitoring()
		if err != nil {
			fmt.Printf("Error in monitoring: %v\n", err)
		}
	}()

	select {}
}
