package cc

import (
	"encoding/json"
	"testing"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
)

func TestCCExtensionName(t *testing.T) {
	ext := CCExtension{}
	if ext.Name() != "CC" {
		t.Errorf("expected name CC, got %s", ext.Name())
	}
}

func TestCCExtensionIsCompatible(t *testing.T) {
	ext := CCExtension{}
	if !ext.IsCompatible() {
		t.Error("CC extension should be compatible")
	}
}

func TestCCExtensionPartyParams(t *testing.T) {
	ext := CCExtension{}
	params := ext.PartyParams()
	expected := []string{"role", "contact_list"}

	if len(params) != len(expected) {
		t.Fatalf("expected %d party params, got %d", len(expected), len(params))
	}
	for i, p := range params {
		if p != expected[i] {
			t.Errorf("expected param %s at index %d, got %s", expected[i], i, p)
		}
	}
}

func TestCCExtensionDialogParams(t *testing.T) {
	ext := CCExtension{}
	params := ext.DialogParams()
	expected := []string{"campaign", "interaction_type", "interaction_id", "skill"}

	if len(params) != len(expected) {
		t.Fatalf("expected %d dialog params, got %d", len(expected), len(params))
	}
	for i, p := range params {
		if p != expected[i] {
			t.Errorf("expected param %s at index %d, got %s", expected[i], i, p)
		}
	}
}

func TestCCExtensionEmptyParams(t *testing.T) {
	ext := CCExtension{}
	if ext.AnalysisParams() != nil {
		t.Error("expected nil analysis params")
	}
	if ext.AttachmentParams() != nil {
		t.Error("expected nil attachment params")
	}
	if ext.VConParams() != nil {
		t.Error("expected nil vcon params")
	}
}

func TestCCRegisteredInDefaultRegistry(t *testing.T) {
	// init() should have registered CC in the default registry
	ext, ok := vcon.DefaultRegistry.Get("CC")
	if !ok {
		t.Fatal("CC extension should be registered in default registry")
	}
	if ext.Name() != "CC" {
		t.Errorf("expected name CC, got %s", ext.Name())
	}
}

func TestRegister(t *testing.T) {
	r := vcon.NewExtensionRegistry()
	Register(r)
	ext, ok := r.Get("CC")
	if !ok {
		t.Fatal("CC extension should be registered after Register()")
	}
	if ext.Name() != "CC" {
		t.Errorf("expected name CC, got %s", ext.Name())
	}
}

