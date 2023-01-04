package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type FileBackedStorage struct {
	*MemStorage
	sync.Mutex

	// Path to backing file.
	storePath string

	// Dump recorded data on each push of a record.
	syncMode bool
}

func NewFileBackedStorage(storePath string, syncMode bool) *FileBackedStorage {
	return &FileBackedStorage{
		MemStorage: NewMemStorage(),
		storePath:  storePath,
		syncMode:   syncMode,
	}
}

func (f *FileBackedStorage) Push(ctx context.Context, key string, record Record) error {
	if err := f.MemStorage.Push(ctx, key, record); err != nil {
		return err
	}

	if f.syncMode {
		return f.Dump()
	}

	return nil
}

func (f *FileBackedStorage) PushList(ctx context.Context, keys []string, records []Record) error {
	if err := f.MemStorage.PushList(ctx, keys, records); err != nil {
		return err
	}

	if f.syncMode {
		return f.Dump()
	}

	return nil
}

func (f *FileBackedStorage) Close() error {
	return f.Dump()
}

func (f *FileBackedStorage) Restore() error {
	f.Lock()
	defer f.Unlock()

	logging.Log.Info("Restoring storage data from " + f.storePath)

	file, err := os.Open(f.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Log.Warning("No storage dump found, data restoration is not possible")
			return nil
		}

		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(f.MemStorage); err != nil {
		return err
	}

	logging.Log.Info("Storage data was successfully restored")

	return nil
}

func (f *FileBackedStorage) Dump() error {
	f.Lock()
	defer f.Unlock()

	logging.Log.Info("Pushing storage data to " + f.storePath)

	file, err := os.OpenFile(f.storePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	snapshot := f.MemStorage.Snapshot()

	if err := encoder.Encode(snapshot); err != nil {
		return err
	}

	return nil
}
