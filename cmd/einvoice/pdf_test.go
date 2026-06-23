package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// embeddedFile describes one attachment to embed in a synthetic test PDF.
type embeddedFile struct {
	name    string
	content string
}

// buildTestPDF assembles a minimal, valid PDF/A-3-style document with the
// given files embedded in the catalog's EmbeddedFiles name tree. When nested
// is true the name tree uses an intermediate /Kids node to exercise the
// recursive walk in collectEmbeddedFiles. When files is empty the catalog has
// no /Names entry at all (a plain PDF without attachments).
func buildTestPDF(files []embeddedFile, nested bool) []byte {
	var buf bytes.Buffer

	// Object numbering:
	//   1 Catalog, 2 Pages, then for each file a Filespec and an EmbeddedFile
	//   stream, then optionally name-tree nodes.
	numObjs := 2 + len(files)*2
	if len(files) > 0 {
		numObjs++ // leaf node holding the /Names array
		if nested {
			numObjs++ // intermediate node holding /Kids
		}
	}
	offsets := make([]int, numObjs+1)

	writeObj := func(n int, body string) {
		offsets[n] = buf.Len()
		buf.WriteString(strconv.Itoa(n) + " 0 obj\n" + body + "\nendobj\n")
	}
	writeStream := func(n int, dict string, content string) {
		offsets[n] = buf.Len()
		buf.WriteString(strconv.Itoa(n) + " 0 obj\n<< " + dict +
			" /Length " + strconv.Itoa(len(content)) + " >>\nstream\n")
		buf.WriteString(content)
		buf.WriteString("\nendstream\nendobj\n")
	}

	buf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")

	// Filespec + EmbeddedFile stream pairs start at object 3.
	fileSpecObj := func(i int) int { return 3 + i*2 }
	streamObj := func(i int) int { return 4 + i*2 }
	// Name-tree objects follow the file pairs.
	leafObj := 3 + len(files)*2
	rootObj := leafObj // when not nested, the leaf is the tree root
	if nested {
		rootObj = leafObj + 1
	}

	catalogNames := ""
	if len(files) > 0 {
		catalogNames = fmt.Sprintf(" /Names << /EmbeddedFiles %d 0 R >>", rootObj)
	}
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R"+catalogNames+" >>")
	writeObj(2, "<< /Type /Pages /Kids [] /Count 0 >>")

	for i, f := range files {
		writeObj(fileSpecObj(i), fmt.Sprintf(
			"<< /Type /Filespec /F (%s) /UF (%s) /EF << /F %d 0 R >> >>",
			f.name, f.name, streamObj(i)))
		writeStream(streamObj(i), "/Type /EmbeddedFile /Subtype /text#2Fxml", f.content)
	}

	if len(files) > 0 {
		var names bytes.Buffer
		for i, f := range files {
			fmt.Fprintf(&names, "(%s) %d 0 R ", f.name, fileSpecObj(i))
		}
		writeObj(leafObj, "<< /Names [ "+names.String()+"] >>")
		if nested {
			writeObj(rootObj, fmt.Sprintf("<< /Kids [ %d 0 R ] >>", leafObj))
		}
	}

	xrefOff := buf.Len()
	buf.WriteString("xref\n0 " + strconv.Itoa(numObjs+1) + "\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= numObjs; i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	buf.WriteString("trailer\n<< /Size " + strconv.Itoa(numObjs+1) +
		" /Root 1 0 R >>\nstartxref\n" + strconv.Itoa(xrefOff) + "\n%%EOF\n")
	return buf.Bytes()
}

func writeTestPDF(t *testing.T, files []embeddedFile, nested bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.pdf")
	if err := os.WriteFile(path, buildTestPDF(files, nested), 0o644); err != nil {
		t.Fatalf("write test PDF: %v", err)
	}
	return path
}

func TestExtractXMLFromPDF_KnownName(t *testing.T) {
	want := `<?xml version="1.0"?><rsm:CrossIndustryInvoice>known</rsm:CrossIndustryInvoice>`
	path := writeTestPDF(t, []embeddedFile{
		{name: "other.xml", content: "<wrong/>"},
		{name: "factur-x.xml", content: want},
	}, false)

	got, err := extractXMLFromPDF(path)
	if err != nil {
		t.Fatalf("extractXMLFromPDF() error = %v", err)
	}
	if string(got) != want {
		t.Errorf("extractXMLFromPDF() = %q, want %q", got, want)
	}
}

func TestExtractXMLFromPDF_NestedNameTree(t *testing.T) {
	want := `<rsm:CrossIndustryInvoice>nested</rsm:CrossIndustryInvoice>`
	path := writeTestPDF(t, []embeddedFile{{name: "ZUGFeRD-invoice.xml", content: want}}, true)

	got, err := extractXMLFromPDF(path)
	if err != nil {
		t.Fatalf("extractXMLFromPDF() error = %v", err)
	}
	if string(got) != want {
		t.Errorf("extractXMLFromPDF() = %q, want %q", got, want)
	}
}

func TestExtractXMLFromPDF_XMLFallback(t *testing.T) {
	want := `<Invoice>fallback</Invoice>`
	// No known invoice filename present; the .xml fallback must pick it up.
	path := writeTestPDF(t, []embeddedFile{{name: "custom-invoice.xml", content: want}}, false)

	got, err := extractXMLFromPDF(path)
	if err != nil {
		t.Fatalf("extractXMLFromPDF() error = %v", err)
	}
	if string(got) != want {
		t.Errorf("extractXMLFromPDF() = %q, want %q", got, want)
	}
}

func TestExtractXMLFromPDF_NoEmbeddedFiles(t *testing.T) {
	path := writeTestPDF(t, nil, false)

	_, err := extractXMLFromPDF(path)
	if err == nil {
		t.Fatal("extractXMLFromPDF() error = nil, want error for PDF without attachments")
	}
}

func TestExtractXMLFromPDF_NoXMLAttachment(t *testing.T) {
	// An embedded file that is not XML and has no known invoice name.
	path := writeTestPDF(t, []embeddedFile{{name: "logo.png", content: "PNGDATA"}}, false)

	_, err := extractXMLFromPDF(path)
	if err == nil {
		t.Fatal("extractXMLFromPDF() error = nil, want error for PDF without invoice XML")
	}
}
