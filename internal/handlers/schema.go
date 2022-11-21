package handlers

import "github.com/alkurbatov/metrics-collector/internal/metrics"

type UpdateCounterReq struct {
	name  string
	value metrics.Counter
}

type UpdateGaugeReq struct {
	name  string
	value metrics.Gauge
}
