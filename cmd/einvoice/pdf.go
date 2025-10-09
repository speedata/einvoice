package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// extractXMLFromPDF extracts the embedded invoice XML from a ZUGFeRD/Factur-X PDF.
// It searches for commonly-named embedded XML files and returns the first match.
//
// ZUGFeRD/Factur-X PDFs embed the invoice XML as a PDF attachment (PDF/A-3).
// The XML file is typically named factur-x.xml, ZUGFeRD-invoice.xml, or similar.
func extractXMLFromPDF(filename string) ([]byte, error) {
	// Open PDF file
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	// Extract all attachments in-memory
	attachments, err := api.ExtractAttachmentsRaw(f, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract attachments from PDF: %w", err)
	}

	if len(attachments) == 0 {
		return nil, fmt.Errorf("PDF contains no embedded files (not a ZUGFeRD/Factur-X invoice)")
	}

	// Known invoice XML filenames in order of preference
	knownNames := []string{
		"factur-x.xml",        // Factur-X standard
		"ZUGFeRD-invoice.xml", // ZUGFeRD 2.x
		"zugferd-invoice.xml", // ZUGFeRD 1.x
		"xrechnung.xml",       // XRechnung
	}

	// First pass: search for exact matches with known filenames
	for _, attachment := range attachments {
		for _, knownName := range knownNames {
			if attachment.FileName == knownName {
				return readAttachment(attachment)
			}
		}
	}

	// Second pass: any .xml file as fallback
	for _, attachment := range attachments {
		if strings.HasSuffix(strings.ToLower(attachment.FileName), ".xml") {
			return readAttachment(attachment)
		}
	}

	return nil, fmt.Errorf("PDF contains no invoice XML attachment")
}

// readAttachment reads all data from an attachment's io.Reader.
func readAttachment(attachment model.Attachment) ([]byte, error) {
	data, err := io.ReadAll(attachment)
	if err != nil {
		return nil, fmt.Errorf("failed to read attachment %q: %w", attachment.FileName, err)
	}
	return data, nil
}
