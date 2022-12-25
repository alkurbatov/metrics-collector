package handlers

import "errors"

var (
	errMetricNotFound       = errors.New("metric not found")
	errMetricNotImplemented = errors.New("support of metric type is not implemented")
	errRecordKindDontMatch  = errors.New("kind of recorded metric doesn't match request")
	errInvalidSignature     = errors.New("invalid signature")
)
