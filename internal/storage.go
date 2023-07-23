package internal

import (
	"github.com/cornelk/hashmap"
)

type MemStorage struct {
	gauge   *hashmap.Map[string, float64]
	counter *hashmap.Map[string, int64]
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   hashmap.New[string, float64](),
		counter: hashmap.New[string, int64](),
	}
}

func (m *MemStorage) GaugeUpdate(key string, value float64) {
	m.gauge.GetOrInsert(key, value)
}

func (m *MemStorage) CounterUpdate(key string, value int64) {
	current, ok := m.counter.Get(key)
	if !ok {
		m.counter.Set(key, value)
	} else {
		m.counter.Set(key, current+value)
	}
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
