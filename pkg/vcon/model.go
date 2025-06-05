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
	Version     string       `json:"vcon"` // must be SpecVersion
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
	Addr     string   `json:"addr,omitempty"` // tel:, sip:, mailto:, etc.
	E164     string   `json:"e164,omitempty"`
	Name     string   `json:"name,omitempty"`
	Location string   `json:"loc,omitempty"`
	Roles    []string `json:"role,omitempty"`
	Passport string   `json:"passport,omitempty"` // STIR PASSporT JWT
}

// Dialog is an interaction (call leg, chat, etc.).
type Dialog struct {
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	MediaType   string     `json:"media_type,omitempty"`
	Originator  int        `json:"orig_party,omitempty"`
	DestParties []int      `json:"dest_parties,omitempty"`
	Content     *FileRef   `json:"content,omitempty"`
	Interaction string     `json:"interaction_id,omitempty"`
	// future: Passport []Passport
}

// Analysis holds machine-generated artefacts.
type Analysis struct {
	Type    string   `json:"type"` // e.g. asr-json, llm-summary
	Vendor  string   `json:"vendor,omitempty"`
	Product string   `json:"product,omitempty"`
	Content *FileRef `json:"content,omitempty"`
}

// Attachment is an arbitrary file linked to a dialog/party.
type Attachment struct {
	FileRef
	DialogIdx int `json:"dialog,omitempty"`
	PartyIdx  int `json:"party,omitempty"`
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
		Version:   SpecVersion,
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
