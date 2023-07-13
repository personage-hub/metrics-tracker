package internal

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (m *MemStorage) GaugeUpdate(key string, value float64) {
	if m.gauge == nil {
		m.gauge = make(map[string]float64)
	}

	if _, ok := m.gauge[key]; ok {
		m.gauge[key] = value
	} else {
		m.gauge[key] = value
	}
}

func (m *MemStorage) CounterUpdate(key string, value int64) {
	if m.counter == nil {
		m.counter = make(map[string]int64)
	}
	if oldValue, ok := m.counter[key]; ok {
		m.counter[key] = oldValue + value
	} else {
		m.counter[key] = value
	}
}

func (m *MemStorage) GaugeMap() map[string]float64 {
	return m.gauge
}

func (m *MemStorage) CounterMap() map[string]int64 {
	return m.counter
}

func (m *MemStorage) GetGaugeMetric(metricName string) float64 {
	return m.gauge[metricName]
}

func (m *MemStorage) GetCounterMetric(metricName string) int64 {
	return m.counter[metricName]
}
