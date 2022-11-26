package storage

import "github.com/alkurbatov/metrics-collector/internal/metrics"

type Record struct {
	Name  string
	Value metrics.Metric
}
