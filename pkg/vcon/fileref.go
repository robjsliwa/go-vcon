package vcon

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
)

// VerifyIntegrity compares body bytes to the stored SHA-512 hash.
// Caller is responsible for fetching external URL content.
func (f *FileRef) VerifyIntegrity(body []byte) error {
	if f.ContentHash == "" {
		return nil // nothing to verify
	}
	sum := sha512.Sum512(body)
	want := f.ContentHash
	got := base64.RawURLEncoding.EncodeToString(sum[:])
	if want != got {
		return errors.New("content hash mismatch")
	}
	return nil
}
