package main

import (
	"bytes"
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type MonitoringClient struct {
	Client        *http.Client
	Config        Config
	Storage       *storage.MemStorage
	MetricFuncMap map[string]func(*runtime.MemStats) float64
}

func NewMonitoringClient(client *http.Client, config Config) *MonitoringClient {
	metricFuncMap := map[string]func(*runtime.MemStats) float64{
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
	return &MonitoringClient{
		Client:        client,
		Config:        config,
		Storage:       storage.NewMemStorage(),
		MetricFuncMap: metricFuncMap,
	}
}

func (mc *MonitoringClient) SendMetric(metricType, metricName, metricValue string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", mc.Config.ServerAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(""))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", "text/plain")

	start := time.Now()

	resp, err := mc.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "sending metric")
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	logger.Log.Info(
		"HTTP request sent",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.String("status", resp.Status),
		zap.String("duration", duration.String()),
	)

	fmt.Println("Response:", resp.Status)

	return nil
}

func (mc *MonitoringClient) StartMonitoring() error {
	tickerPoll := time.NewTicker(mc.Config.PollInterval)
	tickerReport := time.NewTicker(mc.Config.ReportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-tickerPoll.C:
			err := mc.CollectMetrics()
			if err != nil {
				return errors.Wrap(err, "collecting metrics")
			}
		case <-tickerReport.C:
			err := mc.StartReporting()
			if err != nil {
				return errors.Wrap(err, "reporting metrics")
			}
		}
	}
}

func (mc *MonitoringClient) CollectMetrics() error {
	mc.Storage.GaugeUpdate("RandomValue", rand.Float64())

	var metrics runtime.MemStats
	runtime.ReadMemStats(&metrics)

	for metricName, metricFunc := range mc.MetricFuncMap {
		value := metricFunc(&metrics)
		mc.Storage.GaugeUpdate(metricName, value)
	}
	mc.Storage.CounterUpdate("PollCount", 1)

	return nil
}

func (mc *MonitoringClient) UpdateGaugeMetrics() error {
	for metricName, metricValue := range mc.Storage.GaugeMap() {
		err := mc.SendMetric("gauge", metricName, strconv.FormatFloat(metricValue, 'f', -1, 64))
		if err != nil {
			return errors.Wrap(err, "sending gauge metric")
		}
	}
	return nil
}

func (mc *MonitoringClient) UpdateCounterMetrics() error {
	for metricName, metricValue := range mc.Storage.CounterMap() {
		err := mc.SendMetric("counter", metricName, strconv.FormatInt(metricValue, 10))
		if err != nil {
			return errors.Wrap(err, "sending counter metric")
		}
	}
	return nil
}

func (mc *MonitoringClient) StartReporting() error {
	fmt.Println("Reporting metrics to server...")
	err := mc.UpdateGaugeMetrics()
	if err != nil {
		return errors.Wrap(err, "updating gauge metrics")
	}
	err = mc.UpdateCounterMetrics()
	if err != nil {
		return errors.Wrap(err, "updating counter metrics")
	}
	return nil
}
