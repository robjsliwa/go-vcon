package vcon

import (
	"testing"
	"time"
)

func TestRedact(t *testing.T) {
	v := New("example.com")
	v.Subject = "Sensitive Call"
	v.AddParty(Party{Name: "Alice", Tel: "tel:+12025551234"})
	v.AddParty(Party{Name: "Bob", Tel: "tel:+12025555678"})

	now := time.Now().UTC()
	v.AddDialog(Dialog{
		Type:      "recording",
		StartTime: &now,
		Parties:   []int{0, 1},
		Body:      "sensitive-audio-data",
		Encoding:  "base64url",
		MediaType: "audio/wav",
	})

	redacted, err := v.Redact("audio", func(copy *VCon) error {
		// Remove the audio body but preserve the dialog structure
		copy.Dialog[0].Body = ""
		copy.Dialog[0].Encoding = ""
		return nil
	})

	if err != nil {
		t.Fatalf("redact error: %v", err)
	}

	// Redacted copy should have the redacted field set
	if redacted.Redacted == nil {
		t.Fatal("expected redacted field to be set")
	}
	if redacted.Redacted.UUID != v.UUID {
		t.Errorf("expected redacted UUID %s, got %s", v.UUID, redacted.Redacted.UUID)
	}
	if redacted.Redacted.Type != "audio" {
		t.Errorf("expected redaction type 'audio', got %s", redacted.Redacted.Type)
	}

	// Redacted copy should have a different UUID
	if redacted.UUID == v.UUID {
		t.Error("redacted copy should have a different UUID")
	}

	// Redacted copy should have empty body
	if redacted.Dialog[0].Body != "" {
		t.Error("redacted dialog body should be empty")
	}

	// Original should be unchanged
	if v.Dialog[0].Body != "sensitive-audio-data" {
		t.Error("original dialog body should be unchanged")
	}

	// Party count should be preserved
	if len(redacted.Parties) != 2 {
		t.Errorf("expected 2 parties in redacted, got %d", len(redacted.Parties))
	}
}

func TestRedactWithURL(t *testing.T) {
	v := New("example.com")
	v.AddParty(Party{Name: "Alice"})

	hash := ContentHashList{ComputeSHA512([]byte("original-data"))}

	redacted, err := v.Redact("full", func(copy *VCon) error {
		return nil
	}, WithRedactedURL("https://archive.example.com/vcon/123", hash))

	if err != nil {
		t.Fatalf("redact error: %v", err)
	}

	if redacted.Redacted.URL != "https://archive.example.com/vcon/123" {
		t.Errorf("expected URL, got %s", redacted.Redacted.URL)
	}
	if redacted.Redacted.ContentHash.IsEmpty() {
		t.Error("expected content hash to be set")
	}
}

func TestSetRedacted(t *testing.T) {
	v := New("example.com")
	v.SetRedacted("original-uuid", "audio")

	if v.Redacted == nil {
		t.Fatal("expected redacted to be set")
	}
	if v.Redacted.UUID != "original-uuid" {
		t.Errorf("expected UUID original-uuid, got %s", v.Redacted.UUID)
	}
	if v.Redacted.Type != "audio" {
		t.Errorf("expected type audio, got %s", v.Redacted.Type)
	}
}
