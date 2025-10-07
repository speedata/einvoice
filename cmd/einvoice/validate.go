package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/speedata/einvoice"
)

// Result represents the validation result for JSON output
type Result struct {
	File       string       `json:"file"`
	Valid      bool         `json:"valid"`
	Invoice    *InvoiceRef  `json:"invoice,omitempty"`
	Violations []Violation  `json:"violations,omitempty"`
	Error      string       `json:"error,omitempty"`
}

// Violation represents a business rule violation
type Violation struct {
	Rule        string   `json:"rule"`
	Fields      []string `json:"fields,omitempty"`
	Description string   `json:"description,omitempty"`
	Text        string   `json:"text"`
}

// InvoiceRef contains basic invoice metadata
type InvoiceRef struct {
	Number string `json:"number,omitempty"`
	Date   string `json:"date,omitempty"`
	Total  string `json:"total,omitempty"`
}

func runValidate(args []string) int {
	// Parse flags for the validate subcommand
	validateFlags := flag.NewFlagSet("validate", flag.ExitOnError)
	var format string
	var verbose bool
	var profile string
	validateFlags.StringVar(&format, "format", "text", "Output format: text, json")
	validateFlags.BoolVar(&verbose, "verbose", false, "Show detailed rule descriptions and all fields")
	validateFlags.StringVar(&profile, "profile", "en16931", "Validation profile: en16931, peppol")
	validateFlags.Usage = validateUsage
	_ = validateFlags.Parse(args)

	// Require exactly one file argument
	if validateFlags.NArg() != 1 {
		validateUsage()
		return exitError
	}

	filename := validateFlags.Arg(0)

	// Validate profile flag
	if profile != "en16931" && profile != "peppol" {
		fmt.Fprintf(os.Stderr, "Error: unknown profile %q (use 'en16931' or 'peppol')\n", profile)
		return exitError
	}

	// Validate the invoice
	result := validateInvoice(filename, profile)

	// Output results
	switch format {
	case "json":
		outputJSON(result)
	case "text":
		outputText(result, verbose)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format %q (use 'text' or 'json')\n", format)
		return exitError
	}

	// Return appropriate exit code
	if result.Error != "" {
		return exitError
	}
	if !result.Valid {
		return exitViolations
	}
	return exitOK
}

func validateInvoice(filename string, profile string) Result {
	result := Result{
		File: filename,
	}

	// Parse the invoice XML
	invoice, err := einvoice.ParseXMLFile(filename)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to parse invoice: %v", err)
		return result
	}

	// Extract basic invoice metadata
	result.Invoice = &InvoiceRef{
		Number: invoice.InvoiceNumber,
		Date:   invoice.InvoiceDate.Format("2006-01-02"),
		Total:  invoice.GrandTotal.String(),
	}

	// Validate the invoice according to the specified profile
	var validationErr error
	switch profile {
	case "peppol":
		validationErr = invoice.ValidatePEPPOL()
	default: // "en16931" or any other value defaults to EN 16931
		validationErr = invoice.Validate()
	}

	if validationErr == nil {
		result.Valid = true
		return result
	}

	// Extract violations
	if ve, ok := validationErr.(*einvoice.ValidationError); ok {
		result.Valid = false
		semanticErrors := ve.Violations()
		result.Violations = make([]Violation, len(semanticErrors))
		for i, se := range semanticErrors {
			result.Violations[i] = Violation{
				Rule:        se.Rule.Code,
				Fields:      se.Rule.Fields,
				Description: se.Rule.Description,
				Text:        se.Text,
			}
		}
	} else {
		result.Error = fmt.Sprintf("Validation failed: %v", validationErr)
	}

	return result
}

func outputText(result Result, verbose bool) {
	if result.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		return
	}

	if result.Valid {
		fmt.Printf("✓ Invoice %s is valid\n", result.Invoice.Number)
		return
	}

	fmt.Printf("✗ Invoice %s has %d violation(s):\n", result.Invoice.Number, len(result.Violations))
	for _, violation := range result.Violations {
		if verbose {
			// Verbose mode: show full details
			fmt.Printf("  - %s: %s\n", violation.Rule, violation.Text)
			if violation.Description != "" {
				fmt.Printf("    Specification: %s\n", violation.Description)
			}
			if len(violation.Fields) > 0 {
				fmt.Printf("    Fields: %s\n", formatFields(violation.Fields))
			}
		} else {
			// Normal mode: show primary field inline
			primaryField := ""
			if len(violation.Fields) > 0 {
				primaryField = fmt.Sprintf(" (%s)", violation.Fields[0])
			}
			fmt.Printf("  - %s%s: %s\n", violation.Rule, primaryField, violation.Text)
		}
	}
}

// formatFields joins field identifiers with commas
func formatFields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	result := fields[0]
	for i := 1; i < len(fields); i++ {
		result += ", " + fields[i]
	}
	return result
}

func outputJSON(result Result) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

func validateUsage() {
	fmt.Fprintf(os.Stderr, `Usage: einvoice validate [options] <file.xml>

Validates an electronic invoice against business rules.

Options:
  --format string   Output format: text, json (default "text")
  --profile string  Validation profile: en16931, peppol (default "en16931")
  --verbose         Show detailed rule descriptions and all fields
  --help            Show this help message

Profiles:
  en16931  Validate against EN 16931 business rules only
  peppol   Validate against EN 16931 + PEPPOL BIS Billing 3.0 rules

Exit codes:
  0  Invoice is valid
  1  Invoice has validation violations
  2  Error occurred (file not found, parse error, etc.)

Examples:
  einvoice validate invoice.xml
  einvoice validate --verbose invoice.xml
  einvoice validate --profile peppol invoice.xml
  einvoice validate --format json --profile peppol invoice.xml
`)
}
