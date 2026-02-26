package vcon

import (
	"encoding/json"
	"errors"
)

// VConForm represents the serialization form of a vCon per Section 5.4.
type VConForm int

const (
	VConFormUnknown   VConForm = iota
	VConFormUnsigned           // Plain JSON object
	VConFormSigned             // JWS General JSON Serialization
	VConFormEncrypted          // JWE JSON Serialization
)

// String returns a human-readable name for the form.
func (f VConForm) String() string {
	switch f {
	case VConFormUnsigned:
		return "unsigned"
	case VConFormSigned:
		return "signed"
	case VConFormEncrypted:
		return "encrypted"
	default:
		return "unknown"
	}
}

// DetectForm inspects raw JSON bytes and determines whether the data
// represents an unsigned vCon, a signed vCon (JWS), or an encrypted
// vCon (JWE). It does not validate the content, only checks for
// structural markers.
func DetectForm(data []byte) (VConForm, error) {
	if len(data) == 0 {
		return VConFormUnknown, errors.New("empty data")
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return VConFormUnknown, err
	}

	// JWE has "ciphertext" field
	if _, ok := m["ciphertext"]; ok {
		return VConFormEncrypted, nil
	}

	// JWS General JSON has "signatures" array
	if _, ok := m["signatures"]; ok {
		return VConFormSigned, nil
	}

	// Unsigned vCon has "uuid" and/or "parties"
	if _, ok := m["uuid"]; ok {
		return VConFormUnsigned, nil
	}
	if _, ok := m["parties"]; ok {
		return VConFormUnsigned, nil
	}

	return VConFormUnknown, nil
}
