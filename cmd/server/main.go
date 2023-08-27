package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"

	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config := parseFlags()

	log, err := logger.Initialize(config.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	s := storage.NewMemStorage()
	d := storage.DumpFile{Path: config.FileStorage}

	if _, err := os.Stat(config.FileStorage); os.IsNotExist(err) {
		log.Warn("File does not exist, skipping restore.")
	} else {
		if err := d.RestoreData(s); err != nil {
			log.Fatal("Fail restore data from dump", zap.Error(err))
		}
	}

	syncSave := config.StoreInterval == 0
	if !syncSave {
		go func() {
			err := storage.PeriodicSave(&d, s, config.StoreInterval)
			if err != nil {
				log.Fatal("Fail saving data to dump", zap.Error(err))
			}
		}()
	}

	log.Info("Running server", zap.String("address", config.ServerAddress))
	server := NewServer(s, &d, syncSave, log)
	r := chi.NewRouter()
	r.Use(requestWithLogging(server.logger))
	r.Use(gzipHandler)
	r.Mount("/", server.MetricRoute())

	err = http.ListenAndServe(config.ServerAddress, r)

	if err != nil {
		fmt.Printf("Error in server: #{err}")
	}

}
