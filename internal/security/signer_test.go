package security_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/stretchr/testify/require"
)

const secret = "abc"

func TestSignRequest(t *testing.T) {
	tt := []struct {
		name     string
		req      schema.MetricReq
		ok       bool
		expected string
	}{
		{
			name:     "Should sign update counter request",
			req:      schema.NewUpdateCounterReq("PollCount", 10),
			ok:       true,
			expected: "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
		},
		{
			name:     "Should sign update gauge request",
			req:      schema.NewUpdateGaugeReq("Alloc", 123.456),
			ok:       true,
			expected: "63e1e3ffc75258f015fec0eca2d2fdbacc0e7df559be0adef92d43c2133c5cf7",
		},
		{
			name: "Should fail on missing delta",
			req:  schema.MetricReq{ID: "PollCount", MType: entity.Counter},
		},
		{
			name: "Should fail on missing value",
			req:  schema.MetricReq{ID: "Alloc", MType: entity.Gauge},
		},
		{
			name: "Should fail on unexpected metric type",
			req:  schema.MetricReq{ID: "Alloc", MType: "???"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			signer := security.NewSigner(secret)

			err := signer.SignRequest(&tc.req)
			if !tc.ok {
				require.Error(err)
				return
			}

			require.NoError(err)
			require.Equal(tc.expected, tc.req.Hash)
		})
	}
}

func TestVerifySignature(t *testing.T) {
	tt := []struct {
		name  string
		req   schema.MetricReq
		hash  string
		ok    bool
		valid bool
	}{
		{
			name:  "Should verify update counter request",
			req:   schema.NewUpdateCounterReq("PollCount", 10),
			hash:  "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
			ok:    true,
			valid: true,
		},
		{
			name:  "Should sign update gauge request",
			req:   schema.NewUpdateGaugeReq("Alloc", 123.456),
			hash:  "63e1e3ffc75258f015fec0eca2d2fdbacc0e7df559be0adef92d43c2133c5cf7",
			ok:    true,
			valid: true,
		},
		{
			name: "Should be invalid if hash is missing",
			req:  schema.NewUpdateGaugeReq("Alloc", 123.456),
			hash: "",
		},
		{
			name: "Should be invalid if signature is in unexpected encoding",
			req:  schema.MetricReq{ID: "Alloc", MType: "???"},
			hash: "xxx",
		},
		{
			name: "Should be invalid if cannot calculate signature",
			req:  schema.MetricReq{ID: "Alloc", MType: "???"},
			hash: "63e1e3ffc75258f015fec0eca2d2fdbacc0e7df559be0adef92d43c2133c5cf7",
		},
		{
			name: "Should be invalid if counter hashes doesn't match",
			req:  schema.NewUpdateCounterReq("PollCount", 11),
			hash: "4141413a7878783a313233",
			ok:   true,
		},
		{
			name: "Should be invalid if gauge hashes doesn't match",
			req:  schema.NewUpdateGaugeReq("Alloc", 123.454),
			hash: "4141413a7878783a313233",
			ok:   true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			signer := security.NewSigner(secret)

			tc.req.Hash = tc.hash

			valid, err := signer.VerifySignature(&tc.req)
			if !tc.ok {
				require.Error(err)
				require.False(valid)

				return
			}

			require.NoError(err)
			require.Equal(tc.valid, valid)
		})
	}
}
