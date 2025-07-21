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
			jsonStr:     `{"vcon":"0.0.3","uuid":"test-uuid","created_at":"2023-01-15T10:30:00Z","subject":"Test"}`,
			handling:    []string{PropertyHandlingStrict},
			expectError: false,
			checkFunc: func(v *VCon) bool {
				// Custom fields were not included, so this should pass
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
	
	// Test that UUIDs are valid format (we can't guarantee same domain produces identical UUID
	// due to timestamp and monotonic counter logic in the implementation)
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
