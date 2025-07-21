package vcon

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDialogTypes(t *testing.T) {
	tests := []struct {
		dialogType string
		valid      bool
	}{
		{"recording", true},
		{"text", true},
		{"transfer", true},
		{"incomplete", true},
		{"email", true},
		{"chat", true},
		{"invalid_type", true}, // Any string is valid for dialog type
	}

	for _, tt := range tests {
		t.Run(tt.dialogType, func(t *testing.T) {
			dialog := Dialog{
				Type:      tt.dialogType,
				StartTime: &time.Time{},
			}

			jsonData, err := json.Marshal(dialog)
			if err != nil {
				t.Errorf("failed to marshal dialog: %v", err)
			}

			var unmarshaled Dialog
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Errorf("failed to unmarshal dialog: %v", err)
			}

			if unmarshaled.Type != tt.dialogType {
				t.Errorf("expected type %s, got %s", tt.dialogType, unmarshaled.Type)
			}
		})
	}
}

func TestDialogMIMETypes(t *testing.T) {
	tests := []struct {
		mimeType string
		category string
	}{
		{MIMETypePlainText, "text"},
		{MIMETypeAudioWav, "audio"},
		{MIMETypeAudioMpeg, "audio"},
		{MIMETypeVideoMP4, "video"},
		{MIMETypeRFC822, "message"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {		dialog := Dialog{
			Type:      "recording",
			StartTime: &time.Time{},
			MediaType: tt.mimeType,
		}

		_, err := json.Marshal(dialog)
		if err != nil {
			t.Errorf("failed to marshal dialog: %v", err)
		}

			if tt.category == "audio" {
				found := false
				for _, audioType := range AudioMIMETypes {
					if audioType == tt.mimeType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("MIME type %s should be in AudioMIMETypes", tt.mimeType)
				}
			}

			if tt.category == "video" {
				found := false
				for _, videoType := range VideoMIMETypes {
					if videoType == tt.mimeType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("MIME type %s should be in VideoMIMETypes", tt.mimeType)
				}
			}
		})
	}
}

func TestDialogWithBody(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	dialog := Dialog{
		Type:      "text",
		StartTime: &startTime,
		Body:      "This is a test message",
		MediaType: MIMETypePlainText,
		Parties:   []int{0, 1},
	}

	jsonData, err := json.Marshal(dialog)
	if err != nil {
		t.Fatalf("failed to marshal dialog: %v", err)
	}

	var unmarshaled Dialog
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal dialog: %v", err)
	}

	if unmarshaled.Body != dialog.Body {
		t.Errorf("expected body %s, got %s", dialog.Body, unmarshaled.Body)
	}

	if unmarshaled.MediaType != dialog.MediaType {
		t.Errorf("expected mediatype %s, got %s", dialog.MediaType, unmarshaled.MediaType)
	}
}

func TestDialogWithURL(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	dialog := Dialog{
		Type:        "recording",
		StartTime:   &startTime,
		URL:         "https://example.com/recording.wav",
		Filename:    "recording.wav",
		MediaType:   MIMETypeAudioWav,
		Duration:    120.5,
		ContentHash: "sha256hash",
		Algorithm:   "sha256",
	}

	jsonData, err := json.Marshal(dialog)
	if err != nil {
		t.Fatalf("failed to marshal dialog: %v", err)
	}

	var unmarshaled Dialog
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal dialog: %v", err)
	}

	if unmarshaled.URL != dialog.URL {
		t.Errorf("expected URL %s, got %s", dialog.URL, unmarshaled.URL)
	}

	if unmarshaled.Duration != dialog.Duration {
		t.Errorf("expected duration %f, got %f", dialog.Duration, unmarshaled.Duration)
	}

	if unmarshaled.ContentHash != dialog.ContentHash {
		t.Errorf("expected content_hash %s, got %s", dialog.ContentHash, unmarshaled.ContentHash)
	}
}

func TestDialogTransfer(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	dialog := Dialog{
		Type:           "transfer",
		StartTime:      &startTime,
		Transferee:     0,
		Transferor:     1,
		TransferTarget: 2,
		Original:       3,
		Consultation:   4,
		TargetDialog:   5,
	}

	jsonData, err := json.Marshal(dialog)
	if err != nil {
		t.Fatalf("failed to marshal transfer dialog: %v", err)
	}

	var unmarshaled Dialog
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal transfer dialog: %v", err)
	}

	if unmarshaled.Transferee != dialog.Transferee {
		t.Errorf("expected transferee %d, got %d", dialog.Transferee, unmarshaled.Transferee)
	}

	if unmarshaled.Transferor != dialog.Transferor {
		t.Errorf("expected transferor %d, got %d", dialog.Transferor, unmarshaled.Transferor)
	}

	if unmarshaled.TransferTarget != dialog.TransferTarget {
		t.Errorf("expected transfer_target %d, got %d", dialog.TransferTarget, unmarshaled.TransferTarget)
	}
}

