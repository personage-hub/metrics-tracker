package main

import (
	"github.com/personage-hub/metrics-tracker/internal/db"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/personage-hub/metrics-tracker/internal/dumper"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/middlewares"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config := parseFlags()
	log, err := logger.Initialize(config.FlagLogLevel)
	if err != nil {
		panic(err)
	}
	var database db.Database
	var s storage.Storage

	if config.DatabaseDSN == "" {
		d := dumper.NewDumper(config.FileStorage)
		s, err = storage.NewMemStorage(d, config.Restore)
		if err != nil {
			log.Error("skipping restore due to error", zap.Error(err))
		} else {
			log.Info("restore successfully complete")
		}

		go storage.PeriodicSave(s, d, config.StoreInterval)
	} else {
		database, err = db.CreateAndConnect(config.DatabaseDSN)
		if err != nil {
			log.Error("DB error", zap.Error(err))
		}
		err = database.CreateTables()
		if err != nil {
			log.Error("DB error", zap.Error(err))
		}
		s = storage.NewDBStorage(&database, log)

	}

	log.Info("Running server", zap.String("address", config.ServerAddress))
	server := NewServer(s, database, log)
	r := chi.NewRouter()
	r.Use(middlewares.RequestWithLogging(server.logger))
	r.Use(middlewares.GzipHandler)
	r.Mount("/", server.MetricRoute())
	err = http.ListenAndServe(config.ServerAddress, r)

	if err != nil {
		log.Fatal("Error in server:", zap.Error(err))
	}

}
