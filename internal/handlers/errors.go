package handlers

import "errors"

var errMetricNotFound = errors.New("metric not found")
var errMetricNotImplemented = errors.New("support of metric type is not implemented")
