package security_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

const (
	publicKeyPath  entity.FilePath = "../../build/keys/public.pem"
	privateKeyPath entity.FilePath = "../../build/keys/private.pem"
)

func encrypt(t *testing.T, msg string) *bytes.Buffer {
	t.Helper()
	require := require.New(t)

	publicKey, err := security.NewPublicKey(publicKeyPath)
	require.NoError(err)

	input := bytes.NewBufferString(msg)
	encrypted, err := security.Encrypt(io.Reader(input), publicKey)
	require.NoError(err)

	return encrypted
}

func sendEchoRequest(t *testing.T, encryptedPayload *bytes.Buffer) (int, []byte) {
	t.Helper()
	require := require.New(t)

	privateKey, err := security.NewPrivateKey(privateKeyPath)
	require.NoError(err)

	router := chi.NewRouter()
	router.Use(security.DecryptRequest(privateKey))
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		body, hErr := io.ReadAll(r.Body)
		require.NoError(hErr)

		_, hErr = w.Write(body)
		require.NoError(hErr)
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, encryptedPayload)
	require.NoError(err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(err)

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(err)

	return resp.StatusCode, respBody
}

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

			encrypted := encrypt(t, tc.msg)

			privateKey, err := security.NewPrivateKey(privateKeyPath)
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

func TestDecryptRequest(t *testing.T) {
	require := require.New(t)
	msg := "Hello, gopher"
	encrypted := encrypt(t, msg)

	status, resp := sendEchoRequest(t, encrypted)

	require.Equal(http.StatusOK, status)
	require.Equal(msg, string(resp))
}

func TestDecryptRequestOnBadData(t *testing.T) {
	require := require.New(t)
	msg := bytes.NewBufferString("123")

	status, _ := sendEchoRequest(t, msg)

	require.Equal(http.StatusBadRequest, status)
}
