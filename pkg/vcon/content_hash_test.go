package vcon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseContentHash(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantAlg string
		wantH   string
		wantErr bool
	}{
		{
			name:    "valid sha512",
			input:   "sha512-abc123",
			wantAlg: "sha512",
			wantH:   "abc123",
		},
		{
			name:    "valid sha256",
			input:   "sha256-xyz",
			wantAlg: "sha256",
			wantH:   "xyz",
		},
		{
			name:    "missing separator",
			input:   "sha512abc",
			wantErr: true,
		},
		{
			name:    "empty algorithm",
			input:   "-abc123",
			wantErr: true,
		},
		{
			name:    "empty hash",
			input:   "sha512-",
			wantErr: true,
		},
		{
			name:    "hash containing hyphens",
			input:   "sha512-abc-def-ghi",
			wantAlg: "sha512",
			wantH:   "abc-def-ghi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := ParseContentHash(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAlg, ch.Algorithm)
			assert.Equal(t, tt.wantH, ch.Hash)
		})
	}
}

func TestContentHashString(t *testing.T) {
	ch := ContentHash{Algorithm: "sha512", Hash: "abc123"}
	assert.Equal(t, "sha512-abc123", ch.String())
}

func TestComputeSHA512(t *testing.T) {
	data := []byte("hello world")
	ch := ComputeSHA512(data)
	assert.Equal(t, "sha512", ch.Algorithm)
	assert.NotEmpty(t, ch.Hash)

	// Verify round-trip
	assert.True(t, ch.Verify(data))
	assert.False(t, ch.Verify([]byte("different data")))
}

func TestContentHashVerify(t *testing.T) {
	data := []byte("test data")
	ch := ComputeSHA512(data)

	assert.True(t, ch.Verify(data))
	assert.False(t, ch.Verify([]byte("wrong data")))

	// Unknown algorithm returns false
	ch2 := ContentHash{Algorithm: "unknown", Hash: "abc"}
	assert.False(t, ch2.Verify(data))
}

func TestContentHashIsZero(t *testing.T) {
	assert.True(t, ContentHash{}.IsZero())
	assert.False(t, ContentHash{Algorithm: "sha512", Hash: "abc"}.IsZero())
}

func TestContentHashListMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		list     ContentHashList
		expected string
	}{
		{
			name:     "single item serializes as string",
			list:     ContentHashList{{Algorithm: "sha512", Hash: "abc"}},
			expected: `"sha512-abc"`,
		},
		{
			name:     "multiple items serialize as array",
			list:     ContentHashList{{Algorithm: "sha512", Hash: "abc"}, {Algorithm: "sha256", Hash: "xyz"}},
			expected: `["sha512-abc","sha256-xyz"]`,
		},
		{
			name:     "empty list serializes as null",
			list:     ContentHashList{},
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.list)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestContentHashListUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ContentHashList
		wantErr bool
	}{
		{
			name:  "single string",
			input: `"sha512-abc"`,
			want:  ContentHashList{{Algorithm: "sha512", Hash: "abc"}},
		},
		{
			name:  "array of strings",
			input: `["sha512-abc","sha256-xyz"]`,
			want: ContentHashList{
				{Algorithm: "sha512", Hash: "abc"},
				{Algorithm: "sha256", Hash: "xyz"},
			},
		},
		{
			name:    "invalid format",
			input:   `123`,
			wantErr: true,
		},
		{
			name:    "invalid hash format in array",
			input:   `["nohyphen"]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var list ContentHashList
			err := json.Unmarshal([]byte(tt.input), &list)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, list)
		})
	}
}

func TestContentHashListContainsAlgorithm(t *testing.T) {
	list := ContentHashList{
		{Algorithm: "sha512", Hash: "abc"},
		{Algorithm: "sha256", Hash: "xyz"},
	}
	assert.True(t, list.ContainsAlgorithm("sha512"))
	assert.True(t, list.ContainsAlgorithm("sha256"))
	assert.False(t, list.ContainsAlgorithm("md5"))
}

func TestContentHashListFirst(t *testing.T) {
	list := ContentHashList{{Algorithm: "sha512", Hash: "abc"}}
	assert.Equal(t, ContentHash{Algorithm: "sha512", Hash: "abc"}, list.First())

	empty := ContentHashList{}
	assert.True(t, empty.First().IsZero())
}

func TestContentHashListIsEmpty(t *testing.T) {
	assert.True(t, ContentHashList{}.IsEmpty())
	assert.True(t, ContentHashList(nil).IsEmpty())
	assert.False(t, ContentHashList{{Algorithm: "sha512", Hash: "abc"}}.IsEmpty())
}

func TestContentHashListRoundTrip(t *testing.T) {
	// Test that a struct with ContentHashList can round-trip through JSON
	type container struct {
		Hash ContentHashList `json:"content_hash,omitempty"`
	}

	// Single hash
	c1 := container{Hash: ContentHashList{ComputeSHA512([]byte("test"))}}
	data, err := json.Marshal(c1)
	require.NoError(t, err)

	var c2 container
	require.NoError(t, json.Unmarshal(data, &c2))
	assert.Equal(t, c1.Hash[0].Algorithm, c2.Hash[0].Algorithm)
	assert.Equal(t, c1.Hash[0].Hash, c2.Hash[0].Hash)
}
