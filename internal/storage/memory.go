package storage

import "sync"

type MemStorage struct {
	Data map[string]Record `json:"records"`
	sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Data: make(map[string]Record),
	}
}

func (m *MemStorage) Push(key string, record Record) error {
	m.Lock()
	defer m.Unlock()

	m.Data[key] = record
	return nil
}

func (m *MemStorage) Get(key string) (Record, bool) {
	m.RLock()
	defer m.RUnlock()

	record, ok := m.Data[key]
	return record, ok
}

func (m *MemStorage) GetAll() []Record {
	m.RLock()
	defer m.RUnlock()

	rv := make([]Record, len(m.Data))

	i := 0
	for _, v := range m.Data {
		rv[i] = v
		i++
	}

	return rv
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