func TestDialogWithPartyHistory(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	history := []PartyHistory{
		{
			Party: 0,
			Event: string(PartyEventJoin),
			Time:  startTime,
		},
		{
			Party: 1,
			Event: string(PartyEventJoin),
			Time:  startTime.Add(5 * time.Second),
		},
		{
			Party: 0,
			Event: string(PartyEventDrop),
			Time:  startTime.Add(60 * time.Second),
		},
	}

	dialog := Dialog{
		Type:         "conference",
		StartTime:    &startTime,
		Duration:     65.0,
		Parties:      []int{0, 1},
		PartyHistory: history,
	}

	jsonData, err := json.Marshal(dialog)
	if err != nil {
		t.Fatalf("failed to marshal dialog with party history: %v", err)
	}

	var unmarshaled Dialog
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal dialog with party history: %v", err)
	}

	if len(unmarshaled.PartyHistory) != 3 {
		t.Errorf("expected 3 party history entries, got %d", len(unmarshaled.PartyHistory))
	}

	if unmarshaled.PartyHistory[0].Event != string(PartyEventJoin) {
		t.Errorf("expected first event to be join, got %s", unmarshaled.PartyHistory[0].Event)
	}

	if unmarshaled.PartyHistory[2].Event != string(PartyEventDrop) {
		t.Errorf("expected last event to be drop, got %s", unmarshaled.PartyHistory[2].Event)
	}
}

func TestDialogWithEncoding(t *testing.T) {
	tests := []string{"base64", "base64url", "json", "none"}
	
	for _, encoding := range tests {
		t.Run(encoding, func(t *testing.T) {
			startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
			
			dialog := Dialog{
				Type:      "text",
				StartTime: &startTime,
				Body:      "encoded content",
				Encoding:  encoding,
				MediaType: MIMETypePlainText,
			}

			jsonData, err := json.Marshal(dialog)
			if err != nil {
				t.Fatalf("failed to marshal dialog with encoding %s: %v", encoding, err)
			}

			var unmarshaled Dialog
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal dialog with encoding %s: %v", encoding, err)
			}

			if unmarshaled.Encoding != encoding {
				t.Errorf("expected encoding %s, got %s", encoding, unmarshaled.Encoding)
			}

			// Check that encoding is in ValidEncodings
			found := false
			for _, validEnc := range ValidEncodings {
				if validEnc == encoding {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("encoding %s should be in ValidEncodings", encoding)
			}
		})
	}
}

func TestDialogPartiesInterface(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	tests := []struct {
		name    string
		parties interface{}
	}{
		{
			name:    "single party as int",
			parties: 0,
		},
		{
			name:    "multiple parties as slice",
			parties: []int{0, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := Dialog{
				Type:      "text",
				StartTime: &startTime,
				Parties:   tt.parties,
			}

			jsonData, err := json.Marshal(dialog)
			if err != nil {
				t.Fatalf("failed to marshal dialog: %v", err)
			}

			var unmarshaled Dialog
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal dialog: %v", err)
			}

			// The unmarshaled parties will be a float64 or []interface{} due to JSON unmarshaling
			if unmarshaled.Parties == nil {
				t.Error("expected parties to be preserved")
			}
		})
	}
}

func TestDialogOmitEmpty(t *testing.T) {
	startTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	
	// Test minimal dialog
	dialog := Dialog{
		Type:      "text",
		StartTime: &startTime,
	}

	jsonData, err := json.Marshal(dialog)
	if err != nil {
		t.Fatalf("failed to marshal minimal dialog: %v", err)
	}

	jsonStr := string(jsonData)

	// Required fields should be present
	if !strings.Contains(jsonStr, `"type":"text"`) {
		t.Error("expected type to be present")
	}

	if !strings.Contains(jsonStr, `"start":`) {
		t.Error("expected start time to be present")
	}

	// Optional fields should be omitted when empty
	unwantedFields := []string{
		"duration", "parties", "originator", "mediatype", "filename",
		"body", "encoding", "url", "content_hash", "alg", "signature",
		"disposition", "party_history", "transferee", "transferor",
		"transfer_target", "original", "consultation", "target_dialog",
	}

	for _, unwanted := range unwantedFields {
		if strings.Contains(jsonStr, `"`+unwanted+`":`) {
			t.Errorf("expected empty field %s to be omitted, but found in JSON: %s", unwanted, jsonStr)
		}
	}
}

func TestSupportedMIMETypes(t *testing.T) {
	expectedTypes := []string{
		MIMETypePlainText,
		MIMETypeAudioWav,
		MIMETypeAudioWav2,
		MIMETypeAudioWave,
		MIMETypeAudioMpeg,
		MIMETypeAudioMP3,
		MIMETypeAudioOgg,
		MIMETypeAudioWebm,
		MIMETypeAudioM4a,
		MIMETypeAudioAAC,
		MIMETypeVideoMP4,
		MIMETypeVideoOgg,
		MIMETypeMultipart,
		MIMETypeRFC822,
	}

	if len(SupportedMIMETypes) != len(expectedTypes) {
		t.Errorf("expected %d supported MIME types, got %d", len(expectedTypes), len(SupportedMIMETypes))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, supportedType := range SupportedMIMETypes {
			if supportedType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MIME type %s to be in SupportedMIMETypes", expectedType)
		}
	}
}

func TestValidEncodings(t *testing.T) {
	expected := []string{"base64", "base64url", "json", "none"}
	
	if len(ValidEncodings) != len(expected) {
		t.Errorf("expected %d valid encodings, got %d", len(expected), len(ValidEncodings))
	}

	for _, expectedEnc := range expected {
		found := false
		for _, validEnc := range ValidEncodings {
			if validEnc == expectedEnc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected encoding %s to be in ValidEncodings", expectedEnc)
		}
	}
}
