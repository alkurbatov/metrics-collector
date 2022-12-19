package storage

import "errors"

type BrokenStorage struct {
	MemStorage
}

func (b *BrokenStorage) Push(key string, record Record) error {
	return errors.New("failure")
}
