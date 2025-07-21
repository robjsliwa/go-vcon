package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunEmail(t *testing.T) {
	// Reset global variables for testing
	originalGlobalDomain := globalDomain
	originalVConOut := vConOut
	
	defer func() {
		globalDomain = originalGlobalDomain
		vConOut = originalVConOut
	}()

	// Create a temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "email_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Path to the test email file
	testEmailPath := "../../testdata/sample_vcons/test_email.eml"
	absTestEmailPath, err := filepath.Abs(testEmailPath)
	if err != nil {
		t.Fatal(err)
	}

	// Check if test file exists
	if _, err := os.Stat(absTestEmailPath); os.IsNotExist(err) {
		t.Skipf("Test email file not found: %s", absTestEmailPath)
	}

	tests := []struct {
		name        string
		setupFunc   func()
		args        []string
		expectError bool
	}{
		{
			name: "valid email conversion",
			setupFunc: func() {
				globalDomain = "test.example.com"
				vConOut = filepath.Join(tmpDir, "test_email_output.vcon.json")
			},
			args:        []string{absTestEmailPath},
			expectError: false,
		},
		{
			name: "valid email conversion with default output",
			setupFunc: func() {
				globalDomain = "test.example.com"
				vConOut = ""
			},
			args:        []string{absTestEmailPath},
			expectError: false,
		},
		{
			name: "invalid email file",
			setupFunc: func() {
				globalDomain = "test.example.com"
				vConOut = filepath.Join(tmpDir, "test_invalid_output.vcon.json")
			},
			args:        []string{"/nonexistent/file.eml"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			cmd := &cobra.Command{}
			err := runEmail(cmd, tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Determine expected output file
				expectedOutput := vConOut
				if expectedOutput == "" {
					expectedOutput = strings.TrimSuffix(tt.args[0], filepath.Ext(tt.args[0])) + ".vcon.json"
				}

				// Check that the output file was created
				if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
					t.Errorf("expected output file %s was not created", expectedOutput)
				} else {
					// Clean up the default output file if it was created
					if vConOut == "" {
						defer os.Remove(expectedOutput)
					}
				}
			}
		})
	}
}

func TestRunEmailIntegration(t *testing.T) {
	// Path to the test email file
	testEmailPath := "../../testdata/sample_vcons/test_email.eml"
	absTestEmailPath, err := filepath.Abs(testEmailPath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(absTestEmailPath); os.IsNotExist(err) {
		t.Skipf("Test email file not found: %s", absTestEmailPath)
	}

	// Create a temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "email_integration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original values
	originalGlobalDomain := globalDomain
	originalVConOut := vConOut
	
	defer func() {
		globalDomain = originalGlobalDomain
		vConOut = originalVConOut
	}()

	// Set up test values
	globalDomain = "test.example.com"
	vConOut = filepath.Join(tmpDir, "integration_test.vcon.json")

	// Run the email conversion
	cmd := &cobra.Command{}
	err = runEmail(cmd, []string{absTestEmailPath})
	if err != nil {
		t.Fatalf("email conversion failed: %v", err)
	}

	// Verify the output file exists and contains expected content
	if _, err := os.Stat(vConOut); os.IsNotExist(err) {
		t.Fatalf("expected output file %s was not created", vConOut)
	}

	// Read and verify the content contains expected data
	content, err := os.ReadFile(vConOut)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)
	
	// Check for basic vCon structure
	expectedStrings := []string{
		"\"vcon\":",
		"\"uuid\":",
		"\"created_at\":",
		"\"parties\":",
		"\"dialog\":",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("output file does not contain expected JSON structure: %s", expected)
		}
	}

	// Check for email-specific content
	emailSpecificStrings := []string{
		"\"type\": \"email\"",
		"\"mediatype\": \"text/plain\"",
		"mailto:",
	}

	for _, expected := range emailSpecificStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("output file does not contain expected email content: %s", expected)
		}
	}
}

// Test the email parsing logic more specifically
func TestEmailParsingLogic(t *testing.T) {
	// Create a simple test email file
	tmpDir, err := os.MkdirTemp("", "email_parsing_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testEmailContent := `From: Alice <alice@example.com>
To: Bob <bob@example.com>
Cc: Charlie <charlie@example.com>
Subject: Test Email Subject
Date: Mon, 15 Jan 2023 10:30:00 +0000
Message-ID: <test-message-id@example.com>

This is a test email body.
It contains multiple lines
to test the email parsing functionality.
`

	testEmailFile := filepath.Join(tmpDir, "test_simple.eml")
	err = os.WriteFile(testEmailFile, []byte(testEmailContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Save original values
	originalGlobalDomain := globalDomain
	originalVConOut := vConOut
	
	defer func() {
		globalDomain = originalGlobalDomain
		vConOut = originalVConOut
	}()

	// Set up test values
	globalDomain = "test.example.com"
	vConOut = filepath.Join(tmpDir, "parsed_email.vcon.json")

	// Run the email conversion
	cmd := &cobra.Command{}
	err = runEmail(cmd, []string{testEmailFile})
	if err != nil {
		t.Fatalf("email conversion failed: %v", err)
	}

	// Read and verify the content
	content, err := os.ReadFile(vConOut)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify specific parsed content
	expectedContent := []string{
		"Test Email Subject",
		"alice@example.com",
		"bob@example.com",
		"charlie@example.com",
		"test-message-id@example.com",
		"This is a test email body",
		"originator",
		"recipient",
		"cc",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("parsed email does not contain expected content: %s", expected)
		}
	}
}
