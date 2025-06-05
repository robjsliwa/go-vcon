package vcon_test

import (
	"crypto/sha512"
	"encoding/base64"
	"testing"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/stretchr/testify/assert"
)

func TestVerifyIntegrity(t *testing.T) {
	// Calculate the actual hash for "hello world" to use in the test
	testData := []byte("hello world")
	sum := sha512.Sum512(testData)
	correctHash := base64.RawURLEncoding.EncodeToString(sum[:])

	tests := []struct {
		name        string
		contentHash string
		body        []byte
		expectError bool
	}{
		{
			name:        "valid hash",
			contentHash: correctHash,
			body:        []byte("hello world"),
			expectError: false,
		},
		{
			name:        "invalid hash",
			contentHash: "sha512-invalid_hash_value",
			body:        []byte("hello world"),
			expectError: true,
		},
		{
			name:        "empty hash",
			contentHash: "",
			body:        []byte("hello world"),
			expectError: false, // no hash, nothing to verify
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := vcon.FileRef{
				ContentHash: tt.contentHash,
			}

			err := fr.VerifyIntegrity(tt.body)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