func TestGetPartyData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected PartyData
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: PartyData{},
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: PartyData{},
		},
		{
			name: "with role",
			input: map[string]any{
				"role": "agent",
			},
			expected: PartyData{Role: "agent"},
		},
		{
			name: "with role and contact_list",
			input: map[string]any{
				"role":         "customer",
				"contact_list": "VIP",
			},
			expected: PartyData{Role: "customer", ContactList: "VIP"},
		},
		{
			name: "ignores non-CC fields",
			input: map[string]any{
				"role":  "agent",
				"name":  "Alice",
				"other": 42,
			},
			expected: PartyData{Role: "agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPartyData(tt.input)
			if result != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestSetPartyData(t *testing.T) {
	m := map[string]any{"name": "Alice"}
	SetPartyData(m, PartyData{Role: "agent", ContactList: "VIP"})

	if m["role"] != "agent" {
		t.Errorf("expected role agent, got %v", m["role"])
	}
	if m["contact_list"] != "VIP" {
		t.Errorf("expected contact_list VIP, got %v", m["contact_list"])
	}
	if m["name"] != "Alice" {
		t.Error("existing fields should be preserved")
	}
}

func TestSetPartyDataEmpty(t *testing.T) {
	m := map[string]any{"name": "Alice"}
	SetPartyData(m, PartyData{})

	if _, ok := m["role"]; ok {
		t.Error("empty role should not be set")
	}
	if _, ok := m["contact_list"]; ok {
		t.Error("empty contact_list should not be set")
	}
}

func TestGetDialogData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected DialogData
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: DialogData{},
		},
		{
			name: "all fields",
			input: map[string]any{
				"campaign":         "summer_sale",
				"interaction_type": "inbound",
				"interaction_id":   "INT-123",
				"skill":            "billing",
			},
			expected: DialogData{
				Campaign:        "summer_sale",
				InteractionType: "inbound",
				InteractionID:   "INT-123",
				Skill:           "billing",
			},
		},
		{
			name: "partial fields",
			input: map[string]any{
				"campaign": "promo",
			},
			expected: DialogData{Campaign: "promo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDialogData(tt.input)
			if result != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestSetDialogData(t *testing.T) {
	m := map[string]any{"type": "recording"}
	SetDialogData(m, DialogData{
		Campaign:        "promo",
		InteractionType: "outbound",
		InteractionID:   "INT-456",
		Skill:           "sales",
	})

	if m["campaign"] != "promo" {
		t.Errorf("expected campaign promo, got %v", m["campaign"])
	}
	if m["interaction_type"] != "outbound" {
		t.Errorf("expected interaction_type outbound, got %v", m["interaction_type"])
	}
	if m["interaction_id"] != "INT-456" {
		t.Errorf("expected interaction_id INT-456, got %v", m["interaction_id"])
	}
	if m["skill"] != "sales" {
		t.Errorf("expected skill sales, got %v", m["skill"])
	}
	if m["type"] != "recording" {
		t.Error("existing fields should be preserved")
	}
}

func TestMarshalUnmarshalPartyData(t *testing.T) {
	original := PartyData{Role: "agent", ContactList: "Priority"}
	data, err := MarshalPartyData(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	result, err := UnmarshalPartyData(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result != original {
		t.Errorf("expected %+v, got %+v", original, result)
	}
}

func TestMarshalUnmarshalDialogData(t *testing.T) {
	original := DialogData{
		Campaign:        "q4",
		InteractionType: "chat",
		InteractionID:   "CHAT-789",
		Skill:           "support",
	}
	data, err := MarshalDialogData(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	result, err := UnmarshalDialogData(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result != original {
		t.Errorf("expected %+v, got %+v", original, result)
	}
}

func TestAllowedPartyParamsIncludeCC(t *testing.T) {
	// The default registry should include CC party params
	allowed := vcon.DefaultRegistry.AllowedPartyParams()
	ccParams := []string{"role", "contact_list"}
	for _, p := range ccParams {
		if _, ok := allowed[p]; !ok {
			t.Errorf("expected CC param %s in allowed party params", p)
		}
	}
}

func TestAllowedDialogParamsIncludeCC(t *testing.T) {
	// The default registry should include CC dialog params
	allowed := vcon.DefaultRegistry.AllowedDialogParams()
	ccParams := []string{"campaign", "interaction_type", "interaction_id", "skill"}
	for _, p := range ccParams {
		if _, ok := allowed[p]; !ok {
			t.Errorf("expected CC param %s in allowed dialog params", p)
		}
	}
}

func TestCCExtensionVConRoundTrip(t *testing.T) {
	ccJSON := `{
		"vcon": "0.4.0",
		"uuid": "019471e8-2a00-8a96-be3e-580a44cc285f",
		"created_at": "2024-01-15T10:00:00Z",
		"extensions": ["cc"],
		"parties": [
			{"name": "Alice", "tel": "tel:+1234567890", "role": "agent", "contact_list": "VIP"}
		],
		"dialog": [
			{
				"type": "text",
				"start": "2024-01-15T10:00:00Z",
				"parties": [0],
				"body": "hello",
				"encoding": "none",
				"campaign": "summer",
				"interaction_type": "inbound",
				"interaction_id": "INT-1",
				"skill": "billing"
			}
		]
	}`

	// Parse raw JSON to get party/dialog maps
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(ccJSON), &raw); err != nil {
		t.Fatalf("parse raw JSON: %v", err)
	}

	parties := raw["parties"].([]interface{})
	partyMap := parties[0].(map[string]interface{})

	dialogs := raw["dialog"].([]interface{})
	dialogMap := dialogs[0].(map[string]interface{})

	// Default mode: CC fields should be preserved
	allowedParty := vcon.DefaultRegistry.AllowedPartyParams()
	resultParty := vcon.ProcessProperties(partyMap, allowedParty, vcon.PropertyHandlingDefault)
	if resultParty["role"] != "agent" {
		t.Errorf("expected role=agent in default mode, got %v", resultParty["role"])
	}
	if resultParty["contact_list"] != "VIP" {
		t.Errorf("expected contact_list=VIP in default mode, got %v", resultParty["contact_list"])
	}

	allowedDialog := vcon.DefaultRegistry.AllowedDialogParams()
	resultDialog := vcon.ProcessProperties(dialogMap, allowedDialog, vcon.PropertyHandlingDefault)
	if resultDialog["campaign"] != "summer" {
		t.Errorf("expected campaign=summer in default mode, got %v", resultDialog["campaign"])
	}
	if resultDialog["interaction_type"] != "inbound" {
		t.Errorf("expected interaction_type=inbound, got %v", resultDialog["interaction_type"])
	}
	if resultDialog["interaction_id"] != "INT-1" {
		t.Errorf("expected interaction_id=INT-1, got %v", resultDialog["interaction_id"])
	}
	if resultDialog["skill"] != "billing" {
		t.Errorf("expected skill=billing, got %v", resultDialog["skill"])
	}

	// Extract CC data using typed helpers
	pd := GetPartyData(partyMap)
	if pd.Role != "agent" || pd.ContactList != "VIP" {
		t.Errorf("GetPartyData mismatch: %+v", pd)
	}
	dd := GetDialogData(dialogMap)
	if dd.Campaign != "summer" || dd.InteractionType != "inbound" || dd.InteractionID != "INT-1" || dd.Skill != "billing" {
		t.Errorf("GetDialogData mismatch: %+v", dd)
	}

	// Strict mode: CC fields removed (they're extension fields but still in the allowed set)
	// Re-parse since ProcessProperties may modify the map
	var raw2 map[string]interface{}
	if err := json.Unmarshal([]byte(ccJSON), &raw2); err != nil {
		t.Fatalf("parse raw JSON: %v", err)
	}
	parties2 := raw2["parties"].([]interface{})
	partyMap2 := parties2[0].(map[string]interface{})

	// Strict mode with core-only properties (no extension params) should remove CC fields
	corePartyOnly := make(map[string]struct{})
	for k, v := range vcon.AllowedPartyProperties {
		corePartyOnly[k] = v
	}
	resultStrict := vcon.ProcessProperties(partyMap2, corePartyOnly, vcon.PropertyHandlingStrict)
	if _, ok := resultStrict["role"]; ok {
		t.Error("strict mode with core-only props should remove CC 'role' field")
	}
	if _, ok := resultStrict["contact_list"]; ok {
		t.Error("strict mode with core-only props should remove CC 'contact_list' field")
	}

	// BuildFromJSON should succeed (schema allows additional properties)
	v, err := vcon.BuildFromJSON(ccJSON)
	if err != nil {
		t.Fatalf("BuildFromJSON failed: %v", err)
	}

	// Core struct fields should be correct
	if v.Parties[0].Name != "Alice" {
		t.Errorf("expected party name Alice, got %s", v.Parties[0].Name)
	}
	if v.Parties[0].Tel != "tel:+1234567890" {
		t.Errorf("expected party tel tel:+1234567890, got %s", v.Parties[0].Tel)
	}
	if v.Dialog[0].Type != "text" {
		t.Errorf("expected dialog type text, got %s", v.Dialog[0].Type)
	}

	// CC fields are not on Go structs, so they are dropped during json.Unmarshal.
	// Verify they are accessible via the raw JSON intermediate representation.
	// Re-marshal and re-parse as raw JSON to confirm core fields round-trip.
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var rawRT map[string]interface{}
	if err := json.Unmarshal(data, &rawRT); err != nil {
		t.Fatalf("unmarshal raw round-trip: %v", err)
	}
	rtParties := rawRT["parties"].([]interface{})
	rtParty := rtParties[0].(map[string]interface{})
	if rtParty["name"] != "Alice" {
		t.Errorf("round-trip party name mismatch: %v", rtParty["name"])
	}
	// CC fields (role, contact_list) are expected to be absent after Go struct round-trip
	// since they are not part of the Party struct — they live at the map level only
	if _, ok := rtParty["role"]; ok {
		t.Log("note: CC 'role' field survived Go struct round-trip (unexpected but not an error)")
	}
}
