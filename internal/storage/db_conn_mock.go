package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type DBConnMock struct {
	mock.Mock
}

func NewDBConnMock() *DBConnMock {
	return new(DBConnMock)
}

func (m *DBConnMock) Config() *pgx.ConnConfig {
	args := m.Called()
	return args.Get(0).(*pgx.ConnConfig)
}

func (m *DBConnMock) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *DBConnMock) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
