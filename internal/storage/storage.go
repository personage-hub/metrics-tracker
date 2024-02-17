package storage

import (
	"github.com/personage-hub/metrics-tracker/internal/dumper"
	"go.uber.org/zap"
	"log"
	"time"
)

type Storage interface {
	GaugeUpdate(key string, value float64)
	CounterUpdate(key string, value int64)
	GaugeMap() map[string]float64
	CounterMap() map[string]int64
	GetGaugeMetric(metricName string) (float64, bool)
	GetCounterMetric(metricName string) (int64, bool)
	CheckKeeper() bool
}

func PeriodicSave(s Storage, keeper dumper.Dumper, saveInterval int64) {
	tickerSave := time.NewTicker(time.Duration(saveInterval) * time.Second)
	defer tickerSave.Stop()

	for range tickerSave.C {
		gaugeData := s.GaugeMap()
		counterData := s.CounterMap()
		err := keeper.SaveData(gaugeData, counterData)
		if err != nil {
			log.Fatal("fail saving data to dump", zap.Error(err))
		}
	}
}
