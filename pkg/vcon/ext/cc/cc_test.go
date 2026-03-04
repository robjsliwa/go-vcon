package cc

import (
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
