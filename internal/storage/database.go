package storage

import (
	"context"
	"errors"
	"strconv"
)

type DatabaseStorage struct {
	db DBConn
}

func NewDatabaseStorage(db DBConn) *DatabaseStorage {
	return &DatabaseStorage{db: db}
}

func (d *DatabaseStorage) Push(key string, record Record) error {
	return errors.New("not implemented")
}

func (d *DatabaseStorage) Get(key string) (Record, bool) {
	return Record{}, false
}

func (d *DatabaseStorage) GetAll() []Record {
	return make([]Record, 0)
}

func (d *DatabaseStorage) String() string {
	cfg := d.db.Config()
	return "database storage at " + cfg.Host + ":" + strconv.FormatInt(int64(cfg.Port), 10) + "/" + cfg.Database
}

func (d *DatabaseStorage) Ping(ctx context.Context) error {
	return d.db.Ping(ctx)
}

func (d *DatabaseStorage) Close() error {
	return d.db.Close(context.Background())
}
