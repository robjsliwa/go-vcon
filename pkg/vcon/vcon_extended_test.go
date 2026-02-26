package vcon

import (
	"encoding/json"
	"testing"
)

func TestProcessProperties(t *testing.T) {
	testObj := map[string]interface{}{
		"allowed_prop":     "value1",
		"another_allowed":  "value2",
		"non_standard":     "value3",
		"custom_field":     "value4",
	}

	allowedProps := map[string]struct{}{
		"allowed_prop":    {},
		"another_allowed": {},
		"meta":           {},
	}

	tests := []struct {
		name     string
		mode     string
		expected map[string]interface{}
	}{
		{
			name: "default mode keeps all properties",
			mode: PropertyHandlingDefault,
			expected: map[string]interface{}{
				"allowed_prop":    "value1",
				"another_allowed": "value2",
				"non_standard":    "value3",
				"custom_field":    "value4",
			},
		},
		{
			name: "strict mode removes non-standard properties",
			mode: PropertyHandlingStrict,
			expected: map[string]interface{}{
				"allowed_prop":    "value1",
				"another_allowed": "value2",
			},
		},
		{
			name: "meta mode moves non-standard to meta",
			mode: PropertyHandlingMeta,
			expected: map[string]interface{}{
				"allowed_prop":    "value1",
				"another_allowed": "value2",
				"meta": map[string]interface{}{
					"non_standard": "value3",
					"custom_field": "value4",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessProperties(testObj, allowedProps, tt.mode)

			// Check all expected keys are present
			for key, expectedValue := range tt.expected {
				if key == "meta" && tt.mode == PropertyHandlingMeta {
					resultMeta, ok := result["meta"].(map[string]interface{})
					if !ok {
						t.Errorf("expected meta to be a map, got %T", result["meta"])
						continue
					}
					expectedMeta := expectedValue.(map[string]interface{})
					for metaKey, metaValue := range expectedMeta {
						if resultMeta[metaKey] != metaValue {
							t.Errorf("expected meta[%s] = %v, got %v", metaKey, metaValue, resultMeta[metaKey])
						}
					}
				} else {
					if result[key] != expectedValue {
						t.Errorf("expected %s = %v, got %v", key, expectedValue, result[key])
					}
				}
			}
		})
	}
}

func TestNewWithPropertyHandling(t *testing.T) {
	domain := "test.example.com"

	tests := []struct {
		name             string
		propertyHandling []string
		expectedHandling string
	}{
		{
			name:             "default handling when not specified",
			propertyHandling: []string{},
			expectedHandling: PropertyHandlingDefault,
		},
		{
			name:             "strict handling",
			propertyHandling: []string{PropertyHandlingStrict},
			expectedHandling: PropertyHandlingStrict,
		},
		{
			name:             "meta handling",
			propertyHandling: []string{PropertyHandlingMeta},
			expectedHandling: PropertyHandlingMeta,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := New(domain, tt.propertyHandling...)

			if vcon.propertyHandling != tt.expectedHandling {
				t.Errorf("expected property handling %s, got %s", tt.expectedHandling, vcon.propertyHandling)
			}

			if vcon.Vcon != SpecVersion {
				t.Errorf("expected vcon version %s, got %s", SpecVersion, vcon.Vcon)
			}

			if vcon.UUID == "" {
				t.Error("expected non-empty UUID")
			}

			if vcon.CreatedAt.IsZero() {
				t.Error("expected non-zero created time")
			}

			if vcon.Parties == nil || vcon.Dialog == nil || vcon.Analysis == nil || vcon.Attachments == nil {
				t.Error("expected non-nil slices")
			}
		})
	}
}

func TestBuildFromJSON(t *testing.T) {
	// Create a valid JSON string
	validVCon := New("test.example.com")
	validVCon.Subject = "Test Subject"
	validVCon.Parties = []Party{
		{Name: "Alice", Tel: "tel:+15551234567"},
		{Name: "Bob", Mailto: "mailto:bob@example.com"},
	}
	validJSON := validVCon.ToJSON()

	tests := []struct {
		name        string
		jsonStr     string
		handling    []string
		expectError bool
		checkFunc   func(*VCon) bool
	}{
		{
			name:        "valid json with default handling",
			jsonStr:     validJSON,
			handling:    []string{},
			expectError: false,
			checkFunc: func(v *VCon) bool {
				return v.Subject == "Test Subject" && len(v.Parties) == 2
			},
		},
		{
			name:        "valid json with strict handling",
			jsonStr:     validJSON,
			handling:    []string{PropertyHandlingStrict},
			expectError: false,
			checkFunc: func(v *VCon) bool {
				return v.propertyHandling == PropertyHandlingStrict
			},
		},
		{
			name:        "invalid json",
			jsonStr:     `{"invalid": json}`,
			handling:    []string{},
			expectError: true,
			checkFunc:   nil,
		},
		{
			name:        "json with custom properties in strict mode",
			jsonStr:     `{"vcon":"0.4.0","uuid":"550e8400-e29b-41d4-a716-446655440000","created_at":"2023-01-15T10:30:00Z","parties":[],"subject":"Test"}`,
			handling:    []string{PropertyHandlingStrict},
			expectError: false,
			checkFunc: func(v *VCon) bool {
				return v.propertyHandling == PropertyHandlingStrict && v.Subject == "Test"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildFromJSON(tt.jsonStr, tt.handling...)

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

			if tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("check function failed for test case")
			}
		})
	}
}

