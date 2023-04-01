package entity_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestFilePathParsing(t *testing.T) {
	tt := []struct {
		name string
		src  string
		ok   bool
	}{
		{
			name: "Path to existing file",
			src:  "../../go.sum",
			ok:   true,
		},
		{
			name: "Path to not existing file",
			src:  "./xxx.test",
			ok:   false,
		},
		{
			name: "Empty string",
			src:  "",
			ok:   false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			fp := new(entity.FilePath)

			err := fp.Set(tc.src)

			if !tc.ok {
				require.Error(err)
				return
			}

			require.NoError(err)
			require.Equal(tc.src, fp.String())
		})
	}
}

func TestFilePathTypeMatchesString(t *testing.T) {
	fp := entity.FilePath("../../go.sum")

	require.Equal(t, "string", fp.Type())
}
