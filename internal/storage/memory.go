package storage

type MemStorage struct {
	data map[string]Record
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]Record),
	}
}

func (m *MemStorage) Push(key string, record Record) {
	m.data[key] = record
}

func (m *MemStorage) Get(key string) (Record, bool) {
	record, ok := m.data[key]
	return record, ok
}

func (m *MemStorage) GetAll() []Record {
	rv := make([]Record, len(m.data))

	i := 0
	for _, v := range m.data {
		rv[i] = v
		i++
	}

	return rv
}
