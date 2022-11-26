package metrics

import "fmt"

type Metric interface {
	Kind() string
	String() string
}

type Counter int64

func (c Counter) Kind() string {
	return "counter"
}

func (c Counter) String() string {
	return fmt.Sprintf("%d", c)
}

type Gauge float64

func (g Gauge) Kind() string {
	return "gauge"
}

func (g Gauge) String() string {
	return fmt.Sprintf("%.3f", g)
}
