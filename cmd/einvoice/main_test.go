package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateInvoice(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		wantValid     bool
		wantError     bool
		wantViolation bool
	}{
		{
			name:      "valid invoice from testcases",
			filename:  "../../testcases/zugferd_2p0_EN16931_1_Teilrechnung.xml",
			wantValid: false, // We don't know if this is valid without checking
			wantError: false,
		},
		{
			name:      "non-existent file",
			filename:  "non_existent_file.xml",
			wantValid: false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateInvoice(tt.filename)

			if tt.wantError {
				if result.Error == "" {
					t.Errorf("validateInvoice() expected error, got none")
				}
				return
			}

			if result.Error != "" {
				t.Errorf("validateInvoice() unexpected error: %v", result.Error)
				return
			}

			// For valid test case file, just verify we got some result
			if result.Invoice == nil {
				t.Errorf("validateInvoice() invoice metadata is nil")
			}
		})
	}
}

func TestValidateInvoice_MalformedXML(t *testing.T) {
	// Create a temporary malformed XML file
	tmpfile, err := os.CreateTemp("", "malformed-*.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("<invalid>xml</wrong>")); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result := validateInvoice(tmpfile.Name())

	if result.Error == "" {
		t.Error("validateInvoice() expected error for malformed XML, got none")
	}

	if !strings.Contains(result.Error, "parse") {
		t.Errorf("validateInvoice() error should mention parsing, got: %v", result.Error)
	}
}

func TestOutputJSON(t *testing.T) {
	result := Result{
		File:  "test.xml",
		Valid: false,
		Invoice: &InvoiceRef{
			Number: "INV-001",
			Date:   "2024-01-15",
			Total:  "1000.00",
		},
		Violations: []Violation{
			{Rule: "BR-1", Text: "violation 1"},
			{Rule: "BR-2", Text: "violation 2"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputJSON(result)

	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	// Verify it's valid JSON
	var decoded Result
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Errorf("outputJSON() produced invalid JSON: %v", err)
	}

	// Verify content
	if decoded.File != "test.xml" {
		t.Errorf("outputJSON() file = %v, want %v", decoded.File, "test.xml")
	}
	if decoded.Valid != false {
		t.Errorf("outputJSON() valid = %v, want %v", decoded.Valid, false)
	}
	if len(decoded.Violations) != 2 {
		t.Errorf("outputJSON() violations count = %v, want %v", len(decoded.Violations), 2)
	}
}

func TestOutputText(t *testing.T) {
	tests := []struct {
		name       string
		result     Result
		wantStderr bool
		wantOutput string
	}{
		{
			name: "valid invoice",
			result: Result{
				File:  "test.xml",
				Valid: true,
				Invoice: &InvoiceRef{
					Number: "INV-001",
				},
			},
			wantStderr: false,
			wantOutput: "✓ Invoice INV-001 is valid",
		},
		{
			name: "invalid invoice with violations",
			result: Result{
				File:  "test.xml",
				Valid: false,
				Invoice: &InvoiceRef{
					Number: "INV-002",
				},
				Violations: []Violation{
					{Rule: "BR-1", Text: "violation"},
				},
			},
			wantStderr: false,
			wantOutput: "✗ Invoice INV-002 has 1 violation(s):",
		},
		{
			name: "error case",
			result: Result{
				File:  "test.xml",
				Error: "file not found",
			},
			wantStderr: true,
			wantOutput: "Error: file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			rOut, wOut, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr

			outputText(tt.result)

			wOut.Close()
			wErr.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var bufOut, bufErr strings.Builder
			io.Copy(&bufOut, rOut)
			io.Copy(&bufErr, rErr)

			output := bufOut.String()
			errOutput := bufErr.String()

			if tt.wantStderr {
				if !strings.Contains(errOutput, tt.wantOutput) {
					t.Errorf("outputText() stderr = %q, want to contain %q", errOutput, tt.wantOutput)
				}
			} else {
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("outputText() stdout = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestIntegration_ValidFile(t *testing.T) {
	// Test with the actual test case file if it exists
	testFile := filepath.Join("..", "..", "testcases", "zugferd_2p0_EN16931_1_Teilrechnung.xml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found, skipping integration test")
	}

	result := validateInvoice(testFile)

	// Should not have a fatal error
	if result.Error != "" {
		t.Errorf("Integration test failed with error: %v", result.Error)
	}

	// Should have invoice metadata
	if result.Invoice == nil {
		t.Error("Integration test: invoice metadata is nil")
	} else {
		if result.Invoice.Number == "" {
			t.Error("Integration test: invoice number is empty")
		}
		if result.Invoice.Total == "" {
			t.Error("Integration test: invoice total is empty")
		}
	}

	// Log the result for manual inspection
	t.Logf("Invoice %s validation result: valid=%v, violations=%d",
		result.Invoice.Number, result.Valid, len(result.Violations))
	if len(result.Violations) > 0 {
		t.Logf("Violations found:")
		for _, v := range result.Violations {
			t.Logf("  - %s", v)
		}
	}
}
