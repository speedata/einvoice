package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInvoiceFile(t *testing.T) {
	// Use the same test file pattern as validate_test.go
	testFile := filepath.Join("..", "..", "testcases", "zugferd_2p0_EN16931_1_Teilrechnung.xml")

	tests := []struct {
		name    string
		file    string
		wantErr bool
		errMsg  string
		skipIf  bool
	}{
		{
			name:    "Valid XML file from testcases",
			file:    testFile,
			wantErr: false,
			skipIf:  !fileExists(testFile),
		},
		{
			name:    "Non-existent file",
			file:    "nonexistent.xml",
			wantErr: true,
			errMsg:  "no such file",
		},
		{
			name:    "Empty filename",
			file:    "",
			wantErr: true,
			errMsg:  "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIf {
				t.Skip("Test file not found, skipping test")
			}

			invoice, err := parseInvoiceFile(tt.file)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseInvoiceFile() expected error containing %q, got nil", tt.errMsg)
					return
				}
				// Just check if error message contains expected substring
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("parseInvoiceFile() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseInvoiceFile() unexpected error: %v", err)
				return
			}

			if invoice == nil {
				t.Error("parseInvoiceFile() returned nil invoice without error")
			}
		})
	}
}

func TestParseInvoiceFile_UnsupportedFormat(t *testing.T) {
	// Create a temporary file with unsupported content
	tmpfile, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write JSON content (unsupported format)
	content := []byte(`{"invoice": "data"}`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = parseInvoiceFile(tmpfile.Name())
	if err == nil {
		t.Error("parseInvoiceFile() expected error for unsupported format, got nil")
		return
	}

	expectedMsg := "unsupported file format"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("parseInvoiceFile() error = %q, want error containing %q", err.Error(), expectedMsg)
	}
}

func TestParseInvoiceFile_EmptyFile(t *testing.T) {
	// Create an empty temporary file
	tmpfile, err := os.CreateTemp("", "empty*.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	_, err = parseInvoiceFile(tmpfile.Name())
	if err == nil {
		t.Error("parseInvoiceFile() expected error for empty file, got nil")
		return
	}

	// Empty file should be rejected as unsupported format
	expectedMsg := "unsupported file format"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("parseInvoiceFile() error = %q, want error containing %q", err.Error(), expectedMsg)
	}
}

// TestParseInvoiceFile_PDF is a placeholder for PDF parsing tests.
// Real ZUGFeRD/Factur-X PDF files are needed for complete testing.
// See cmd/einvoice/testdata/README.md for instructions on adding test PDFs.
func TestParseInvoiceFile_PDF(t *testing.T) {
	testdataDir := "testdata"

	// Check if testdata directory exists
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skip("testdata directory not found - PDF tests skipped (see testdata/README.md)")
		return
	}

	// Look for PDF test files
	pdfFiles, err := filepath.Glob(filepath.Join(testdataDir, "*.pdf"))
	if err != nil {
		t.Fatal(err)
	}

	if len(pdfFiles) == 0 {
		t.Skip("no PDF test files found in testdata/ - PDF tests skipped (see testdata/README.md)")
		return
	}

	// Test each PDF file found
	for _, pdfFile := range pdfFiles {
		t.Run(filepath.Base(pdfFile), func(t *testing.T) {
			invoice, err := parseInvoiceFile(pdfFile)
			if err != nil {
				t.Errorf("parseInvoiceFile() error = %v, want success for valid ZUGFeRD PDF", err)
				return
			}

			if invoice == nil {
				t.Error("parseInvoiceFile() returned nil invoice without error")
			}

			// Basic sanity checks
			if invoice.InvoiceNumber == "" {
				t.Error("parseInvoiceFile() invoice has empty InvoiceNumber")
			}
		})
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// contains checks if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
