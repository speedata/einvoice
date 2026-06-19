package main

import (
	"fmt"
	"strings"

	pdf "github.com/speedata/pdfdisassembler"
)

// knownInvoiceXMLNames lists the embedded XML filenames used by the common
// hybrid invoice standards, in order of preference.
var knownInvoiceXMLNames = []string{
	"factur-x.xml",        // Factur-X / ZUGFeRD 2.x standard
	"ZUGFeRD-invoice.xml", // ZUGFeRD 2.x (legacy name)
	"zugferd-invoice.xml", // ZUGFeRD 1.x
	"xrechnung.xml",       // XRechnung
}

// extractXMLFromPDF extracts the embedded invoice XML from a ZUGFeRD/Factur-X PDF.
// It searches for commonly-named embedded XML files and returns the first match.
//
// ZUGFeRD/Factur-X PDFs embed the invoice XML as a PDF attachment (PDF/A-3)
// stored in the catalog's EmbeddedFiles name tree. We use the read-only
// pdfdisassembler parser, which walks that name tree directly without the
// strict PDF-version validation that rejects PDF/A-3 features (e.g.
// AFRelationship) on files declaring an older header version.
func extractXMLFromPDF(filename string) ([]byte, error) {
	r, err := pdf.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Collect all embedded files from the catalog's EmbeddedFiles name tree.
	files := r.EmbeddedFiles()
	if len(files) == 0 {
		return nil, fmt.Errorf("PDF contains no embedded files (not a ZUGFeRD/Factur-X invoice)")
	}

	attachments := make(map[string]*pdf.Dict, len(files))
	for _, f := range files {
		attachments[f.Name] = f.Spec
	}

	// First pass: exact matches with known invoice filenames.
	for _, name := range knownInvoiceXMLNames {
		if fileSpec, ok := attachments[name]; ok {
			if data, err := readEmbeddedFile(fileSpec); err == nil {
				return data, nil
			}
		}
	}

	// Second pass: any .xml file as fallback.
	for name, fileSpec := range attachments {
		if strings.HasSuffix(strings.ToLower(name), ".xml") {
			if data, err := readEmbeddedFile(fileSpec); err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("PDF contains no invoice XML attachment")
}

// readEmbeddedFile returns the decoded content of an embedded file from its
// file specification dictionary. The embedded stream lives in /EF, preferring
// /F (the standard file) over /UF (the Unicode file name variant).
func readEmbeddedFile(fileSpec *pdf.Dict) ([]byte, error) {
	ef, ok := fileSpec.Dict("EF")
	if !ok {
		return nil, fmt.Errorf("file specification has no embedded file stream")
	}
	for _, key := range []string{"F", "UF"} {
		if stream, ok := ef.Stream(key); ok {
			data, err := stream.Content()
			if err != nil {
				return nil, fmt.Errorf("failed to decode embedded file: %w", err)
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("file specification has no embedded file stream")
}
