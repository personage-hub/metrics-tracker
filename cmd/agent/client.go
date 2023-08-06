package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/personage-hub/metrics-tracker/internal/logger"
	"github.com/personage-hub/metrics-tracker/internal/metrics"
	project_constants "github.com/personage-hub/metrics-tracker/internal/project_constants"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
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

func (mc *MonitoringClient) compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gzw, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, errors.Wrap(err, "failed init compress writer.")
	}
	_, err = gzw.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write compress data to buffer.")
	}
	defer gzw.Close()

	return b.Bytes(), nil
}

func (mc *MonitoringClient) SendMetric(metric metrics.Metrics) error {
	url := fmt.Sprintf("http://%s/update/", mc.Config.ServerAddress)

	data, err := easyjson.Marshal(metric)
	if err != nil {
		return errors.Wrap(err, "Converting data for request")
	}
	compressData, err := mc.compress(data)
	if err != nil {
		return errors.Wrap(err, "Compress data for request")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(compressData))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", project_constants.ContentTypeJSON)
	req.Header.Set("Content-Encoding", project_constants.Compression)
	req.Header.Set("Accept-Encoding", project_constants.Compression)

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
			_ = mc.CollectMetrics()
		case <-tickerReport.C:
			_ = mc.StartReporting()
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
		m := metrics.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &metricValue,
		}
		err := mc.SendMetric(m)
		if err != nil {
			return errors.Wrap(err, "sending gauge metric")
		}
	}
	return nil
}

func (mc *MonitoringClient) UpdateCounterMetrics() error {
	for metricName, metricValue := range mc.Storage.CounterMap() {
		m := metrics.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &metricValue,
		}
		err := mc.SendMetric(m)
		if err != nil {
			return errors.Wrap(err, "sending counter metric")
		}
	}
	return nil
}

func (mc *MonitoringClient) StartReporting() error {
	fmt.Println("Reporting metrics to server...")
	_ = mc.UpdateGaugeMetrics()
	_ = mc.UpdateCounterMetrics()
	return nil
}
