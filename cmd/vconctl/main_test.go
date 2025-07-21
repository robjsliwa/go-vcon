package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
)

func TestParseParty(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		expected *vcon.Party
	}{
		{
			name: "name only",
			spec: "John Doe",
			expected: &vcon.Party{
				Name: "John Doe",
			},
		},
		{
			name: "name with tel",
			spec: "John Doe,tel:+15551234567",
			expected: &vcon.Party{
				Name: "John Doe",
				Tel:  "tel:+15551234567",
			},
		},
		{
			name: "name with mailto",
			spec: "John Doe,mailto:john@example.com",
			expected: &vcon.Party{
				Name:   "John Doe",
				Mailto: "mailto:john@example.com",
			},
		},
		{
			name: "name with non-standard contact",
			spec: "John Doe,skype:johndoe",
			expected: &vcon.Party{
				Name: "John Doe",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseParty(tt.spec)
			if result.Name != tt.expected.Name {
				t.Errorf("expected Name %s, got %s", tt.expected.Name, result.Name)
			}
			if result.Tel != tt.expected.Tel {
				t.Errorf("expected Tel %s, got %s", tt.expected.Tel, result.Tel)
			}
			if result.Mailto != tt.expected.Mailto {
				t.Errorf("expected Mailto %s, got %s", tt.expected.Mailto, result.Mailto)
			}
		})
	}
}

func TestGetDate(t *testing.T) {
	// Create a temporary file for testing file modification time
	tmpFile, err := os.CreateTemp("", "test_date")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Set a specific modification time
	testTime := time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)
	err = os.Chtimes(tmpFile.Name(), testTime, testTime)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		flag         string
		path         string
		expectedTime time.Time
	}{
		{
			name:         "valid RFC3339 flag",
			flag:         "2023-12-25T10:30:00Z",
			path:         tmpFile.Name(),
			expectedTime: time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
		},
		{
			name:         "invalid flag uses file mtime",
			flag:         "invalid-date",
			path:         tmpFile.Name(),
			expectedTime: testTime,
		},
		{
			name: "no flag no file uses now",
			flag: "",
			path: "/nonexistent/file",
			// We'll check this is close to now
		},
		{
			name:         "empty flag uses file mtime",
			flag:         "",
			path:         tmpFile.Name(),
			expectedTime: testTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDate(tt.flag, tt.path)
			
			if tt.name == "no flag no file uses now" {
				// Check that the result is within the last few seconds
				now := time.Now()
				if result.Before(now.Add(-5*time.Second)) || result.After(now.Add(5*time.Second)) {
					t.Errorf("expected time close to now, got %v", result)
				}
			} else {
				if !result.Equal(tt.expectedTime) {
					t.Errorf("expected %v, got %v", tt.expectedTime, result)
				}
			}
		})
	}
}

func TestWriteVconFile(t *testing.T) {
	// Create a test vCon
	v := vcon.New("test.example.com")
	v.Subject = "Test Subject"

	// Create a temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "vcon_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		out         string
		src         string
		expectedOut string
	}{
		{
			name:        "explicit output path",
			out:         tmpDir + "/explicit.json",
			src:         "input.wav",
			expectedOut: tmpDir + "/explicit.json",
		},
		{
			name:        "default output path from wav",
			out:         "",
			src:         tmpDir + "/test.wav",
			expectedOut: tmpDir + "/test.vcon.json",
		},
		{
			name:        "default output path from eml",
			out:         "",
			src:         tmpDir + "/test.eml",
			expectedOut: tmpDir + "/test.vcon.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeVconFile(v, tt.out, tt.src)
			if err != nil {
				t.Errorf("writeVconFile failed: %v", err)
			}

			// Check that the file was created
			if _, err := os.Stat(tt.expectedOut); os.IsNotExist(err) {
				t.Errorf("expected output file %s was not created", tt.expectedOut)
			}

			// Check that the file contains valid JSON
			content, err := os.ReadFile(tt.expectedOut)
			if err != nil {
				t.Errorf("failed to read output file: %v", err)
			}

			if !strings.Contains(string(content), "Test Subject") {
				t.Errorf("output file does not contain expected content")
			}
		})
	}
}

func TestFetchIfRemote(t *testing.T) {
	// Create a temporary file for local test
	tmpFile, err := os.CreateTemp("", "test_local")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name        string
		src         string
		expectError bool
	}{
		{
			name:        "local file",
			src:         tmpFile.Name(),
			expectError: false,
		},
		{
			name:        "http url (will fail but shouldn't crash)",
			src:         "http://invalid-url-that-should-not-exist.com/file.wav",
			expectError: true,
		},
		{
			name:        "https url (will fail but shouldn't crash)",
			src:         "https://invalid-url-that-should-not-exist.com/file.wav",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup, err := fetchIfRemote(tt.src)
			if cleanup != nil {
				defer cleanup()
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for %s but got none", tt.src)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.src, err)
				}
				if path == "" {
					t.Errorf("expected non-empty path for %s", tt.src)
				}
			}
		})
	}
}
