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
	validateFlags.StringVar(&format, "format", "text", "Output format: text, json")
	validateFlags.BoolVar(&verbose, "verbose", false, "Show detailed rule descriptions and all fields")
	validateFlags.Usage = validateUsage
	_ = validateFlags.Parse(args)

	// Require exactly one file argument
	if validateFlags.NArg() != 1 {
		validateUsage()
		return exitError
	}

	filename := validateFlags.Arg(0)

	// Validate the invoice
	result := validateInvoice(filename)

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

func validateInvoice(filename string) Result {
	result := Result{
		File: filename,
	}

	// Parse the invoice (XML or PDF)
	invoice, err := parseInvoiceFile(filename)
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

	// Validate the invoice
	// Validate() automatically detects which rules to apply based on:
	// - Specification identifier (BT-24) for PEPPOL detection
	// - Seller country for country-specific rules
	validationErr := invoice.Validate()

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
	fmt.Fprintf(os.Stderr, `Usage: einvoice validate [options] <file>

Validates an electronic invoice against business rules.

Supports both XML and ZUGFeRD/Factur-X PDF formats.

The validator automatically detects which rules to apply based on:
  - Specification identifier (BT-24) for PEPPOL BIS Billing 3.0
  - Seller country for country-specific rules (DK, IT, NL, NO, SE)

All invoices are validated against EN 16931 core rules. Additional validation
rules (PEPPOL, country-specific) are applied automatically when detected.

Options:
  --format string   Output format: text, json (default "text")
  --verbose         Show detailed rule descriptions and all fields
  --help            Show this help message

Exit codes:
  0  Invoice is valid
  1  Error occurred (file not found, parse error, etc.)
  2  Invoice has validation violations

Examples:
  einvoice validate invoice.xml
  einvoice validate invoice.pdf
  einvoice validate --verbose invoice.xml
  einvoice validate --format json invoice.pdf
`)
}
