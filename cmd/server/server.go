package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/personage-hub/metrics-tracker/internal"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	storage *internal.MemStorage
}

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

func NewServer(storage *internal.MemStorage) *Server {
	return &Server{
		storage: storage,
	}
}

func (s *Server) metricsList(metricType MetricType) []string {
	var list []string
	switch metricType {
	case Gauge:
		for metricName := range s.storage.GaugeMap() {
			list = append(list, metricName)
		}
	case Counter:
		for metricName := range s.storage.CounterMap() {
			list = append(list, metricName)
		}
	}

	return list
}

func (s *Server) metricsHandle(rw http.ResponseWriter, r *http.Request) {
	gaugeList := s.metricsList(Gauge)
	counterList := s.metricsList(Counter)

	result := "Gauge list: " +
		strings.Join(gaugeList, ", ") +
		"\n" +
		"Counter list: " +
		strings.Join(counterList, ", ")

	_, _ = io.WriteString(rw, result)
}

func (s *Server) updateMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		res.Write([]byte("Method not allowed"))
		return
	}
	metricType := MetricType(strings.ToLower(chi.URLParam(req, "metricType")))
	metricName := chi.URLParam(req, "metricName")
	metricValue := chi.URLParam(req, "metricValue")
	if metricName == "" {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte(fmt.Errorf("metric unkwon").Error()))
		return
	}
	switch metricType {
	case Gauge:
		floatValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Errorf("invalid metric type: %s", metricType).Error()))
			return
		}
		s.storage.GaugeUpdate(metricName, floatValue)

	case Counter:
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

func (s *Server) metricGet(writer http.ResponseWriter, request *http.Request) {
	metricType := MetricType(strings.ToLower(chi.URLParam(request, "metricType")))
	metricName := chi.URLParam(request, "metricName")

	switch metricType {
	case Gauge:
		value, ok := s.storage.GetGaugeMetric(metricName)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Преобразуем значение к строке
		valueStr := fmt.Sprintf("%v", value)
		writer.Write([]byte(valueStr))

	case Counter:
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
		r.Get("/", s.metricsHandle)
		r.Route("/value", func(r chi.Router) {
			r.Route("/{metricType}", func(r chi.Router) {
				r.Get("/{metricName}", s.metricGet)
			})
		})
	})

	r.Post("/update/{metricType}/{metricName}/{metricValue}", s.updateMetric)

	return http.ListenAndServe(c.ServerAddress, r)
}
