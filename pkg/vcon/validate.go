package vcon

import (
	_ "embed"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schema/vcon.json
var rawSchema []byte

var compiledSchema = jsonschema.MustCompileString("schema", string(rawSchema))

// Validate runs JSON-Schema validation + custom cross-checks.
func (v *VCon) Validate() error {
	if err := compiledSchema.Validate(v); err != nil {
		return err
	}

	for _, dlg := range v.Dialog {
		if dlg.Originator >= len(v.Parties) {
			return ErrInvalidIndex
		}
		for _, idx := range dlg.DestParties {
			if idx >= len(v.Parties) {
				return ErrInvalidIndex
			}
		}
	}
	return nil
}

var ErrInvalidIndex = &jsonschema.ValidationError{
	Message: "dialog references out-of-range party index",
}
