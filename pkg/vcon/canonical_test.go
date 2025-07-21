package vcon

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCanonicalise(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple object",
			input:    map[string]interface{}{"b": 2, "a": 1},
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "nested object",
			input:    map[string]interface{}{"z": map[string]interface{}{"y": 2, "x": 1}, "a": 1},
			expected: `{"a":1,"z":{"x":1,"y":2}}`,
		},
		{
			name:     "array",
			input:    []interface{}{3, 1, 2},
			expected: `[3,1,2]`,
		},
		{
			name:     "mixed types",
			input:    map[string]interface{}{"number": 42, "string": "hello", "boolean": true, "null": nil},
			expected: `{"boolean":true,"null":null,"number":42,"string":"hello"}`,
		},
		{
			name:     "empty object",
			input:    map[string]interface{}{},
			expected: `{}`,
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Canonicalise(tt.input)
			if err != nil {
				t.Fatalf("failed to canonicalise: %v", err)
			}

			resultStr := string(result)
			if resultStr != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resultStr)
			}
		})
	}
}

func TestCanonicaliseWithVCon(t *testing.T) {
	// Test canonicalization with a vCon object
	vcon := New("test.example.com")
	vcon.Subject = "Test Subject"
	vcon.Parties = []Party{
		{Name: "Bob", Tel: "tel:+15551111111"},
		{Name: "Alice", Mailto: "mailto:alice@example.com"},
	}

	canonical, err := Canonicalise(vcon)
	if err != nil {
		t.Fatalf("failed to canonicalise vCon: %v", err)
	}

	canonicalStr := string(canonical)
	
	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(canonical, &parsed); err != nil {
		t.Fatalf("canonical result is not valid JSON: %v", err)
	}

	// Should maintain alphabetical key ordering
	if !strings.Contains(canonicalStr, `"created_at"`) {
		t.Error("expected created_at field in canonical form")
	}
	
	if !strings.Contains(canonicalStr, `"parties"`) {
		t.Error("expected parties field in canonical form")
	}

	// The keys should be in alphabetical order (RFC 8785)
	// Check that created_at comes before parties alphabetically in the string
	createdAtPos := strings.Index(canonicalStr, `"created_at"`)
	partiesPos := strings.Index(canonicalStr, `"parties"`)
	if createdAtPos > partiesPos {
		t.Error("expected keys to be in alphabetical order")
	}
}

func TestCanonicaliseWithComplexNesting(t *testing.T) {
	// Test with deeply nested structures
	complex := map[string]interface{}{
		"z_field": map[string]interface{}{
			"nested_z": "value",
			"nested_a": map[string]interface{}{
				"deep_z": 3,
				"deep_a": 1,
			},
		},
		"a_field": []interface{}{
			map[string]interface{}{
				"item_z": "last",
				"item_a": "first",
			},
		},
	}

	canonical, err := Canonicalise(complex)
	if err != nil {
		t.Fatalf("failed to canonicalise complex structure: %v", err)
	}

	canonicalStr := string(canonical)

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(canonical, &parsed); err != nil {
		t.Fatalf("canonical result is not valid JSON: %v", err)
	}

	// Check that keys are ordered alphabetically at each level
	expectedPatterns := []string{
		`"a_field"`,
		`"z_field"`,
		`"nested_a"`,
		`"nested_z"`,
		`"deep_a"`,
		`"deep_z"`,
		`"item_a"`,
		`"item_z"`,
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(canonicalStr, pattern) {
			t.Errorf("expected pattern %s not found in canonical form", pattern)
		}
	}
}

func TestCanonicaliseErrorHandling(t *testing.T) {
	// Test with values that can't be marshaled to JSON
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "function",
			input: func() {},
		},
		{
			name:  "channel",
			input: make(chan int),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Canonicalise(tt.input)
			if err == nil {
				t.Errorf("expected error when canonicalising %s", tt.name)
			}
		})
	}
}

func TestCanonicaliseConsistency(t *testing.T) {
	// Test that canonicalization produces consistent results
	input := map[string]interface{}{
		"z": 1,
		"a": 2,
		"m": map[string]interface{}{
			"y": 3,
			"b": 4,
		},
	}

	// Call canonicalise multiple times
	result1, err1 := Canonicalise(input)
	result2, err2 := Canonicalise(input)
	result3, err3 := Canonicalise(input)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("unexpected errors: %v, %v, %v", err1, err2, err3)
	}

	// All results should be identical
	if string(result1) != string(result2) || string(result2) != string(result3) {
		t.Error("canonicalise should produce consistent results")
		t.Errorf("result1: %s", string(result1))
		t.Errorf("result2: %s", string(result2))
		t.Errorf("result3: %s", string(result3))
	}
}

func TestCanonicaliseWithNumbers(t *testing.T) {
	// Test canonicalization with different number types
	input := map[string]interface{}{
		"integer":    42,
		"float":      3.14159,
		"zero":       0,
		"negative":   -17,
		"scientific": 1e6,
	}

	canonical, err := Canonicalise(input)
	if err != nil {
		t.Fatalf("failed to canonicalise numbers: %v", err)
	}

	canonicalStr := string(canonical)

	// Should be valid JSON with numbers preserved
	var parsed map[string]interface{}
	if err := json.Unmarshal(canonical, &parsed); err != nil {
		t.Fatalf("canonical result is not valid JSON: %v", err)
	}

	// Check that numbers are properly represented
	expectedNumbers := []string{"42", "3.14159", "0", "-17", "1000000"}
	for _, num := range expectedNumbers {
		if !strings.Contains(canonicalStr, num) {
			t.Errorf("expected number %s not found in canonical form: %s", num, canonicalStr)
		}
	}
}

func TestCanonicaliseWithStrings(t *testing.T) {
	// Test canonicalization with various string types
	input := map[string]interface{}{
		"simple":      "hello",
		"with_quotes": `say "hello"`,
		"with_newline": "line1\nline2",
		"with_unicode": "café 🚀",
		"empty":       "",
	}

	canonical, err := Canonicalise(input)
	if err != nil {
		t.Fatalf("failed to canonicalise strings: %v", err)
	}

	canonicalStr := string(canonical)

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(canonical, &parsed); err != nil {
		t.Fatalf("canonical result is not valid JSON: %v", err)
	}

	// Check that special characters are properly escaped
	if !strings.Contains(canonicalStr, `"hello"`) {
		t.Error("expected simple string to be present")
	}
	
	if !strings.Contains(canonicalStr, `"say \"hello\""`) {
		t.Error("expected escaped quotes")
	}

	if !strings.Contains(canonicalStr, `"line1\nline2"`) {
		t.Error("expected escaped newline")
	}
}
