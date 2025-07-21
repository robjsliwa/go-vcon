package vcon

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPartyEventType(t *testing.T) {
	tests := []struct {
		eventType PartyEventType
		expected  string
	}{
		{PartyEventJoin, "join"},
		{PartyEventDrop, "drop"},
		{PartyEventHold, "hold"},
		{PartyEventUnhold, "unhold"},
		{PartyEventMute, "mute"},
		{PartyEventUnmute, "unmute"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.eventType))
			}
		})
	}
}

func TestPartyJSONSerialization(t *testing.T) {
	party := Party{
		Tel:      "tel:+15551234567",
		Mailto:   "mailto:alice@example.com",
		Name:     "Alice Smith",
		Role:     "originator",
		UUID:     "test-uuid-123",
		Timezone: "America/New_York",
	}

	// Test marshaling
	jsonData, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("failed to marshal party: %v", err)
	}

	jsonStr := string(jsonData)
	expectedFields := []string{
		`"tel":"tel:+15551234567"`,
		`"mailto":"mailto:alice@example.com"`,
		`"name":"Alice Smith"`,
		`"role":"originator"`,
		`"uuid":"test-uuid-123"`,
		`"timezone":"America/New_York"`,
	}

	for _, expected := range expectedFields {
		if !contains(jsonStr, expected) {
			t.Errorf("expected JSON to contain %s, got %s", expected, jsonStr)
		}
	}

	// Test unmarshaling
	var unmarshaled Party
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal party: %v", err)
	}

	if unmarshaled.Tel != party.Tel {
		t.Errorf("expected Tel %s, got %s", party.Tel, unmarshaled.Tel)
	}
	if unmarshaled.Mailto != party.Mailto {
		t.Errorf("expected Mailto %s, got %s", party.Mailto, unmarshaled.Mailto)
	}
	if unmarshaled.Name != party.Name {
		t.Errorf("expected Name %s, got %s", party.Name, unmarshaled.Name)
	}
}

func TestPartyWithCivicAddress(t *testing.T) {
	civicAddr := &CivicAddress{
		Country: "US",
		PC:      "12345",
		A3:      "New York",
		STS:     "123 Main St",
	}

	party := Party{
		Name:         "Bob Johnson",
		CivicAddress: civicAddr,
	}

	// Test marshaling
	jsonData, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("failed to marshal party with civic address: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Party
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal party with civic address: %v", err)
	}

	if unmarshaled.CivicAddress == nil {
		t.Fatal("expected CivicAddress to be preserved")
	}

	if unmarshaled.CivicAddress.Country != civicAddr.Country {
		t.Errorf("expected Country %s, got %s", civicAddr.Country, unmarshaled.CivicAddress.Country)
	}
}

func TestPartyOmitEmpty(t *testing.T) {
	// Test that empty fields are omitted
	party := Party{
		Name: "Minimal Party",
	}

	jsonData, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("failed to marshal minimal party: %v", err)
	}

	jsonStr := string(jsonData)

	// These fields should not be present when empty
	unwantedFields := []string{
		"tel", "stir", "mailto", "validation", "gmlpos",
		"civicaddress", "timezone", "uuid", "role", "contact_list",
	}

	for _, unwanted := range unwantedFields {
		if contains(jsonStr, `"`+unwanted+`":`) {
			t.Errorf("expected empty field %s to be omitted, but found in JSON: %s", unwanted, jsonStr)
		}
	}

	// Name should be present
	if !contains(jsonStr, `"name":"Minimal Party"`) {
		t.Errorf("expected name to be present in JSON: %s", jsonStr)
	}
}

func TestPartyHistory(t *testing.T) {
	// This tests the PartyHistory structure that would be used in Dialog
	history := []PartyHistory{
		{
			Party: 0,
			Event: string(PartyEventJoin),
			Time:  time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			Party: 0,
			Event: string(PartyEventDrop),
			Time:  time.Date(2023, 1, 15, 10, 45, 0, 0, time.UTC),
		},
	}

	// Test marshaling
	jsonData, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("failed to marshal party history: %v", err)
	}

	// Test unmarshaling
	var unmarshaled []PartyHistory
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal party history: %v", err)
	}

	if len(unmarshaled) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(unmarshaled))
	}

	if unmarshaled[0].Event != string(PartyEventJoin) {
		t.Errorf("expected first event to be join, got %s", unmarshaled[0].Event)
	}

	if unmarshaled[1].Event != string(PartyEventDrop) {
		t.Errorf("expected second event to be drop, got %s", unmarshaled[1].Event)
	}
}

func TestPartyValidation(t *testing.T) {
	// Test that parties with different validation states work correctly
	parties := []Party{
		{
			Name:       "Verified User",
			Tel:        "tel:+15551234567",
			Validation: "verified",
		},
		{
			Name:       "Unverified User",
			Mailto:     "mailto:user@example.com",
			Validation: "unverified",
		},
		{
			Name: "No Validation Info",
			Tel:  "tel:+15559876543",
		},
	}

	// Test that all parties serialize correctly
	for i, party := range parties {
		jsonData, err := json.Marshal(party)
		if err != nil {
			t.Errorf("failed to marshal party %d: %v", i, err)
			continue
		}

		var unmarshaled Party
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Errorf("failed to unmarshal party %d: %v", i, err)
			continue
		}

		if unmarshaled.Name != party.Name {
			t.Errorf("party %d: expected Name %s, got %s", i, party.Name, unmarshaled.Name)
		}

		if unmarshaled.Validation != party.Validation {
			t.Errorf("party %d: expected Validation %s, got %s", i, party.Validation, unmarshaled.Validation)
		}
	}
}

func TestPartyContactList(t *testing.T) {
	party := Party{
		Name:        "Conference Organizer",
		ContactList: "participants.json",
		Role:        "moderator",
	}

	jsonData, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("failed to marshal party with contact list: %v", err)
	}

	var unmarshaled Party
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal party with contact list: %v", err)
	}

	if unmarshaled.ContactList != party.ContactList {
		t.Errorf("expected ContactList %s, got %s", party.ContactList, unmarshaled.ContactList)
	}

	if unmarshaled.Role != party.Role {
		t.Errorf("expected Role %s, got %s", party.Role, unmarshaled.Role)
	}
}

func TestPartyRoles(t *testing.T) {
	// Test common party roles
	roles := []string{
		"originator",
		"recipient",
		"moderator",
		"participant",
		"observer",
		"cc",
		"bcc",
	}

	for _, role := range roles {
		party := Party{
			Name: "Test User",
			Role: role,
		}

		jsonData, err := json.Marshal(party)
		if err != nil {
			t.Errorf("failed to marshal party with role %s: %v", role, err)
			continue
		}

		var unmarshaled Party
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Errorf("failed to unmarshal party with role %s: %v", role, err)
			continue
		}

		if unmarshaled.Role != role {
			t.Errorf("expected role %s, got %s", role, unmarshaled.Role)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
