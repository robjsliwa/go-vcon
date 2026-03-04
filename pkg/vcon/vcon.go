package vcon

import (
	"crypto/sha1"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema/vcon.json
var vconSchema []byte

// SpecVersion is the draft version this library targets.
const SpecVersion = "0.4.0"

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
		"group": {}, "redacted": {}, "amended": {}, "parties": {},
		"dialog": {}, "attachments": {}, "analysis": {},
		"extensions": {}, "critical": {},
	}

	AllowedPartyProperties = map[string]struct{}{
		"tel": {}, "stir": {}, "mailto": {}, "name": {}, "validation": {},
		"gmlpos": {}, "civicaddress": {}, "uuid": {},
		"sip": {}, "did": {},
	}

	AllowedDialogProperties = map[string]struct{}{
		"type": {}, "start": {}, "duration": {}, "parties": {}, "originator": {},
		"mediatype": {}, "filename": {}, "body": {}, "encoding": {},
		"url": {}, "content_hash": {},
		"disposition": {}, "party_history": {}, "transferee": {}, "transferor": {},
		"transfer_target": {}, "original": {}, "consultation": {}, "target_dialog": {},
		"application": {}, "message_id": {}, "session_id": {},
	}

	AllowedAttachmentProperties = map[string]struct{}{
		"body": {}, "encoding": {}, "url": {}, "content_hash": {}, "dialog": {},
		"party": {}, "start": {}, "mediatype": {}, "filename": {}, "purpose": {},
	}

	AllowedAnalysisProperties = map[string]struct{}{
		"type": {}, "dialog": {}, "mediatype": {}, "filename": {}, "vendor": {},
		"product": {}, "schema": {}, "body": {}, "encoding": {}, "url": {},
		"content_hash": {},
	}
)

// Global for UUID8 timestamp tracking
var lastV8Timestamp int64

// Core Types

// VCon is the top-level container.
type VCon struct {
	Vcon        string            `json:"vcon,omitempty"`
	UUID        string            `json:"uuid"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
	Subject     string            `json:"subject,omitempty"`
	Group       []json.RawMessage `json:"group,omitempty"`
	Redacted    *RedactedObject   `json:"redacted,omitempty"`
	Amended     *AmendedObject    `json:"amended,omitempty"`
	Extensions  []string          `json:"extensions,omitempty"`
	Critical    []string          `json:"critical,omitempty"`
	Parties     []Party           `json:"parties"`
	Dialog      []Dialog          `json:"dialog,omitempty"`
	Analysis    []Analysis        `json:"analysis,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`

	// Internal fields
	propertyHandling string             `json:"-"`
	registry         *ExtensionRegistry `json:"-"`
}

// Analysis holds machine-generated artefacts.
type Analysis struct {
	Type        string          `json:"type"`
	Dialog      interface{}     `json:"dialog,omitempty"`
	MediaType   string          `json:"mediatype,omitempty"`
	Filename    string          `json:"filename,omitempty"`
	Vendor      string          `json:"vendor,omitempty"`
	Product     string          `json:"product,omitempty"`
	Schema      interface{}     `json:"schema,omitempty"`
	Body        interface{}     `json:"body,omitempty"`
	Encoding    string          `json:"encoding,omitempty"`
	URL         string          `json:"url,omitempty"`
	ContentHash ContentHashList `json:"content_hash,omitempty"`
}

// ProcessProperties handles properties based on the provided mode.
// The optional registry parameter merges extension params into the allowed set.
func ProcessProperties(obj map[string]interface{}, allowedProps map[string]struct{}, mode string, registry ...*ExtensionRegistry) map[string]interface{} {
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

// VConOption configures a VCon.
type VConOption func(*VCon)

// WithRegistry sets a custom extension registry on a VCon.
func WithRegistry(r *ExtensionRegistry) VConOption {
	return func(v *VCon) {
		v.registry = r
	}
}

// New creates an empty, valid container with property handling options.
func New(domain string, propertyHandling ...string) *VCon {
	handling := PropertyHandlingDefault
	if len(propertyHandling) > 0 {
		handling = propertyHandling[0]
	}

	vcon := &VCon{
		Vcon:             SpecVersion,
		UUID:             UUID8DomainName(domain),
		CreatedAt:        time.Now().UTC(),
		Parties:          []Party{},
		Dialog:           []Dialog{},
		Analysis:         []Analysis{},
		Attachments:      []Attachment{},
		propertyHandling: handling,
		registry:         DefaultRegistry,
	}
	return vcon
}

func validateAgainstSchema(rawMap map[string]interface{}) error {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)

	var schemaData interface{}
	if err := json.Unmarshal(vconSchema, &schemaData); err != nil {
		return err
	}
	if err := compiler.AddResource("vcon.schema.json", schemaData); err != nil {
		return err
	}
	schema, err := compiler.Compile("vcon.schema.json")
	if err != nil {
		return err
	}
	if err := schema.Validate(rawMap); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

func processNestedSlices(m map[string]interface{}, handling string) {
	sliceProps := []struct {
		key     string
		allowed map[string]struct{}
	}{
		{"parties", AllowedPartyProperties},
		{"dialog", AllowedDialogProperties},
		{"attachments", AllowedAttachmentProperties},
		{"analysis", AllowedAnalysisProperties},
	}
	for _, sp := range sliceProps {
		items, ok := m[sp.key].([]interface{})
		if !ok {
			continue
		}
		processed := make([]interface{}, len(items))
		for i, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				processed[i] = ProcessProperties(itemMap, sp.allowed, handling)
			} else {
				processed[i] = item
			}
		}
		m[sp.key] = processed
	}
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

	// Auto-detect v0.0.3 and migrate
	if ver, ok := rawMap["vcon"].(string); ok && ver == "0.0.3" {
		migrateV003ToV040(rawMap)
	}

	if err := validateAgainstSchema(rawMap); err != nil {
		return nil, err
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

	processNestedSlices(processedMap, handling)

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
	vcon.registry = DefaultRegistry
	return &vcon, nil
}

