package dumper

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type DumpFile struct {
	Path string
}

type FileStorage struct {
	CounterData map[string]int64
	GaugeData   map[string]float64
}

func NewDumper(path string) *DumpFile {
	return &DumpFile{
		Path: path,
	}
}

func (file *DumpFile) SaveData(gaugeMap map[string]float64, counterMap map[string]int64) error {
	fileStorage := FileStorage{
		CounterData: counterMap,
		GaugeData:   gaugeMap,
	}

	data, err := json.MarshalIndent(fileStorage, "", "  ")
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
	fileStorage := FileStorage{}
	if err := json.Unmarshal(data, &fileStorage); err != nil {
		return nil, nil, err
	}
	return fileStorage.GaugeData, fileStorage.CounterData, nil
}

func (file *DumpFile) CheckHealth() bool {
	var stat syscall.Statfs_t

	err := syscall.Statfs(file.Path, &stat)
	if err != nil {
		return false
	}

	all := stat.Blocks

	free := stat.Bfree

	usagePercentage := (1 - float64(free)/float64(all)) * 100

	return usagePercentage < 90
}
