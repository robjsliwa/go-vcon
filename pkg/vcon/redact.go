package vcon

import (
	"encoding/json"
)

// RedactOption configures the redacted object.
type RedactOption func(*RedactedObject)

// WithRedactedURL sets the URL and content hash for the original vCon.
func WithRedactedURL(url string, hash ContentHashList) RedactOption {
	return func(r *RedactedObject) {
		r.URL = url
		r.ContentHash = hash
	}
}

// Redact creates a redacted copy of this VCon. The redactFn modifies the
// deep copy to remove sensitive data. Per spec Section 4.1.8, empty array
// placeholders should preserve indices.
func (v *VCon) Redact(redactionType string, redactFn func(*VCon) error, opts ...RedactOption) (*VCon, error) {
	// Deep copy via JSON round-trip
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var copy VCon
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, err
	}

	// Apply the redaction function
	if err := redactFn(&copy); err != nil {
		return nil, err
	}

	// Set the redacted field
	redacted := &RedactedObject{
		UUID: v.UUID,
		Type: redactionType,
	}
	for _, opt := range opts {
		opt(redacted)
	}
	copy.Redacted = redacted

	// Generate a new UUID for the redacted copy
	copy.UUID = UUID8DomainName("redacted." + v.UUID)

	return &copy, nil
}

// SetRedacted marks this vCon as a redacted version of another vCon.
func (v *VCon) SetRedacted(uuid, redactionType string, opts ...RedactOption) {
	v.Redacted = &RedactedObject{
		UUID: uuid,
		Type: redactionType,
	}
	for _, opt := range opts {
		opt(v.Redacted)
	}
}
