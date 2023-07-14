package main

import (
	"bytes"
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal"
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
	go startPolling(storage)
	startReporting(storage)
}

func startPolling(s *internal.MemStorage) {
	for {
		collectMetrics(s)
		time.Sleep(pollInterval)
	}
}

func collectMetrics(s *internal.MemStorage) {
	s.GaugeUpdate("RandomValue", rand.Float64())

	var metrics runtime.MemStats
	runtime.ReadMemStats(&metrics)

	for metricName, metricFunc := range metricFuncMap {
		value := metricFunc(&metrics)
		s.GaugeUpdate(metricName, value)
	}
	s.CounterUpdate("PollCount", 1)
}

func updateGaugeMetrics(metrics map[string]float64) {
	for metricName, metricValue := range metrics {
		go sendMetric("gauge", metricName, strconv.FormatFloat(metricValue, 'f', -1, 64))
	}
}

func updateCounterMetrics(metrics map[string]int64) {
	for metricName, metricValue := range metrics {
		go sendMetric("gauge", metricName, strconv.FormatInt(metricValue, 10))
	}
}

func sendMetric(metricType, metricName, metricValue string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(""))
	if err != nil {
		fmt.Println("Error creating request:", err)
		panic(err)
	}

	req.Header.Set("Content-Type", "text/plain")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending metric:", err)
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response:", resp.Status)
}

func startReporting(s *internal.MemStorage) {
	for {
		fmt.Println("Reporting metrics to server...")
		updateGaugeMetrics(s.GaugeMap())
		updateCounterMetrics(s.CounterMap())
		time.Sleep(reportInterval)
	}
}
