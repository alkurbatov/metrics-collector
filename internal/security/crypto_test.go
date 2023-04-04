package security_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptMessage(t *testing.T) {
	tt := []struct {
		name string
		msg  string
	}{
		{
			name: "Encrypt and decrypt short meessage",
			msg:  "This is test message",
		},
		{
			name: "Encrypt and decrypt long meessage",
			msg:  strings.Repeat("12345", 800),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			publicKey, err := security.NewPublicKey("../../build/keys/public.pem")
			require.NoError(err)

			privateKey, err := security.NewPrivateKey("../../build/keys/private.pem")
			require.NoError(err)

			input := bytes.NewBufferString(tc.msg)
			encrypted, err := security.Encrypt(io.Reader(input), publicKey)
			require.NoError(err)

			decrypted, err := security.Decrypt(io.Reader(encrypted), privateKey)
			require.NoError(err)

			require.Equal(tc.msg, decrypted.String())
		})
	}
}

func TestLoadInvalidPublicKey(t *testing.T) {
	_, err := security.NewPublicKey("xxx")
	require.Error(t, err)
}

func TestLoadInvalidPrivateKey(t *testing.T) {
	_, err := security.NewPrivateKey("xxx")
	require.Error(t, err)
}
