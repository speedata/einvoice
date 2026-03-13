package main

import (
	"fmt"
	"io"
	"os"
	"slices"
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
	defer func() { _ = f.Close() }()

	// Read PDF structure first, then validate with version adjustment.
	// Some older ZUGFeRD PDFs (e.g. v2.01) declare PDF 1.3 but use
	// PDF/A-3 features like AFRelationship (which requires at least 1.4
	// in pdfcpu's relaxed validation mode). We bump the header version
	// to 1.4 before validation to handle these files.
	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.EXTRACTATTACHMENTS
	ctx, err := api.ReadContext(f, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}
	if ctx.HeaderVersion != nil && *ctx.HeaderVersion < model.V14 {
		v := model.V14
		ctx.HeaderVersion = &v
	}
	if err := api.ValidateContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate PDF: %w", err)
	}

	attachments, err := ctx.ExtractAttachments(nil)
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
		if slices.Contains(knownNames, attachment.FileName) {
			return readAttachment(attachment)
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
