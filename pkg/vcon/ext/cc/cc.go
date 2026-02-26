// Package cc implements the Contact Center (CC) vCon extension
// as defined in draft-ietf-vcon-cc-extension-01.
//
// The CC extension adds party-level and dialog-level parameters
// for contact center scenarios. It is a compatible extension
// (does not need to be listed in the critical[] array).
//
// Party parameters: role, contact_list
// Dialog parameters: campaign, interaction_type, interaction_id, skill
package cc

import (
	"encoding/json"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
)

// Name is the extension token registered with IANA.
const Name = "CC"

// CCExtension implements vcon.Extension for the Contact Center extension.
type CCExtension struct{}

func (e CCExtension) Name() string       { return Name }
func (e CCExtension) IsCompatible() bool { return true }

func (e CCExtension) PartyParams() []string {
	return []string{"role", "contact_list"}
}

func (e CCExtension) DialogParams() []string {
	return []string{"campaign", "interaction_type", "interaction_id", "skill"}
}

func (e CCExtension) AnalysisParams() []string   { return nil }
func (e CCExtension) AttachmentParams() []string  { return nil }
func (e CCExtension) VConParams() []string        { return nil }

// Register adds the CC extension to the given registry.
func Register(r *vcon.ExtensionRegistry) {
	r.Register(CCExtension{})
}

func init() {
	Register(vcon.DefaultRegistry)
}

// PartyData holds CC extension fields for a Party.
type PartyData struct {
	Role        string `json:"role,omitempty"`
	ContactList string `json:"contact_list,omitempty"`
}

// DialogData holds CC extension fields for a Dialog.
type DialogData struct {
	Campaign        string `json:"campaign,omitempty"`
	InteractionType string `json:"interaction_type,omitempty"`
	InteractionID   string `json:"interaction_id,omitempty"`
	Skill           string `json:"skill,omitempty"`
}

// GetPartyData extracts CC extension fields from a Party's extra properties map.
func GetPartyData(m map[string]any) PartyData {
	if m == nil {
		return PartyData{}
	}
	var d PartyData
	if v, ok := m["role"].(string); ok {
		d.Role = v
	}
	if v, ok := m["contact_list"].(string); ok {
		d.ContactList = v
	}
	return d
}

// SetPartyData writes CC extension fields into a map.
func SetPartyData(m map[string]any, data PartyData) {
	if data.Role != "" {
		m["role"] = data.Role
	}
	if data.ContactList != "" {
		m["contact_list"] = data.ContactList
	}
}

// GetDialogData extracts CC extension fields from a Dialog's extra properties map.
func GetDialogData(m map[string]any) DialogData {
	if m == nil {
		return DialogData{}
	}
	var d DialogData
	if v, ok := m["campaign"].(string); ok {
		d.Campaign = v
	}
	if v, ok := m["interaction_type"].(string); ok {
		d.InteractionType = v
	}
	if v, ok := m["interaction_id"].(string); ok {
		d.InteractionID = v
	}
	if v, ok := m["skill"].(string); ok {
		d.Skill = v
	}
	return d
}

// SetDialogData writes CC extension fields into a map.
func SetDialogData(m map[string]any, data DialogData) {
	if data.Campaign != "" {
		m["campaign"] = data.Campaign
	}
	if data.InteractionType != "" {
		m["interaction_type"] = data.InteractionType
	}
	if data.InteractionID != "" {
		m["interaction_id"] = data.InteractionID
	}
	if data.Skill != "" {
		m["skill"] = data.Skill
	}
}

// MarshalPartyData serializes PartyData to JSON.
func MarshalPartyData(data PartyData) ([]byte, error) {
	return json.Marshal(data)
}

// UnmarshalPartyData deserializes PartyData from JSON.
func UnmarshalPartyData(b []byte) (PartyData, error) {
	var d PartyData
	err := json.Unmarshal(b, &d)
	return d, err
}

// MarshalDialogData serializes DialogData to JSON.
func MarshalDialogData(data DialogData) ([]byte, error) {
	return json.Marshal(data)
}

// UnmarshalDialogData deserializes DialogData from JSON.
func UnmarshalDialogData(b []byte) (DialogData, error) {
	var d DialogData
	err := json.Unmarshal(b, &d)
	return d, err
}
