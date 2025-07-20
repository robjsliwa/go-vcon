package vcon_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	v := vcon.New("example.com")
	assert.Equal(t, vcon.SpecVersion, v.Vcon)
	assert.NotEqual(t, uuid.UUID{}, v.UUID)
	assert.False(t, v.CreatedAt.IsZero())
}

func TestRoundTrip(t *testing.T) {
	// Create a new vcon for testing
	v := vcon.New("example.com")
	v.Subject = "demo"

	idx := v.AddParty(vcon.Party{Name: "Alice"})
	assert.Equal(t, 0, idx)
	
	now := time.Now().UTC()
	v.AddDialog(vcon.Dialog{
        StartTime: &now,
        Originator: 0,
        Type: "text",
        MediaType: "audio/wav",
        ContentHash: "test-hash",
        Body: "Hello Alice!",
        Parties: 1,
        Encoding: "base64",
    })

	// Test the JSON marshaling and unmarshaling
	data, err := json.Marshal(v)
	require.NoError(t, err)
	
	var out vcon.VCon
	err = json.Unmarshal(data, &out)
	require.NoError(t, err)
	
	// Re-enable validation now that we've fixed the issues
	err = out.Validate()
	require.NoError(t, err, "validate: %v", err)

	// Verify the data was preserved
	assert.Equal(t, v.Subject, out.Subject)
	assert.Equal(t, v.UUID, out.UUID)
	assert.Equal(t, v.Vcon, out.Vcon)
	assert.Equal(t, len(v.Parties), len(out.Parties))
	assert.Equal(t, len(v.Dialog), len(out.Dialog))
}

func TestAddParty(t *testing.T) {
	v := vcon.New("example.com")
	
	idx1 := v.AddParty(vcon.Party{Name: "Alice"})
	idx2 := v.AddParty(vcon.Party{Name: "Bob"})
	
	assert.Equal(t, 0, idx1)
	assert.Equal(t, 1, idx2)
	assert.Equal(t, 2, len(v.Parties))
	assert.Equal(t, "Alice", v.Parties[0].Name)
	assert.Equal(t, "Bob", v.Parties[1].Name)
}

func TestAddDialog(t *testing.T) {
	v := vcon.New("example.com")
	
	now := time.Now().UTC()
	idx := v.AddDialog(vcon.Dialog{
		StartTime: &now, 
		Duration: (5 * time.Second).Seconds(),
		MediaType: "audio/wav",
		ContentHash: "test-hash",
	})
	
	assert.Equal(t, 0, idx)
	assert.Equal(t, 1, len(v.Dialog))
	assert.Equal(t, "audio/wav", v.Dialog[0].MediaType)
}

func TestAddAnalysis(t *testing.T) {
	v := vcon.New("example.com")
	
	idx := v.AddAnalysis(vcon.Analysis{
		Type: "transcript",
		Vendor: "test-vendor",
		Product: "test-product",
		ContentHash: "test-content",
	})
	
	assert.Equal(t, 0, idx)
	assert.Equal(t, 1, len(v.Analysis))
	assert.Equal(t, "transcript", v.Analysis[0].Type)
	assert.Equal(t, "test-vendor", v.Analysis[0].Vendor)
	assert.Equal(t, "test-product", v.Analysis[0].Product)
}

