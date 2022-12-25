package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/schema"
)

type Secret string

func (s Secret) String() string {
	return strings.Repeat("*", len(s))
}

var (
	errIncompleteRequest  = errors.New("metrics value not set")
	errNotSigned          = errors.New("request not signed")
	errMetricNotSupported = errors.New("unsupported metric type")
)

type Signer struct {
	secret []byte
}

func NewSigner(secret Secret) *Signer {
	key := []byte(secret)
	if len(key) < 32 {
		logging.Log.Warning("Insecure signature: secret key is shorter than 32 bytes!")
	}

	return &Signer{secret: key}
}

func (s *Signer) calculateSignature(req *schema.MetricReq) ([]byte, error) {
	mac := hmac.New(sha256.New, s.secret)

	var msg string

	switch req.MType {
	case "counter":
		if req.Delta == nil {
			return nil, errIncompleteRequest
		}

		msg = fmt.Sprintf("%s:%s:%d", req.ID, req.MType, *req.Delta)

	case "gauge":
		if req.Value == nil {
			return nil, errIncompleteRequest
		}

		msg = fmt.Sprintf("%s:%s:%f", req.ID, req.MType, *req.Value)

	default:
		return nil, errMetricNotSupported
	}

	mac.Write([]byte(msg))

	return mac.Sum(nil), nil
}

func (s *Signer) SignRequest(req *schema.MetricReq) error {
	digest, err := s.calculateSignature(req)
	if err != nil {
		return err
	}

	req.Hash = hex.EncodeToString(digest)

	return nil
}

func (s *Signer) VerifySignature(req *schema.MetricReq) (bool, error) {
	if len(req.Hash) == 0 {
		return false, errNotSigned
	}

	expected, err := hex.DecodeString(req.Hash)
	if err != nil {
		return false, err
	}

	digest, err := s.calculateSignature(req)
	if err != nil {
		return false, err
	}

	return hmac.Equal(digest, expected), nil
}
