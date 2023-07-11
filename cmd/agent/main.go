package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverAddress  = "http://localhost:8080"
)

var metricFuncMap = map[string]func(m *runtime.MemStats) float64{
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
		return float64(m.GCCPUFraction)
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
	go startReporting()
	startPolling()
}

func startPolling() {
	for {
		collectMetrics()
		time.Sleep(pollInterval)
	}
}

func collectMetrics() {
	updateGaugeMetric("RandomValue", rand.Float64())

	metrics := runtime.MemStats{}
	runtime.ReadMemStats(&metrics)
	incrementCounterMetric("PollCount")

	for metricName, metricFunc := range metricFuncMap {
		value := metricFunc(&metrics)
		updateGaugeMetric(metricName, value)
	}
}

func updateGaugeMetric(name string, value float64) {
	sendMetric("gauge", name, strconv.FormatFloat(value, 'f', -1, 64))
}

func incrementCounterMetric(name string) {
	sendMetric("counter", name, "1")
}

func sendMetric(metricType, metricName, metricValue string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(""))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending metric:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Response:", resp.Status)
	fmt.Println("Body:", string(body))
}

func startReporting() {
	for {
		time.Sleep(reportInterval)
		fmt.Println("Reporting metrics to server...")
		// Добавьте здесь логику для отправки отчета на сервер
	}
}
