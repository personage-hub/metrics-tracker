package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/personage-hub/metrics-tracker/internal/metrics"
	"github.com/personage-hub/metrics-tracker/internal/storage"
)

type Server struct {
	storage *storage.MemStorage
}

func NewServer(storage *storage.MemStorage) *Server {
	return &Server{
		storage: storage,
	}
}

func (s *Server) updateMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		res.Write([]byte("Method not allowed"))
		return
	}

	var metric metrics.Metrics

	err := easyjson.UnmarshalFromReader(req.Body, &metric)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Invalid request payload"))
		return
	}

	switch metric.MType {
	case "gauge":
		if metric.Value != nil {
			s.storage.GaugeUpdate(metric.ID, *metric.Value)
		} else {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Missing value for gauge metric"))
			return
		}
	case "counter":
		if metric.Delta != nil {
			s.storage.CounterUpdate(metric.ID, *metric.Delta)
		} else {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Missing delta for counter metric"))
			return
		}
	default:
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Invalid metric type"))
		return
	}

	data, _ := easyjson.Marshal(metric)
	res.WriteHeader(http.StatusOK)
	res.Write(data)
}

func (s *Server) getMetric(writer http.ResponseWriter, request *http.Request) {
	var metric metrics.Metrics
	err := easyjson.UnmarshalFromReader(request.Body, &metric)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Invalid request payload"))
		return
	}

	switch metric.MType {
	case "gauge":
		value, ok := s.storage.GetGaugeMetric(metric.ID)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		metric.Value = &value
	case "counter":
		value, ok := s.storage.GetCounterMetric(metric.ID)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		metric.Delta = &value
	default:
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Invalid metric type"))
		return
	}

	data, _ := easyjson.Marshal(metric)
	writer.Write(data) // send back the retrieved metric
}

func (s *Server) Run(c Config) error {
	r := chi.NewRouter()
	r.Use(requestWithLogging)
	r.Route("/", func(r chi.Router) {
		r.Post("/update/", s.updateMetric)
		r.Post("/value/", s.getMetric)
	})

	return http.ListenAndServe(c.ServerAddress, r)
}
