package storage

import "errors"

type BrokenStorage struct {
}

func (b BrokenStorage) Push(key string, record Record) error {
	return errors.New("failure")
}

func (b BrokenStorage) Get(key string) (Record, bool) {
	return Record{}, false
}

func (b BrokenStorage) GetAll() []Record {
	return make([]Record, 0)
}
