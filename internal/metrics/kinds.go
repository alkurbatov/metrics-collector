package metrics

import (
	"strconv"
)

type Metric interface {
	Kind() string
	String() string
}

type Counter int64

func (c Counter) Kind() string {
	return "counter"
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

type Gauge float64

func (g Gauge) Kind() string {
	return "gauge"
}

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', 3, 64)
}
