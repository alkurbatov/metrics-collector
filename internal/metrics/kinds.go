package metrics

import (
	"strconv"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

type Metric interface {
	Kind() string
	String() string
}

type Counter int64

func (c Counter) Kind() string {
	return entity.Counter
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

type Gauge float64

func (g Gauge) Kind() string {
	return entity.Gauge
}

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}
