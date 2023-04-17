// Package security implements security-related features such as
// signature creation and verification, secrets processing etc.
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

// A Signer provides signature generation and verification functionality.
type Signer struct {
	secret []byte
}

// NewSigner creates new Signer object with the given secret.
// The secret is used to generate/verify payload signature.
func NewSigner(secret Secret) *Signer {
	return &Signer{secret: []byte(secret)}
}

// CalculateSignature generates signature for provided payload.
func (s *Signer) CalculateSignature(name string, data metrics.Metric) (string, error) {
	mac := hmac.New(sha256.New, s.secret)

	var msg string

	switch v := data.(type) {
	case metrics.Counter:
		msg = fmt.Sprintf("%s:%s:%d", name, v.Kind(), v)

	case metrics.Gauge:
		msg = fmt.Sprintf("%s:%s:%f", name, v.Kind(), v)

	default:
		return "", fmt.Errorf("security - CalculateSignature - data.Value.(type): %w", entity.ErrMetricNotImplemented)
	}

	mac.Write([]byte(msg))
	digest := mac.Sum(nil)

	return hex.EncodeToString(digest), nil
}

// CalculateRecordSignature generates signature for provided record.
func (s *Signer) CalculateRecordSignature(data storage.Record) (string, error) {
	return s.CalculateSignature(data.Name, data.Value)
}

// VerifySignature checks signature of provided payload.
func (s *Signer) VerifySignature(name string, data metrics.Metric, hash string) (bool, error) {
	if len(hash) == 0 {
		return false, fmt.Errorf("security - VerifySignature - len(hash): %w", entity.ErrNotSigned)
	}

	expected, err := s.CalculateSignature(name, data)
	if err != nil {
		return false, fmt.Errorf("security - VerifySignature - s.calculateSignature: %w", err)
	}

	return expected == hash, nil
}

// VerifyRecordSignature checks signature of provided record.
func (s *Signer) VerifyRecordSignature(data storage.Record, hash string) (bool, error) {
	return s.VerifySignature(data.Name, data.Value, hash)
}
