package app_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetAddressParsing(t *testing.T) {
	tt := []struct {
		name string
		src  string
		ok   bool
	}{
		{
			name: "All interfaces and port",
			src:  "0.0.0.0:8080",
			ok:   true,
		},
		{
			name: "Localhost and port",
			src:  "0.0.0.0:8080",
			ok:   true,
		},
		{
			name: "Some IP and port",
			src:  "10.20.0.10:3000",
			ok:   true,
		},
		{
			name: "Some hostname and port",
			src:  "collector:8888",
			ok:   true,
		},
		{
			name: "Missing delimiter",
			src:  "collector8888",
			ok:   false,
		},
		{
			name: "Missing port",
			src:  "collector",
			ok:   false,
		},
		{
			name: "Missing host or address",
			src:  "1000",
			ok:   false,
		},
		{
			name: "Empty string",
			src:  "",
			ok:   false,
		},
		{
			name: "Only delimiter",
			src:  ":",
			ok:   false,
		},
		{
			name: "Wrong format",
			src:  "xxx/100",
			ok:   false,
		},
		{
			name: "Invalid port value",
			src:  "localhost:1_",
			ok:   false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			addr := new(app.NetAddress)

			err := addr.Set(tc.src)

			if !tc.ok {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.src, addr.String())
		})
	}
}

func TestNetAddressTypeMatchesString(t *testing.T) {
	addr := app.NetAddress("0.0.0.0:8080")

	require.Equal(t, "string", addr.Type())
}
