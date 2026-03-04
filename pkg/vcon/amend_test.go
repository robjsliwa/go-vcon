package vcon

import (
	"testing"
	"time"
)

func TestAmend(t *testing.T) {
	v := New("example.com")
	v.Subject = "Original Call"
	v.AddParty(Party{Name: "Alice"})

	now := time.Now().UTC()
	v.AddDialog(Dialog{
		Type:      "recording",
		StartTime: &now,
		Parties:   []int{0},
		MediaType: "audio/wav",
	})

	amended, err := v.Amend(func(copy *VCon) error {
		// Add a transcript analysis
		copy.AddAnalysis(Analysis{
			Type:      "transcript",
			Dialog:    []int{0},
			MediaType: "text/plain",
			Vendor:    "TranscriptCo",
			Product:   "AutoTranscribe v1.0",
			Body:      "Hello, this is the transcript.",
			Encoding:  "none",
		})
		return nil
	})

	if err != nil {
		t.Fatalf("amend error: %v", err)
	}

	// Amended copy should have the amended field set
	if amended.Amended == nil {
		t.Fatal("expected amended field to be set")
	}
	if amended.Amended.UUID != v.UUID {
		t.Errorf("expected amended UUID %s, got %s", v.UUID, amended.Amended.UUID)
	}

	// Amended copy should have a different UUID
	if amended.UUID == v.UUID {
		t.Error("amended copy should have a different UUID")
	}

	// Amended copy should have the added analysis
	if len(amended.Analysis) != 1 {
		t.Fatalf("expected 1 analysis in amended, got %d", len(amended.Analysis))
	}
	if amended.Analysis[0].Type != "transcript" {
		t.Errorf("expected analysis type transcript, got %s", amended.Analysis[0].Type)
	}

	// Original should not have the analysis
	if len(v.Analysis) != 0 {
		t.Error("original should not have analysis added")
	}
}

func TestAmendWithURL(t *testing.T) {
	v := New("example.com")
	v.AddParty(Party{Name: "Alice"})

	hash := ContentHashList{ComputeSHA512([]byte("original-data"))}

	amended, err := v.Amend(func(copy *VCon) error {
		return nil
	}, WithAmendedURL("https://archive.example.com/vcon/456", hash))

	if err != nil {
		t.Fatalf("amend error: %v", err)
	}

	if amended.Amended.URL != "https://archive.example.com/vcon/456" {
		t.Errorf("expected URL, got %s", amended.Amended.URL)
	}
	if amended.Amended.ContentHash.IsEmpty() {
		t.Error("expected content hash to be set")
	}
}

func TestSetAmended(t *testing.T) {
	v := New("example.com")
	v.SetAmended("original-uuid")

	if v.Amended == nil {
		t.Fatal("expected amended to be set")
	}
	if v.Amended.UUID != "original-uuid" {
		t.Errorf("expected UUID original-uuid, got %s", v.Amended.UUID)
	}
}

func TestAmendPreservesOriginal(t *testing.T) {
	v := New("example.com")
	v.Subject = "Do Not Modify"
	v.AddParty(Party{Name: "Alice"})
	v.AddParty(Party{Name: "Bob"})

	amended, err := v.Amend(func(copy *VCon) error {
		copy.Subject = "Modified Subject"
		return nil
	})

	if err != nil {
		t.Fatalf("amend error: %v", err)
	}

	// Original should be untouched
	if v.Subject != "Do Not Modify" {
		t.Errorf("original subject was modified: %s", v.Subject)
	}

	// Amended should have new subject
	if amended.Subject != "Modified Subject" {
		t.Errorf("expected amended subject 'Modified Subject', got %s", amended.Subject)
	}
}
