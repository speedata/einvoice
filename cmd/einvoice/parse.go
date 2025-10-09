package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/speedata/einvoice"
)

// parseInvoiceFile parses an invoice from an XML file or ZUGFeRD/Factur-X PDF.
// It automatically detects the file type based on content (magic bytes) and
// routes to the appropriate parser.
//
// Supported formats:
//   - Plain XML invoice files
//   - ZUGFeRD/Factur-X PDF files with embedded XML
func parseInvoiceFile(filename string) (*einvoice.Invoice, error) {
	// Open file for type detection
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// Read magic bytes to detect file type
	header := make([]byte, 512)
	_, err = f.Read(header)
	_ = f.Close() // Ignore close error on read-only file
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect file type based on content
	isPDF := bytes.HasPrefix(header, []byte("%PDF"))
	isXML := bytes.HasPrefix(header, []byte("<?xml")) ||
		bytes.HasPrefix(header, []byte("<"))

	// Route to appropriate parser
	if isXML {
		return einvoice.ParseXMLFile(filename)
	}

	if isPDF {
		// Extract embedded XML from PDF
		xmlBytes, err := extractXMLFromPDF(filename)
		if err != nil {
			return nil, err
		}
		// Parse the extracted XML
		return einvoice.ParseReader(bytes.NewReader(xmlBytes))
	}

	return nil, fmt.Errorf("unsupported file format (expected XML or PDF)")
}
