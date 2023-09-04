package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/personage-hub/metrics-tracker/internal/dumper"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/middlewares"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	config := parseFlags()
	log, err := logger.Initialize(config.FlagLogLevel)
	if err != nil {
		panic(err)
	}
	syncSave := config.StoreInterval == 0

	d := dumper.NewDumper(config.FileStorage)
	s, err := storage.NewMemStorage(d, config.Restore, syncSave)

	if err != nil {
		panic(err)
	}

	if !syncSave {
		s.PeriodicSave(config.StoreInterval)
	}

	log.Info("Running server", zap.String("address", config.ServerAddress))
	server := NewServer(s, log)
	r := chi.NewRouter()
	r.Use(middlewares.RequestWithLogging(server.logger))
	r.Use(middlewares.GzipHandler)
	r.Mount("/", server.MetricRoute())

	err = http.ListenAndServe(config.ServerAddress, r)

	if err != nil {
		log.Fatal("Error in server:", zap.Error(err))
	}

}
