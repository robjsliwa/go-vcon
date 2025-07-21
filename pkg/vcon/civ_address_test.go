package vcon

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewCivicAddress(t *testing.T) {
	addr := NewCivicAddress()
	if addr == nil {
		t.Fatal("NewCivicAddress should return non-nil address")
	}

	// All fields should be empty by default
	if addr.Country != "" {
		t.Errorf("expected empty Country, got %s", addr.Country)
	}
	if addr.A1 != "" {
		t.Errorf("expected empty A1, got %s", addr.A1)
	}
	if addr.PC != "" {
		t.Errorf("expected empty PC, got %s", addr.PC)
	}
}

func TestCivicAddressSerialization(t *testing.T) {
	addr := &CivicAddress{
		Country: "US",
		A1:      "CA", // State
		A3:      "San Francisco", // City
		PC:      "94105", // Postal code
		STS:     "Main Street", // Street
		HNO:     "123", // House number
	}

	// Test marshaling
	jsonData, err := json.Marshal(addr)
	if err != nil {
		t.Fatalf("failed to marshal civic address: %v", err)
	}

	jsonStr := string(jsonData)
	expectedFields := []string{
		`"country":"US"`,
		`"a1":"CA"`,
		`"a3":"San Francisco"`,
		`"pc":"94105"`,
		`"sts":"Main Street"`,
		`"hno":"123"`,
	}

	for _, expected := range expectedFields {
		if !strings.Contains(jsonStr, expected) {
			t.Errorf("expected JSON to contain %s, got %s", expected, jsonStr)
		}
	}

	// Test unmarshaling
	var unmarshaled CivicAddress
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal civic address: %v", err)
	}

	if unmarshaled.Country != addr.Country {
		t.Errorf("expected Country %s, got %s", addr.Country, unmarshaled.Country)
	}
	if unmarshaled.A1 != addr.A1 {
		t.Errorf("expected A1 %s, got %s", addr.A1, unmarshaled.A1)
	}
	if unmarshaled.A3 != addr.A3 {
		t.Errorf("expected A3 %s, got %s", addr.A3, unmarshaled.A3)
	}
	if unmarshaled.PC != addr.PC {
		t.Errorf("expected PC %s, got %s", addr.PC, unmarshaled.PC)
	}
}

func TestCivicAddressOmitEmpty(t *testing.T) {
	// Test that empty fields are omitted from JSON
	addr := &CivicAddress{
		Country: "US",
		A3:      "New York",
		// Other fields left empty
	}

	jsonData, err := json.Marshal(addr)
	if err != nil {
		t.Fatalf("failed to marshal civic address: %v", err)
	}

	jsonStr := string(jsonData)

	// Present fields
	if !strings.Contains(jsonStr, `"country":"US"`) {
		t.Error("expected country to be present")
	}
	if !strings.Contains(jsonStr, `"a3":"New York"`) {
		t.Error("expected a3 to be present")
	}

	// Absent fields should not be present
	unwantedFields := []string{
		"a1", "a2", "a4", "a5", "a6", "prd", "pod", "sts", 
		"hno", "hns", "lmk", "loc", "flr", "nam", "pc",
	}

	for _, unwanted := range unwantedFields {
		if strings.Contains(jsonStr, `"`+unwanted+`":`) {
			t.Errorf("expected empty field %s to be omitted, but found in JSON: %s", unwanted, jsonStr)
		}
	}
}

func TestCivicAddressToMap(t *testing.T) {
	addr := &CivicAddress{
		Country: "CA",
		A1:      "ON", // Province
		A3:      "Toronto", // City
		PC:      "M5V 3M8", // Postal code
		STS:     "King Street West", // Street
		HNO:     "290", // House number
		FLR:     "15", // Floor
		NAM:     "CN Tower Area", // Location name
	}

	resultMap := addr.ToMap()
	
	expectedEntries := map[string]string{
		"country": "CA",
		"a1":      "ON",
		"a3":      "Toronto", 
		"pc":      "M5V 3M8",
		"sts":     "King Street West",
		"hno":     "290",
		"flr":     "15",
		"nam":     "CN Tower Area",
	}

	for key, expectedValue := range expectedEntries {
		if value, exists := resultMap[key]; !exists {
			t.Errorf("expected key %s to be present in map", key)
		} else if value != expectedValue {
			t.Errorf("expected %s = %s, got %s", key, expectedValue, value)
		}
	}

	// Empty fields should not be in the map
	emptyFields := []string{"a2", "a4", "a5", "a6", "prd", "pod", "hns", "lmk", "loc"}
	for _, field := range emptyFields {
		if _, exists := resultMap[field]; exists {
			t.Errorf("expected empty field %s to not be in map", field)
		}
	}
}

