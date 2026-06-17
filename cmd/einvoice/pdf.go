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

	cat, err := r.Catalog()
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF catalog: %w", err)
	}

	// Locate the EmbeddedFiles name tree: /Root/Names/EmbeddedFiles.
	names, ok := cat.Dict("Names")
	if !ok {
		return nil, fmt.Errorf("PDF contains no embedded files (not a ZUGFeRD/Factur-X invoice)")
	}
	efTree, ok := names.Dict("EmbeddedFiles")
	if !ok {
		return nil, fmt.Errorf("PDF contains no embedded files (not a ZUGFeRD/Factur-X invoice)")
	}

	// Collect all (filename -> file specification dict) pairs from the tree.
	attachments := map[string]*pdf.Dict{}
	collectEmbeddedFiles(r, efTree, attachments)
	if len(attachments) == 0 {
		return nil, fmt.Errorf("PDF contains no embedded files (not a ZUGFeRD/Factur-X invoice)")
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

// collectEmbeddedFiles walks an EmbeddedFiles name tree node recursively,
// adding each (name -> file specification dict) pair to dst. Name trees are
// either intermediate nodes with a /Kids array or leaf nodes with a /Names
// array holding alternating name/value entries.
func collectEmbeddedFiles(r *pdf.Reader, node *pdf.Dict, dst map[string]*pdf.Dict) {
	if node == nil {
		return
	}
	if kids, ok := node.Array("Kids"); ok {
		for _, kid := range kids {
			if child, err := r.ResolveDict(kid); err == nil {
				collectEmbeddedFiles(r, child, dst)
			}
		}
	}
	if entries, ok := node.Array("Names"); ok {
		// Alternating [name1, fileSpec1, name2, fileSpec2, ...].
		for i := 0; i+1 < len(entries); i += 2 {
			name, ok := entries[i].(pdf.String)
			if !ok {
				continue
			}
			if fileSpec, err := r.ResolveDict(entries[i+1]); err == nil {
				dst[string(name)] = fileSpec
			}
		}
	}
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
