package main

import (
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config := parseFlags()
	s := storage.NewMemStorage()
	d := storage.DumpFile{
		Path: config.FileStorage,
	}
	syncSave := config.StoreInterval == 0
	if !syncSave {
		go func() {
			err := storage.PeriodicSave(&d, *s, config.StoreInterval)
			if err != nil {
				panic(err)
			}
		}()
	}
	if config.Restore {
		if err := d.RestoreData(s); err != nil {
			panic(err)
		}
	}
	if err := logger.Initialize(config.FlagLogLevel); err != nil {
		panic(err)
	}
	logger.Log.Info("Running server", zap.String("address", config.ServerAddress))
	server := NewServer(s, &d, syncSave)
	if err := server.Run(config); err != nil {
		panic(err)
	}
}
