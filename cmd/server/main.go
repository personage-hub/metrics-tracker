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
		for metricName, _ := range s.GaugeMap() {
			list = append(list, metricName)
		}
	case Counter:
		for metricName, _ := range s.CounterMap() {
			list = append(list, metricName)
		}
	}

	return list
}

func metricsHandle(rw http.ResponseWriter, r *http.Request) {
	s := r.Context().Value("storage").(*internal.MemStorage)
	gaugeList := metricsList(Gauge, s)
	counterList := metricsList(Counter, s)

	result := "Gauge list: " +
		strings.Join(gaugeList, ", ") +
		"\n" +
		"Counter list: " +
		strings.Join(counterList, ", ")

	_, _ = io.WriteString(rw, result)
}

func updateMetric(res http.ResponseWriter, req *http.Request, storage *internal.MemStorage) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		res.Write([]byte("Method not allowed"))
		return
	}
	urlParts := strings.Split(req.URL.Path, "/")
	if len(urlParts) != 5 {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := MetricType(urlParts[2])
	metricName := urlParts[3]
	metricValue := urlParts[4]
	switch metricType {
	case Gauge:
		floatValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Errorf("invalid metric type: %s", metricType).Error()))
			return
		}
		storage.GaugeUpdate(metricName, floatValue)

	case Counter:
		intValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Errorf("invalid metric type: %s", metricType).Error()))
			return
		}
		storage.CounterUpdate(metricName, intValue)

	default:
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Errorf("unknown metric type: %s", metricType).Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
}

func metricGet(writer http.ResponseWriter, request *http.Request) {
	s := request.Context().Value("storage").(*internal.MemStorage)
	metricType := MetricType(strings.ToLower(chi.URLParam(request, "metricType")))
	metricName := strings.ToLower(chi.URLParam(request, "metricName"))

	switch metricType {
	case Gauge:
		value, ok := s.GetGaugeMetric(metricName)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Преобразуем значение к строке
		valueStr := fmt.Sprintf("%v", value)
		response := fmt.Sprintf("Имя метрики: %s, Значение метрики: %s", metricName, valueStr)
		writer.Write([]byte(response))

	case Counter:
		value, ok := s.GetCounterMetric(metricName)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		// Преобразуем значение к строке
		valueStr := fmt.Sprintf("%v", value)
		response := fmt.Sprintf("Имя метрики: %s, Значение метрики: %s", metricName, valueStr)
		writer.Write([]byte(response))

	default:
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
}

func StorageMiddleware(s *internal.MemStorage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "storage", s)
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

	r.Post("/update", func(res http.ResponseWriter, req *http.Request) {
		updateMetric(res, req, storage)
	})

	return http.ListenAndServe(`:8080`, r)
}
