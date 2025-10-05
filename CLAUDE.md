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
- Detected violations stored in `Invoice.Violations`

**Calculation (`calculate.go`)**
- `UpdateApplicableTradeTax(exemptReason)`: Recalculates VAT breakdown from line items and document-level allowances/charges per BR-45
- `UpdateTotals()`: Recalculates all monetary totals per business rules:
  - BR-CO-10: LineTotal = sum of line net amounts (BT-106)
  - BR-CO-13: TaxBasisTotal = LineTotal - AllowanceTotal + ChargeTotal (BT-109)
  - BR-CO-15: GrandTotal = TaxBasisTotal + TaxTotal (BT-112)
  - BR-CO-16: DuePayableAmount = GrandTotal - TotalPrepaid + RoundingAmount (BT-115)

**Validation (`check.go`)**
- `check()`: Validates invoice against EN 16931 business rules (BR-1 to BR-45+)
- Returns `SemanticError` structs with rule number, affected fields, and description
- Called automatically by `ParseReader()`
- Methods: `checkBR()`, `checkBRO()`, `checkOther()`

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
- Violations are accumulated in `Invoice.Violations`, not errors
- Parsing succeeds even with violations (allows partial data recovery)

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
