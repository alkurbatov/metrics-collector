package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/schema"
)

func signError(reason error) error {
	return fmt.Errorf("failed to sign request: %w", reason)
}

func verifyError(reason error) error {
	return fmt.Errorf("failed to verify request signature: %w", reason)
}

type Signer struct {
	secret []byte
}

func NewSigner(secret Secret) *Signer {
	return &Signer{secret: []byte(secret)}
}

func (s *Signer) calculateSignature(req *schema.MetricReq) ([]byte, error) {
	mac := hmac.New(sha256.New, s.secret)

	var msg string

	switch req.MType {
	case entity.Counter:
		if req.Delta == nil {
			return nil, entity.ErrIncompleteRequest
		}

		msg = fmt.Sprintf("%s:%s:%d", req.ID, req.MType, *req.Delta)

	case entity.Gauge:
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

func (s *Signer) SignRequest(req *schema.MetricReq) error {
	digest, err := s.calculateSignature(req)
	if err != nil {
		return signError(err)
	}

	req.Hash = hex.EncodeToString(digest)

	return nil
}

func (s *Signer) VerifySignature(req *schema.MetricReq) (bool, error) {
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
