# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library for reading, writing, and validating electronic invoices (XML) conforming to the EN 16931 standard. The library supports ZUGFeRD/Factur-X formats (CII - Cross Industry Invoice). UBL support is not yet implemented.

## Commands

### Testing
- Run all tests: `go test`
- Run tests with verbose output: `go test -v`
- Run a specific test: `go test -run TestName`
- Run example tests: `go test -run Example`

### Building
- Build the library: `go build`
- Get dependencies: `go mod download`
- Tidy dependencies: `go mod tidy`

### Formatting and Linting
- Format code: `go fmt ./...`
- Run vet: `go vet ./...`

## Architecture

### Core Data Model (`model.go`)

The library uses a hierarchical data structure to represent invoices:

- **Invoice**: Main container with profile type, dates, parties, totals, and line items
- **Party**: Represents Seller, Buyer, PayeeTradeParty, and ShipTo entities
- **InvoiceLine**: Individual line items with pricing, quantities, and tax information
- **TradeTax**: VAT breakdown for each tax category and rate
- **AllowanceCharge**: Document and line-level allowances/charges
- **PaymentMeans**: Payment details (IBAN, BIC, credit card info)

### Profile Types
Invoice profiles are ordered by completeness (from minimum to extended):
- CProfileMinimum
- CProfileBasicWL (basic without lines)
- CProfileBasic
- CProfileEN16931 (main standard)
- CProfileExtended
- CProfileXRechnung

The `is()` helper function in `writer.go` checks if an invoice's profile meets a minimum level.

### Parsing (`parser.go`)

- Uses `cxpath` (XPath library) for XML parsing
- Parses CII (Cross Industry Invoice) format only
- Main entry point: `ParseXMLFile(filename)` or parsing from `io.Reader`
- Converts XML to the Go struct model
- Automatically runs validation checks after parsing

### Writing (`writer.go`)

- Generates EN 16931 compliant XML output
- Uses `etree` library for XML construction
- Main entry point: `Invoice.Write(io.Writer)`
- Supports CII format output
- Profile-aware: includes/excludes elements based on invoice profile

### Validation (`check.go`)

Implements EN 16931 business rules (BR-1 to BR-45):
- BR rules: Core validation rules
- BR-O rules: Extended rules for specific profiles
- Returns violations in `Invoice.Violations []SemanticError`
- Violations are populated automatically during parsing

Business rules check for:
- Required fields based on profile
- Tax calculations and consistency
- VAT category validations
- Document reference requirements
- Payment terms requirements

### Calculation (`calculate.go`)

Helper functions for invoice calculations:

- `UpdateApplicableTradeTax(exemptReason map[string]string)`: Recalculates tax breakdown from line items, groups by category code and tax rate
- `UpdateTotals()`: Calculates monetary summation (line total, tax total, grand total, due payable)

These functions should be called when programmatically building invoices.

## Key Dependencies

- `github.com/shopspring/decimal`: Precise decimal arithmetic for monetary values
- `github.com/beevik/etree`: XML document construction for writing
- `github.com/speedata/cxpath`: XPath evaluation for parsing XML

## Important Implementation Notes

### Decimal Precision
All monetary amounts use `decimal.Decimal` type, never float64. This ensures precision in financial calculations.

### Date Handling
Dates are stored as `time.Time` but formatted as "YYYYMMDD" (format 102) in XML output.

### Tax Calculations
- Line items contain tax category, rate, and totals
- `UpdateApplicableTradeTax()` aggregates line taxes by category+rate
- Tax amounts are rounded to 2 decimal places
- Zero-rate taxes require exemption reasons

### Profile Awareness
When writing XML, the output structure varies by profile. Use the `is(profileType, invoice)` pattern to conditionally include elements based on minimum required profile.

## Test Data

Test cases are in the `testcases/` directory with sample ZUGFeRD XML files for validation testing.
