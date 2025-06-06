package vcon

import (
	_ "embed"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schema/vcon.json
var rawSchema []byte

var compiledSchema = jsonschema.MustCompileString("schema", string(rawSchema))

// Validate runs JSON-Schema validation + custom cross-checks.
func (v *VCon) Validate() error {
	// Convert our struct to a map for validation to handle field name differences
	// between Go struct fields and JSON schema fields
	data, err := marshalToMap(v)
	if err != nil {
		return err
	}

	if err := compiledSchema.Validate(data); err != nil {
		return err
	}

	return nil
}

// marshalToMap converts the VCon struct to a map for validation,
// ensuring field names match the JSON schema
func marshalToMap(v *VCon) (map[string]interface{}, error) {
	// First marshal to JSON bytes
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Then unmarshal to a map
	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}

var ErrInvalidIndex = &jsonschema.ValidationError{
	Message: "dialog references out-of-range party index",
}
