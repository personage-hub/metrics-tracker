package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/personage-hub/metrics-tracker/internal"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type MetricType string

type storageKey int

const (
	storageKeyContextKey storageKey = iota
)

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

func main() {
	storage := internal.NewMemStorage()
	if err := run(storage); err != nil {
		panic(err)
	}
}

func metricsList(metricType MetricType, s *internal.MemStorage) []string {
	var list []string
	switch metricType {
	case Gauge:
		for metricName := range s.GaugeMap() {
			list = append(list, metricName)
		}
	case Counter:
		for metricName := range s.CounterMap() {
			list = append(list, metricName)
		}
	}

	return list
}

func metricsHandle(rw http.ResponseWriter, r *http.Request) {
	s := r.Context().Value(storageKeyContextKey).(*internal.MemStorage)
	gaugeList := metricsList(Gauge, s)
	counterList := metricsList(Counter, s)

	result := "Gauge list: " +
		strings.Join(gaugeList, ", ") +
		"\n" +
		"Counter list: " +
		strings.Join(counterList, ", ")

	_, _ = io.WriteString(rw, result)
}

func updateMetric(res http.ResponseWriter, req *http.Request) {
	s := req.Context().Value(storageKeyContextKey).(*internal.MemStorage)
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
		s.GaugeUpdate(metricName, floatValue)

	case Counter:
		intValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Errorf("invalid metric type: %s", metricType).Error()))
			return
		}
		s.CounterUpdate(metricName, intValue)

	default:
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Errorf("unknown metric type: %s", metricType).Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
}

func metricGet(writer http.ResponseWriter, request *http.Request) {
	s := request.Context().Value(storageKeyContextKey).(*internal.MemStorage)
	metricType := MetricType(strings.ToLower(chi.URLParam(request, "metricType")))
	metricName := chi.URLParam(request, "metricName")

	switch metricType {
	case Gauge:
		value, ok := s.GetGaugeMetric(metricName)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Преобразуем значение к строке
		valueStr := fmt.Sprintf("%v", value)
		writer.Write([]byte(valueStr))

	case Counter:
		value, ok := s.GetCounterMetric(metricName)
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

func StorageMiddleware(s *internal.MemStorage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), storageKeyContextKey, s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func run(storage *internal.MemStorage) error {
	r := chi.NewRouter()

	r.Use(StorageMiddleware(storage))

	r.Route("/", func(r chi.Router) {
		r.Get("/", metricsHandle)
		r.Route("/value", func(r chi.Router) {
			r.Route("/{metricType}", func(r chi.Router) {
				r.Get("/{metricName}", metricGet)
			})
		})
	})

	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateMetric)

	return http.ListenAndServe(serverAddress, r)
}
