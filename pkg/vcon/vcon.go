package vcon

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SpecVersion is the draft version this library targets.
const SpecVersion = "0.0.3"

// Property handling modes
const (
	PropertyHandlingDefault = "default" // Keep non-standard properties
	PropertyHandlingStrict  = "strict"  // Remove non-standard properties
	PropertyHandlingMeta    = "meta"    // Move non-standard properties to meta
)

// Allowed properties for validation
var (
	AllowedVConProperties = map[string]struct{}{
		"vcon": {}, "uuid": {}, "created_at": {}, "updated_at": {}, "subject": {},
		"group": {}, "redacted": {}, "appended": {}, "parties": {},
		"dialog": {}, "attachments": {}, "analysis": {}, "meta": {},
	}

	AllowedPartyProperties = map[string]struct{}{
		"tel": {}, "stir": {}, "mailto": {}, "name": {}, "validation": {},
		"gmlpos": {}, "civicaddress": {}, "timezone": {}, "uuid": {},
		"role": {}, "contact_list": {}, "meta": {},
	}

	AllowedDialogProperties = map[string]struct{}{
		"type": {}, "start": {}, "duration": {}, "parties": {}, "originator": {},
		"mediatype": {}, "filename": {}, "body": {}, "encoding": {},
		"url": {}, "content_hash": {}, "alg": {}, "signature": {},
		"disposition": {}, "party_history": {}, "transferee": {}, "transferor": {},
		"transfer_target": {}, "original": {}, "consultation": {}, "target_dialog": {},
		"campaign": {}, "interaction_type": {}, "interaction_id": {}, "skill": {},
		"application": {}, "message_id": {}, "meta": {},
	}

	AllowedAttachmentProperties = map[string]struct{}{
		"body": {}, "encoding": {}, "url": {}, "content_hash": {}, "dialog": {},
		"party": {}, "start": {}, "mediatype": {}, "filename": {}, "meta": {},
	}

	AllowedAnalysisProperties = map[string]struct{}{
		"type": {}, "dialog": {}, "mediatype": {}, "filename": {}, "vendor": {},
		"product": {}, "schema": {}, "body": {}, "encoding": {}, "url": {},
		"content_hash": {}, "meta": {},
	}
)

// Global for UUID8 timestamp tracking
var lastV8Timestamp int64

// Core Types

// VCon is the top-level container.
type VCon struct {
	Vcon        string       `json:"vcon"`
	UUID        string       `json:"uuid"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
	Subject     string       `json:"subject,omitempty"`
	Group       interface{}  `json:"group,omitempty"`
	Redacted    interface{}  `json:"redacted,omitempty"`
	Appended    bool         `json:"appended,omitempty"`
	Parties     []Party      `json:"parties,omitempty"`
	Dialog      []Dialog     `json:"dialog,omitempty"`
	Analysis    []Analysis   `json:"analysis,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Meta        interface{}  `json:"meta,omitempty"`
	
	// Internal property handling mode
	propertyHandling string `json:"-"`
}