func TestUUID8DomainName(t *testing.T) {
	domain1 := "example.com"
	domain2 := "different.com"

	uuid1 := UUID8DomainName(domain1)
	uuid2 := UUID8DomainName(domain2)

	if uuid1 == uuid2 {
		t.Error("different domains should generate different UUIDs")
	}

	if len(uuid1) != 36 { // Standard UUID format length
		t.Errorf("expected UUID length 36, got %d", len(uuid1))
	}

	uuid1Again := UUID8DomainName(domain1)
	if len(uuid1Again) != 36 {
		t.Errorf("expected UUID length 36 for repeated call, got %d", len(uuid1Again))
	}
}

func TestUUID8Time(t *testing.T) {
	custom1 := uint64(12345)
	custom2 := uint64(67890)

	uuid1 := UUID8Time(custom1)
	uuid2 := UUID8Time(custom2)

	if uuid1 == uuid2 {
		t.Error("different custom bits should generate different UUIDs")
	}

	if len(uuid1) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(uuid1))
	}

	// Test monotonic nature - subsequent calls should generate different UUIDs
	uuid3 := UUID8Time(custom1)
	if uuid1 == uuid3 {
		t.Error("subsequent calls should generate different UUIDs due to timestamp")
	}
}

func TestVConToJSON(t *testing.T) {
	vcon := New("test.example.com")
	vcon.Subject = "Test Subject"
	vcon.Parties = []Party{{Name: "Alice"}}

	jsonStr := vcon.ToJSON()

	if jsonStr == "" {
		t.Error("expected non-empty JSON string")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("ToJSON produced invalid JSON: %v", err)
	}

	if parsed["subject"] != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got %v", parsed["subject"])
	}
}

func TestVConToMap(t *testing.T) {
	vcon := New("test.example.com")
	vcon.Subject = "Test Subject"
	vcon.Parties = []Party{{Name: "Alice"}}

	resultMap := vcon.ToMap()

	if resultMap == nil {
		t.Error("expected non-nil map")
	}

	if resultMap["subject"] != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got %v", resultMap["subject"])
	}

	if resultMap["vcon"] != SpecVersion {
		t.Errorf("expected vcon version %s, got %v", SpecVersion, resultMap["vcon"])
	}
}

func TestVConAddParty(t *testing.T) {
	vcon := New("test.example.com")

	initialLen := len(vcon.Parties)

	party := Party{Name: "Alice", Tel: "tel:+15551234567"}
	index := vcon.AddParty(party)

	if len(vcon.Parties) != initialLen+1 {
		t.Errorf("expected parties length %d, got %d", initialLen+1, len(vcon.Parties))
	}

	if index != initialLen {
		t.Errorf("expected index %d, got %d", initialLen, index)
	}

	if vcon.Parties[index].Name != "Alice" {
		t.Errorf("expected party name 'Alice', got %s", vcon.Parties[index].Name)
	}
}

func TestProcessPropertiesWithNilInput(t *testing.T) {
	result := ProcessProperties(nil, AllowedVConProperties, PropertyHandlingDefault)
	if result != nil {
		t.Errorf("expected nil result for nil input, got %v", result)
	}
}

func TestProcessPropertiesWithExistingMeta(t *testing.T) {
	testObj := map[string]interface{}{
		"allowed_prop":  "value1",
		"custom_field":  "value2",
		"meta": map[string]interface{}{
			"existing_meta": "existing_value",
		},
	}

	allowedProps := map[string]struct{}{
		"allowed_prop": {},
		"meta":        {},
	}

	result := ProcessProperties(testObj, allowedProps, PropertyHandlingMeta)

	metaValue, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected meta to be a map, got %T", result["meta"])
	}

	if metaValue["existing_meta"] != "existing_value" {
		t.Errorf("expected existing_meta to be preserved, got %v", metaValue["existing_meta"])
	}

	if metaValue["custom_field"] != "value2" {
		t.Errorf("expected custom_field to be moved to meta, got %v", metaValue["custom_field"])
	}
}

