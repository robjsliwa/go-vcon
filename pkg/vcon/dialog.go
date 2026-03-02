package vcon

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

// MIME types constants
const (
	MIMETypePlainText = "text/plain"
	MIMETypeAudioWav  = "audio/x-wav"
	MIMETypeAudioWav2 = "audio/wav"
	MIMETypeAudioWave = "audio/wave"
	MIMETypeAudioMpeg = "audio/mpeg"
	MIMETypeAudioMP3  = "audio/mp3"
	MIMETypeAudioOgg  = "audio/ogg"
	MIMETypeAudioWebm = "audio/webm"
	MIMETypeAudioM4a  = "audio/x-m4a"
	MIMETypeAudioAAC  = "audio/aac"
	MIMETypeVideoMP4  = "video/x-mp4"
	MIMETypeVideoOgg  = "video/ogg"
	MIMETypeMultipart = "multipart/mixed"
	MIMETypeRFC822    = "message/rfc822"
)

// Valid encoding types (v0.4.0: "base64" removed, only "base64url", "json", "none")
var ValidEncodings = []string{"base64url", "json", "none"}

// All supported MIME types
var SupportedMIMETypes = []string{
	MIMETypePlainText,
	MIMETypeAudioWav,
	MIMETypeAudioWav2,
	MIMETypeAudioWave,
	MIMETypeAudioMpeg,
	MIMETypeAudioMP3,
	MIMETypeAudioOgg,
	MIMETypeAudioWebm,
	MIMETypeAudioM4a,
	MIMETypeAudioAAC,
	MIMETypeVideoMP4,
	MIMETypeVideoOgg,
	MIMETypeMultipart,
	MIMETypeRFC822,
}

// Audio MIME types
var AudioMIMETypes = []string{
	MIMETypeAudioWav,
	MIMETypeAudioWav2,
	MIMETypeAudioWave,
	MIMETypeAudioMpeg,
	MIMETypeAudioMP3,
	MIMETypeAudioOgg,
	MIMETypeAudioWebm,
	MIMETypeAudioM4a,
	MIMETypeAudioAAC,
}

// Video MIME types
var VideoMIMETypes = []string{
	MIMETypeVideoMP4,
	MIMETypeVideoOgg,
}

// Dialog is an interaction (call leg, chat, etc.)
type Dialog struct {
	Type         string          `json:"type"`  // recording, text, transfer, incomplete
	StartTime    *time.Time      `json:"start"` // Required
	Duration     float64         `json:"duration,omitempty"`
	Parties      interface{}     `json:"parties,omitempty"` // int or []int
	Originator   int             `json:"originator,omitempty"`
	MediaType    string          `json:"mediatype,omitempty"` // MIME type
	Filename     string          `json:"filename,omitempty"`
	Body         string          `json:"body,omitempty"`
	Encoding     string          `json:"encoding,omitempty"`     // e.g., "base64url"
	URL          string          `json:"url,omitempty"`          // For external data
	ContentHash  ContentHashList `json:"content_hash,omitempty"` // SHA-512 hash(es)
	Disposition  string          `json:"disposition,omitempty"`
	PartyHistory []PartyHistory  `json:"party_history,omitempty"`
	SessionID    interface{}     `json:"session_id,omitempty"` // SessionId or []SessionId

	// Dialog Transfer fields (int or []int per v0.4.0)
	Transferee     int         `json:"transferee,omitempty"`
	Transferor     int         `json:"transferor,omitempty"`
	TransferTarget *IntOrSlice `json:"transfer_target,omitempty"`
	Original       *IntOrSlice `json:"original,omitempty"`
	Consultation   *IntOrSlice `json:"consultation,omitempty"`
	TargetDialog   *IntOrSlice `json:"target_dialog,omitempty"`

	// Additional fields
	Application string `json:"application,omitempty"`
	MessageID   string `json:"message_id,omitempty"`
}

// DialogOption is a function that configures a Dialog
type DialogOption func(*Dialog)

// NewDialog creates a new Dialog with the required fields
func NewDialog(dialogType string, start time.Time, parties interface{}, opts ...DialogOption) *Dialog {
	dialog := &Dialog{
		Type:      dialogType,
		StartTime: &start,
		Parties:   parties,
	}

	// Apply all provided options
	for _, opt := range opts {
		opt(dialog)
	}

	return dialog
}

