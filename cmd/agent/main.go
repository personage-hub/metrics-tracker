package main

import (
	"bytes"
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal"
	"github.com/pkg/errors"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

var metricFuncMap = map[string]func(*runtime.MemStats) float64{
	"Alloc": func(m *runtime.MemStats) float64 {
		return float64(m.Alloc)
	},
	"BuckHashSys": func(m *runtime.MemStats) float64 {
		return float64(m.BuckHashSys)
	},
	"Frees": func(m *runtime.MemStats) float64 {
		return float64(m.Frees)
	},
	"GCCPUFraction": func(m *runtime.MemStats) float64 {
		return m.GCCPUFraction
	},
	"GCSys": func(m *runtime.MemStats) float64 {
		return float64(m.GCSys)
	},
	"HeapAlloc": func(m *runtime.MemStats) float64 {
		return float64(m.HeapAlloc)
	},
	"HeapIdle": func(m *runtime.MemStats) float64 {
		return float64(m.HeapIdle)
	},
	"HeapInuse": func(m *runtime.MemStats) float64 {
		return float64(m.HeapInuse)
	},
	"HeapObjects": func(m *runtime.MemStats) float64 {
		return float64(m.HeapObjects)
	},
	"HeapReleased": func(m *runtime.MemStats) float64 {
		return float64(m.HeapReleased)
	},
	"HeapSys": func(m *runtime.MemStats) float64 {
		return float64(m.HeapSys)
	},
	"LastGC": func(m *runtime.MemStats) float64 {
		return float64(m.LastGC)
	},
	"Lookups": func(m *runtime.MemStats) float64 {
		return float64(m.Lookups)
	},
	"MCacheInuse": func(m *runtime.MemStats) float64 {
		return float64(m.MCacheInuse)
	},
	"MCacheSys": func(m *runtime.MemStats) float64 {
		return float64(m.MCacheSys)
	},
	"MSpanInuse": func(m *runtime.MemStats) float64 {
		return float64(m.MSpanInuse)
	},
	"MSpanSys": func(m *runtime.MemStats) float64 {
		return float64(m.MSpanSys)
	},
	"Mallocs": func(m *runtime.MemStats) float64 {
		return float64(m.Mallocs)
	},
	"NextGC": func(m *runtime.MemStats) float64 {
		return float64(m.NextGC)
	},
	"NumForcedGC": func(m *runtime.MemStats) float64 {
		return float64(m.NumForcedGC)
	},
	"NumGC": func(m *runtime.MemStats) float64 {
		return float64(m.NumGC)
	},
	"OtherSys": func(m *runtime.MemStats) float64 {
		return float64(m.OtherSys)
	},
	"PauseTotalNs": func(m *runtime.MemStats) float64 {
		return float64(m.PauseTotalNs)
	},
	"StackInuse": func(m *runtime.MemStats) float64 {
		return float64(m.StackInuse)
	},
	"StackSys": func(m *runtime.MemStats) float64 {
		return float64(m.StackSys)
	},
	"Sys": func(m *runtime.MemStats) float64 {
		return float64(m.Sys)
	},
	"TotalAlloc": func(m *runtime.MemStats) float64 {
		return float64(m.TotalAlloc)
	},
}

func main() {
	storage := internal.NewMemStorage()

	config := parseFlag()

	go func() {
		err := startMonitoring(storage, config)
		if err != nil {
			fmt.Printf("Error in monitoring: %v\n", err)
		}
	}()

	select {}
}

func startMonitoring(s *internal.MemStorage, c Config) error {
	tickerPoll := time.NewTicker(c.PollInterval)
	tickerReport := time.NewTicker(c.ReportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-tickerPoll.C:
			err := collectMetrics(s)
			if err != nil {
				return errors.Wrap(err, "collecting metrics")
			}
		case <-tickerReport.C:
			err := startReporting(c.ServerAddress, s)
			if err != nil {
				return errors.Wrap(err, "reporting metrics")
			}
		}
	}
}

func collectMetrics(s *internal.MemStorage) error {
	s.GaugeUpdate("RandomValue", rand.Float64())

	var metrics runtime.MemStats
	runtime.ReadMemStats(&metrics)

	for metricName, metricFunc := range metricFuncMap {
		value := metricFunc(&metrics)
		s.GaugeUpdate(metricName, value)
	}
	s.CounterUpdate("PollCount", 1)

	return nil
}

func updateGaugeMetrics(serverAddress string, metrics map[string]float64) error {
	for metricName, metricValue := range metrics {
		err := sendMetric(serverAddress, "gauge", metricName, strconv.FormatFloat(metricValue, 'f', -1, 64))
		if err != nil {
			return errors.Wrap(err, "sending gauge metric")
		}
	}
	return nil
}

func updateCounterMetrics(serverAddress string, metrics map[string]int64) error {
	for metricName, metricValue := range metrics {
		err := sendMetric(serverAddress, "counter", metricName, strconv.FormatInt(metricValue, 10))
		if err != nil {
			return errors.Wrap(err, "sending counter metric")
		}
	}
	return nil
}

func sendMetric(serverAddress string, metricType, metricName, metricValue string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", serverAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(""))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", "text/plain")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "sending metric")
	}
	defer resp.Body.Close()

	fmt.Println("Response:", resp.Status)

	return nil
}

func startReporting(serverAddress string, s *internal.MemStorage) error {
	fmt.Println("Reporting metrics to server...")
	err := updateGaugeMetrics(serverAddress, s.GaugeMap())
	if err != nil {
		return errors.Wrap(err, "updating gauge metrics")
	}
	err = updateCounterMetrics(serverAddress, s.CounterMap())
	if err != nil {
		return errors.Wrap(err, "updating counter metrics")
	}
	return nil
}
