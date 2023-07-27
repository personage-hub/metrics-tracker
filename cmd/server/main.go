package main

import (
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config := parseFlags()

	storage := storage.NewMemStorage()

	if err := logger.Initialize(config.FlagLogLevel); err != nil {
		panic(err)
	}

	logger.Log.Info("Running server", zap.String("address", config.ServerAddress))

	server := NewServer(storage)

	if err := server.Run(config); err != nil {
		panic(err)
	}
}
