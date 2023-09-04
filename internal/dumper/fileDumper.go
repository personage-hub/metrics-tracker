package dumper

import (
	"encoding/json"
	"fmt"
	"os"
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

func (file *DumpFile) SaveData(gaugeMap map[string]float64, counterMap map[string]int64) error {
	file.FileStorage.GaugeData = gaugeMap
	file.FileStorage.CounterData = counterMap
	data, err := json.MarshalIndent(file.FileStorage, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(file.Path, data, 0666); err != nil {
		return err
	}
	return nil
}

func (file *DumpFile) RestoreData() (map[string]float64, map[string]int64, error) {

	if _, err := os.Stat(file.Path); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("file does not exist, skipping restore. %w", err)
	}
	data, err := os.ReadFile(file.Path)
	if err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(data, &file.FileStorage); err != nil {
		return nil, nil, err
	}
	return file.FileStorage.GaugeData, file.FileStorage.CounterData, nil
}