// migrateV003ToV040 converts a v0.0.3 raw map to v0.4.0 format in-place.
// migrateSliceItems applies a migration function to each map item in a JSON array field.
func migrateSliceItems(m map[string]interface{}, key string, fn func(map[string]interface{})) {
	items, ok := m[key].([]interface{})
	if !ok {
		return
	}
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			fn(itemMap)
		}
	}
}

// migrateEncodingAndHash converts "base64" to "base64url" and fixes content_hash format.
func migrateEncodingAndHash(m map[string]interface{}) {
	if enc, ok := m["encoding"].(string); ok && enc == "base64" {
		m["encoding"] = "base64url"
	}
	migrateContentHash(m)
}

func migrateV003ToV040(m map[string]interface{}) {
	m["vcon"] = "0.4.0"

	delete(m, "appended")
	delete(m, "meta")

	if _, ok := m["parties"]; !ok {
		m["parties"] = []interface{}{}
	}

	migrateSliceItems(m, "dialog", func(dm map[string]interface{}) {
		delete(dm, "alg")
		delete(dm, "signature")
		delete(dm, "meta")
		migrateEncodingAndHash(dm)
		delete(dm, "campaign")
		delete(dm, "interaction_type")
		delete(dm, "interaction_id")
		delete(dm, "skill")
	})

	migrateSliceItems(m, "parties", func(pm map[string]interface{}) {
		delete(pm, "role")
		delete(pm, "contact_list")
		delete(pm, "timezone")
		delete(pm, "meta")
	})

	migrateSliceItems(m, "attachments", func(am map[string]interface{}) {
		delete(am, "meta")
		migrateEncodingAndHash(am)
	})

	migrateSliceItems(m, "analysis", func(am map[string]interface{}) {
		delete(am, "meta")
		migrateEncodingAndHash(am)
	})
}

// migrateContentHash converts content_hash from old "alg:hash" format to "alg-hash".
func migrateContentHash(m map[string]interface{}) {
	ch, ok := m["content_hash"].(string)
	if !ok || ch == "" {
		return
	}
	m["content_hash"] = strings.ReplaceAll(ch, ":", "-")
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

func (v *VCon) validateCoreFields() []string {
	var errs []string
	if v.UUID == "" {
		errs = append(errs, "missing required field: uuid")
	}
	if v.CreatedAt.IsZero() {
		errs = append(errs, "missing required field: created_at")
	}
	return errs
}

func (v *VCon) validateMutualExclusion() []string {
	count := 0
	if v.Redacted != nil {
		count++
	}
	if v.Amended != nil {
		count++
	}
	if len(v.Group) > 0 {
		count++
	}
	if count > 1 {
		return []string{"redacted, amended, and group are mutually exclusive"}
	}
	return nil
}

func (v *VCon) validateCriticalExtensions() []string {
	if len(v.Critical) == 0 {
		return nil
	}
	reg := v.registry
	if reg == nil {
		reg = DefaultRegistry
	}
	if err := reg.ValidateCritical(v.Critical); err != nil {
		return []string{fmt.Sprintf("critical extension validation: %s", err)}
	}
	return nil
}

func (v *VCon) validateDialogs() []string {
	var errs []string
	for i, dialog := range v.Dialog {
		if parties, ok := dialog.Parties.([]int); ok {
			for _, partyIdx := range parties {
				if partyIdx < 0 || partyIdx >= len(v.Parties) {
					errs = append(errs, fmt.Sprintf("dialog at index %d references invalid party index: %d", i, partyIdx))
				}
			}
		}
		if dialog.Type == "" {
			errs = append(errs, fmt.Sprintf("dialog at index %d missing required field: type", i))
		}
		if dialog.StartTime == nil {
			errs = append(errs, fmt.Sprintf("dialog at index %d missing required field: start", i))
		}
	}
	return errs
}

func (v *VCon) validateAnalysis() []string {
	var errs []string
	for i, analysis := range v.Analysis {
		if dialogs, ok := analysis.Dialog.([]int); ok {
			for _, dialogIdx := range dialogs {
				if dialogIdx < 0 || dialogIdx >= len(v.Dialog) {
					errs = append(errs, fmt.Sprintf("analysis at index %d references invalid dialog index: %d", i, dialogIdx))
				}
			}
		}
	}
	return errs
}

func (v *VCon) allValidationErrors() []string {
	var errs []string
	errs = append(errs, v.validateCoreFields()...)
	errs = append(errs, v.validateMutualExclusion()...)
	errs = append(errs, v.validateCriticalExtensions()...)
	errs = append(errs, v.validateDialogs()...)
	errs = append(errs, v.validateAnalysis()...)
	return errs
}

// Validate validates the VCon structure
func (v *VCon) Validate() error {
	if errs := v.allValidationErrors(); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}
	return nil
}

// IsValid validates the VCon and returns if it's valid and any errors
func (v *VCon) IsValid() (bool, []string) {
	errs := v.allValidationErrors()
	return len(errs) == 0, errs
}
