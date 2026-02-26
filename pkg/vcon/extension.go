package vcon

import (
	"fmt"
	"maps"
	"strings"
	"sync"
)

// Extension defines a vCon extension per Section 2.5 of the spec.
// Extensions add new parameters to vCon object types.
type Extension interface {
	// Name returns the extension token (e.g., "CC").
	// Listed in the vCon extensions[] array.
	Name() string

	// IsCompatible returns true if this extension is compatible
	// (adds fields without altering existing semantics).
	// Incompatible extensions MUST be listed in critical[].
	IsCompatible() bool

	// PartyParams returns parameter names this extension adds to Party objects.
	PartyParams() []string

	// DialogParams returns parameter names this extension adds to Dialog objects.
	DialogParams() []string

	// AnalysisParams returns parameter names this extension adds to Analysis objects.
	AnalysisParams() []string

	// AttachmentParams returns parameter names this extension adds to Attachment objects.
	AttachmentParams() []string

	// VConParams returns parameter names this extension adds to the top-level VCon object.
	VConParams() []string
}

// ExtensionRegistry manages registered extensions.
// Thread-safe for concurrent use.
type ExtensionRegistry struct {
	mu         sync.RWMutex
	extensions map[string]Extension
}

// DefaultRegistry is the global default registry with built-in extensions pre-registered.
var DefaultRegistry = NewExtensionRegistry()

// NewExtensionRegistry creates a new empty extension registry.
func NewExtensionRegistry() *ExtensionRegistry {
	return &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}
}

// Register adds an extension to the registry.
func (r *ExtensionRegistry) Register(ext Extension) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.extensions[ext.Name()] = ext
}

// Get retrieves an extension by name.
func (r *ExtensionRegistry) Get(name string) (Extension, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ext, ok := r.extensions[name]
	return ext, ok
}

// Has checks whether an extension is registered.
func (r *ExtensionRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.extensions[name]
	return ok
}

// AllowedPartyParams returns core params merged with all registered extension params.
func (r *ExtensionRegistry) AllowedPartyParams() map[string]struct{} {
	result := make(map[string]struct{})
	maps.Copy(result, AllowedPartyProperties)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ext := range r.extensions {
		for _, p := range ext.PartyParams() {
			result[p] = struct{}{}
		}
	}
	return result
}

// AllowedDialogParams returns core params merged with all registered extension params.
func (r *ExtensionRegistry) AllowedDialogParams() map[string]struct{} {
	result := make(map[string]struct{})
	maps.Copy(result, AllowedDialogProperties)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ext := range r.extensions {
		for _, p := range ext.DialogParams() {
			result[p] = struct{}{}
		}
	}
	return result
}

// AllowedAnalysisParams returns core params merged with all registered extension params.
func (r *ExtensionRegistry) AllowedAnalysisParams() map[string]struct{} {
	result := make(map[string]struct{})
	maps.Copy(result, AllowedAnalysisProperties)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ext := range r.extensions {
		for _, p := range ext.AnalysisParams() {
			result[p] = struct{}{}
		}
	}
	return result
}

// AllowedAttachmentParams returns core params merged with all registered extension params.
func (r *ExtensionRegistry) AllowedAttachmentParams() map[string]struct{} {
	result := make(map[string]struct{})
	maps.Copy(result, AllowedAttachmentProperties)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ext := range r.extensions {
		for _, p := range ext.AttachmentParams() {
			result[p] = struct{}{}
		}
	}
	return result
}

// AllowedVConParams returns core params merged with all registered extension params.
func (r *ExtensionRegistry) AllowedVConParams() map[string]struct{} {
	result := make(map[string]struct{})
	maps.Copy(result, AllowedVConProperties)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ext := range r.extensions {
		for _, p := range ext.VConParams() {
			result[p] = struct{}{}
		}
	}
	return result
}

// ValidateCritical checks that all names in the critical[] array
// are registered in this registry. Returns error listing unsupported ones.
func (r *ExtensionRegistry) ValidateCritical(critical []string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var unsupported []string
	for _, name := range critical {
		if _, ok := r.extensions[name]; !ok {
			unsupported = append(unsupported, name)
		}
	}
	if len(unsupported) > 0 {
		return fmt.Errorf("unsupported critical extensions: %s", strings.Join(unsupported, ", "))
	}
	return nil
}

// RegisteredNames returns the names of all registered extensions.
func (r *ExtensionRegistry) RegisteredNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.extensions))
	for name := range r.extensions {
		names = append(names, name)
	}
	return names
}
