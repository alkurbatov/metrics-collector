package services

import "github.com/alkurbatov/metrics-collector/internal/storage"

type Recorder interface {
	PushCounter(name, rawValue string) error
	PushGauge(name, rawValue string) error
	GetRecord(kind, name string) (storage.Record, bool)
	ListRecords() []storage.Record
}
