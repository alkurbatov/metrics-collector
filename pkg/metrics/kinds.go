package metrics

import (
	"strconv"
)

const (
	KindCounter = "counter"
	KindGauge   = "gauge"
)

var _ Metric = (*Counter)(nil)
var _ Metric = (*Gauge)(nil)

// A Metric is common representation of all supported metrics kinds.
type Metric interface {
	// Kind return kind of this metric, e.g. "counter".
	Kind() string

	// String provides string representation of metrics value.
	String() string
}

// Counter represents always growing numeric metric, e.g. count of HTTP requests.
type Counter int64

func (c Counter) Kind() string {
	return KindCounter
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

// Gauge represents numeric metric value which could go up or down,
// e.g. count of currently allocated virtual memory.
type Gauge float64

func (g Gauge) Kind() string {
	return KindGauge
}

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}
