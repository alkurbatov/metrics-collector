// Package security implements security-related features such as
// signature creation and verification, secrets processing etc.
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

func signError(reason error) error {
	return fmt.Errorf("failed to sign request: %w", reason)
}

func verifyError(reason error) error {
	return fmt.Errorf("failed to verify request signature: %w", reason)
}

// A Signer provides signature generation and verification functionality.
type Signer struct {
	secret []byte
}

// NewSigner creates new Signer object with the given secret.
// The secret is used to generate/verify payload signature.
func NewSigner(secret Secret) *Signer {
	return &Signer{secret: []byte(secret)}
}

func (s *Signer) calculateSignature(req *metrics.MetricReq) ([]byte, error) {
	mac := hmac.New(sha256.New, s.secret)

	var msg string

	switch req.MType {
	case metrics.KindCounter:
		if req.Delta == nil {
			return nil, entity.ErrIncompleteRequest
		}

		msg = fmt.Sprintf("%s:%s:%d", req.ID, req.MType, *req.Delta)

	case metrics.KindGauge:
		if req.Value == nil {
			return nil, entity.ErrIncompleteRequest
		}

		msg = fmt.Sprintf("%s:%s:%f", req.ID, req.MType, *req.Value)

	default:
		return nil, entity.MetricNotImplementedError(req.MType) //nolint: wrapcheck
	}

	mac.Write([]byte(msg))

	return mac.Sum(nil), nil
}

// SignRequest generates signature for provided payload.
func (s *Signer) SignRequest(req *metrics.MetricReq) error {
	digest, err := s.calculateSignature(req)
	if err != nil {
		return signError(err)
	}

	req.Hash = hex.EncodeToString(digest)

	return nil
}

// VerifySignature checks signature of provided payload.
func (s *Signer) VerifySignature(req *metrics.MetricReq) (bool, error) {
	if len(req.Hash) == 0 {
		return false, verifyError(entity.ErrNotSigned)
	}

	expected, err := hex.DecodeString(req.Hash)
	if err != nil {
		return false, verifyError(err)
	}

	digest, err := s.calculateSignature(req)
	if err != nil {
		return false, verifyError(err)
	}

	return hmac.Equal(digest, expected), nil
}
