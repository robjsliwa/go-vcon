package vcon

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// ValidAttachmentEncodings defines the allowed encoding types for attachments
var ValidAttachmentEncodings = []string{"base64", "base64url", "json", "none"}

// AttachmentType represents a specific type of attachment
type AttachmentType string

const (
	// AttachmentTypeTags is for tag collections
	AttachmentTypeTags AttachmentType = "tags"
	// AttachmentTypeMetadata is for general metadata
	AttachmentTypeMetadata AttachmentType = "metadata"
	// AttachmentTypeDocument is for document attachments
	AttachmentTypeDocument AttachmentType = "document"
)

// NewAttachment creates a new Attachment with the specified type, body, and encoding
func NewAttachment(attachmentType string, body interface{}, encoding string) (*Attachment, error) {
	// Validate encoding
	validEncoding := false
	for _, enc := range ValidAttachmentEncodings {
		if encoding == enc {
			validEncoding = true
			break
		}
	}
	
	if !validEncoding {
		return nil, fmt.Errorf("invalid encoding: %s", encoding)
	}
	
	// Convert body to string if it's not already
	var bodyStr string
	switch b := body.(type) {
	case string:
		bodyStr = b
	default:
		// For JSON encoding, convert to JSON string
		if encoding == "json" {
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
			}
			bodyStr = string(data)
		} else {
			bodyStr = fmt.Sprintf("%v", body)
		}
	}
	
	// Create the attachment
	att := &Attachment{
		Body:     bodyStr,
		Encoding: encoding,
		Meta:     make(map[string]interface{}),
	}
	
	return att, nil
}

// GetBody retrieves the body content, converting from the encoded format if necessary
func (a *Attachment) GetBody() (interface{}, error) {
	switch a.Encoding {
	case "base64":
		// Decode base64 if needed
		decoded, err := base64.StdEncoding.DecodeString(a.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 body: %w", err)
		}
		return string(decoded), nil
	
	case "base64url":
		// Decode base64url if needed
		decoded, err := base64.URLEncoding.DecodeString(a.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64url body: %w", err)
		}
		return string(decoded), nil
	
	case "json":
		// Parse JSON if needed
		var result interface{}
		if err := json.Unmarshal([]byte(a.Body), &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body: %w", err)
		}
		return result, nil
	
	default: // "none" or any other encoding
		return a.Body, nil
	}
}