// WithMediaType sets the media type for a Dialog
func WithMediaType(mediaType string) DialogOption {
	return func(d *Dialog) {
		d.MediaType = mediaType
	}
}

// WithBody sets the body content for a Dialog
func WithBody(body string) DialogOption {
	return func(d *Dialog) {
		d.Body = body
	}
}

// WithEncoding sets the encoding for a Dialog
func WithEncoding(encoding string) DialogOption {
	return func(d *Dialog) {
		d.Encoding = encoding
	}
}

// WithURL sets the URL for external content in a Dialog
func WithURL(url string) DialogOption {
	return func(d *Dialog) {
		d.URL = url
	}
}

// WithOriginator sets the originator party index for a Dialog
func WithOriginator(originator int) DialogOption {
	return func(d *Dialog) {
		d.Originator = originator
	}
}

func (d *Dialog) addContentHashToMap(result map[string]interface{}) {
	if d.ContentHash.IsEmpty() {
		return
	}
	if len(d.ContentHash) == 1 {
		result["content_hash"] = d.ContentHash[0].String()
	} else {
		strs := make([]string, len(d.ContentHash))
		for i, ch := range d.ContentHash {
			strs[i] = ch.String()
		}
		result["content_hash"] = strs
	}
}

func (d *Dialog) addPartyHistoryToMap(result map[string]interface{}) {
	if len(d.PartyHistory) == 0 {
		return
	}
	partyHistory := make([]map[string]interface{}, len(d.PartyHistory))
	for i, ph := range d.PartyHistory {
		phMap := map[string]interface{}{
			"party": ph.Party,
			"event": ph.Event,
			"time":  ph.Time.Format(time.RFC3339),
		}
		if ph.Button != "" {
			phMap["button"] = ph.Button
		}
		partyHistory[i] = phMap
	}
	result["party_history"] = partyHistory
}

func (d *Dialog) addTransferFieldsToMap(result map[string]interface{}) {
	if d.Transferee != 0 {
		result["transferee"] = d.Transferee
	}
	if d.Transferor != 0 {
		result["transferor"] = d.Transferor
	}
	transferFields := []struct {
		key string
		val *IntOrSlice
	}{
		{"transfer_target", d.TransferTarget},
		{"original", d.Original},
		{"consultation", d.Consultation},
		{"target_dialog", d.TargetDialog},
	}
	for _, f := range transferFields {
		if f.val != nil && !f.val.IsZero() {
			result[f.key] = f.val.value
		}
	}
}

// ToMap converts the Dialog to a map, excluding empty fields
func (d *Dialog) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	if d.StartTime == nil {
		now := time.Now().UTC()
		d.StartTime = &now
	}

	if d.Type != "" {
		result["type"] = d.Type
	}
	result["start"] = d.StartTime.Format(time.RFC3339)

	if d.Parties != nil {
		result["parties"] = d.Parties
	}
	if d.Originator != 0 {
		result["originator"] = d.Originator
	}
	if d.MediaType != "" {
		result["mediatype"] = d.MediaType
	}
	if d.Filename != "" {
		result["filename"] = d.Filename
	}
	if d.Body != "" {
		result["body"] = d.Body
	}
	if d.Encoding != "" {
		result["encoding"] = d.Encoding
	}
	if d.URL != "" {
		result["url"] = d.URL
	}
	d.addContentHashToMap(result)
	if d.Disposition != "" {
		result["disposition"] = d.Disposition
	}
	d.addPartyHistoryToMap(result)
	d.addTransferFieldsToMap(result)
	if d.Duration > 0 {
		result["duration"] = d.Duration
	}

	return result
}

// ToDict converts the Dialog to a map (alias for ToMap)
func (d *Dialog) ToDict() map[string]interface{} {
	return d.ToMap()
}

