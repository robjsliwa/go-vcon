package vcon

import (
	"time"

	"crypto/x509"

	"github.com/google/uuid"
)

// SpecVersion is the draft version this library targets.
const SpecVersion = "0.0.3"

// Core Types

// VCon is the top-level container.
type VCon struct {
	Vcon        string       `json:"vcon"` // must be SpecVersion
	UUID        uuid.UUID    `json:"uuid"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
	Subject     string       `json:"subject,omitempty"`
	Group       string       `json:"group,omitempty"`
	Redacted    bool         `json:"redacted,omitempty"`
	Appended    bool         `json:"appended,omitempty"`
	Parties     []Party      `json:"parties,omitempty"`
	Dialog      []Dialog     `json:"dialog,omitempty"`
	Analysis    []Analysis   `json:"analysis,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Party represents a participant.
type Party struct {
	Tel          string        `json:"tel,omitempty"` // tel URL
	Stir         string        `json:"stir,omitempty"` // STIR PASSporT JWT
	Mailto       string        `json:"mailto,omitempty"`
	Name         string        `json:"name,omitempty"`
	Validation   string        `json:"validation,omitempty"`
	GmlPos       string        `json:"gmlpos,omitempty"`
	CivicAddress *CivicAddress `json:"civicaddress,omitempty"`
	Timezone     string        `json:"timezone,omitempty"`
	UUID         string        `json:"uuid,omitempty"`
	Role         string        `json:"role,omitempty"`
	ContactList  string        `json:"contact_list,omitempty"`
}

// CivicAddress contains civic address information for a party's location
type CivicAddress struct {
	Country     string `json:"country,omitempty"`
	A1          string `json:"a1,omitempty"`
	A2          string `json:"a2,omitempty"`
	A3          string `json:"a3,omitempty"`
	A4          string `json:"a4,omitempty"`
	A5          string `json:"a5,omitempty"`
	A6          string `json:"a6,omitempty"`
	PRD         string `json:"prd,omitempty"`
	POD         string `json:"pod,omitempty"`
	STS         string `json:"sts,omitempty"`
	HNO         string `json:"hno,omitempty"`
	HNS         string `json:"hns,omitempty"`
	LMK         string `json:"lmk,omitempty"`
	LOC         string `json:"loc,omitempty"`
	FLR         string `json:"flr,omitempty"`
	NAM         string `json:"nam,omitempty"`
	PC          string `json:"pc,omitempty"`
}

// Dialog is an interaction (call leg, chat, etc.).
type Dialog struct {
	Type          string      `json:"type"` // recording, text, transfer, incomplete
	StartTime     *time.Time  `json:"start"`
	Duration      float64     `json:"duration,omitempty"`
	Parties       interface{} `json:"parties,omitempty"` // int or []int or []interface{} with int or []int
	Originator    int         `json:"originator,omitempty"`
	MediaType     string      `json:"mediatype,omitempty"`
	Filename      string      `json:"filename,omitempty"`
	Body          string      `json:"body,omitempty"`
	Encoding      string      `json:"encoding,omitempty"`
	URL           string      `json:"url,omitempty"`
	ContentHash   string      `json:"content_hash,omitempty"`
	Disposition   string      `json:"disposition,omitempty"`
	PartyHistory  []PartyHistory `json:"party_history,omitempty"`
	// Dialog Transfer fields
	Transferee    int    `json:"transferee,omitempty"`
	Transferor    int    `json:"transferor,omitempty"`
	TransferTarget int   `json:"transfer-target,omitempty"`
	Original      int    `json:"original,omitempty"`
	Consultation  int    `json:"consultation,omitempty"`
	TargetDialog  int    `json:"target-dialog,omitempty"`
	// Additional fields
	Campaign        string `json:"campaign,omitempty"`
	InteractionType string `json:"interaction_type,omitempty"`
	InteractionID   string `json:"interaction_id,omitempty"`
	Skill           string `json:"skill,omitempty"`
	Application     string `json:"application,omitempty"`
	MessageID       string `json:"message_id,omitempty"`
}

// PartyHistory represents a party joining/leaving/status change event
type PartyHistory struct {
	Party int       `json:"party"`
	Event string    `json:"event"` // join, drop, hold, unhold, mute, unmute
	Time  time.Time `json:"time"`
}

// Analysis holds machine-generated artefacts.
type Analysis struct {
	Type        string      `json:"type"` // e.g. asr-json, llm-summary
	Dialog      interface{} `json:"dialog,omitempty"` // int or []int
	MediaType   string      `json:"mediatype,omitempty"`
	Filename    string      `json:"filename,omitempty"`
	Vendor      string      `json:"vendor,omitempty"`
	Product     string      `json:"product,omitempty"`
	Schema      string      `json:"schema,omitempty"`
	Body        interface{} `json:"body,omitempty"`
	Encoding    string      `json:"encoding,omitempty"`
	URL         string      `json:"url,omitempty"`
	ContentHash string      `json:"content_hash,omitempty"`
}

// Attachment is an arbitrary file linked to a dialog/party.
type Attachment struct {
	FileRef
	DialogIdx  int       `json:"dialog,omitempty"`
	PartyIdx   int       `json:"party"`
	StartTime  time.Time `json:"start"`
	MediaType  string    `json:"mediatype,omitempty"`
	Filename   string    `json:"filename,omitempty"`
}

// FileRef stores inline or external binary/text.
type FileRef struct {
	Body        string `json:"body,omitempty"`         // base64
	Encoding    string `json:"encoding,omitempty"`     // "base64"
	URL         string `json:"url,omitempty"`          // https://…
	ContentHash string `json:"content_hash,omitempty"` // SHA-512 (base64url)
}

// SignOptions contains options for vCon signing
type SignOptions struct {
    Certificates []*x509.Certificate
    ExtraHeaders map[string]interface{}
}

// Convenience API

// New creates an empty, valid container.
func New() *VCon {
	return &VCon{
		Vcon:   SpecVersion,
		UUID:      uuid.New(),
		CreatedAt: time.Now().UTC(),
	}
}

// Add* helpers
func (v *VCon) AddParty(p Party) int {
	v.Parties = append(v.Parties, p)
	return len(v.Parties) - 1
}

func (v *VCon) AddDialog(d Dialog) int {
	v.Dialog = append(v.Dialog, d)
	return len(v.Dialog) - 1
}

func (v *VCon) AddAnalysis(a Analysis) int {
	v.Analysis = append(v.Analysis, a)
	return len(v.Analysis) - 1
}

func (v *VCon) AddAttachment(att Attachment) int {
	v.Attachments = append(v.Attachments, att)
	return len(v.Attachments) - 1
}
