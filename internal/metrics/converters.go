package metrics

import (
	"strconv"
)

func ToCounter(value string) (Counter, error) {
	rawValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return Counter(rawValue), nil
}

func ToGauge(value string) (Gauge, error) {
	rawValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return Gauge(rawValue), nil
}
