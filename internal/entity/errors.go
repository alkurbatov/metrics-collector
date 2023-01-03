package entity

import "errors"

var (
	ErrHealthCheckNotSupported = errors.New("storage doesn't support healthcheck")
	ErrIncompleteRequest       = errors.New("metrics value not set")
	ErrInvalidSignature        = errors.New("invalid signature")
	ErrMetricInvalidName       = errors.New("metric name contains invalid characters")
	ErrMetricLongName          = errors.New("metric name is too long")
	ErrMetricNotFound          = errors.New("metric not found")
	ErrNotSigned               = errors.New("request not signed")
	ErrRecordKindDontMatch     = errors.New("kind of recorded metric doesn't match request")
	ErrRestoreNoSource         = errors.New("state restoration was requested, but path to store file is not set")
	ErrUnexpected              = errors.New("unexpected error")
)

type MetricNotImplementedError struct {
	Kind string
}

func (e *MetricNotImplementedError) Error() string {
	return "support of metric type " + e.Kind + " is not implemented"
}
