package compression

import (
	"compress/gzip"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/rs/zerolog"
)

type Compressor struct {
	http.ResponseWriter

	logger *zerolog.Logger

	// Only Gzip is supported
	encoder *gzip.Writer

	// Supported values of Content-Type header
	supportedContent map[string]struct{}
}

func NewCompressor(w http.ResponseWriter, logger *zerolog.Logger) *Compressor {
	supportedContent := make(map[string]struct{}, 2)
	supportedContent["application/json"] = struct{}{}
	supportedContent["text/html; charset=utf-8"] = struct{}{}

	return &Compressor{
		ResponseWriter:   w,
		logger:           logger,
		supportedContent: supportedContent,
	}
}

func (c *Compressor) Write(resp []byte) (int, error) {
	contentType := c.Header().Get("Content-Type")
	if _, ok := c.supportedContent[contentType]; !ok {
		c.logger.Debug().Msg("Compression not supported for " + contentType)
		return c.ResponseWriter.Write(resp)
	}

	if c.encoder == nil {
		encoder, err := gzip.NewWriterLevel(c.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			c.logger.Error().Err(err).Msg("")
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
		logger := logging.GetLogger(r.Context())

		if !isGzipEncoded(r.Header.Get("Accept-Encoding")) {
			logger.Debug().Msg("Compression not supported by client")

			next.ServeHTTP(w, r)
			return
		}

		compressor := NewCompressor(w, logger)
		defer compressor.Close()

		next.ServeHTTP(compressor, r)
	})
}
