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
			name:         "valid base64 encoding",
			attachType:   "document",
			body:         "SGVsbG8gV29ybGQ=",
			encoding:     "base64",
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
	expected := []string{"base64", "base64url", "json", "none"}
	
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
		DialogIdx:   1,
		PartyIdx:    0,
		StartTime:   startTime,
		MediaType:   "text/plain",
		Filename:    "test.txt",
		ContentHash: "sha256hash",
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

	if unmarshaled.DialogIdx != attachment.DialogIdx {
		t.Errorf("expected dialog index %d, got %d", attachment.DialogIdx, unmarshaled.DialogIdx)
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
}

func TestAttachmentWithURL(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	attachment := Attachment{
		URL:         "https://example.com/document.pdf",
		DialogIdx:   0,
		PartyIdx:    1,
		StartTime:   startTime,
		MediaType:   "application/pdf",
		Filename:    "document.pdf",
		ContentHash: "sha256hash",
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

func TestAttachmentWithMeta(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	meta := map[string]interface{}{
		"custom_field": "custom_value",
		"priority":     "high",
		"tags":         []string{"important", "document"},
	}

	attachment := Attachment{
		Body:      "content",
		Encoding:  "none",
		PartyIdx:  0,
		StartTime: startTime,
		Meta:      meta,
	}

	// Test marshaling
	jsonData, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("failed to marshal attachment with meta: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Attachment
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal attachment with meta: %v", err)
	}

	if unmarshaled.Meta == nil {
		t.Fatal("expected meta to be preserved")
	}

	metaMap, ok := unmarshaled.Meta.(map[string]interface{})
	if !ok {
		t.Fatalf("expected meta to be a map, got %T", unmarshaled.Meta)
	}

	if metaMap["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field to be preserved in meta")
	}
}

func TestAttachmentOmitEmpty(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	// Minimal attachment with only required fields
	attachment := Attachment{
		PartyIdx:  0,
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

	if !strings.Contains(jsonStr, `"start":`) {
		t.Error("expected start time to be present")
	}

	// Optional fields should be omitted when empty
	unwantedFields := []string{
		"body", "encoding", "url", "content_hash", "dialog", 
		"mediatype", "filename", "meta",
	}

	for _, unwanted := range unwantedFields {
		if strings.Contains(jsonStr, `"`+unwanted+`":`) {
			t.Errorf("expected empty field %s to be omitted, but found in JSON: %s", unwanted, jsonStr)
		}
	}
}

func TestNewAttachmentWithBase64(t *testing.T) {
	// Test creating attachment with base64-encoded content
	base64Content := "SGVsbG8sIFdvcmxkIQ==" // base64 encoding of "Hello, World!"

	attachment, err := NewAttachment(string(AttachmentTypeDocument), base64Content, "base64")
	if err != nil {
		t.Fatalf("failed to create base64 attachment: %v", err)
	}

	if attachment.Body != base64Content {
		t.Errorf("expected body %s, got %s", base64Content, attachment.Body)
	}

	if attachment.Encoding != "base64" {
		t.Errorf("expected encoding base64, got %s", attachment.Encoding)
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
	validEncodings := []string{"base64", "base64url", "json", "none"}
	invalidEncodings := []string{"base32", "hex", "invalid", ""}

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
