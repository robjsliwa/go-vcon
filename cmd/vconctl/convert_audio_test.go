package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// checkFFProbeAvailable checks if ffprobe is available in the system
func checkFFProbeAvailable() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}

func TestRunAudio(t *testing.T) {
	// Skip test if ffprobe is not available
	if !checkFFProbeAvailable() {
		t.Skip("ffprobe not available in PATH - skipping audio conversion tests")
	}
	
	// Reset global variables for testing
	originalGlobalDomain := globalDomain
	originalAudioInput := audioInput
	originalAudioParties := audioParties
	originalAudioDate := audioDate
	originalVConOut := vConOut
	
	defer func() {
		globalDomain = originalGlobalDomain
		audioInput = originalAudioInput
		audioParties = originalAudioParties
		audioDate = originalAudioDate
		vConOut = originalVConOut
	}()

	// Create a temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "audio_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Path to the test audio file
	testAudioPath := "../../testdata/sample_vcons/1745501752.21.wav"
	absTestAudioPath, err := filepath.Abs(testAudioPath)
	if err != nil {
		t.Fatal(err)
	}

	// Check if test file exists
	if _, err := os.Stat(absTestAudioPath); os.IsNotExist(err) {
		t.Skipf("Test audio file not found: %s", absTestAudioPath)
	}

	tests := []struct {
		name        string
		setupFunc   func()
		expectError bool
	}{
		{
			name: "valid audio conversion with parties",
			setupFunc: func() {
				globalDomain = "test.example.com"
				audioInput = absTestAudioPath
				audioParties = []string{"Alice,tel:+15551234567", "Bob,mailto:bob@example.com"}
				audioDate = "2023-01-15T10:30:00Z"
				vConOut = filepath.Join(tmpDir, "test_output.vcon.json")
			},
			expectError: false,
		},
		{
			name: "valid audio conversion without explicit date",
			setupFunc: func() {
				globalDomain = "test.example.com"
				audioInput = absTestAudioPath
				audioParties = []string{"Alice"}
				audioDate = ""
				vConOut = filepath.Join(tmpDir, "test_output2.vcon.json")
			},
			expectError: false,
		},
		{
			name: "invalid audio file",
			setupFunc: func() {
				globalDomain = "test.example.com"
				audioInput = "/nonexistent/file.wav"
				audioParties = []string{"Alice"}
				audioDate = ""
				vConOut = filepath.Join(tmpDir, "test_output3.vcon.json")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			cmd := &cobra.Command{}
			err := runAudio(cmd, []string{})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check that the output file was created
				if _, err := os.Stat(vConOut); os.IsNotExist(err) {
					t.Errorf("expected output file %s was not created", vConOut)
				}
			}
		})
	}
}

func TestRunAudioIntegration(t *testing.T) {
	// This test requires ffprobe to be available
	// Skip if ffprobe is not available in the system
	if !checkFFProbeAvailable() {
		t.Skip("ffprobe not available, skipping integration test")
	}

	testAudioPath := "../../testdata/sample_vcons/1745501752.21.wav"
	absTestAudioPath, err := filepath.Abs(testAudioPath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(absTestAudioPath); os.IsNotExist(err) {
		t.Skipf("Test audio file not found: %s", absTestAudioPath)
	}

	// Create a temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "audio_integration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original values
	originalGlobalDomain := globalDomain
	originalAudioInput := audioInput
	originalAudioParties := audioParties
	originalAudioDate := audioDate
	originalVConOut := vConOut
	
	defer func() {
		globalDomain = originalGlobalDomain
		audioInput = originalAudioInput
		audioParties = originalAudioParties
		audioDate = originalAudioDate
		vConOut = originalVConOut
	}()

	// Set up test values
	globalDomain = "test.example.com"
	audioInput = absTestAudioPath
	audioParties = []string{"Test Speaker,tel:+15551234567"}
	audioDate = "2023-01-15T10:30:00Z"
	vConOut = filepath.Join(tmpDir, "integration_test.vcon.json")

	// Run the audio conversion
	cmd := &cobra.Command{}
	err = runAudio(cmd, []string{})
	if err != nil {
		t.Fatalf("audio conversion failed: %v", err)
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
	expectedStrings := []string{
		"1745501752.21.wav",
		"Test Speaker",
		"tel:+15551234567",
		"2023-01-15T10:30:00Z",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("output file does not contain expected string: %s", expected)
		}
	}
}

func TestRunAudioWithoutFFProbe(t *testing.T) {
	// Create a temporary test that simulates missing ffprobe
	// by overriding the checkFFProbeAvailable function behavior
	
	// Save original function (we can't actually override it easily in Go, 
	// so we'll just test that the logic would work)
	
	// Test what would happen if ffprobe was not available
	// This simulates the GitHub Actions environment
	
	// Create a function that simulates ffprobe not being available
	oldPath := os.Getenv("PATH")
	defer func() {
		os.Setenv("PATH", oldPath)
	}()
	
	// Set PATH to a directory that doesn't contain ffprobe
	tempDir := t.TempDir()
	os.Setenv("PATH", tempDir)
	
	// Now test that checkFFProbeAvailable returns false
	if checkFFProbeAvailable() {
		t.Error("Expected checkFFProbeAvailable to return false with empty PATH")
	}
	
	t.Log("Successfully verified that checkFFProbeAvailable returns false when ffprobe is not in PATH")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
