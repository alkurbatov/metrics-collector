package storage

import (
	"time"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type Storage interface {
	Push(key string, record Record) error
	Get(key string) (Record, bool)
	GetAll() []Record
}

func NewDataStore(path string, storeInterval time.Duration) Storage {
	if len(path) == 0 {
		logging.Log.Debug("Attached in-memory storage")
		return NewMemStorage()
	}

	logging.Log.Debug("Attached file-backed memory storage at " + path)
	return NewFileBackedStorage(path, storeInterval == 0)
}
