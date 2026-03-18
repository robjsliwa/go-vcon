package vcon

import (
	"testing"
)

func TestDetectFormUnsigned(t *testing.T) {
	data := []byte(`{"uuid":"550e8400-e29b-41d4-a716-446655440000","created_at":"2023-01-15T10:30:00Z","parties":[]}`)
	form, err := DetectForm(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form != VConFormUnsigned {
		t.Errorf("expected unsigned, got %s", form)
	}
}

func TestDetectFormSigned(t *testing.T) {
	data := []byte(`{"payload":"eyJ0ZXN0IjoidmFsdWUifQ","signatures":[{"protected":"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9","signature":"abc123"}]}`)
	form, err := DetectForm(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form != VConFormSigned {
		t.Errorf("expected signed, got %s", form)
	}
}

func TestDetectFormSignedFlattened(t *testing.T) {
	data := []byte(`{"payload":"eyJ0ZXN0IjoidmFsdWUifQ","protected":"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9","signature":"abc123"}`)
	form, err := DetectForm(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form != VConFormSigned {
		t.Errorf("expected signed, got %s", form)
	}
}

func TestDetectFormSignaturWithoutPayloadIsNotSigned(t *testing.T) {
	// "signature" without "payload" should not be detected as signed
	data := []byte(`{"signature":"abc123","uuid":"test"}`)
	form, err := DetectForm(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form == VConFormSigned {
		t.Errorf("should not detect as signed without payload field")
	}
}

func TestDetectFormEncrypted(t *testing.T) {
	data := []byte(`{"protected":"eyJ0eXAiOiJKV1QiLCJlbmMiOiJBMjU2Q0JDLUhTNTEyIn0","ciphertext":"encrypted_data","iv":"test_iv","tag":"test_tag"}`)
	form, err := DetectForm(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form != VConFormEncrypted {
		t.Errorf("expected encrypted, got %s", form)
	}
}

func TestDetectFormEmpty(t *testing.T) {
	_, err := DetectForm([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestDetectFormInvalidJSON(t *testing.T) {
	_, err := DetectForm([]byte(`{invalid json}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDetectFormUnknown(t *testing.T) {
	data := []byte(`{"some_random_field":"value"}`)
	form, err := DetectForm(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form != VConFormUnknown {
		t.Errorf("expected unknown, got %s", form)
	}
}

func TestVConFormString(t *testing.T) {
	tests := []struct {
		form     VConForm
		expected string
	}{
		{VConFormUnknown, "unknown"},
		{VConFormUnsigned, "unsigned"},
		{VConFormSigned, "signed"},
		{VConFormEncrypted, "encrypted"},
	}

	for _, tt := range tests {
		if tt.form.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.form.String())
		}
	}
}
