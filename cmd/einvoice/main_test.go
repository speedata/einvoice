package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "no arguments",
			args:     []string{"einvoice"},
			wantExit: exitError,
		},
		{
			name:     "unknown command",
			args:     []string{"einvoice", "unknown"},
			wantExit: exitError,
		},
		{
			name:     "validate command exists",
			args:     []string{"einvoice", "validate"},
			wantExit: exitError, // Will fail because no file argument, but tests command exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args and restore after test
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Capture stderr to suppress output during tests
			oldStderr := os.Stderr
			_, w, _ := os.Pipe()
			os.Stderr = w

			// Set os.Args for the test
			os.Args = tt.args

			exitCode := run()

			w.Close()
			os.Stderr = oldStderr

			if exitCode != tt.wantExit {
				t.Errorf("run() exit code = %v, want %v", exitCode, tt.wantExit)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	usage()

	w.Close()
	os.Stderr = oldStderr

	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	// Verify usage output contains expected elements
	expectedStrings := []string{
		"Usage:",
		"einvoice",
		"command",
		"validate",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("usage() output missing %q, got: %v", expected, output)
		}
	}
}

func TestSubcommandDispatch(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantExit    int
		wantStderr  string
	}{
		{
			name:       "dispatch to validate",
			args:       []string{"einvoice", "validate"},
			wantExit:   exitError, // No file provided
			wantStderr: "Usage: einvoice validate",
		},
		{
			name:       "unknown command shows error",
			args:       []string{"einvoice", "invalid"},
			wantExit:   exitError,
			wantStderr: `unknown command "invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args and restore after test
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Set os.Args for the test
			os.Args = tt.args

			exitCode := run()

			w.Close()
			os.Stderr = oldStderr

			var buf strings.Builder
			io.Copy(&buf, r)
			stderrOutput := buf.String()

			if exitCode != tt.wantExit {
				t.Errorf("run() exit code = %v, want %v", exitCode, tt.wantExit)
			}

			if tt.wantStderr != "" && !strings.Contains(stderrOutput, tt.wantStderr) {
				t.Errorf("run() stderr should contain %q, got: %v", tt.wantStderr, stderrOutput)
			}
		})
	}
}
