package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

type DBConnPoolMock struct {
	mock.Mock
}

func NewDBConnPoolMock() *DBConnPoolMock {
	return new(DBConnPoolMock)
}

func (m *DBConnPoolMock) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).(*pgxpool.Conn), args.Error(1)
}

func (m *DBConnPoolMock) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *DBConnPoolMock) Close() {
	_ = m.Called()
}
