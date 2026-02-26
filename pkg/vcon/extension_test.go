package vcon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testExtension is a mock extension for testing.
type testExtension struct {
	name         string
	compatible   bool
	partyParams  []string
	dialogParams []string
}

func (e testExtension) Name() string            { return e.name }
func (e testExtension) IsCompatible() bool       { return e.compatible }
func (e testExtension) PartyParams() []string    { return e.partyParams }
func (e testExtension) DialogParams() []string   { return e.dialogParams }
func (e testExtension) AnalysisParams() []string { return nil }
func (e testExtension) AttachmentParams() []string { return nil }
func (e testExtension) VConParams() []string     { return nil }

func TestExtensionRegistryRegisterAndGet(t *testing.T) {
	r := NewExtensionRegistry()

	ext := testExtension{
		name:         "TEST",
		compatible:   true,
		partyParams:  []string{"custom_field"},
		dialogParams: []string{"custom_dialog_field"},
	}

	r.Register(ext)

	got, ok := r.Get("TEST")
	require.True(t, ok)
	assert.Equal(t, "TEST", got.Name())
	assert.True(t, got.IsCompatible())
}

func TestExtensionRegistryHas(t *testing.T) {
	r := NewExtensionRegistry()
	assert.False(t, r.Has("TEST"))

	r.Register(testExtension{name: "TEST"})
	assert.True(t, r.Has("TEST"))
}

func TestExtensionRegistryValidateCritical(t *testing.T) {
	r := NewExtensionRegistry()
	r.Register(testExtension{name: "A"})
	r.Register(testExtension{name: "B"})

	// All critical extensions are registered
	assert.NoError(t, r.ValidateCritical([]string{"A", "B"}))

	// Missing critical extension
	err := r.ValidateCritical([]string{"A", "C"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "C")

	// Empty critical list
	assert.NoError(t, r.ValidateCritical(nil))
	assert.NoError(t, r.ValidateCritical([]string{}))
}

func TestExtensionRegistryAllowedPartyParams(t *testing.T) {
	r := NewExtensionRegistry()
	r.Register(testExtension{
		name:        "TEST",
		partyParams: []string{"custom_field", "another_field"},
	})

	allowed := r.AllowedPartyParams()

	// Should include core params
	_, hasTel := allowed["tel"]
	assert.True(t, hasTel)

	// Should include extension params
	_, hasCustom := allowed["custom_field"]
	assert.True(t, hasCustom)
	_, hasAnother := allowed["another_field"]
	assert.True(t, hasAnother)
}

func TestExtensionRegistryAllowedDialogParams(t *testing.T) {
	r := NewExtensionRegistry()
	r.Register(testExtension{
		name:         "TEST",
		dialogParams: []string{"custom_dialog_field"},
	})

	allowed := r.AllowedDialogParams()

	// Should include core params
	_, hasType := allowed["type"]
	assert.True(t, hasType)

	// Should include extension params
	_, hasCustom := allowed["custom_dialog_field"]
	assert.True(t, hasCustom)
}

func TestExtensionRegistryRegisteredNames(t *testing.T) {
	r := NewExtensionRegistry()
	r.Register(testExtension{name: "A"})
	r.Register(testExtension{name: "B"})

	names := r.RegisteredNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "A")
	assert.Contains(t, names, "B")
}

func TestExtensionRegistryOverwrite(t *testing.T) {
	r := NewExtensionRegistry()
	r.Register(testExtension{name: "TEST", partyParams: []string{"v1_field"}})
	r.Register(testExtension{name: "TEST", partyParams: []string{"v2_field"}})

	ext, ok := r.Get("TEST")
	require.True(t, ok)
	assert.Equal(t, []string{"v2_field"}, ext.PartyParams())
}
