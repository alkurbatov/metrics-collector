package compression

import (
	"compress/gzip"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type Compressor struct {
	http.ResponseWriter

	// Only Gzip is supported
	encoder *gzip.Writer

	// Supported values of Content-Type header
	supportedContent map[string]struct{}
}

func NewCompressor(w http.ResponseWriter) *Compressor {
	supportedContent := make(map[string]struct{}, 2)
	supportedContent["application/json"] = struct{}{}
	supportedContent["text/html; charset=utf-8"] = struct{}{}

	return &Compressor{
		ResponseWriter:   w,
		supportedContent: supportedContent,
	}
}

func (c *Compressor) Write(resp []byte) (int, error) {
	contentType := c.Header().Get("Content-Type")
	if _, ok := c.supportedContent[contentType]; !ok {
		logging.Log.Debug("Compression not supported for " + contentType)
		return c.ResponseWriter.Write(resp)
	}

	if c.encoder == nil {
		encoder, err := gzip.NewWriterLevel(c.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			logging.Log.Error(err)
			return 0, err
		}

		c.encoder = encoder
	}

	c.Header().Set("Content-Encoding", "gzip")

	return c.encoder.Write(resp)
}

func (c *Compressor) Close() {
	if c.encoder != nil {
		c.encoder.Close()
	}
}

func CompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isGzipEncoded(r.Header.Get("Accept-Encoding")) {
			logging.Log.Debug("Compression not supported by client")

			next.ServeHTTP(w, r)
			return
		}

		compressor := NewCompressor(w)
		defer compressor.Close()

		next.ServeHTTP(compressor, r)
	})
}
