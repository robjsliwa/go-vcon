package vcon

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ContentHash represents a content hash in the format "algorithm-base64url_encoded_hash"
// as defined in Section 2.2 of the vCon spec.
type ContentHash struct {
	Algorithm string // lowercase, no hyphens (e.g., "sha512")
	Hash      string // base64url-encoded hash value (no padding)
}

// ParseContentHash parses a content hash string in the format "algorithm-hash".
func ParseContentHash(s string) (ContentHash, error) {
	alg, hash, found := strings.Cut(s, "-")
	if !found {
		return ContentHash{}, fmt.Errorf("invalid content_hash format: missing '-' separator in %q", s)
	}
	if alg == "" {
		return ContentHash{}, fmt.Errorf("invalid content_hash format: empty algorithm in %q", s)
	}
	if hash == "" {
		return ContentHash{}, fmt.Errorf("invalid content_hash format: empty hash in %q", s)
	}
	return ContentHash{Algorithm: alg, Hash: hash}, nil
}

// ComputeSHA512 computes a SHA-512 content hash for the given data.
func ComputeSHA512(data []byte) ContentHash {
	h := sha512.Sum512(data)
	return ContentHash{
		Algorithm: "sha512",
		Hash:      base64.RawURLEncoding.EncodeToString(h[:]),
	}
}

// String returns the "algorithm-hash" string representation.
func (ch ContentHash) String() string {
	return ch.Algorithm + "-" + ch.Hash
}

// Verify recomputes the hash of data and compares it with the stored hash.
// Currently supports sha512.
func (ch ContentHash) Verify(data []byte) bool {
	switch ch.Algorithm {
	case "sha512":
		h := sha512.Sum512(data)
		expected := base64.RawURLEncoding.EncodeToString(h[:])
		return expected == ch.Hash
	default:
		return false
	}
}

// IsZero returns true if the ContentHash is empty.
func (ch ContentHash) IsZero() bool {
	return ch.Algorithm == "" && ch.Hash == ""
}

// ContentHashList holds one or more content hashes.
// Per spec, content_hash can be a single string or an array of strings.
type ContentHashList []ContentHash

// MarshalJSON serializes as a single string if one element, or array if multiple.
func (l ContentHashList) MarshalJSON() ([]byte, error) {
	if len(l) == 0 {
		return json.Marshal(nil)
	}
	if len(l) == 1 {
		return json.Marshal(l[0].String())
	}
	strs := make([]string, len(l))
	for i, ch := range l {
		strs[i] = ch.String()
	}
	return json.Marshal(strs)
}

// UnmarshalJSON handles both a single string and an array of strings.
func (l *ContentHashList) UnmarshalJSON(data []byte) error {
	// Try single string first
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		ch, err := ParseContentHash(single)
		if err != nil {
			return err
		}
		*l = ContentHashList{ch}
		return nil
	}

	// Try array of strings
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("content_hash must be a string or array of strings: %w", err)
	}
	result := make(ContentHashList, len(arr))
	for i, s := range arr {
		ch, err := ParseContentHash(s)
		if err != nil {
			return err
		}
		result[i] = ch
	}
	*l = result
	return nil
}

// ContainsAlgorithm checks if any hash in the list uses the given algorithm.
func (l ContentHashList) ContainsAlgorithm(alg string) bool {
	for _, ch := range l {
		if ch.Algorithm == alg {
			return true
		}
	}
	return false
}

// First returns the first content hash, or a zero value if empty.
func (l ContentHashList) First() ContentHash {
	if len(l) == 0 {
		return ContentHash{}
	}
	return l[0]
}

// IsEmpty returns true if the list has no entries.
func (l ContentHashList) IsEmpty() bool {
	return len(l) == 0
}