// AddExternalData adds external data to the dialog
func (d *Dialog) AddExternalData(urlStr string, filename string, mimeType string) error {
	// Validate the URL
	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Make HTTP request to fetch content
	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("failed to fetch external data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch external data: HTTP status %d", resp.StatusCode)
	}

	// Set the URL
	d.URL = urlStr

	// Set the content type/MIME type
	if mimeType != "" {
		d.MediaType = mimeType
	} else {
		d.MediaType = resp.Header.Get("Content-Type")
	}

	// Set the filename if provided, otherwise extract from URL
	if filename != "" {
		d.Filename = filename
	} else {
		parsedURL, _ := url.Parse(urlStr)
		d.Filename = path.Base(parsedURL.Path)
	}

	// Read the body to calculate hash
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Calculate SHA-512 hash
	d.ContentHash = ContentHashList{ComputeSHA512(body)}

	return nil
}

// AddInlineData adds inline data to the dialog
func (d *Dialog) AddInlineData(body string, filename string, mimeType string) error {
	// Validate the encoding
	if d.Encoding != "" && !isValidEncoding(d.Encoding) {
		return fmt.Errorf("invalid encoding: %s", d.Encoding)
	}

	d.Body = body
	d.MediaType = mimeType
	d.Filename = filename

	// Set default encoding if not specified
	if d.Encoding == "" {
		d.Encoding = "base64url"
	}

	// Calculate SHA-512 hash
	d.ContentHash = ContentHashList{ComputeSHA512([]byte(body))}

	return nil
}

// Helper to validate encoding
func isValidEncoding(encoding string) bool {
	for _, valid := range ValidEncodings {
		if encoding == valid {
			return true
		}
	}
	return false
}

// IsExternalData checks if the dialog is an external data dialog
func (d *Dialog) IsExternalData() bool {
	return d.URL != ""
}

// IsInlineData checks if the dialog is an inline data dialog
func (d *Dialog) IsInlineData() bool {
	return !d.IsExternalData() && d.Body != ""
}

// IsText checks if the dialog is a text dialog
func (d *Dialog) IsText() bool {
	return d.MediaType == MIMETypePlainText
}

// IsAudio checks if the dialog is an audio dialog
func (d *Dialog) IsAudio() bool {
	for _, audioType := range AudioMIMETypes {
		if d.MediaType == audioType {
			return true
		}
	}
	return false
}

// IsVideo checks if the dialog is a video dialog
func (d *Dialog) IsVideo() bool {
	for _, videoType := range VideoMIMETypes {
		if d.MediaType == videoType {
			return true
		}
	}
	return false
}

// IsEmail checks if the dialog is an email dialog
func (d *Dialog) IsEmail() bool {
	return d.MediaType == MIMETypeRFC822
}

// IsExternalDataChanged checks if external data has changed by comparing hashes
func (d *Dialog) IsExternalDataChanged() (bool, error) {
	if !d.IsExternalData() || d.ContentHash.IsEmpty() {
		return false, nil
	}

	// Fetch the content again to compare hash
	resp, err := http.Get(d.URL)
	if err != nil {
		return true, fmt.Errorf("failed to fetch external data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return true, fmt.Errorf("failed to fetch external data: HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return true, fmt.Errorf("failed to read response body: %w", err)
	}

	// Verify using the first hash
	return !d.ContentHash.First().Verify(body), nil
}

// ToInlineData converts the dialog from external data to inline data
func (d *Dialog) ToInlineData() error {
	if !d.IsExternalData() {
		return errors.New("dialog is not external data")
	}

	// Fetch the content
	resp, err := http.Get(d.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch external data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch external data: HTTP status %d", resp.StatusCode)
	}

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Set the body as base64url encoded content
	d.Body = encodeBase64URL(body)
	d.Encoding = "base64url"

	// Set media type if not already set
	if d.MediaType == "" {
		d.MediaType = resp.Header.Get("Content-Type")
	}

	// Set the filename if not already set
	if d.Filename == "" {
		parsedURL, _ := url.Parse(d.URL)
		d.Filename = path.Base(parsedURL.Path)
	}

	// Calculate SHA-512 hash
	d.ContentHash = ContentHashList{ComputeSHA512(body)}

	// Remove the URL since this is now inline data
	d.URL = ""

	return nil
}

// encodeBase64URL encodes data using base64url encoding without padding
func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// FromMap creates a Dialog from a map
func DialogFromMap(data map[string]interface{}) (*Dialog, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	var dialog Dialog
	if err := json.Unmarshal(jsonData, &dialog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to Dialog: %w", err)
	}

	return &dialog, nil
}
