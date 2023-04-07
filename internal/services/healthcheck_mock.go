package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ HealthCheck = (*HealthCheckMock)(nil)

type HealthCheckMock struct {
	mock.Mock
}

func (m *HealthCheckMock) CheckStorage(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
