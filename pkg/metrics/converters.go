package metrics

import (
	"fmt"
	"strconv"
)

// ToCounter creates new Counter metric object from string value.
func ToCounter(value string) (Counter, error) {
	rawValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert to counter: %w", err)
	}

	return Counter(rawValue), nil
}

// ToGauge creates new Gauge metric object from string value.
func ToGauge(value string) (Gauge, error) {
	rawValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert to gauge: %w", err)
	}

	return Gauge(rawValue), nil
}
