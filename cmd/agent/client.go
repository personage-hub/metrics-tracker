package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"

	"github.com/mailru/easyjson"
	"github.com/personage-hub/metrics-tracker/internal/consts"
	"github.com/personage-hub/metrics-tracker/internal/metrics"
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
	Storage       storage.Storage
	MetricFuncMap map[string]func(*runtime.MemStats) float64
	logger        *zap.Logger
}

func NewMonitoringClient(client *http.Client, logger *zap.Logger, config Config) *MonitoringClient {
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
		logger:        logger,
	}
}

func (mc *MonitoringClient) compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gzw, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, errors.Wrap(err, "failed init compress writer.")
	}

	if _, err = gzw.Write(data); err != nil {
		return nil, errors.Wrap(err, "failed to write compress data to buffer.")
	}

	if err = gzw.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close compress writer.")
	}

	return b.Bytes(), nil
}

func (mc *MonitoringClient) SendMetric(metric metrics.Metrics) error {
	url := fmt.Sprintf("http://%s/update/", mc.Config.ServerAddress)

	data, err := easyjson.Marshal(metric)
	if err != nil {
		return errors.Wrap(err, "Converting data for request")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", consts.ContentTypeJSON)

	start := time.Now()

	resp, err := mc.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "sending metric")
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	mc.logger.Info(
		"HTTP request sent",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.String("status", resp.Status),
		zap.String("duration", duration.String()),
	)

	fmt.Println("Response:", resp.Status)

	return nil
}

func (mc *MonitoringClient) StartMonitoring(ctx context.Context) {
	tickerPoll := time.NewTicker(mc.Config.PollInterval)
	tickerReport := time.NewTicker(mc.Config.ReportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	doneCh := make(chan bool)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(doneCh)
				return
			case <-tickerPoll.C:
				_ = mc.CollectMetrics()
			case <-tickerReport.C:
				_ = mc.StartReporting()
			}
		}
	}()

	<-doneCh
	return
}

func (mc *MonitoringClient) CollectMetrics() error {
	mc.Storage.GaugeUpdate("RandomValue", rand.Float64())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	for metricName, metricFunc := range mc.MetricFuncMap {
		value := metricFunc(&m)
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
