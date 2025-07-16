package vcon

import (
	"encoding/json"

	jc "github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

// Canonicalise returns RFC 8785‑canonical JSON bytes for any Go value.
func Canonicalise(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jc.Transform(raw)
}