// Analysis holds machine-generated artefacts.
type Analysis struct {
	Type        string      `json:"type"`
	Dialog      interface{} `json:"dialog,omitempty"`
	MediaType   string      `json:"mediatype,omitempty"`
	Filename    string      `json:"filename,omitempty"`
	Vendor      string      `json:"vendor,omitempty"`
	Product     string      `json:"product,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
	Body        interface{} `json:"body,omitempty"`
	Encoding    string      `json:"encoding,omitempty"`
	URL         string      `json:"url,omitempty"`
	ContentHash string      `json:"content_hash,omitempty"`
	Meta        interface{} `json:"meta,omitempty"`
}

// Attachment is a file linked to a dialog/party.
type Attachment struct {
	Body        string      `json:"body,omitempty"`
	Encoding    string      `json:"encoding,omitempty"`
	URL         string      `json:"url,omitempty"`
	ContentHash string      `json:"content_hash,omitempty"`
	DialogIdx   int         `json:"dialog,omitempty"`
	PartyIdx    int         `json:"party"`
	StartTime   time.Time   `json:"start"`
	MediaType   string      `json:"mediatype,omitempty"`
	Filename    string      `json:"filename,omitempty"`
	Meta        interface{} `json:"meta,omitempty"`
}

// ProcessProperties handles properties based on the provided mode
func ProcessProperties(obj map[string]interface{}, allowedProps map[string]struct{}, mode string) map[string]interface{} {
	if obj == nil {
		return nil
	}

	result := make(map[string]interface{})
	nonStandard := make(map[string]interface{})

	// Separate standard and non-standard properties
	for k, v := range obj {
		_, isAllowed := allowedProps[k]
		if isAllowed {
			result[k] = v
		} else {
			nonStandard[k] = v
		}
	}

	// Handle non-standard properties based on mode
	switch mode {
	case PropertyHandlingStrict:
		// Ignore non-standard properties
	case PropertyHandlingMeta:
		// Move non-standard properties to meta
		if len(nonStandard) > 0 {
			meta, exists := result["meta"]
			if !exists {
				meta = make(map[string]interface{})
			}
			metaMap, ok := meta.(map[string]interface{})
			if !ok {
				metaMap = make(map[string]interface{})
			}
			for k, v := range nonStandard {
				metaMap[k] = v
			}
			result["meta"] = metaMap
		}
	default: // PropertyHandlingDefault
		// Keep non-standard properties
		for k, v := range nonStandard {
			result[k] = v
		}
	}

	return result
}

// New creates an empty, valid container with property handling options.
func New(propertyHandling ...string) *VCon {
	handling := PropertyHandlingDefault
	if len(propertyHandling) > 0 {
		handling = propertyHandling[0]
	}

	vcon := &VCon{
		Vcon:             SpecVersion,
		UUID:             UUID8DomainName("strolid.com"),
		CreatedAt:        time.Now().UTC(),
		Parties:          []Party{},
		Dialog:           []Dialog{},
		Analysis:         []Analysis{},
		Attachments:      []Attachment{},
		propertyHandling: handling,
	}
	return vcon
}

// BuildFromJSON creates a VCon from a JSON string
func BuildFromJSON(jsonStr string, propertyHandling ...string) (*VCon, error) {
	handling := PropertyHandlingDefault
	if len(propertyHandling) > 0 {
		handling = propertyHandling[0]
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process top-level properties
	processedMap := ProcessProperties(rawMap, AllowedVConProperties, handling)

	// Handle created_at if it's a string
	if createdAt, ok := processedMap["created_at"].(string); ok {
		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("invalid created_at format: %w", err)
		}
		processedMap["created_at"] = parsedTime
	}

	// Process nested structures
	if parties, ok := processedMap["parties"].([]interface{}); ok {
		processedParties := make([]interface{}, len(parties))
		for i, party := range parties {
			if partyMap, ok := party.(map[string]interface{}); ok {
				processedParties[i] = ProcessProperties(partyMap, AllowedPartyProperties, handling)
			} else {
				processedParties[i] = party
			}
		}
		processedMap["parties"] = processedParties
	}

	if dialogs, ok := processedMap["dialog"].([]interface{}); ok {
		processedDialogs := make([]interface{}, len(dialogs))
		for i, dialog := range dialogs {
			if dialogMap, ok := dialog.(map[string]interface{}); ok {
				processedDialogs[i] = ProcessProperties(dialogMap, AllowedDialogProperties, handling)
			} else {
				processedDialogs[i] = dialog
			}
		}
		processedMap["dialog"] = processedDialogs
	}

	if attachments, ok := processedMap["attachments"].([]interface{}); ok {
		processedAttachments := make([]interface{}, len(attachments))
		for i, attachment := range attachments {
			if attachmentMap, ok := attachment.(map[string]interface{}); ok {
				processedAttachments[i] = ProcessProperties(attachmentMap, AllowedAttachmentProperties, handling)
			} else {
				processedAttachments[i] = attachment
			}
		}
		processedMap["attachments"] = processedAttachments
	}

	if analyses, ok := processedMap["analysis"].([]interface{}); ok {
		processedAnalyses := make([]interface{}, len(analyses))
		for i, analysis := range analyses {
			if analysisMap, ok := analysis.(map[string]interface{}); ok {
				processedAnalyses[i] = ProcessProperties(analysisMap, AllowedAnalysisProperties, handling)
			} else {
				processedAnalyses[i] = analysis
			}
		}
		processedMap["analysis"] = processedAnalyses
	}

	// Marshal back to JSON and then to VCon
	processedJSON, err := json.Marshal(processedMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal processed map: %w", err)
	}

	var vcon VCon
	if err := json.Unmarshal(processedJSON, &vcon); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to VCon: %w", err)
	}

	vcon.propertyHandling = handling
	return &vcon, nil
}

// UUID8DomainName generates a UUID8 using a domain name
func UUID8DomainName(domain string) string {
	// SHA1 hash the domain name
	hasher := sha1.New()
	hasher.Write([]byte(domain))
	dnSHA1 := hasher.Sum(nil)

	// Get upper 64 bits of the hash
	hashUpper64 := dnSHA1[0:8]
	var int64Val uint64
	for _, b := range hashUpper64 {
		int64Val = (int64Val << 8) | uint64(b)
	}

	return UUID8Time(int64Val)
}

// UUID8Time generates a UUID8 using a timestamp and custom bits
func UUID8Time(customC62Bits uint64) string {
	now := time.Now().UnixNano()
	
	// Ensure timestamp is monotonically increasing
	if now <= lastV8Timestamp {
		now = lastV8Timestamp + 1
	}
	lastV8Timestamp = now

	
	// Create UUID v7 format: timestamp_ms + rand
	// Then modify version bits to make it UUID v8
	uuidV7, err := uuid.NewV7()
	if err != nil {
		// Fallback to V4 if V7 fails
		uuidV7 = uuid.New()
	}
	uuidBytes := uuidV7[:]
	
	// Set the version to 8
	uuidBytes[6] = (uuidBytes[6] & 0x0F) | 0x80
	
	// Create UUID from the bytes
	uuidObj, _ := uuid.FromBytes(uuidBytes)
	uuidStr := uuidObj.String()
	
	return uuidStr
}

// ToJSON serializes the VCon to a JSON string
func (v *VCon) ToJSON() string {
	data, _ := json.Marshal(v)
	return string(data)
}

// ToMap converts the VCon to a map
func (v *VCon) ToMap() map[string]interface{} {
	var result map[string]interface{}
	data, _ := json.Marshal(v)
	json.Unmarshal(data, &result)
	return result
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

// FindPartyIndex finds the index of a party with a matching property value
func (v *VCon) FindPartyIndex(by string, val interface{}) int {
	for i, party := range v.Parties {
		partyMap := structToMap(party)
		if partyVal, ok := partyMap[by]; ok && partyVal == val {
			return i
		}
	}
	return -1
}

// FindDialogByProperty finds a dialog with a matching property value
func (v *VCon) FindDialogByProperty(by string, val interface{}) *Dialog {
	for _, dialog := range v.Dialog {
		dialogMap := structToMap(dialog)
		if dialogVal, ok := dialogMap[by]; ok && dialogVal == val {
			return &dialog
		}
	}
	return nil
}

// FindAttachmentByType finds an attachment by its type
func (v *VCon) FindAttachmentByType(attachmentType string) map[string]interface{} {
	for _, att := range v.Attachments {
		if att.Encoding == attachmentType {
			return structToMap(att)
		}
	}
	return nil
}

// FindAnalysisByType finds an analysis entry by its type
func (v *VCon) FindAnalysisByType(analysisType string) map[string]interface{} {
	for _, analysis := range v.Analysis {
		if analysis.Type == analysisType {
			return structToMap(analysis)
		}
	}
	return nil
}

// AddTag adds a tag to the VCon
func (v *VCon) AddTag(tagName string, tagValue string) {
	tagsAttachment := v.FindAttachmentByType("tags")
	if tagsAttachment == nil {
		// Create new tags attachment
		v.AddAttachment(Attachment{
			Encoding: "tags",
			Body:     fmt.Sprintf("%s:%s", tagName, tagValue),
		})
		return
	}

	// Add tag to existing tags
	currentTags, ok := tagsAttachment["body"].(string)
	if !ok {
		tagsAttachment["body"] = fmt.Sprintf("%s:%s", tagName, tagValue)
	} else {
		tagsAttachment["body"] = fmt.Sprintf("%s,%s:%s", currentTags, tagName, tagValue)
	}
}

// GetTag gets a tag value by its name
func (v *VCon) GetTag(tagName string) string {
	tagsAttachment := v.FindAttachmentByType("tags")
	if tagsAttachment == nil {
		return ""
	}

	tags, ok := tagsAttachment["body"].(string)
	if !ok {
		return ""
	}

	// Parse tags
	tagPairs := parseTags(tags)
	return tagPairs[tagName]
}

// Helper to parse tags
func parseTags(tagString string) map[string]string {
	result := make(map[string]string)
	for _, tag := range strings.Split(tagString, ",") {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// Helper to convert struct to map
func structToMap(obj interface{}) map[string]interface{} {
	data, _ := json.Marshal(obj)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// SaveToFile saves the VCon to a file
func (v *VCon) SaveToFile(filePath string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal VCon: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadFromFile loads a VCon from a file
func LoadFromFile(filePath string, propertyHandling ...string) (*VCon, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return BuildFromJSON(string(data), propertyHandling...)
}

// LoadFromURL loads a VCon from a URL
func LoadFromURL(url string, propertyHandling ...string) (*VCon, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return BuildFromJSON(string(data), propertyHandling...)
}

// PostToURL posts a VCon to a URL
func (v *VCon) PostToURL(url string, headers map[string]string) (*http.Response, error) {
	jsonData := v.ToJSON()
	
	req, err := http.NewRequest("POST", url, strings.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{}
	return client.Do(req)
}

// Validate validates the VCon structure
func (v *VCon) Validate() error {
	// Check required fields
	if v.Vcon == "" {
		return fmt.Errorf("missing required field: vcon")
	}
	if v.UUID == "" {
		return fmt.Errorf("missing required field: uuid")
	}
	if v.CreatedAt.IsZero() {
		return fmt.Errorf("missing required field: created_at")
	}

	// Validate dialogs
	for i, dialog := range v.Dialog {
		// Check if dialog references valid parties
		if parties, ok := dialog.Parties.([]int); ok {
			for _, partyIdx := range parties {
				if partyIdx < 0 || partyIdx >= len(v.Parties) {
					return fmt.Errorf("dialog at index %d references invalid party index: %d", i, partyIdx)
				}
			}
		}
		
		// Check required dialog fields
		if dialog.Type == "" {
			return fmt.Errorf("dialog at index %d missing required field: type", i)
		}
		if dialog.StartTime == nil {
			return fmt.Errorf("dialog at index %d missing required field: start", i)
		}
	}

	// Validate analysis
	for i, analysis := range v.Analysis {
		// Check if analysis references valid dialogs
		if dialogs, ok := analysis.Dialog.([]int); ok {
			for _, dialogIdx := range dialogs {
				if dialogIdx < 0 || dialogIdx >= len(v.Dialog) {
					return fmt.Errorf("analysis at index %d references invalid dialog index: %d", i, dialogIdx)
				}
			}
		}
	}

	return nil
}

// IsValid validates the VCon and returns if it's valid and any errors
func (v *VCon) IsValid() (bool, []string) {
	var errors []string
	
	// Check required fields
	if v.Vcon == "" {
		errors = append(errors, "missing required field: vcon")
	}
	if v.UUID == "" {
		errors = append(errors, "missing required field: uuid")
	}
	if v.CreatedAt.IsZero() {
		errors = append(errors, "missing required field: created_at")
	}

	// Validate dialogs
	for i, dialog := range v.Dialog {
		// Check if dialog references valid parties
		if parties, ok := dialog.Parties.([]int); ok {
			for _, partyIdx := range parties {
				if partyIdx < 0 || partyIdx >= len(v.Parties) {
					errors = append(errors, fmt.Sprintf("dialog at index %d references invalid party index: %d", i, partyIdx))
				}
			}
		}
		
		// Check required dialog fields
		if dialog.Type == "" {
			errors = append(errors, fmt.Sprintf("dialog at index %d missing required field: type", i))
		}
		if dialog.StartTime == nil {
			errors = append(errors, fmt.Sprintf("dialog at index %d missing required field: start", i))
		}
	}

	// Validate analysis
	for i, analysis := range v.Analysis {
		// Check if analysis references valid dialogs
		if dialogs, ok := analysis.Dialog.([]int); ok {
			for _, dialogIdx := range dialogs {
				if dialogIdx < 0 || dialogIdx >= len(v.Dialog) {
					errors = append(errors, fmt.Sprintf("analysis at index %d references invalid dialog index: %d", i, dialogIdx))
				}
			}
		}
	}

	return len(errors) == 0, errors
}
