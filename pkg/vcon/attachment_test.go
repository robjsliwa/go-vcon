package vcon

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAttachmentTypes(t *testing.T) {
	tests := []AttachmentType{
		AttachmentTypeTags,
		AttachmentTypeMetadata,
		AttachmentTypeDocument,
	}

	expected := []string{
		"tags",
		"metadata",
		"document",
	}

	for i, attachmentType := range tests {
		if string(attachmentType) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], string(attachmentType))
		}
	}
}

func TestNewAttachment(t *testing.T) {
	tests := []struct {
		name         string
		attachType   string
		body         interface{}
		encoding     string
		expectError  bool
		expectedType string
	}{
		{
			name:         "valid base64url encoding",
			attachType:   "document",
			body:         "SGVsbG8gV29ybGQ",
			encoding:     "base64url",
			expectError:  false,
			expectedType: "document",
		},
		{
			name:         "valid json encoding",
			attachType:   "metadata",
			body:         map[string]interface{}{"key": "value"},
			encoding:     "json",
			expectError:  false,
			expectedType: "metadata",
		},
		{
			name:        "invalid encoding",
			attachType:  "document",
			body:        "content",
			encoding:    "invalid_encoding",
			expectError: true,
		},
		{
			name:         "none encoding",
			attachType:   "tags",
			body:         "plain text",
			encoding:     "none",
			expectError:  false,
			expectedType: "tags",
		},
		{
			name:        "base64 encoding is no longer valid",
			attachType:  "document",
			body:        "SGVsbG8gV29ybGQ=",
			encoding:    "base64",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment, err := NewAttachment(tt.attachType, tt.body, tt.encoding)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if attachment == nil {
				t.Fatal("expected non-nil attachment")
			}

			if attachment.Encoding != tt.encoding {
				t.Errorf("expected encoding %s, got %s", tt.encoding, attachment.Encoding)
			}
		})
	}
}

func TestValidAttachmentEncodings(t *testing.T) {
	expected := []string{"base64url", "json", "none"}

	if len(ValidAttachmentEncodings) != len(expected) {
		t.Errorf("expected %d valid encodings, got %d", len(expected), len(ValidAttachmentEncodings))
	}

	for _, expectedEnc := range expected {
		found := false
		for _, validEnc := range ValidAttachmentEncodings {
			if validEnc == expectedEnc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected encoding %s to be in ValidAttachmentEncodings", expectedEnc)
		}
	}
}

func TestAttachmentSerialization(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

	attachment := Attachment{
		Body:        "test content",
		Encoding:    "none",
		DialogIdx:   IntPtr(1),
		PartyIdx:    0,
		StartTime:   startTime,
		MediaType:   "text/plain",
		Filename:    "test.txt",
		ContentHash: ContentHashList{{Algorithm: "sha512", Hash: "abc123"}},
		Purpose:     "transcript",
	}

	// Test marshaling
	jsonData, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("failed to marshal attachment: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Attachment
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal attachment: %v", err)
	}

	if unmarshaled.Body != attachment.Body {
		t.Errorf("expected body %s, got %s", attachment.Body, unmarshaled.Body)
	}

	if unmarshaled.Encoding != attachment.Encoding {
		t.Errorf("expected encoding %s, got %s", attachment.Encoding, unmarshaled.Encoding)
	}

	if unmarshaled.DialogIdx == nil || *unmarshaled.DialogIdx != *attachment.DialogIdx {
		t.Errorf("expected dialog index %v, got %v", attachment.DialogIdx, unmarshaled.DialogIdx)
	}

	if unmarshaled.PartyIdx != attachment.PartyIdx {
		t.Errorf("expected party index %d, got %d", attachment.PartyIdx, unmarshaled.PartyIdx)
	}

	if unmarshaled.MediaType != attachment.MediaType {
		t.Errorf("expected mediatype %s, got %s", attachment.MediaType, unmarshaled.MediaType)
	}

	if unmarshaled.Filename != attachment.Filename {
		t.Errorf("expected filename %s, got %s", attachment.Filename, unmarshaled.Filename)
	}

	if unmarshaled.Purpose != attachment.Purpose {
		t.Errorf("expected purpose %s, got %s", attachment.Purpose, unmarshaled.Purpose)
	}
}

