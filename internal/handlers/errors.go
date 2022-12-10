package handlers

import "errors"

var errMetricNotFound = errors.New("metric not found")
var errMetricNotImplemented = errors.New("support of metric type is not implemented")
var errRecordKindDontMatch = errors.New("kind of recorded metric doesn't match request")
