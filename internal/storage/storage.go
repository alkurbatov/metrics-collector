package storage

type Storage interface {
	Push(key string, record Record)
	Get(key string) (Record, bool)
	GetAll() []Record
}
