package vcon

import (
	"encoding/json"
	"fmt"
)

// RedactedObject marks a vCon as a redacted version of another vCon.
// Per spec Section 4.1.8.
type RedactedObject struct {
	UUID        string          `json:"uuid"`
	Type        string          `json:"type"`
	URL         string          `json:"url,omitempty"`
	ContentHash ContentHashList `json:"content_hash,omitempty"`
}

// AmendedObject marks a vCon as amending a prior version.
// Per spec Section 4.1.9.
type AmendedObject struct {
	UUID        string          `json:"uuid,omitempty"`
	URL         string          `json:"url,omitempty"`
	ContentHash ContentHashList `json:"content_hash,omitempty"`
}

// SessionId represents a dialog session identifier with local and remote components.
type SessionId struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

// IntOrSlice handles fields that can be either a single int or []int.
// Used for transfer_target, original, consultation, target_dialog fields.
type IntOrSlice struct {
	value any // int or []int
}

// NewIntValue creates an IntOrSlice with a single int value.
func NewIntValue(v int) *IntOrSlice {
	return &IntOrSlice{value: v}
}

// NewIntSliceValue creates an IntOrSlice with a slice of int values.
func NewIntSliceValue(v []int) *IntOrSlice {
	return &IntOrSlice{value: v}
}

// AsInt returns the value as an int if it is a single value.
func (v IntOrSlice) AsInt() (int, bool) {
	i, ok := v.value.(int)
	return i, ok
}

// AsSlice returns the value as a slice. If the value is a single int,
// it is wrapped in a single-element slice.
func (v IntOrSlice) AsSlice() []int {
	switch val := v.value.(type) {
	case int:
		return []int{val}
	case []int:
		return val
	default:
		return nil
	}
}

// IsZero returns true if the IntOrSlice has no value set.
func (v IntOrSlice) IsZero() bool {
	return v.value == nil
}

// MarshalJSON serializes as int or []int.
func (v IntOrSlice) MarshalJSON() ([]byte, error) {
	if v.value == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(v.value)
}

// UnmarshalJSON handles both int and []int.
func (v *IntOrSlice) UnmarshalJSON(data []byte) error {
	// Try single int
	var single int
	if err := json.Unmarshal(data, &single); err == nil {
		v.value = single
		return nil
	}

	// Try array of ints
	var arr []int
	if err := json.Unmarshal(data, &arr); err == nil {
		v.value = arr
		return nil
	}

	// Try float (JSON numbers can be float)
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		v.value = int(f)
		return nil
	}

	return fmt.Errorf("IntOrSlice: expected int or []int, got %s", string(data))
}
