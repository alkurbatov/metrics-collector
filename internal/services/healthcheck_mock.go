package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type HealthCheckMock struct {
	mock.Mock
}

func (m *HealthCheckMock) CheckStorage(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
