package vcon

import (
	"encoding/json"
	"time"
)

// PartyEventType represents the type of event in a party's history
type PartyEventType string

const (
	// PartyEventJoin indicates a party joined the conversation
	PartyEventJoin PartyEventType = "join"
	// PartyEventDrop indicates a party left the conversation
	PartyEventDrop PartyEventType = "drop"
	// PartyEventHold indicates a party was put on hold
	PartyEventHold PartyEventType = "hold"
	// PartyEventUnhold indicates a party was taken off hold
	PartyEventUnhold PartyEventType = "unhold"
	// PartyEventMute indicates a party was muted
	PartyEventMute PartyEventType = "mute"
	// PartyEventUnmute indicates a party was unmuted
	PartyEventUnmute PartyEventType = "unmute"
)

// Party represents a participant in a vCon.
type Party struct {
	// Telephone number of the party (tel URL)
	Tel string `json:"tel,omitempty"`
	// STIR PASSporT JWT
	Stir string `json:"stir,omitempty"`
	// Email address of the party
	Mailto string `json:"mailto,omitempty"`
	// Display name of the party
	Name string `json:"name,omitempty"`
	// Validation information of the party
	Validation string `json:"validation,omitempty"`
	// GML position of the party
	GmlPos string `json:"gmlpos,omitempty"`
	// Civic address of the party
	CivicAddress *CivicAddress `json:"civicaddress,omitempty"`
	// Timezone of the party
	Timezone string `json:"timezone,omitempty"`
	// UUID of the party
	UUID string `json:"uuid,omitempty"`
	// Role of the party
	Role string `json:"role,omitempty"`
	// Contact list of the party
	ContactList string `json:"contact_list,omitempty"`
	// Additional attributes (not directly serialized)
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// PartyOption is a function that configures a Party
type PartyOption func(*Party)

// NewParty creates a new Party with the given options
func NewParty(opts ...PartyOption) *Party {
	p := &Party{
		Meta: make(map[string]interface{}),
	}
	
	// Apply all provided options
	for _, opt := range opts {
		opt(p)
	}
	
	return p
}

// WithTel sets the telephone number for a Party
func WithTel(tel string) PartyOption {
	return func(p *Party) {
		p.Tel = tel
	}
}

// WithName sets the display name for a Party
func WithName(name string) PartyOption {
	return func(p *Party) {
		p.Name = name
	}
}

// WithMailto sets the email address for a Party
func WithMailto(mailto string) PartyOption {
	return func(p *Party) {
		p.Mailto = mailto
	}
}

// WithRole sets the role for a Party
func WithRole(role string) PartyOption {
	return func(p *Party) {
		p.Role = role
	}
}

// WithCivicAddress sets the civic address for a Party
func WithCivicAddress(address *CivicAddress) PartyOption {
	return func(p *Party) {
		p.CivicAddress = address
	}
}

// WithMeta adds a metadata entry to a Party
func WithMeta(key string, value interface{}) PartyOption {
	return func(p *Party) {
		if p.Meta == nil {
			p.Meta = make(map[string]interface{})
		}
		p.Meta[key] = value
	}
}

// ToMap converts the Party to a map, excluding empty fields
func (p *Party) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	
	if p.Tel != "" {
		result["tel"] = p.Tel
	}
	if p.Stir != "" {
		result["stir"] = p.Stir
	}
	if p.Mailto != "" {
		result["mailto"] = p.Mailto
	}
	if p.Name != "" {
		result["name"] = p.Name
	}
	if p.Validation != "" {
		result["validation"] = p.Validation
	}
	if p.GmlPos != "" {
		result["gmlpos"] = p.GmlPos
	}
	if p.CivicAddress != nil {
		result["civicaddress"] = p.CivicAddress.ToMap()
	}
	if p.Timezone != "" {
		result["timezone"] = p.Timezone
	}
	if p.UUID != "" {
		result["uuid"] = p.UUID
	}
	if p.Role != "" {
		result["role"] = p.Role
	}
	if p.ContactList != "" {
		result["contact_list"] = p.ContactList
	}
	
	// Add any Meta items
	for k, v := range p.Meta {
		if v != nil {
			result[k] = v
		}
	}
	
	return result
}

// SetFromMap sets Party fields from a map
func (p *Party) SetFromMap(data map[string]interface{}) {
	// Initialize Meta if it's nil
	if p.Meta == nil {
		p.Meta = make(map[string]interface{})
	}
	
	// Handle standard fields
	if v, ok := data["tel"].(string); ok {
		p.Tel = v
	}
	if v, ok := data["stir"].(string); ok {
		p.Stir = v
	}
	if v, ok := data["mailto"].(string); ok {
		p.Mailto = v
	}
	if v, ok := data["name"].(string); ok {
		p.Name = v
	}
	if v, ok := data["validation"].(string); ok {
		p.Validation = v
	}
	if v, ok := data["gmlpos"].(string); ok {
		p.GmlPos = v
	}
	if v, ok := data["civicaddress"].(map[string]interface{}); ok {
		civicAddressMap := make(map[string]string)
		for k, val := range v {
			if strVal, ok := val.(string); ok {
				civicAddressMap[k] = strVal
			}
		}
		if p.CivicAddress == nil {
			p.CivicAddress = NewCivicAddress()
		}
		p.CivicAddress.SetFromMap(civicAddressMap)
	}
	if v, ok := data["timezone"].(string); ok {
		p.Timezone = v
	}
	if v, ok := data["uuid"].(string); ok {
		p.UUID = v
	}
	if v, ok := data["role"].(string); ok {
		p.Role = v
	}
	if v, ok := data["contact_list"].(string); ok {
		p.ContactList = v
	}
	
	// Any other fields get added to Meta
	for k, v := range data {
		switch k {
		case "tel", "stir", "mailto", "name", "validation", "gmlpos",
			 "civicaddress", "timezone", "uuid", "role", "contact_list", "meta":
			// Skip fields we've already processed
			continue
		default:
			p.Meta[k] = v
		}
	}
}

// ToDict converts the Party to a dictionary
func (p *Party) ToDict() map[string]interface{} {
	raw, _ := json.Marshal(p)
	var result map[string]interface{}
	json.Unmarshal(raw, &result)
	return result
}

// PartyHistory represents a party joining/leaving/status change event
type PartyHistory struct {
	// Index of the party
	Party int `json:"party"`
	// Event type (e.g. "join", "leave", "hold")
	Event string `json:"event"`
	// Time of the event
	Time time.Time `json:"time"`
}

// NewPartyHistory creates a new PartyHistory instance
func NewPartyHistory(party int, event PartyEventType, t time.Time) *PartyHistory {
	return &PartyHistory{
		Party: party,
		Event: string(event),
		Time:  t,
	}
}

// ToMap converts the PartyHistory to a map
func (ph *PartyHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"party": ph.Party,
		"event": ph.Event,
		"time":  ph.Time,
	}
}
