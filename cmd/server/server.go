package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

func (s *Server) updateMetricV2(res http.ResponseWriter, req *http.Request) {
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

func (s *Server) getMetricV2(writer http.ResponseWriter, request *http.Request) {
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

func (s *Server) updateMetricV1(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		res.Write([]byte("Method not allowed"))
		return
	}
	metricType := strings.ToLower(chi.URLParam(req, "metricType"))
	metricName := chi.URLParam(req, "metricName")
	metricValue := chi.URLParam(req, "metricValue")
	if metricName == "" {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte(fmt.Errorf("metric unkwon").Error()))
		return
	}
	switch metricType {
	case "gauge":
		floatValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Errorf("invalid metric type: %s", metricType).Error()))
			return
		}
		s.storage.GaugeUpdate(metricName, floatValue)

	case "counter":
		intValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Errorf("invalid metric type: %s", metricType).Error()))
			return
		}
		s.storage.CounterUpdate(metricName, intValue)

	default:
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Errorf("unknown metric type: %s", metricType).Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (s *Server) metricGetV1(writer http.ResponseWriter, request *http.Request) {
	metricType := strings.ToLower(chi.URLParam(request, "metricType"))
	metricName := chi.URLParam(request, "metricName")

	switch metricType {
	case "gauge":
		value, ok := s.storage.GetGaugeMetric(metricName)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Преобразуем значение к строке
		valueStr := fmt.Sprintf("%v", value)
		writer.Write([]byte(valueStr))

	case "counter":
		value, ok := s.storage.GetCounterMetric(metricName)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Преобразуем значение к строке
		valueStr := fmt.Sprintf("%v", value)
		writer.Write([]byte(valueStr))

	default:
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *Server) Run(c Config) error {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/api/v1/update/{metricType}/{metricName}/{metricValue}", s.updateMetricV1)
		r.Get("/api/v1/value/{metricType}/{metricName}", s.metricGetV1)
		r.Post("/api/v2/update/", s.updateMetricV2)
		r.Post("/api/v2/value/", s.getMetricV2)
	})

	return http.ListenAndServe(c.ServerAddress, r)
}
