package storage

import (
	"time"
)

type Storage interface {
	Push(key string, record Record) error
	Get(key string) (Record, bool)
	GetAll() []Record
	String() string
}

func NewDataStore(path string, storeInterval time.Duration) Storage {
	if len(path) == 0 {
		return NewMemStorage()
	}

	return NewFileBackedStorage(path, storeInterval == 0)
}
