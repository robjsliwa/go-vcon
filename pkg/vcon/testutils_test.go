package vcon_test

import (
	"crypto/sha512"
	"encoding/base64"
)

// calculateHash calculates a vCon-compatible content hash for a byte slice
func calculateHash(data []byte) string {
	sum := sha512.Sum512(data)
	hash := base64.RawURLEncoding.EncodeToString(sum[:])
	return "sha512-" + hash
}
