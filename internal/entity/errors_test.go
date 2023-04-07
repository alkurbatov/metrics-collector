package entity_test

import (
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
