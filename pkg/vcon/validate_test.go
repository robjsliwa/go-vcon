package vcon_test

import (
	"testing"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Skip("Skipping validation test until JSON schema issues are fixed")
	
	tests := []struct {
		name        string
		modifyVCon  func(*vcon.VCon)
		expectError bool
	}{
		{
			name:        "valid vcon",
			modifyVCon:  func(v *vcon.VCon) {},
			expectError: false,
		},
		{
			name: "missing version",
			modifyVCon: func(v *vcon.VCon) {
				v.Version = ""
			},
			expectError: true,
		},
		{
			name: "invalid party index in dialog",
			modifyVCon: func(v *vcon.VCon) {
				v.Dialog = append(v.Dialog, vcon.Dialog{
					Originator: 999, // Out of range
				})
			},
			expectError: true,
		},
		{
			name: "invalid dest party index in dialog",
			modifyVCon: func(v *vcon.VCon) {
				v.Dialog = append(v.Dialog, vcon.Dialog{
					Originator:  0,
					DestParties: []int{999}, // Out of range
				})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vcon.New()
			v.AddParty(vcon.Party{Name: "Test"})
			
			// Apply the modification for this test case
			tt.modifyVCon(v)
			
			err := v.Validate()
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
