package vcon

import (
	"encoding/json"
)

// AmendOption configures the amended object.
type AmendOption func(*AmendedObject)

// WithAmendedURL sets the URL and content hash for the original vCon.
func WithAmendedURL(url string, hash ContentHashList) AmendOption {
	return func(a *AmendedObject) {
		a.URL = url
		a.ContentHash = hash
	}
}

// Amend creates an amended copy of this VCon with additional data.
// The amendFn modifies the deep copy to add analysis, transcripts, etc.
func (v *VCon) Amend(amendFn func(*VCon) error, opts ...AmendOption) (*VCon, error) {
	// Deep copy via JSON round-trip
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var copy VCon
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, err
	}

	// Apply the amendment function
	if err := amendFn(&copy); err != nil {
		return nil, err
	}

	// Set the amended field
	amended := &AmendedObject{
		UUID: v.UUID,
	}
	for _, opt := range opts {
		opt(amended)
	}
	copy.Amended = amended

	// Generate a new UUID for the amended copy
	copy.UUID = UUID8DomainName("amended." + v.UUID)

	return &copy, nil
}

// SetAmended marks this vCon as amending a prior version.
func (v *VCon) SetAmended(uuid string, opts ...AmendOption) {
	v.Amended = &AmendedObject{
		UUID: uuid,
	}
	for _, opt := range opts {
		opt(v.Amended)
	}
}
