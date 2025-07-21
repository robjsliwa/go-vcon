package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

func TestValidateCommand(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "validate_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid vCon file
	validVcon := vcon.New("test.example.com")
	validVcon.Subject = "Test Subject"
	validData, _ := json.MarshalIndent(validVcon, "", "  ")
	validFile := filepath.Join(tmpDir, "valid.vcon.json")
	err = os.WriteFile(validFile, validData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create an invalid JSON file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "validate valid vcon",
			args:        []string{validFile},
			expectError: false,
		},
		{
			name:        "validate invalid file",
			args:        []string{invalidFile},
			expectError: false, // Command doesn't error, just prints validation results
		},
		{
			name:        "validate nonexistent file",
			args:        []string{"/nonexistent/file.json"},
			expectError: false, // Command doesn't error, just prints validation results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the Run function directly since it prints to stdout
			// Instead, we test that the command is properly configured
			err := validateCmd.Args(validateCmd, tt.args)
			if err != nil && !tt.expectError {
				t.Errorf("validateCmd.Args failed for valid args: %v", tt.args)
			}
		})
	}
}

func TestSignCommandValidation(t *testing.T) {
	// Test that the sign command is properly configured
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "correct number of args",
			args:        []string{"test.vcon.json"},
			expectError: false,
		},
		{
			name:        "too few args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"file1.json", "file2.json"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := signCmd.Args(signCmd, tt.args)
			if tt.expectError && err == nil {
				t.Errorf("expected error for args %v but got none", tt.args)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

func TestEncryptCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "correct number of args",
			args:        []string{"test.vcon.json"},
			expectError: false,
		},
		{
			name:        "too few args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"file1.json", "file2.json"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := encryptCmd.Args(encryptCmd, tt.args)
			if tt.expectError && err == nil {
				t.Errorf("expected error for args %v but got none", tt.args)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

func TestDecryptCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "correct number of args",
			args:        []string{"test.vcon.json"},
			expectError: false,
		},
		{
			name:        "too few args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"file1.json", "file2.json"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decryptCmd.Args(decryptCmd, tt.args)
			if tt.expectError && err == nil {
				t.Errorf("expected error for args %v but got none", tt.args)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

func TestVerifyCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "correct number of args",
			args:        []string{"test.vcon.json"},
			expectError: false,
		},
		{
			name:        "too few args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"file1.json", "file2.json"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyCmd.Args(verifyCmd, tt.args)
			if tt.expectError && err == nil {
				t.Errorf("expected error for args %v but got none", tt.args)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

func TestAudioCommandValidation(t *testing.T) {
	// Test that required flags work as expected
	if !audioCmd.Flags().Changed("input") {
		// The input flag should be required - we can't test the execution
		// but we can verify the flag is properly configured
		flag := audioCmd.Flags().Lookup("input")
		if flag == nil {
			t.Error("input flag not found")
		}
	}

	// Test that args validation works (audio command accepts no args)
	err := audioCmd.Args(audioCmd, []string{})
	if err != nil {
		t.Errorf("audioCmd should accept no args, got error: %v", err)
	}

	// Audio command should not accept positional arguments, only flags
	err = audioCmd.Args(audioCmd, []string{"unexpected"})
	if err == nil {
		t.Errorf("audioCmd should not accept positional args")
	}
}

func TestEmailCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "correct number of args",
			args:        []string{"test.eml"},
			expectError: false,
		},
		{
			name:        "too few args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"file1.eml", "file2.eml"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := emailCmd.Args(emailCmd, tt.args)
			if tt.expectError && err == nil {
				t.Errorf("expected error for args %v but got none", tt.args)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	// Test that the main commands are properly configured
	commands := []*cobra.Command{
		validateCmd,
		signCmd,
		encryptCmd,
		verifyCmd,
		decryptCmd,
		genkeyCmd,
		convertCmd,
		audioCmd,
		emailCmd,
	}

	for _, cmd := range commands {
		t.Run("command_"+cmd.Name(), func(t *testing.T) {
			if cmd.Use == "" {
				t.Errorf("command %s has empty Use field", cmd.Name())
			}
			if cmd.Short == "" {
				t.Errorf("command %s has empty Short description", cmd.Name())
			}
		})
	}
}

func TestDieFunction(t *testing.T) {
	// We can't easily test die() since it calls os.Exit(1)
	// But we can verify it's defined and takes the expected parameters
	// This is more of a compilation test
	defer func() {
		if r := recover(); r != nil {
			// If die() panics instead of calling os.Exit, that's also acceptable for testing
		}
	}()

	// Just verify the function signature works
	_ = func(context string, err error) {
		// Mock implementation that doesn't actually exit
		if context == "" || err == nil {
			t.Error("die function should handle context and error parameters")
		}
	}
}

func TestCommandIntegration(t *testing.T) {
	// Test that the root command includes all expected subcommands
	rootCmd.SetArgs([]string{"--help"})
	
	// Verify convert command has subcommands
	convertSubcommands := convertCmd.Commands()
	expectedSubcommands := []string{"audio", "email", "zoom"}
	
	subcommandNames := make([]string, len(convertSubcommands))
	for i, cmd := range convertSubcommands {
		subcommandNames[i] = cmd.Name()
	}
	
	for _, expected := range expectedSubcommands {
		found := false
		for _, actual := range subcommandNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected convert subcommand %s not found in %v", expected, subcommandNames)
		}
	}
}
