package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

var _ Storage = (*FileBackedStorage)(nil)

func dumpError(reason error) error {
	return fmt.Errorf("data dump failed: %w", reason)
}

func restoreError(reason error) error {
	return fmt.Errorf("data restore failed: %w", reason)
}

// FileBackedStorage implements in-memory metrics storage with ability to
// dump/restore metrics data to/from disk.
type FileBackedStorage struct {
	*MemStorage
	sync.Mutex

	// Path to backing file.
	storePath string

	// Dump recorded data on each push of a record.
	syncMode bool
}

// NewFileBackedStorage creates new instance of FileBackedStorage.
func NewFileBackedStorage(storePath string, syncMode bool) *FileBackedStorage {
	return &FileBackedStorage{
		MemStorage: NewMemStorage(),
		storePath:  storePath,
		syncMode:   syncMode,
	}
}

// Push records metric data.
func (f *FileBackedStorage) Push(ctx context.Context, key string, record Record) error {
	if err := f.MemStorage.Push(ctx, key, record); err != nil {
		return err
	}

	if f.syncMode {
		return f.Dump(ctx)
	}

	return nil
}

// PushBatch records list of metrics data.
func (f *FileBackedStorage) PushBatch(ctx context.Context, data map[string]Record) error {
	if err := f.MemStorage.PushBatch(ctx, data); err != nil {
		return err
	}

	if f.syncMode {
		return f.Dump(ctx)
	}

	return nil
}

// Close dumps all stored data to disk. The storage can be restored from this dump later.
func (f *FileBackedStorage) Close(ctx context.Context) error {
	return f.Dump(ctx)
}

// Restore reads previously stored data from disk and populates the storage.
func (f *FileBackedStorage) Restore() (err error) {
	f.Lock()
	defer f.Unlock()

	log.Info().Msg("Restoring storage data from " + f.storePath)

	file, err := os.Open(f.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn().Msg("No storage dump found, data restoration is not possible")
			return nil
		}

		return restoreError(err)
	}

	defer func() {
		if dErr := file.Close(); err == nil && dErr != nil {
			err = restoreError(dErr)
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(f.MemStorage); err != nil {
		return restoreError(err)
	}

	log.Info().Msg("Storage data was successfully restored")

	return nil
}

// Dump writes all stored data to disk. The storage can be restored from this dump later.
func (f *FileBackedStorage) Dump(ctx context.Context) (err error) {
	f.Lock()
	defer f.Unlock()

	log.Ctx(ctx).Info().Msg("Pushing storage data to " + f.storePath)

	file, err := os.OpenFile(f.storePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return dumpError(err)
	}

	defer func() {
		if dErr := file.Close(); err == nil && dErr != nil {
			err = dumpError(dErr)
		}
	}()

	encoder := json.NewEncoder(file)
	snapshot := f.Snapshot()

	if err := encoder.Encode(snapshot); err != nil {
		return dumpError(err)
	}

	return nil
}
