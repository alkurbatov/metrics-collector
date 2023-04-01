package compression

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var gzipWritersPool = sync.Pool{
	New: func() interface{} {
		// NB (alkurbatov): It seems that NewWriterLevel only returns error
		// on a bad level. We are guaranteeing that the level is valid
		// so it is okay to ignore the returned error.
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

// A Compressor implements data compression using grip encoder.
type Compressor struct {
	http.ResponseWriter

	logger *zerolog.Logger

	// Only Gzip is supported
	encoder *gzip.Writer

	// Supported values of Content-Type header
	supportedContent map[string]struct{}
}

// NewCompressor creatse new Compressor instance.
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

// Write compresses response content data in case of supported type.
// The content type should be specified in the Content-Type header in advance.
func (c *Compressor) Write(resp []byte) (int, error) {
	contentType := c.Header().Get("Content-Type")
	if _, ok := c.supportedContent[contentType]; !ok {
		c.logger.Debug().Msg("Compression not supported for " + contentType)
		return c.ResponseWriter.Write(resp)
	}

	if c.encoder == nil {
		encoder := gzipWritersPool.Get().(*gzip.Writer)
		encoder.Reset(c.ResponseWriter)

		c.encoder = encoder
	}

	c.Header().Set("Content-Encoding", "gzip")

	return c.encoder.Write(resp)
}

// Close dumps internal buffers and finishes compression.
// Must be called before end of response processing, otherwise part of data can be lost.
func (c *Compressor) Close() {
	if c.encoder == nil {
		return
	}

	if err := c.encoder.Close(); err != nil {
		c.logger.Error().Err(err).Msg("Compressor - Close - c.encoder.Close")
	}

	gzipWritersPool.Put(c.encoder)
}

// CompressResponse is net/http middleware executing gzip compression
// is gzip is supported by client and response belongs to supported type.
func CompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())

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

// Pack compresses bytes using gzip algorithm.
func Pack(data []byte) (*bytes.Buffer, error) {
	rv := new(bytes.Buffer)

	encoder := gzipWritersPool.Get().(*gzip.Writer)
	defer gzipWritersPool.Put(encoder)

	encoder.Reset(rv)

	if _, err := encoder.Write(data); err != nil {
		return nil, err
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return rv, nil
}
