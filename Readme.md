[![Go Reference](https://pkg.go.dev/badge/github.com/speedata/einvoice.svg)](https://pkg.go.dev/github.com/speedata/einvoice)
[![Go Report Card](https://goreportcard.com/badge/github.com/speedata/einvoice)](https://goreportcard.com/report/github.com/speedata/einvoice)
[![Releases](https://img.shields.io/github/v/release/speedata/einvoice?include_prereleases)](https://github.com/speedata/einvoice/releases/latest)

# einvoice - a Go library to read, write and verify electronic invoices

**Work in progress**

This library will be used to read, write and verify electronic invoices (XML) which conform to the EN 16931 standard.

## Installation

    go get github.com/speedata/einvoice

## Usage

```go
invoice, err := einvoice.ParseXMLFile(filename)
if err != nil {
	...
}
// now invoice contains all the information from the XML file
// check for validation violations
err = invoice.Validate()
if err != nil {
	var valErr *einvoice.ValidationError
	if errors.As(err, &valErr) {
		for _, v := range valErr.Violations() {
			fmt.Printf("Error: %s - %s\n", v.Rule.Code, v.Text)
		}
		for _, w := range valErr.Warnings() {
			fmt.Printf("Warning: %s - %s\n", w.Rule.Code, w.Text)
		}
	}
} else {
	// Validation passed - check for recommendations
	for _, w := range invoice.Warnings() {
		fmt.Printf("Recommendation: %s - %s\n", w.Rule.Code, w.Text)
	}
}
```

Building and validating an invoice programmatically:

```go
import "errors"

func buildInvoice() error {
	inv := &einvoice.Invoice{
		InvoiceNumber: "INV-001",
		// ... set all required fields
	}

	// Validate before writing
	if err := inv.Validate(); err != nil {
		var valErr *einvoice.ValidationError
		if errors.As(err, &valErr) {
			for _, v := range valErr.Violations() {
				fmt.Printf("Rule %s: %s\n", v.Rule, v.Text)
			}
		}
		return err
	}

	return inv.Write(os.Stdout)
}
```

Round-trip: parsing and writing back:

```go
func dothings() error {
	inv, err := einvoice.ParseXMLFile("...")
	if err != nil {
		return err
	}

	// Write back in the same format (CII or UBL)
	return inv.Write(os.Stdout)
}
```

### Intelligent Validation with Auto-Detection

The `Validate()` method automatically detects and applies the appropriate validation rules:

- **EN 16931 Core Rules**: Always validated for all invoices
- **PEPPOL BIS Billing 3.0**: Auto-detected based on specification identifier (BT-24)
- **Country-Specific Rules**: Auto-detected based on seller country (future: DK, IT, NL, NO, SE)

Example of a PEPPOL invoice being automatically validated:

```go
inv := &einvoice.Invoice{
	GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0",
	BPSpecifiedDocumentContextParameter: "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
	// ... other fields
}

// Automatically validates both EN 16931 AND PEPPOL rules
if err := inv.Validate(); err != nil {
	// Handle validation errors
}
```

No need to call separate validation methods - `Validate()` handles everything automatically!

### Warnings vs Errors

The validation distinguishes between **errors** (hard requirements) and **warnings** (recommendations):

- **Errors**: Violations of "must" requirements cause `Validate()` to return an error
- **Warnings**: Violations of "should" recommendations don't fail validation but are reported

```go
err := invoice.Validate()
if err == nil {
	// Validation passed - but check for recommendations
	if invoice.HasWarnings() {
		for _, w := range invoice.Warnings() {
			fmt.Printf("Recommendation: %s - %s\n", w.Rule.Code, w.Text)
		}
	}
}
```

For example, BR-DE-21 recommends that German sellers use XRechnung specification identifier. This is reported as a warning for German sellers using other profiles (Factur-X, PEPPOL), but doesn't fail validation.

There is a [dedicated example](https://pkg.go.dev/github.com/speedata/einvoice#example-Invoice.Write) in [the documentation](https://pkg.go.dev/github.com/speedata/einvoice).

## Command Line Tool

A CLI tool is available for validating invoices from the command line.

### Installation

```bash
go install github.com/speedata/einvoice/cmd/einvoice@latest
```

### Usage

Validate an invoice and display violations in human-readable format:

```bash
einvoice validate invoice.xml
```

Output validation results as JSON (useful for automation and CI/CD):

```bash
einvoice validate --format json invoice.xml
```

### Exit Codes

- `0` - Invoice is valid (no violations)
- `1` - Error occurred (file not found, parse error, etc.)
- `2` - Invoice has validation violations

These exit codes make it easy to integrate the validator into shell scripts and CI/CD pipelines:

```bash
if einvoice validate invoice.xml; then
    echo "Invoice is valid!"
else
    echo "Validation failed"
fi
```

## Current status

* **Reading and writing** of EN 16931 invoices in both formats:
  - **ZUGFeRD/Factur-X** (CII format)
  - **UBL 2.1** (Invoice and CreditNote)
* Intelligent validation with auto-detection:
  - **EN 16931 Core Rules**: BR-1 to BR-65, BR-CO-*, BR-DEC-*
  - **VAT Category Rules**: BR-S-*, BR-AE-*, BR-E-*, BR-Z-*, BR-G-*, BR-IC-*, BR-IG-*, BR-IP-*, BR-O-*
  - **PEPPOL BIS Billing 3.0**: Auto-detected and validated (PEPPOL-EN16931-R*)
  - Single `Validate()` method handles all rule sets automatically
* XML output for all ZUGFeRD profiles (Minimum, BasicWL, Basic, EN16931, Extended, XRechnung)
* UBL 2.1 Invoice and CreditNote output with full EN 16931 compliance
* Profile detection based on specification identifier URN (BT-24)
* Format auto-detection when parsing (automatically recognizes CII or UBL)
* Round-trip support: parse and write back in the same format

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup and prerequisites
- Running tests and checking coverage
- Code style and conventions
- Pull request process
- How to work with test fixtures

## Test Fixtures

Test fixtures are organized by profile and format in the [testdata/](testdata/) directory. See [testdata/README.md](testdata/README.md) for:
- Directory structure and organization
- Fixture sources and provenance (from official EN 16931 and PEPPOL test suites)
- How to add new fixtures
- Usage patterns in tests

