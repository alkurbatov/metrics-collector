package security

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/rs/zerolog/log"
)

// PublicKey is RSA key used to encrypt data.
type PublicKey *rsa.PublicKey

// PrivateKey is RSA key used to decrypt data.
type PrivateKey *rsa.PrivateKey

// newRawKey reads key in PEM format from file.
func newRawKey(path entity.FilePath) (*pem.Block, error) {
	rawKey, err := os.ReadFile(path.String())
	if err != nil {
		return nil, fmt.Errorf("security - newRawKey - os.ReadFile: %w", err)
	}

	key, _ := pem.Decode(rawKey)
	if key == nil {
		return nil, fmt.Errorf("security - newRawKey - pem.Decode: %w", entity.ErrBadKeyFile)
	}

	return key, nil
}

// NewPublicKey reads RSA public key from file.
func NewPublicKey(path entity.FilePath) (PublicKey, error) {
	block, err := newRawKey(path)
	if err != nil {
		return nil, err
	}

	rawKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("security - newPublicKey - x509.ParsePKIXPublicKey: %w", err)
	}

	key, ok := rawKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("security - newPublicKey - .(*rsa.PublicKey): %w", entity.ErrNotSupportedKey)
	}

	return key, nil
}

// Encrypt encrypts the provided message with RSA algorithm.
func Encrypt(src io.Reader, key PublicKey) (*bytes.Buffer, error) {
	msg := new(bytes.Buffer)

	// NB (alkurbatov): As the message could be large, thus
	// we have to split it in chunks to bypass RSA key limitations.
	chunkSize := (*rsa.PublicKey)(key).Size() - 2*sha256.New().Size() - 2
	chunk := make([]byte, chunkSize)

	for {
		n, err := src.Read(chunk)

		if n > 0 {
			// NB (alkurbatov): If len(message) < chunkSize, avoid encryption and sending of trailing zeroes.
			if n != len(chunk) {
				chunk = chunk[:n]
			}

			encryptedChunk, encErr := rsa.EncryptOAEP(
				sha256.New(),
				rand.Reader,
				key,
				chunk,
				nil)
			if err != nil {
				return nil, fmt.Errorf("security - Encrypt - rsa.EncryptOAEP: %w", encErr)
			}

			msg.Write(encryptedChunk)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("security - Encrypt - reader.Read: %w", err)
		}
	}

	return msg, nil
}

// NewPrivateKey reads RSA private key from file.
func NewPrivateKey(path entity.FilePath) (PrivateKey, error) {
	block, err := newRawKey(path)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("security - NewPrivateKey - x509.ParsePKCS1PrivateKey: %w", err)
	}

	return key, nil
}

// Decrypt decrypts the provided message using RSA algorithm.
func Decrypt(src io.Reader, key PrivateKey) (*bytes.Buffer, error) {
	msg := new(bytes.Buffer)

	// NB (alkurbatov): As the message could be large, thus
	// we have to read it in chunks to bypass RSA key limitations.
	chunkSize := key.PublicKey.Size()
	chunk := make([]byte, chunkSize)

	for {
		n, err := src.Read(chunk)

		if n > 0 {
			// NB (alkurbatov): If len(message) < chunkSize, cut off possible garbage.
			if n != len(chunk) {
				chunk = chunk[:n]
			}

			decryptedChunk, decErr := rsa.DecryptOAEP(
				sha256.New(),
				nil,
				key,
				chunk,
				nil)
			if decErr != nil {
				return nil, fmt.Errorf("security - Decrypt - rsa.DecryptOAEP: %w", decErr)
			}

			msg.Write(decryptedChunk)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("security - Decrypt - src.Read: %w", err)
		}
	}

	return msg, nil
}

// DecryptRequest is a HTTP middleware that decrypts request's body
// using RSA algorithm.
func DecryptRequest(key PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger := log.Ctx(r.Context())

			msg, err := Decrypt(r.Body, key)
			if err != nil {
				logger.Error().Err(err).Msg("security - DecryptRequest - Decrypt")
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			r.Body = io.NopCloser(bytes.NewReader(msg.Bytes()))
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
