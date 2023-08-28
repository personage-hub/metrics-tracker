package dumper

import (
	"encoding/json"
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal/storage"
	"os"
	"time"
)

type DumpFile struct {
	Path        string
	FileStorage struct {
		CounterData map[string]int64
		GaugeData   map[string]float64
	}
}

func NewDumper(path string) *DumpFile {
	return &DumpFile{
		Path: path,
	}
}

func (file *DumpFile) SaveData(s storage.Storage) error {
	file.FileStorage.GaugeData = s.GaugeMap()
	file.FileStorage.CounterData = s.CounterMap()
	data, err := json.MarshalIndent(file.FileStorage, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(file.Path, data, 0666); err != nil {
		return err
	}
	return nil
}

func (file *DumpFile) RestoreData(s storage.Storage) error {
	data, err := os.ReadFile(file.Path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &file.FileStorage); err != nil {
		return err
	}
	for metricName, metricValue := range file.FileStorage.GaugeData {
		s.GaugeUpdate(metricName, metricValue)
	}
	for metricName, metricValue := range file.FileStorage.CounterData {
		s.CounterUpdate(metricName, metricValue)
	}
	return nil
}

func PeriodicSave(dumper Dumper, storage storage.Storage, interval int64) error {
	for {
		if interval == 0 {
			return nil
		}

		time.Sleep(time.Duration(interval) * time.Second)

		if err := dumper.SaveData(storage); err != nil {
			return fmt.Errorf("failed to dump data: %w", err)
		}
	}
}
