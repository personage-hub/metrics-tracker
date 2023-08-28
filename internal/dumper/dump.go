package dumper

import "github.com/personage-hub/metrics-tracker/internal/storage"

type Dumper interface {
	SaveData(s storage.Storage) error
	RestoreData(s storage.Storage) error
}
