# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library for reading, writing, and verifying electronic invoices (XML) conforming to the EN 16931 standard. It supports ZUGFeRD/Factur-X Cross Industry Invoice (CII) format. UBL format is not yet supported.

## Common Commands

### Testing
```bash
# Run all tests
go test

# Run tests in verbose mode
go test -v

# Run a specific test
go test -run TestName

# Run tests with coverage
go test -cover
```

### Building
```bash
# This is a library package, not a standalone binary
# To use it in another project:
go get github.com/speedata/einvoice
```

## Architecture

### Core Components

**Data Model (`model.go`)**
- `Invoice`: Main struct containing all invoice data per EN 16931
- `Party`: Represents buyer, seller, payee, ship-to parties
- `InvoiceLine`: Individual line items on the invoice
- `TradeTax`: VAT breakdown per category (BG-23)
- `AllowanceCharge`: Discounts/charges at document or line level
- Profile types: `CProfileMinimum`, `CProfileBasicWL`, `CProfileBasic`, `CProfileEN16931`, `CProfileExtended`, `CProfileXRechnung`

**Parsing (`parser.go`)**
- `ParseXMLFile(filename)`: Reads ZUGFeRD/Factur-X XML files
- `ParseReader(io.Reader)`: Parses from any reader
- Uses XPath-based parsing via `github.com/speedata/cxpath`
- Automatically validates business rules during parsing

**Calculation (`calculate.go`)**
- `UpdateApplicableTradeTax(exemptReason)`: Recalculates VAT breakdown from line items and document-level allowances/charges per BR-45
- `UpdateTotals()`: Recalculates all monetary totals per business rules:
  - BR-CO-10: LineTotal = sum of line net amounts (BT-106)
  - BR-CO-13: TaxBasisTotal = LineTotal - AllowanceTotal + ChargeTotal (BT-109)
  - BR-CO-15: GrandTotal = TaxBasisTotal + TaxTotal (BT-112)
  - BR-CO-16: DuePayableAmount = GrandTotal - TotalPrepaid + RoundingAmount (BT-115)

**Validation (Organized by domain)**
The validation logic is split across multiple focused files for maintainability:

- `check.go`: Main orchestrator containing core business rules (BR-1 to BR-65) and public `Validate()` method
- `validation.go`: ValidationError type with methods for accessing violations
- `check_vat_standard.go`: Standard rated VAT validations (BR-S-1 to BR-S-10)
- `check_vat_reverse.go`: Reverse charge VAT validations (BR-AE-1 to BR-AE-10)
- `check_vat_exempt.go`: Exempt from VAT validations (BR-E-1 to BR-E-10)
- `check_vat_zero.go`: Zero rated VAT validations (BR-Z-1 to BR-Z-10)
- `check_vat_export.go`: Export outside EU validations (BR-G-1 to BR-G-10)
- `check_vat_intracommunity.go`: Intra-community supply validations (BR-IC-1 to BR-IC-12)
- `check_vat_igic.go`: IGIC (Canary Islands) validations (BR-IG-1 to BR-IG-10)
- `check_vat_ipsi.go`: IPSI (Ceuta/Melilla) validations (BR-IP-1 to BR-IP-10)
- `check_vat_notsubject.go`: Not subject to VAT validations (BR-O-1 to BR-O-14)

Each validation file contains a single method (e.g., `checkVATStandard()`) with comprehensive documentation explaining:
- The tax category purpose and requirements
- All business rules implemented in that file
- Field references (BT-/BG-) per EN 16931 specification

**Validation API:**
- Public: `Invoice.Validate() error` - validates and returns `ValidationError` if violations exist
- Private: `Invoice.violations` field - use `Validate()` or deprecated `Violations()` accessor
- Automatically runs during parsing; call explicitly when building invoices programmatically

**Business Rules (`rules/` package)**
- Auto-generated from official EN 16931 schematron specifications
- Source: [ConnectingEurope/eInvoicing-EN16931](https://github.com/ConnectingEurope/eInvoicing-EN16931)
- Current version: v1.3.14.1
- 203 rules extracted from schematron XML
- Generation tool: `cmd/genrules` - See [cmd/genrules/README.md](cmd/genrules/README.md)
- Regenerate with: `cd rules && go generate`

**Rule Structure:**
```go
type Rule struct {
    Code        string      // EN 16931 rule code (e.g., "BR-01", "BR-S-08")
    Fields      []string    // BT-/BG- identifiers from semantic model
    Description string      // Official specification requirement text
}
```

**Rule Naming:**
- `BR-01` → `rules.BR1` (remove leading zeros)
- `BR-S-08` → `rules.BRS8` (remove dashes and zeros)
- `BR-CO-14` → `rules.BRCO14` (remove all dashes)

**Custom Rules:**
The `rules/en16931.go` file includes custom rules not in the official schematron:
- `Check`: Line total calculation validation
- `BR34`, `BR35`, `BR39`, `BR40`: Non-negative amount validations
- `BRIG1-10`: Aliases for IGIC rules (Canary Islands - official: BR-AF-*)
- `BRIP1-10`: Aliases for IPSI rules (Ceuta/Melilla - official: BR-AG-*)

⚠️ When regenerating rules, manually preserve the custom rules section marked by the comment banner.

**Writing (`writer.go`)**
- `Invoice.Write(io.Writer)`: Outputs ZUGFeRD/Factur-X XML
- Uses `github.com/beevik/etree` for XML generation
- Profile-aware: outputs fields based on `Invoice.Profile` level
- Helper functions: `formatPercent()`, `addTimeUDT()`, `addTimeQDT()`

### Key Design Patterns

**Profile Hierarchy**
Profiles are ordered by inclusiveness: Extended > EN16931 > Basic > BasicWL > Minimum. The `is()` function checks if an invoice profile meets a minimum level. Higher profiles include all fields from lower profiles.

**Business Rule Validation**
- Rules are named per EN 16931 spec: BR-1, BR-CO-10, BR-S-8, etc.
- `Validate()` returns `ValidationError` containing violations, or nil if valid
- Parsing succeeds even with violations (allows partial data recovery)
- Access violations via `ValidationError.Violations()` or deprecated `Invoice.Violations()`

**Decimal Precision**
All monetary amounts use `github.com/shopspring/decimal` for exact arithmetic. Tax calculations round to 2 decimal places. VAT percentage formatting removes trailing zeros via regex.

**XML Namespaces**
- `rsm`: CrossIndustryInvoice:100
- `ram`: ReusableAggregateBusinessInformationEntity:100
- `udt`: UnqualifiedDataType:100
- `qdt`: QualifiedDataType:100

### BT/BG Field References

The codebase uses EN 16931 Business Term (BT-) and Business Group (BG-) notation extensively:
- BT-1: Invoice number
- BT-106: Sum of invoice line net amounts
- BT-110: Invoice total VAT amount
- BG-23: VAT breakdown
- BG-25: Invoice line

Comments in code reference these terms. When modifying calculations or validation, preserve these references for traceability to the specification.

### Common Pitfalls

**Calculation Dependencies**
Always call `UpdateApplicableTradeTax()` before `UpdateTotals()` when modifying invoice data. TradeTaxes must be current before total calculations.

**Time Parsing**
Dates use format "20060102" (YYYYMMDD). Zero time values are treated as "not set" rather than errors.

**ChargeIndicator Boolean**
In `AllowanceCharge`, `ChargeIndicator=false` means allowance/discount, `true` means charge. This affects sign in calculations.