func TestCivicAddressToMapWithEmptyAddress(t *testing.T) {
	addr := NewCivicAddress()
	resultMap := addr.ToMap()
	
	if len(resultMap) != 0 {
		t.Errorf("expected empty map for empty address, got %d entries", len(resultMap))
	}
}

func TestCivicAddressAllFields(t *testing.T) {
	// Test all possible fields
	addr := &CivicAddress{
		Country: "US",
		A1:      "NY",
		A2:      "New York County",
		A3:      "New York",
		A4:      "Manhattan",
		A5:      "10001",
		A6:      "Empire State Building",
		PRD:     "Suite 1000",
		POD:     "PO Box 123",
		STS:     "5th Avenue",
		HNO:     "350",
		HNS:     "Empire State Building",
		LMK:     "Near Times Square",
		LOC:     "Midtown",
		FLR:     "86",
		NAM:     "Observation Deck",
		PC:      "10118",
	}

	// Test ToMap includes all fields
	resultMap := addr.ToMap()
	expectedCount := 17 // All fields should be present

	if len(resultMap) != expectedCount {
		t.Errorf("expected %d fields in map, got %d", expectedCount, len(resultMap))
	}

	// Test JSON serialization includes all fields
	jsonData, err := json.Marshal(addr)
	if err != nil {
		t.Fatalf("failed to marshal full civic address: %v", err)
	}

	jsonStr := string(jsonData)
	allFields := []string{
		"country", "a1", "a2", "a3", "a4", "a5", "a6",
		"prd", "pod", "sts", "hno", "hns", "lmk", "loc", "flr", "nam", "pc",
	}

	for _, field := range allFields {
		if !strings.Contains(jsonStr, `"`+field+`":`) {
			t.Errorf("expected field %s to be in JSON", field)
		}
	}
}

func TestCivicAddressInternationalAddresses(t *testing.T) {
	tests := []struct {
		name string
		addr *CivicAddress
	}{
		{
			name: "UK Address",
			addr: &CivicAddress{
				Country: "GB",
				A3:      "London",
				PC:      "SW1A 1AA",
				STS:     "Buckingham Palace Road",
				HNO:     "1",
			},
		},
		{
			name: "Japanese Address",
			addr: &CivicAddress{
				Country: "JP",
				A1:      "Tokyo",
				A3:      "Shibuya",
				PC:      "150-0002",
				STS:     "Shibuya",
			},
		},
		{
			name: "German Address",
			addr: &CivicAddress{
				Country: "DE",
				A1:      "Bavaria",
				A3:      "Munich",
				PC:      "80539",
				STS:     "Marienplatz",
				HNO:     "8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON round trip
			jsonData, err := json.Marshal(tt.addr)
			if err != nil {
				t.Fatalf("failed to marshal %s address: %v", tt.name, err)
			}

			var unmarshaled CivicAddress
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal %s address: %v", tt.name, err)
			}

			if unmarshaled.Country != tt.addr.Country {
				t.Errorf("%s: expected Country %s, got %s", tt.name, tt.addr.Country, unmarshaled.Country)
			}

			// Test ToMap conversion
			resultMap := tt.addr.ToMap()
			if resultMap["country"] != tt.addr.Country {
				t.Errorf("%s: expected country in map to be %s, got %s", tt.name, tt.addr.Country, resultMap["country"])
			}
		})
	}
}

func TestCivicAddressPartialData(t *testing.T) {
	// Test addresses with only some fields populated (common in real usage)
	tests := []struct {
		name string
		addr *CivicAddress
	}{
		{
			name: "Country and City only",
			addr: &CivicAddress{
				Country: "FR",
				A3:      "Paris",
			},
		},
		{
			name: "Street address only",
			addr: &CivicAddress{
				STS: "Champs-Élysées",
				HNO: "101",
			},
		},
		{
			name: "Postal code only",
			addr: &CivicAddress{
				PC: "75008",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should successfully serialize and deserialize
			jsonData, err := json.Marshal(tt.addr)
			if err != nil {
				t.Fatalf("failed to marshal %s: %v", tt.name, err)
			}

			var unmarshaled CivicAddress
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal %s: %v", tt.name, err)
			}

			// ToMap should only include non-empty fields
			resultMap := tt.addr.ToMap()
			
			// Count non-empty fields in original
			expectedEntries := 0
			if tt.addr.Country != "" { expectedEntries++ }
			if tt.addr.A3 != "" { expectedEntries++ }
			if tt.addr.STS != "" { expectedEntries++ }
			if tt.addr.HNO != "" { expectedEntries++ }
			if tt.addr.PC != "" { expectedEntries++ }

			if len(resultMap) != expectedEntries {
				t.Errorf("%s: expected %d entries in map, got %d", tt.name, expectedEntries, len(resultMap))
			}
		})
	}
}
