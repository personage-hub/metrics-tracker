package storage

import (
	"fmt"
	"github.com/cornelk/hashmap"
	"github.com/personage-hub/metrics-tracker/internal/dumper"
	"go.uber.org/zap"
	"log"
	"time"
)

type Storage interface {
	GaugeUpdate(key string, value float64) error
	CounterUpdate(key string, value int64) error
	GaugeMap() map[string]float64
	CounterMap() map[string]int64
	GetGaugeMetric(metricName string) (float64, bool)
	GetCounterMetric(metricName string) (int64, bool)
	PeriodicSave(saveInterval int64)
}

type MemStorage struct {
	gauge    *hashmap.Map[string, float64]
	counter  *hashmap.Map[string, int64]
	syncSave bool
	keeper   dumper.Dumper
}

func NewMemStorage(k dumper.Dumper, restore bool, syncSave bool) (*MemStorage, error) {
	m := &MemStorage{
		gauge:    hashmap.New[string, float64](),
		counter:  hashmap.New[string, int64](),
		syncSave: syncSave,
		keeper:   k,
	}
	if !restore {
		return m, nil
	}
	gaugeData, counterData, err := m.keeper.RestoreData()
	if err != nil {
		fmt.Printf("skipping restore")
		return m, nil
	}
	for metricName, metricValue := range gaugeData {
		_ = m.GaugeUpdate(metricName, metricValue)
	}
	for metricName, metricValue := range counterData {
		_ = m.CounterUpdate(metricName, metricValue)
	}
	return m, nil
}

func (m *MemStorage) GaugeUpdate(key string, value float64) error {
	m.gauge.Set(key, value)
	if m.syncSave {
		gaugeData := m.GaugeMap()
		counterData := m.CounterMap()
		err := m.keeper.SaveData(gaugeData, counterData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MemStorage) CounterUpdate(key string, value int64) error {
	current, ok := m.counter.Get(key)
	if !ok {
		m.counter.Set(key, value)
	} else {
		m.counter.Set(key, current+value)
	}
	if m.syncSave {
		gaugeData := m.GaugeMap()
		counterData := m.CounterMap()
		err := m.keeper.SaveData(gaugeData, counterData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MemStorage) GaugeMap() map[string]float64 {
	resultMap := make(map[string]float64)
	m.gauge.Range(func(key string, value float64) bool {
		resultMap[key] = value
		return true
	})
	return resultMap
}

func (m *MemStorage) CounterMap() map[string]int64 {
	resultMap := make(map[string]int64)
	m.counter.Range(func(key string, value int64) bool {
		resultMap[key] = value
		return true
	})
	return resultMap
}

func (m *MemStorage) GetGaugeMetric(metricName string) (float64, bool) {
	value, ok := m.gauge.Get(metricName)
	if ok {
		return value, true
	}
	return 0, false
}

func (m *MemStorage) GetCounterMetric(metricName string) (int64, bool) {
	value, ok := m.counter.Get(metricName)
	if ok {
		return value, true
	}
	return 0, false
}

func (m *MemStorage) PeriodicSave(saveInterval int64) {
	tickerSave := time.NewTicker(time.Duration(saveInterval) * time.Second)
	defer tickerSave.Stop()
	for {
		select {
		case <-tickerSave.C:
			gaugeData := m.GaugeMap()
			counterData := m.CounterMap()
			err := m.keeper.SaveData(gaugeData, counterData)
			if err != nil {
				if err != nil {
					log.Fatal("fail saving data to dump", zap.Error(err))
				}
			}
		}
	}
}
