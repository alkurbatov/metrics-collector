package services

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/mock"
)

var _ Recorder = (*RecorderMock)(nil)

type RecorderMock struct {
	mock.Mock
}

func (m *RecorderMock) Push(ctx context.Context, record storage.Record) (storage.Record, error) {
	args := m.Called(ctx, record)
	return args.Get(0).(storage.Record), args.Error(1)
}

func (m *RecorderMock) PushList(ctx context.Context, records []storage.Record) ([]storage.Record, error) {
	args := m.Called(ctx, records)
	return args.Get(0).([]storage.Record), args.Error(1)
}

func (m *RecorderMock) Get(ctx context.Context, kind, name string) (storage.Record, error) {
	args := m.Called(ctx, kind, name)
	return args.Get(0).(storage.Record), args.Error(1)
}

func (m *RecorderMock) List(ctx context.Context) ([]storage.Record, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.Record), args.Error(1)
}
