package entity_test

import (
	"net"
	"net/http"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMetricNotImplementedError(t *testing.T) {
	err := entity.MetricNotImplementedError("summary")

	require.ErrorIs(t, err, entity.ErrMetricNotImplemented)
	snaps.MatchSnapshot(t, err.Error())
}

func TestHTTPError(t *testing.T) {
	err := entity.HTTPError(http.StatusInternalServerError, []byte("Something bad has happened"))

	require.ErrorIs(t, err, entity.ErrHTTP)
	snaps.MatchSnapshot(t, err.Error())
}

func TestEncodingNotSupportedError(t *testing.T) {
	err := entity.EncodingNotSupportedError("deflate")

	require.ErrorIs(t, err, entity.ErrEncodingNotSupported)
	snaps.MatchSnapshot(t, err.Error())
}

func TestTestUntrustedSourceError(t *testing.T) {
	tt := []struct {
		name     string
		sourceIP string
	}{
		{
			name: "Empty IP",
		},
		{
			name:     "IPv4 value",
			sourceIP: "192.168.10.12",
		},
		{
			name:     "IPv6 value",
			sourceIP: "2001:470:28:30d:21e:67ff:fe7a:e1da",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := entity.UntrustedSourceError(net.ParseIP(tc.sourceIP))

			require.ErrorIs(t, err, entity.ErrUntrustedSource)
			snaps.MatchSnapshot(t, err.Error())
		})
	}
}
