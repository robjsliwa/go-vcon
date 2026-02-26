package vcon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactedObjectJSON(t *testing.T) {
	ro := RedactedObject{
		UUID: "550e8400-e29b-41d4-a716-446655440000",
		Type: "audio",
		URL:  "https://example.com/original.json",
		ContentHash: ContentHashList{
			{Algorithm: "sha512", Hash: "abc123"},
		},
	}

	data, err := json.Marshal(ro)
	require.NoError(t, err)

	var ro2 RedactedObject
	require.NoError(t, json.Unmarshal(data, &ro2))
	assert.Equal(t, ro.UUID, ro2.UUID)
	assert.Equal(t, ro.Type, ro2.Type)
	assert.Equal(t, ro.URL, ro2.URL)
	assert.Equal(t, 1, len(ro2.ContentHash))
	assert.Equal(t, "sha512", ro2.ContentHash[0].Algorithm)
}

func TestAmendedObjectJSON(t *testing.T) {
	ao := AmendedObject{
		UUID: "550e8400-e29b-41d4-a716-446655440000",
		URL:  "https://example.com/prior.json",
		ContentHash: ContentHashList{
			{Algorithm: "sha512", Hash: "xyz789"},
		},
	}

	data, err := json.Marshal(ao)
	require.NoError(t, err)

	var ao2 AmendedObject
	require.NoError(t, json.Unmarshal(data, &ao2))
	assert.Equal(t, ao.UUID, ao2.UUID)
	assert.Equal(t, ao.URL, ao2.URL)
	assert.Equal(t, 1, len(ao2.ContentHash))
}

func TestAmendedObjectOmitEmpty(t *testing.T) {
	ao := AmendedObject{}
	data, err := json.Marshal(ao)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	_, hasUUID := m["uuid"]
	_, hasURL := m["url"]
	_, hasHash := m["content_hash"]
	assert.False(t, hasUUID)
	assert.False(t, hasURL)
	assert.False(t, hasHash)
}

func TestSessionIdJSON(t *testing.T) {
	sid := SessionId{Local: "abc123", Remote: "xyz789"}

	data, err := json.Marshal(sid)
	require.NoError(t, err)

	var sid2 SessionId
	require.NoError(t, json.Unmarshal(data, &sid2))
	assert.Equal(t, "abc123", sid2.Local)
	assert.Equal(t, "xyz789", sid2.Remote)
}

func TestIntOrSliceSingleValue(t *testing.T) {
	v := NewIntValue(42)

	i, ok := v.AsInt()
	assert.True(t, ok)
	assert.Equal(t, 42, i)

	slice := v.AsSlice()
	assert.Equal(t, []int{42}, slice)

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, "42", string(data))

	var v2 IntOrSlice
	require.NoError(t, json.Unmarshal(data, &v2))
	i2, ok := v2.AsInt()
	assert.True(t, ok)
	assert.Equal(t, 42, i2)
}

func TestIntOrSliceSliceValue(t *testing.T) {
	v := NewIntSliceValue([]int{1, 2, 3})

	_, ok := v.AsInt()
	assert.False(t, ok)

	slice := v.AsSlice()
	assert.Equal(t, []int{1, 2, 3}, slice)

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, "[1,2,3]", string(data))

	var v2 IntOrSlice
	require.NoError(t, json.Unmarshal(data, &v2))
	slice2 := v2.AsSlice()
	assert.Equal(t, []int{1, 2, 3}, slice2)
}

func TestIntOrSliceIsZero(t *testing.T) {
	assert.True(t, IntOrSlice{}.IsZero())
	assert.False(t, NewIntValue(0).IsZero())
	assert.False(t, NewIntSliceValue([]int{}).IsZero())
}

func TestIntOrSliceNil(t *testing.T) {
	v := IntOrSlice{}
	assert.Nil(t, v.AsSlice())
	_, ok := v.AsInt()
	assert.False(t, ok)

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

func TestIntOrSliceUnmarshalInvalid(t *testing.T) {
	var v IntOrSlice
	err := json.Unmarshal([]byte(`"hello"`), &v)
	assert.Error(t, err)
}
