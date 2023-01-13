package metrics

import (
	"fmt"
	"strconv"
)

func ToCounter(value string) (Counter, error) {
	rawValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert to counter: %w", err)
	}

	return Counter(rawValue), nil
}

func ToGauge(value string) (Gauge, error) {
	rawValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert to gauge: %w", err)
	}

	return Gauge(rawValue), nil
}