func TestAttachmentWithURL(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

	attachment := Attachment{
		URL:         "https://example.com/document.pdf",
		DialogIdx:   IntPtr(0),
		PartyIdx:    1,
		StartTime:   startTime,
		MediaType:   "application/pdf",
		Filename:    "document.pdf",
		ContentHash: ContentHashList{{Algorithm: "sha512", Hash: "xyz789"}},
	}

	// Test marshaling
	jsonData, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("failed to marshal attachment with URL: %v", err)
	}

	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "https://example.com/document.pdf") {
		t.Error("expected URL to be present in JSON")
	}

	// Test unmarshaling
	var unmarshaled Attachment
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal attachment with URL: %v", err)
	}

	if unmarshaled.URL != attachment.URL {
		t.Errorf("expected URL %s, got %s", attachment.URL, unmarshaled.URL)
	}
}

func TestAttachmentOmitEmpty(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

	// Minimal attachment with required fields (dialog is now required by IETF schema)
	attachment := Attachment{
		PartyIdx:  0,
		DialogIdx: IntPtr(0),
		StartTime: startTime,
	}

	jsonData, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("failed to marshal minimal attachment: %v", err)
	}

	jsonStr := string(jsonData)

	// Required fields should be present
	if !strings.Contains(jsonStr, `"party":0`) {
		t.Error("expected party index to be present")
	}

	if !strings.Contains(jsonStr, `"dialog":0`) {
		t.Error("expected dialog index to be present")
	}

	if !strings.Contains(jsonStr, `"start":`) {
		t.Error("expected start time to be present")
	}

	// Optional fields should be omitted when empty
	unwantedFields := []string{
		"body", "encoding", "url", "content_hash",
		"mediatype", "filename", "purpose",
	}

	for _, unwanted := range unwantedFields {
		if strings.Contains(jsonStr, `"`+unwanted+`":`) {
			t.Errorf("expected empty field %s to be omitted, but found in JSON: %s", unwanted, jsonStr)
		}
	}
}

func TestNewAttachmentWithBase64URL(t *testing.T) {
	// Test creating attachment with base64url-encoded content
	base64urlContent := "SGVsbG8sIFdvcmxkIQ" // base64url encoding of "Hello, World!" (no padding)

	attachment, err := NewAttachment(string(AttachmentTypeDocument), base64urlContent, "base64url")
	if err != nil {
		t.Fatalf("failed to create base64url attachment: %v", err)
	}

	if attachment.Body != base64urlContent {
		t.Errorf("expected body %s, got %s", base64urlContent, attachment.Body)
	}

	if attachment.Encoding != "base64url" {
		t.Errorf("expected encoding base64url, got %s", attachment.Encoding)
	}
}

func TestNewAttachmentWithJSON(t *testing.T) {
	// Test creating attachment with JSON content
	jsonContent := map[string]interface{}{
		"name":        "test document",
		"size":        1024,
		"permissions": []string{"read", "write"},
	}

	attachment, err := NewAttachment(string(AttachmentTypeMetadata), jsonContent, "json")
	if err != nil {
		t.Fatalf("failed to create JSON attachment: %v", err)
	}

	if attachment.Encoding != "json" {
		t.Errorf("expected encoding json, got %s", attachment.Encoding)
	}

	// The body should be a JSON string
	bodyStr := attachment.Body

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &parsed); err != nil {
		t.Errorf("body is not valid JSON: %v", err)
	}
}

func TestAttachmentEncodingValidation(t *testing.T) {
	validEncodings := []string{"base64url", "json", "none"}
	invalidEncodings := []string{"base64", "base32", "hex", "invalid", ""}

	for _, encoding := range validEncodings {
		t.Run("valid_"+encoding, func(t *testing.T) {
			_, err := NewAttachment("document", "content", encoding)
			if err != nil {
				t.Errorf("expected valid encoding %s to work, got error: %v", encoding, err)
			}
		})
	}

	for _, encoding := range invalidEncodings {
		t.Run("invalid_"+encoding, func(t *testing.T) {
			_, err := NewAttachment("document", "content", encoding)
			if err == nil {
				t.Errorf("expected invalid encoding %s to fail", encoding)
			}
		})
	}
}

func TestAttachmentDialogRequired(t *testing.T) {
	// Attachment missing "dialog" should fail schema validation via BuildFromJSON
	jsonStr := `{
		"vcon": "0.4.0",
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"created_at": "2023-01-15T10:30:00Z",
		"parties": [{"name": "Alice"}],
		"dialog": [{"type": "recording", "start": "2023-01-15T10:30:00Z"}],
		"attachments": [
			{
				"body": "data",
				"encoding": "none",
				"party": 0,
				"start": "2023-01-15T10:30:00Z"
			}
		]
	}`
	_, err := BuildFromJSON(jsonStr)
	if err == nil {
		t.Error("expected error for attachment missing required 'dialog' field")
	}
}
