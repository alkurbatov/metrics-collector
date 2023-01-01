package storage

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

type BrokenStorage struct {
	MemStorage
}

func (b *BrokenStorage) Push(ctx context.Context, key string, record Record) error {
	return entity.ErrUnexpected
}
