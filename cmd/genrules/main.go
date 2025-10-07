// genrules generates Go code from EN 16931 schematron specifications.
//
// This tool parses official schematron files and generates Rule definitions
// for the einvoice/rules package.
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

const version = "0.1.0"

// Command line flags
var (
	sourceFlag  = flag.String("source", "", "Schematron file path or URL (required)")
	outputFlag  = flag.String("output", "rules/en16931.go", "Output file path")
	packageFlag = flag.String("package", "rules", "Target package name")
	versionFlag = flag.String("version", "", "Source file version (e.g., v1.3.14.1)")
	helpFlag    = flag.Bool("help", false, "Show help message")
)

// SchematronSchema represents a schematron schema root element
type SchematronSchema struct {
	XMLName  xml.Name             `xml:"schema"`
	Patterns []SchematronPattern  `xml:"pattern"`
}

// SchematronPattern represents a schematron pattern element
type SchematronPattern struct {
	XMLName xml.Name         `xml:"pattern"`
	ID      string           `xml:"id,attr"`
	Rules   []SchematronRule `xml:"rule"`
}

// SchematronRule represents a schematron rule element
type SchematronRule struct {
	XMLName xml.Name           `xml:"rule"`
	Context string             `xml:"context,attr"`
	Asserts []SchematronAssert `xml:"assert"`
}

// SchematronAssert represents a schematron assert element
type SchematronAssert struct {
	XMLName     xml.Name `xml:"assert"`
	ID          string   `xml:"id,attr"`
	Test        string   `xml:"test,attr"`
	Flag        string   `xml:"flag,attr"`
	Description string   `xml:",chardata"`
}

// Rule represents a parsed business rule ready for code generation
type Rule struct {
	ID          string   // Go identifier (e.g., "BR1", "BRS8", "BRCO14")
	Code        string   // Rule code (e.g., "BR-01", "BR-S-08", "BR-CO-14")
	Fields      []string // BT-/BG- identifiers
	Description string   // Cleaned description text
}

// ByCode implements sort.Interface for []Rule based on Code field
type ByCode []Rule

func (a ByCode) Len() int      { return len(a) }
func (a ByCode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCode) Less(i, j int) bool {
	// Natural sort: BR-1 < BR-2 < BR-10 < BR-CO-1 < BR-S-1
	return compareRuleCodes(a[i].Code, a[j].Code)
}

func main() {
	flag.Parse()

	if *helpFlag || *sourceFlag == "" {
		printUsage()
		if *sourceFlag == "" {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Read schematron source
	schematronData, err := readSource(*sourceFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading source: %v\n", err)
		os.Exit(1)
	}

	// Parse schematron XML
	rules, err := parseSchematron(schematronData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schematron: %v\n", err)
		os.Exit(1)
	}

	// Generate Go code
	output, err := generateCode(rules, *packageFlag, *versionFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Format with gofmt
	formatted, err := formatCode(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: gofmt failed: %v\n", err)
		formatted = output // Use unformatted as fallback
	}

	// Write output file
	if err := writeOutput(*outputFlag, formatted); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d rules to %s\n", len(rules), *outputFlag)
}

// readSource reads the schematron file from a local path or URL
func readSource(source string) ([]byte, error) {
	// Check if source is a URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URL: %w", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP error: %s", resp.Status)
		}

		return io.ReadAll(resp.Body)
	}

	// Read from local file
	return os.ReadFile(source)
}

// parseSchematron parses the schematron XML and extracts business rules
// Handles both full schema documents and single pattern documents
func parseSchematron(data []byte) ([]Rule, error) {
	// Try parsing as full schema first
	var schema SchematronSchema
	var patterns []SchematronPattern

	if err := xml.Unmarshal(data, &schema); err == nil && len(schema.Patterns) > 0 {
		// Successfully parsed as schema with patterns
		patterns = schema.Patterns
	} else {
		// Try parsing as single pattern
		var pattern SchematronPattern
		if err := xml.Unmarshal(data, &pattern); err != nil {
			return nil, fmt.Errorf("failed to parse XML: %w", err)
		}
		patterns = []SchematronPattern{pattern}
	}

	// Use map to deduplicate rules by code
	ruleMap := make(map[string]Rule)
	for _, pattern := range patterns {
		for _, schemaRule := range pattern.Rules {
			for _, assert := range schemaRule.Asserts {
				if assert.ID == "" {
					continue // Skip asserts without IDs
				}

				rule := Rule{
					ID:          ruleCodeToIdentifier(assert.ID),
					Code:        assert.ID,
					Description: cleanDescription(assert.Description),
					Fields:      extractFields(assert.Description),
				}

				// Only add if not already present (first occurrence wins)
				if _, exists := ruleMap[assert.ID]; !exists {
					ruleMap[assert.ID] = rule
				}
			}
		}
	}

	// Convert map to slice
	var rules []Rule
	for _, rule := range ruleMap {
		rules = append(rules, rule)
	}

	// Sort rules by code
	sort.Sort(ByCode(rules))

	return rules, nil
}

// ruleCodeToIdentifier converts a rule code to a valid Go identifier
// Examples:
//   - "BR-01" → "BR1", "BR-S-08" → "BRS8", "BR-CO-14" → "BRCO14"
//   - "PEPPOL-EN16931-R001" → "PEPPOLEN16931R1"
//   - "PEPPOL-COMMON-R040" → "PEPPOLCOMMONR40"
//   - "DK-R-002" → "DKR2"
func ruleCodeToIdentifier(code string) string {
	// Split on dashes
	parts := strings.Split(code, "-")

	var result strings.Builder
	for _, part := range parts {
		// Check if part is all digits
		if len(part) > 0 && isAllDigits(part) {
			// Remove leading zeros from numeric parts
			trimmed := strings.TrimLeft(part, "0")
			if trimmed == "" {
				// All zeros, keep one
				result.WriteString("0")
			} else {
				result.WriteString(trimmed)
			}
		} else {
			// For compound parts like "R001", strip leading zeros from trailing digits
			result.WriteString(stripLeadingZerosFromSuffix(part))
		}
	}

	return result.String()
}

// isAllDigits checks if a string contains only digits
func isAllDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return len(s) > 0
}

