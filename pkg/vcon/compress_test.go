package vcon

import (
	"bytes"
	"testing"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	original := []byte(`{"uuid":"test","parties":[],"dialog":[],"analysis":[],"attachments":[]}`)

	compressed, err := CompressPayload(original)
	if err != nil {
		t.Fatalf("compress error: %v", err)
	}

	if len(compressed) == 0 {
		t.Fatal("compressed data should not be empty")
	}

	decompressed, err := DecompressPayload(compressed)
	if err != nil {
		t.Fatalf("decompress error: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("round trip failed: got %s", string(decompressed))
	}
}

func TestCompressPayloadProducesSmaller(t *testing.T) {
	// Large repetitive data should compress well
	original := bytes.Repeat([]byte("hello world vcon test data "), 100)

	compressed, err := CompressPayload(original)
	if err != nil {
		t.Fatalf("compress error: %v", err)
	}

	if len(compressed) >= len(original) {
		t.Errorf("expected compressed size (%d) < original size (%d)", len(compressed), len(original))
	}
}

func TestDecompressInvalidData(t *testing.T) {
	_, err := DecompressPayload([]byte("not gzip data"))
	if err == nil {
		t.Error("expected error decompressing invalid data")
	}
}
