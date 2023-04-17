package compression_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/compression"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

func sendEchoRequest(
	t *testing.T,
	headers map[string]string,
	payload *bytes.Buffer,
) (int, []byte) {
	t.Helper()
	require := require.New(t)

	router := chi.NewRouter()
	router.Use(compression.DecompressRequest)
	router.Use(compression.CompressResponse)

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		body, hErr := io.ReadAll(r.Body)
		require.NoError(hErr)

		if len(r.Header.Get("Content-Type")) != 0 {
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		}

		_, hErr = w.Write(body)
		require.NoError(hErr)
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, payload)
	require.NoError(err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(err)

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}

	var respBody []byte

	if len(resp.Header.Get("Content-Encoding")) != 0 {
		var gz *gzip.Reader
		gz, err = gzip.NewReader(resp.Body)
		require.NoError(err)

		defer func() {
			_ = gz.Close()
		}()

		respBody, err = io.ReadAll(gz)
		require.NoError(err)

		return resp.StatusCode, respBody
	}

	respBody, err = io.ReadAll(resp.Body)
	require.NoError(err)

	return resp.StatusCode, respBody
}

func TestCompressDecompressJSONMessage(t *testing.T) {
	msg := `{"text": "Hello, gopher"}`
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}

	payload, err := compression.Pack([]byte(msg))
	require.NoError(t, err)

	status, resp := sendEchoRequest(t, headers, payload)

	require.Equal(t, http.StatusOK, status)
	require.Equal(t, msg, string(resp))
}

func TestClientWithoutCompression(t *testing.T) {
	msg := "Hello, gopher"
	headers := make(map[string]string)

	payload := bytes.NewBufferString(msg)
	status, resp := sendEchoRequest(t, headers, payload)

	require.Equal(t, http.StatusOK, status)
	require.Equal(t, msg, string(resp))
}

func TestDecompressFailsOnNotSupportedEncoding(t *testing.T) {
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "deflate",
	}

	payload := bytes.NewBufferString("Compressed with deflate")
	status, _ := sendEchoRequest(t, headers, payload)

	require.Equal(t, http.StatusBadRequest, status)
}

func TestResponseInPlainTextNotCompressed(t *testing.T) {
	msg := "Hello, gopher"
	headers := map[string]string{
		"Content-Type":     "text/plain",
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}

	compressed, err := compression.Pack([]byte(msg))
	require.NoError(t, err)

	status, resp := sendEchoRequest(t, headers, compressed)

	require.Equal(t, http.StatusOK, status)
	require.Equal(t, msg, string(resp))
}

func TestClientAcceptNotSupportedEncoding(t *testing.T) {
	msg := "Plain text"
	headers := map[string]string{
		"Content-Type":    "application/json",
		"Accept-Encoding": "deflate",
	}

	payload := bytes.NewBufferString(msg)
	status, resp := sendEchoRequest(t, headers, payload)

	require.Equal(t, http.StatusOK, status)
	require.Equal(t, msg, string(resp))
}

func TestClientSentPlainTextAsCompressed(t *testing.T) {
	msg := "Plain text"
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
	}

	payload := bytes.NewBufferString(msg)
	status, _ := sendEchoRequest(t, headers, payload)

	require.Equal(t, http.StatusBadRequest, status)
}