func TestVConMutualExclusivity(t *testing.T) {
	v := New("test.example.com")
	v.Redacted = &RedactedObject{UUID: "test-uuid", Type: "audio"}
	v.Amended = &AmendedObject{UUID: "test-uuid-2"}

	err := v.Validate()
	if err == nil {
		t.Error("expected validation error for mutually exclusive redacted and amended")
	}
}

func TestVConExtensionsField(t *testing.T) {
	v := New("test.example.com")
	v.Extensions = []string{"CC", "CUSTOM"}

	jsonStr := v.ToJSON()
	var parsed map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &parsed)

	exts, ok := parsed["extensions"].([]interface{})
	if !ok {
		t.Fatal("expected extensions to be an array")
	}
	if len(exts) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(exts))
	}
}

func TestV003Migration(t *testing.T) {
	// A v0.0.3 vCon with old fields: role, base64 encoding, interaction_id, appended, meta
	v003JSON := `{
		"vcon": "0.0.3",
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"created_at": "2023-01-15T10:30:00Z",
		"appended": false,
		"meta": {"custom": "value"},
		"parties": [
			{"name": "Alice", "tel": "tel:+12025551234", "role": "customer", "timezone": "US/Eastern"},
			{"name": "Bob", "tel": "tel:+12025555678", "role": "agent", "contact_list": "VIP"}
		],
		"dialog": [
			{
				"type": "text",
				"start": "2023-01-15T10:30:00Z",
				"parties": [0, 1],
				"body": "Hello",
				"encoding": "base64",
				"mediatype": "text/plain",
				"alg": "sha256",
				"signature": "abc123",
				"campaign": "summer",
				"interaction_type": "inbound",
				"interaction_id": "INT-1",
				"skill": "billing",
				"meta": {"dialog_meta": true}
			}
		],
		"analysis": [
			{
				"type": "sentiment",
				"body": "positive",
				"encoding": "base64",
				"meta": {"analysis_meta": true}
			}
		],
		"attachments": [
			{
				"body": "data",
				"encoding": "base64",
				"party": 0,
				"start": "2023-01-15T10:30:00Z",
				"meta": {"att_meta": true}
			}
		]
	}`

	v, err := BuildFromJSON(v003JSON)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Version should be migrated to 0.4.0
	if v.Vcon != "0.4.0" {
		t.Errorf("expected version 0.4.0, got %s", v.Vcon)
	}

	// Parties should not have role/contact_list/timezone (removed in migration)
	// Since these fields don't exist on Party struct anymore, they're just dropped
	if len(v.Parties) != 2 {
		t.Errorf("expected 2 parties, got %d", len(v.Parties))
	}

	// Dialog encoding should be migrated from "base64" to "base64url"
	if len(v.Dialog) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(v.Dialog))
	}
	if v.Dialog[0].Encoding != "base64url" {
		t.Errorf("expected encoding base64url after migration, got %s", v.Dialog[0].Encoding)
	}

	// CC extension fields should be removed from dialog
	// (campaign, interaction_type, interaction_id, skill are no longer on Dialog struct)

	// Analysis encoding should be migrated
	if len(v.Analysis) != 1 {
		t.Fatalf("expected 1 analysis, got %d", len(v.Analysis))
	}

	// Attachments encoding should be migrated
	if len(v.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(v.Attachments))
	}
	if v.Attachments[0].Encoding != "base64url" {
		t.Errorf("expected attachment encoding base64url after migration, got %s", v.Attachments[0].Encoding)
	}
}

func TestV003MigrationFromFile(t *testing.T) {
	// Test migration using the comprehensive v0.0.3 fixture
	v, err := LoadFromFile("../../testdata/sample_vcons/comprehensive-vcon.json")
	if err != nil {
		t.Fatalf("load v0.0.3 fixture failed: %v", err)
	}

	// Should have been migrated to 0.4.0
	if v.Vcon != "0.4.0" {
		t.Errorf("expected version 0.4.0, got %s", v.Vcon)
	}

	// Dialogs should have encoding migrated
	for i, d := range v.Dialog {
		if d.Encoding == "base64" {
			t.Errorf("dialog[%d] encoding should not be 'base64' after migration", i)
		}
	}
}
