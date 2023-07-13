package main

import (
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal"
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

func run(storage *internal.MemStorage) error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, func(res http.ResponseWriter, req *http.Request) {
		updateMetric(res, req, storage)
	})
	return http.ListenAndServe(`:8080`, mux)
}
