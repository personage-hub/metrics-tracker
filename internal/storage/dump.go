package storage

type Dumper interface {
	SaveData(s Storage) error
	RestoreData(s Storage) error
}
