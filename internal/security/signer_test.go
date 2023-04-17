package security_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/require"
)

const secret = "abc"

func TestCalculateSignature(t *testing.T) {
	tt := []struct {
		name       string
		metricName string
		data       metrics.Metric
		err        error
		expected   string
	}{
		{
			name:       "Sign counter record",
			metricName: "PollCount",
			data:       metrics.Counter(10),
			err:        nil,
			expected:   "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
		},
		{
			name:       "Sign gauge record",
			metricName: "Alloc",
			data:       metrics.Gauge(123.456),
			err:        nil,
			expected:   "63e1e3ffc75258f015fec0eca2d2fdbacc0e7df559be0adef92d43c2133c5cf7",
		},
		{
			name:       "Signing fails if counter type is unknown",
			metricName: "Alloc",
			data:       nil,
			err:        entity.ErrMetricNotImplemented,
			expected:   "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			signer := security.NewSigner(secret)

			hash, err := signer.CalculateRecordSignature(storage.Record{Name: tc.metricName, Value: tc.data})

			require.ErrorIs(err, tc.err)
			require.Equal(tc.expected, hash)
		})
	}
}

func TestVerifySignature(t *testing.T) {
	tt := []struct {
		name       string
		metricName string
		data       metrics.Metric
		hash       string
		valid      bool
		err        error
	}{
		{
			name:       "Verify signature of counter record",
			metricName: "PollCount",
			data:       metrics.Counter(10),
			hash:       "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
			valid:      true,
			err:        nil,
		},
		{
			name:       "Verify signature of gauge record",
			metricName: "Alloc",
			data:       metrics.Gauge(123.456),
			hash:       "63e1e3ffc75258f015fec0eca2d2fdbacc0e7df559be0adef92d43c2133c5cf7",
			valid:      true,
			err:        nil,
		},
		{
			name:       "Signature is invalid if hash is missing",
			metricName: "Alloc",
			data:       metrics.Gauge(123.456),
			hash:       "",
			valid:      false,
			err:        entity.ErrNotSigned,
		},
		{
			name:       "Signature is invalid if hash in unexpected encoding",
			metricName: "Alloc",
			data:       metrics.Gauge(123.456),
			hash:       "xxx",
			valid:      false,
			err:        nil,
		},
		{
			name:       "Signature is invalid if signer cannot calculate signature",
			metricName: "Alloc",
			data:       nil,
			hash:       "63e1e3ffc75258f015fec0eca2d2fdbacc0e7df559be0adef92d43c2133c5cf7",
			valid:      false,
			err:        entity.ErrMetricNotImplemented,
		},
		{
			name:       "Signature is invalid if counter hashes don't match",
			metricName: "PollCount",
			data:       metrics.Counter(11),
			hash:       "4141413a7878783a313233",
			err:        nil,
			valid:      false,
		},
		{
			name:       "Signature is invalid if gauge hashes don't match",
			metricName: "Alloc",
			data:       metrics.Gauge(123.456),
			hash:       "4141413a7878783a313233",
			err:        nil,
			valid:      false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			signer := security.NewSigner(secret)

			valid, err := signer.VerifyRecordSignature(storage.Record{Name: tc.metricName, Value: tc.data}, tc.hash)

			require.ErrorIs(err, tc.err)
			require.Equal(tc.valid, valid)
		})
	}
}
