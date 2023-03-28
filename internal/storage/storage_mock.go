package storage

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) Push(ctx context.Context, key string, record Record) error {
	args := m.Called(ctx, key, record)
	return args.Error(0)
}

func (m *Mock) PushBatch(ctx context.Context, data map[string]Record) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *Mock) Get(ctx context.Context, key string) (Record, error) {
	args := m.Called(ctx, key)

	return args.Get(0).(Record), args.Error(1)
}

func (m *Mock) GetAll(ctx context.Context) ([]Record, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Record), args.Error(1)
}

func (m *Mock) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