// stripLeadingZerosFromSuffix removes leading zeros from numeric suffix
// Examples: "R001" → "R1", "CL005" → "CL5", "ABC" → "ABC"
func stripLeadingZerosFromSuffix(s string) string {
	// Find where digits start from the right
	i := len(s) - 1
	for i >= 0 && s[i] >= '0' && s[i] <= '9' {
		i--
	}

	// i now points to last non-digit (or -1 if all digits)
	if i < 0 {
		// All digits - shouldn't happen if isAllDigits check is done first
		return strings.ToUpper(s)
	}

	prefix := s[:i+1]
	suffix := s[i+1:]

	if len(suffix) == 0 {
		// No digits
		return strings.ToUpper(s)
	}

	// Remove leading zeros from suffix
	trimmed := strings.TrimLeft(suffix, "0")
	if trimmed == "" {
		// All zeros
		trimmed = "0"
	}

	return strings.ToUpper(prefix) + trimmed
}

// cleanDescription removes rule ID prefixes and normalizes whitespace
// Handles both EN16931 format ([BR-XX]-) and PEPPOL format ([RULE-ID])
func cleanDescription(desc string) string {
	// Find and remove [RULE-ID]- or [RULE-ID] prefix
	// Matches: [BR-XX]-, [IT-R-001], [PEPPOL-EN16931-R001], etc.
	re := regexp.MustCompile(`^\[[^\]]+\]-?\s*`)
	desc = re.ReplaceAllString(desc, "")

	// Normalize whitespace (collapse multiple spaces/newlines to single space)
	desc = strings.Join(strings.Fields(desc), " ")

	// Trim
	desc = strings.TrimSpace(desc)

	return desc
}

// extractFields extracts BT-/BG- identifiers from the description
func extractFields(desc string) []string {
	// Regex pattern to match (BT-nnn) or (BG-nnn)
	re := regexp.MustCompile(`\(B[TG]-\d+\)`)
	matches := re.FindAllString(desc, -1)

	if len(matches) == 0 {
		return nil
	}

	// Remove parentheses and deduplicate
	seen := make(map[string]bool)
	var fields []string
	for _, match := range matches {
		// Remove parentheses: "(BT-24)" → "BT-24"
		field := strings.Trim(match, "()")
		if !seen[field] {
			seen[field] = true
			fields = append(fields, field)
		}
	}

	// Sort alphabetically for consistency
	sort.Strings(fields)

	return fields
}

// compareRuleCodes implements natural sorting for rule codes
func compareRuleCodes(a, b string) bool {
	// Simple implementation: lexicographic comparison works for most cases
	// Could be enhanced for true natural sorting if needed
	return a < b
}

// generateCode generates Go source code from the parsed rules
func generateCode(rules []Rule, packageName, sourceVersion string) ([]byte, error) {
	tmpl := template.Must(template.New("rules").Parse(codeTemplate))

	data := struct {
		Package       string
		SourceVersion string
		Generated     string
		Rules         []Rule
	}{
		Package:       packageName,
		SourceVersion: sourceVersion,
		Generated:     time.Now().UTC().Format(time.RFC3339),
		Rules:         rules,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return []byte(buf.String()), nil
}

// formatCode runs gofmt on the generated code
func formatCode(code []byte) ([]byte, error) {
	cmd := exec.Command("gofmt", "-s")
	cmd.Stdin = strings.NewReader(string(code))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gofmt failed: %w", err)
	}

	return output, nil
}

// writeOutput writes the generated code to the output file
func writeOutput(path string, data []byte) error {
	// Create output directory if needed
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// printUsage prints the usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, `genrules v%s - Generate Go code from EN 16931 schematron specifications

Usage:
  genrules [flags]

Flags:
  --source string    Schematron file path or URL (required)
  --output string    Output file path (default: rules/en16931.go)
  --package string   Target package name (default: rules)
  --version string   Source file version for tracking (optional)
  --help             Show this help message

Examples:
  # Generate from local file
  genrules --source /tmp/EN16931-CII-model-abstract.sch

  # Generate from URL with version tracking
  genrules \
    --source https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/master/cii/schematron/abstract/EN16931-CII-model.sch \
    --version v1.3.14.1 \
    --output rules/en16931.go

  # Generate to different package
  genrules \
    --source /tmp/EN16931-CII-model-abstract.sch \
    --package myrules \
    --output myrules/rules.go

Source: https://github.com/speedata/einvoice
`, version)
}

// codeTemplate is the Go code generation template
const codeTemplate = `// Code generated by genrules from EN16931-CII-model.sch; DO NOT EDIT.
// Source: https://github.com/ConnectingEurope/eInvoicing-EN16931
{{- if .SourceVersion}}
// Version: {{.SourceVersion}}
{{- end}}
// Generated: {{.Generated}}

package {{.Package}}

var (
{{- range .Rules}}
	{{.ID}} = Rule{
		Code:        "{{.Code}}",
		{{- if .Fields}}
		Fields:      []string{ {{range $i, $f := .Fields}}{{if $i}}, {{end}}"{{$f}}"{{end}} },
		{{- else}}
		Fields:      nil,
		{{- end}}
		Description: ` + "`" + `{{.Description}}` + "`" + `,
	}
{{- end}}
)
`
