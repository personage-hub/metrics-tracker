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
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type metric struct {
	metricValue float64
	metricType  string
}

type MonitoringClient struct {
	Client        *http.Client
	Config        Config
	Storage       storage.Storage
	metricStorage chan map[string]metric
	logger        *zap.Logger
}

func NewMonitoringClient(client *http.Client, logger *zap.Logger, config Config) *MonitoringClient {
	return &MonitoringClient{
		Client:        client,
		Config:        config,
		metricStorage: make(chan map[string]metric, 1),
		logger:        logger,
	}
}

func (mc *MonitoringClient) compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gzw, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	if _, err = gzw.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write compress data to buffer: %w", err)
	}

	if err = gzw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compress writer: %w", err)
	}

	return b.Bytes(), nil
}

func (mc *MonitoringClient) SendMetric(metric metrics.Metrics) error {
	url := fmt.Sprintf("http://%s/update/", mc.Config.ServerAddress)

	data, err := easyjson.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed converting data for request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed creating request: %w", err)
	}

	req.Header.Set("Content-Type", consts.ContentTypeJSON)

	start := time.Now()

	resp, err := mc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed send metric: %w", err)
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

	mc.logger.Info("Response:", zap.Int("status", resp.StatusCode))

	return nil
}

func (mc *MonitoringClient) StartMonitoring(ctx context.Context) {
	tickerPoll := time.NewTicker(mc.Config.PollInterval)
	tickerReport := time.NewTicker(mc.Config.ReportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tickerPoll.C:
			mc.CollectMetrics()
		case <-tickerReport.C:
			mc.StartReporting()
		}
	}
}

func (mc *MonitoringClient) CollectMetrics() {
	mc.logger.Info("starting collecting metrics")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var metStorage map[string]metric

	if len(mc.metricStorage) == 0 {
		metStorage = make(map[string]metric)
	} else {
		metStorage = <-mc.metricStorage
	}

	metStorage["Alloc"] = metric{metricValue: float64(m.Alloc), metricType: "gauge"}
	metStorage["BuckHashSys"] = metric{metricValue: float64(m.BuckHashSys), metricType: "gauge"}
	metStorage["Frees"] = metric{metricValue: float64(m.Frees), metricType: "gauge"}
	metStorage["GCCPUFraction"] = metric{metricValue: m.GCCPUFraction, metricType: "gauge"}
	metStorage["GCSys"] = metric{metricValue: float64(m.GCSys), metricType: "gauge"}
	metStorage["HeapAlloc"] = metric{metricValue: float64(m.HeapAlloc), metricType: "gauge"}
	metStorage["HeapIdle"] = metric{metricValue: float64(m.HeapIdle), metricType: "gauge"}
	metStorage["HeapInuse"] = metric{metricValue: float64(m.HeapInuse), metricType: "gauge"}
	metStorage["HeapObjects"] = metric{metricValue: float64(m.HeapObjects), metricType: "gauge"}
	metStorage["HeapReleased"] = metric{metricValue: float64(m.HeapReleased), metricType: "gauge"}
	metStorage["HeapSys"] = metric{metricValue: float64(m.HeapSys), metricType: "gauge"}
	metStorage["LastGC"] = metric{metricValue: float64(m.LastGC), metricType: "gauge"}
	metStorage["Lookups"] = metric{metricValue: float64(m.Lookups), metricType: "gauge"}
	metStorage["MCacheInuse"] = metric{metricValue: float64(m.MCacheInuse), metricType: "gauge"}
	metStorage["MCacheSys"] = metric{metricValue: float64(m.MCacheSys), metricType: "gauge"}
	metStorage["MSpanInuse"] = metric{metricValue: float64(m.MSpanInuse), metricType: "gauge"}
	metStorage["MSpanSys"] = metric{metricValue: float64(m.MSpanSys), metricType: "gauge"}
	metStorage["Mallocs"] = metric{metricValue: float64(m.Mallocs), metricType: "gauge"}
	metStorage["NextGC"] = metric{metricValue: float64(m.NextGC), metricType: "gauge"}
	metStorage["NumForcedGC"] = metric{metricValue: float64(m.NumForcedGC), metricType: "gauge"}
	metStorage["NumGC"] = metric{metricValue: float64(m.NumGC), metricType: "gauge"}
	metStorage["OtherSys"] = metric{metricValue: float64(m.OtherSys), metricType: "gauge"}
	metStorage["PauseTotalNs"] = metric{metricValue: float64(m.PauseTotalNs), metricType: "gauge"}
	metStorage["StackInuse"] = metric{metricValue: float64(m.StackInuse), metricType: "gauge"}
	metStorage["StackSys"] = metric{metricValue: float64(m.StackSys), metricType: "gauge"}
	metStorage["Sys"] = metric{metricValue: float64(m.Sys), metricType: "gauge"}
	metStorage["TotalAlloc"] = metric{metricValue: float64(m.TotalAlloc), metricType: "gauge"}
	metStorage["RandomValue"] = metric{metricValue: rand.Float64(), metricType: "gauge"}

	pollCount, ok := metStorage["PollCount"]
	if !ok {
		metStorage["PollCount"] = metric{metricValue: float64(1), metricType: "counter"}
	} else {
		pollCount.metricValue += float64(1)
		metStorage["PollCount"] = pollCount
	}

	mc.metricStorage <- metStorage
	mc.logger.Info("finish collecting metrics")
}

func (mc *MonitoringClient) StartReporting() {
	mc.logger.Info("Reporting metrics to server...")
	metStorage := <-mc.metricStorage
	for metricName, currentMetric := range metStorage {
		m := metrics.Metrics{
			ID:    metricName,
			MType: currentMetric.metricType,
		}
		switch currentMetric.metricType {
		case "gauge":
			{
				m.Value = &currentMetric.metricValue
			}
		case "counter":
			{
				v := int64(currentMetric.metricValue)
				m.Delta = &v
			}

		}
		err := mc.SendMetric(m)
		if err != nil {
			mc.logger.Error("error sending gauge metric: %w", zap.Error(err))
		}
	}
}
