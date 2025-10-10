package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/speedata/einvoice"
)

// parseInvoiceFile parses an invoice from an XML file or ZUGFeRD/Factur-X PDF.
// It detects the file type based on the file extension and routes to the
// appropriate parser.
//
// Supported formats:
//   - Plain XML invoice files
//   - ZUGFeRD/Factur-X PDF files with embedded XML
func parseInvoiceFile(filename string) (*einvoice.Invoice, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf":
		// Extract embedded XML from PDF
		xmlBytes, err := extractXMLFromPDF(filename)
		if err != nil {
			return nil, err
		}
		// Parse the extracted XML
		return einvoice.ParseReader(bytes.NewReader(xmlBytes))
	case ".xml":
		return einvoice.ParseXMLFile(filename)
	default:
		return nil, fmt.Errorf("unsupported file format (expected XML or PDF)")
	}
}
