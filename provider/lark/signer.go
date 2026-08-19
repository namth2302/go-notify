package lark

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

// Signer generates Lark webhook signature using HMAC-SHA256.
type Signer struct {
	secret string
}

// NewSigner creates a new Lark signer.
func NewSigner(secret string) *Signer {
	return &Signer{secret: secret}
}

// GenerateSignature generates timestamp string and HMAC-SHA256 signature string.
func (s *Signer) GenerateSignature(t time.Time) (timestamp string, sign string, err error) {
	if s.secret == "" {
		return "", "", nil
	}

	ts := strconv.FormatInt(t.Unix(), 10)
	stringToSign := fmt.Sprintf("%s\n%s", ts, s.secret)

	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err = h.Write([]byte(""))
	if err != nil {
		return "", "", fmt.Errorf("hmac write failed: %w", err)
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return ts, signature, nil
}
