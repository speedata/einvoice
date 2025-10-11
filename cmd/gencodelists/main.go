// gencodelists generates code lists for human-readable descriptions
// of codes used in electronic invoicing (UNTDID, UNECE, etc.).
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"
)

const (
	untdidURL     = "https://raw.githubusercontent.com/invopop/gobl/main/data/catalogues/untdid.json"
	uneceURL      = "https://raw.githubusercontent.com/datasets/unece-units-of-measure/main/data/units-of-measure.csv"
	uneceRec21URL = "https://datahub.io/core/unece-package-codes/_r/-/data/data.csv"
	untdid4451URL = "https://www.xrepository.de/api/xrepository/urn:xoev-de:kosit:codeliste:untdid.4451_4/download/UNTDID_4451_4.json"
)

type codeEntry struct {
	Code string
	Name string
}

func main() {
	output := flag.String("output", "", "output file path")
	pkg := flag.String("package", "codelists", "package name")
	flag.Parse()

	if *output == "" {
		log.Fatal("--output flag is required")
	}

	log.Println("Fetching code lists...")

	// Fetch document types
	docTypes, err := fetchDocumentTypes()
	if err != nil {
		log.Fatalf("Failed to fetch document types: %v", err)
	}
	log.Printf("Fetched %d document types", len(docTypes))

	// Fetch UNECE Rec 20 unit codes
	rec20Codes, err := fetchUnitCodes()
	if err != nil {
		log.Fatalf("Failed to fetch Rec 20 unit codes: %v", err)
	}
	log.Printf("Fetched %d Rec 20 unit codes", len(rec20Codes))

	// Fetch UNECE Rec 21 package codes (prefixed with X)
	rec21Codes, err := fetchPackageCodes()
	if err != nil {
		log.Fatalf("Failed to fetch Rec 21 package codes: %v", err)
	}
	log.Printf("Fetched %d Rec 21 package codes", len(rec21Codes))

	// Merge unit codes
	unitCodes := mergeUnitCodes(rec20Codes, rec21Codes)
	log.Printf("Total unit codes: %d", len(unitCodes))

	// Fetch UNTDID 4451 text subject qualifiers
	textSubjectQualifiers, err := fetchTextSubjectQualifiers()
	if err != nil {
		log.Fatalf("Failed to fetch text subject qualifiers: %v", err)
	}
	log.Printf("Fetched %d text subject qualifiers", len(textSubjectQualifiers))

	// Generate Go code
	if err := generateGoCode(*output, *pkg, docTypes, unitCodes, textSubjectQualifiers); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	log.Printf("Generated %s", *output)
}

func fetchDocumentTypes() ([]codeEntry, error) {
	resp, err := http.Get(untdidURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data struct {
		Extensions []struct {
			Key    string `json:"key"`
			Values []struct {
				Code string          `json:"code"`
				Name map[string]string `json:"name"`
			} `json:"values"`
		} `json:"extensions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Find document type extension
	for _, ext := range data.Extensions {
		if ext.Key == "untdid-document-type" {
			var entries []codeEntry
			for _, val := range ext.Values {
				entries = append(entries, codeEntry{
					Code: val.Code,
					Name: val.Name["en"],
				})
			}
			return entries, nil
		}
	}

	return nil, fmt.Errorf("document types not found in UNTDID data")
}

func fetchUnitCodes() ([]codeEntry, error) {
	resp, err := http.Get(uneceURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	reader := csv.NewReader(resp.Body)

	// Read header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var entries []codeEntry

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// CSV format: Status,CommonCode,Name,Description,LevelAndCategory,Symbol,ConversionFactor
		if len(record) < 3 {
			continue
		}

		status := record[0]
		code := strings.TrimSpace(record[1])
		name := strings.TrimSpace(record[2])

		// Skip deprecated codes (Status = "X")
		if status == "X" {
			continue
		}

		// Skip empty codes
		if code == "" || name == "" {
			continue
		}

		entries = append(entries, codeEntry{
			Code: code,
			Name: name,
		})
	}

	return entries, nil
}

func fetchPackageCodes() ([]codeEntry, error) {
	resp, err := http.Get(uneceRec21URL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	reader := csv.NewReader(resp.Body)

	// Read header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var entries []codeEntry

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// CSV format: Status,Code,Name,Description,Numeric code
		if len(record) < 3 {
			continue
		}

		code := strings.TrimSpace(record[1])
		name := strings.TrimSpace(record[2])

		// Skip empty codes
		if code == "" || name == "" {
			continue
		}

		// Per PEPPOL rules: Rec 21 codes are prefixed with "X" to avoid duplication with Rec 20
		// Example: PP becomes XPP
		entries = append(entries, codeEntry{
			Code: "X" + code,
			Name: strings.ToLower(name), // lowercase to match Rec 20 style
		})
	}

	return entries, nil
}

func fetchTextSubjectQualifiers() ([]codeEntry, error) {
	resp, err := http.Get(untdid4451URL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// XRepository JSON structure
	var data struct {
		Daten [][]string `json:"daten"` // Array of [code, name, description]
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var entries []codeEntry
	for _, row := range data.Daten {
		if len(row) < 2 {
			continue
		}
		code := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])

		if code == "" || name == "" {
			continue
		}

		entries = append(entries, codeEntry{
			Code: code,
			Name: name,
		})
	}

	return entries, nil
}

func mergeUnitCodes(rec20, rec21 []codeEntry) []codeEntry {
	seen := make(map[string]bool)
	var merged []codeEntry

	// Add Rec 20 codes first
	for _, entry := range rec20 {
		if !seen[entry.Code] {
			seen[entry.Code] = true
			merged = append(merged, entry)
		}
	}

	// Add Rec 21 codes (only if not already present)
	for _, entry := range rec21 {
		if !seen[entry.Code] {
			seen[entry.Code] = true
			merged = append(merged, entry)
		}
	}

	// Sort by code
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Code < merged[j].Code
	})

	return merged
}

func generateGoCode(output, pkg string, docTypes, unitCodes, textSubjectQualifiers []codeEntry) error {
	tmpl := template.Must(template.New("codelists").Parse(codeTemplate))

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	data := struct {
		Package               string
		DocumentTypes         []codeEntry
		UnitCodes             []codeEntry
		TextSubjectQualifiers []codeEntry
	}{
		Package:               pkg,
		DocumentTypes:         docTypes,
		UnitCodes:             unitCodes,
		TextSubjectQualifiers: textSubjectQualifiers,
	}

	return tmpl.Execute(f, data)
}

const codeTemplate = `// Code generated by gencodelists. DO NOT EDIT.

package {{.Package}}

// documentTypes maps UNTDID 1001 document type codes to human-readable names.
// Source: https://github.com/invopop/gobl (Apache 2.0)
var documentTypes = map[string]string{
{{- range .DocumentTypes}}
	"{{.Code}}": "{{.Name}}",
{{- end}}
}

// unitCodes maps UNECE Recommendation 20 unit codes to human-readable names.
// Source: https://github.com/datasets/unece-units-of-measure
var unitCodes = map[string]string{
{{- range .UnitCodes}}
	"{{.Code}}": "{{.Name}}",
{{- end}}
}

// textSubjectQualifiers maps UNTDID 4451 text subject qualifier codes to human-readable names.
// Source: https://www.xrepository.de (KoSIT - Coordination Office for IT Standards, Germany)
var textSubjectQualifiers = map[string]string{
{{- range .TextSubjectQualifiers}}
	"{{.Code}}": "{{.Name}}",
{{- end}}
}
`
