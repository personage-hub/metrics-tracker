package main

import (
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"go.uber.org/zap"
	"os"
)

func main() {
	config := parseFlags()

	if err := logger.Initialize(config.FlagLogLevel); err != nil {
		panic(err)
	}

	s := storage.NewMemStorage()
	d := storage.DumpFile{Path: config.FileStorage}

	if _, err := os.Stat(config.FileStorage); os.IsNotExist(err) {
		logger.Log.Warn("File does not exist, skipping restore.")
	} else {
		if err := d.RestoreData(s); err != nil {
			logger.Log.Fatal("Fail restore data from dump", zap.Error(err))
		}
	}

	syncSave := config.StoreInterval == 0
	if !syncSave {
		go func() {
			err := storage.PeriodicSave(&d, *s, config.StoreInterval)
			if err != nil {
				logger.Log.Fatal("Fail saving data to dump", zap.Error(err))
			}
		}()
	}

	logger.Log.Info("Running server", zap.String("address", config.ServerAddress))
	server := NewServer(s, &d, syncSave)
	if err := server.Run(config); err != nil {
		panic(err)
	}
}
