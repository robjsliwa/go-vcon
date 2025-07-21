package main

import (
	"os/exec"
	"testing"
)

func TestFFProbeDetection(t *testing.T) {
	// Test our ffprobe detection logic
	available := checkFFProbeAvailable()
	t.Logf("ffprobe available: %v", available)
	
	// Also test with exec.LookPath directly
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Logf("ffprobe not found in PATH: %v", err)
	} else {
		t.Log("ffprobe found in PATH")
	}
	
	// Test with a non-existent command
	_, err = exec.LookPath("nonexistent-command-12345")
	if err != nil {
		t.Logf("nonexistent command correctly not found: %v", err)
	}
}
