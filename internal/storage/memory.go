package storage

import (
	"context"
	"sync"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

type MemStorage struct {
	Data map[string]Record `json:"records"`
	sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Data: make(map[string]Record),
	}
}

func (m *MemStorage) Push(ctx context.Context, key string, record Record) error {
	m.Lock()
	defer m.Unlock()

	m.Data[key] = record

	return nil
}

func (m *MemStorage) PushList(ctx context.Context, keys []string, records []Record) error {
	m.Lock()
	defer m.Unlock()

	for i := range records {
		m.Data[keys[i]] = records[i]
	}

	return nil
}

func (m *MemStorage) Get(ctx context.Context, key string) (Record, error) {
	m.RLock()
	defer m.RUnlock()

	record, ok := m.Data[key]
	if !ok {
		return Record{}, entity.ErrMetricNotFound
	}

	return record, nil
}

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

func (m *MemStorage) Close() error {
	return nil // noop
}

func (m *MemStorage) Snapshot() *MemStorage {
	m.RLock()
	defer m.RUnlock()

	snapshot := make(map[string]Record, len(m.Data))

	for k, v := range m.Data {
		snapshot[k] = v
	}

	return &MemStorage{Data: snapshot}
}
