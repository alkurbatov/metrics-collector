package entity

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrBadAddressFormat        = errors.New("expected address in host:port form")
	ErrBadKeyFile              = errors.New("provided file doesn't contain key in the PEM format")
	ErrEncodingNotSupported    = errors.New("encoding type not supported")
	ErrHTTP                    = errors.New("HTTP request failed")
	ErrHealthCheckNotSupported = errors.New("storage doesn't support healthcheck")
	ErrIncompleteRequest       = errors.New("metrics value not set")
	ErrInvalidSignature        = errors.New("invalid signature")
	ErrMetricInvalidName       = errors.New("metric name contains invalid characters")
	ErrMetricLongName          = errors.New("metric name is too long")
	ErrMetricNotFound          = errors.New("metric not found")
	ErrMetricNotImplemented    = errors.New("metric kind not supported")
	ErrNotSigned               = errors.New("request not signed")
	ErrNotSupportedKey         = errors.New("provided key type is not supported")
	ErrRecordKindDontMatch     = errors.New("kind of recorded metric doesn't match request")
	ErrRestoreNoSource         = errors.New("state restoration was requested, but path to store file is not set")
	ErrTransportNotSupported   = errors.New("transport type not supported")
	ErrUnexpected              = errors.New("unexpected error")
	ErrUntrustedSource         = errors.New("request's source is not trusted")
)

func MetricNotImplementedError(kind string) error {
	return fmt.Errorf("%w (%s)", ErrMetricNotImplemented, kind)
}

func HTTPError(code int, reason []byte) error {
	return fmt.Errorf("%w (%d): %s", ErrHTTP, code, reason)
}

func EncodingNotSupportedError(name string) error {
	return fmt.Errorf("%w (%s)", ErrEncodingNotSupported, name)
}

func UntrustedSourceError(src net.IP) error {
	return fmt.Errorf("%w (%s)", ErrUntrustedSource, src.String())
}

func TransportNotSupportedError(name string) error {
	return fmt.Errorf("%w (%s)", ErrTransportNotSupported, name)
}
