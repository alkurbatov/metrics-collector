package handlers

func parseUpdateCounterReq(parts ...string) (*UpdateCounterReq, error) {
	name := parts[1]
	if err := validateMetricName(name); err != nil {
		return nil, err
	}

	value, err := toCounter(parts[2])
	if err != nil {
		return nil, err
	}

	return &UpdateCounterReq{
		name:  name,
		value: value,
	}, nil
}

func parseUpdateGaugeReq(parts ...string) (*UpdateGaugeReq, error) {
	name := parts[1]
	if err := validateMetricName(name); err != nil {
		return nil, err
	}

	value, err := toGauge(parts[2])
	if err != nil {
		return nil, err
	}

	return &UpdateGaugeReq{
		name:  name,
		value: value,
	}, nil
}
