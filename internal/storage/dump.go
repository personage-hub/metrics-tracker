package storage

type Dumper interface {
	SaveData(s MemStorage) error
	RestoreData(s *MemStorage) error
}
