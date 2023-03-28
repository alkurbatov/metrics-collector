package storage

import (
	"context"
	"sync"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

// MemStorage implements in-memory metrics storage.
type MemStorage struct {
	Data map[string]Record `json:"records"`
	sync.RWMutex
}

// NewMemStorage creates new instance of MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Data: make(map[string]Record),
	}
}

// Push records metric data.
func (m *MemStorage) Push(ctx context.Context, key string, record Record) error {
	m.Lock()
	defer m.Unlock()

	m.Data[key] = record

	return nil
}

// PushBatch records list of metrics data.
func (m *MemStorage) PushBatch(ctx context.Context, data map[string]Record) error {
	m.Lock()
	defer m.Unlock()

	for id, record := range data {
		m.Data[id] = record
	}

	return nil
}

// Get returns stored metrics record.
func (m *MemStorage) Get(ctx context.Context, key string) (Record, error) {
	m.RLock()
	defer m.RUnlock()

	record, ok := m.Data[key]
	if !ok {
		return Record{}, entity.ErrMetricNotFound
	}

	return record, nil
}

// GetAll returns all stored metrics.
func (m *MemStorage) GetAll(ctx context.Context) ([]Record, error) {
	m.RLock()
	defer m.RUnlock()

	rv := make([]Record, len(m.Data))
	i := 0

	for _, v := range m.Data {
		rv[i] = v
		i++
	}

	return rv, nil
}

// Close has no effect on in-memory storage.
func (m *MemStorage) Close(ctx context.Context) error {
	return nil // noop
}

// Snapshot creates independent copy of in-memory storage.
func (m *MemStorage) Snapshot() *MemStorage {
	m.RLock()
	defer m.RUnlock()

	snapshot := make(map[string]Record, len(m.Data))

	for k, v := range m.Data {
		snapshot[k] = v
	}

	return &MemStorage{Data: snapshot}
}
