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

**Profile Detection:**
Profiles are detected automatically via `GuidelineSpecifiedDocumentContextParameter` (BT-24) URN string:
- `urn:factur-x.eu:1p0:minimum` → Minimum profile
- `urn:factur-x.eu:1p0:basicwl` → Basic WL profile
- `urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic` → Basic profile
- `urn:cen.eu:en16931:2017` → EN 16931 profile
- `urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended` → Extended profile
- `urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0` → XRechnung profile
- `urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0` → PEPPOL BIS Billing 3.0

Helper methods: `IsMinimum()`, `IsBasicWL()`, `IsBasic()`, `IsEN16931()`, `IsExtended()`, `IsXRechnung()`, `ProfileLevel()` (returns int 0-5)

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
The validation logic is split across multiple focused files for maintainability. All validation files and functions follow the `validate_*.go` and `validate*()` naming pattern:

- `validation.go`: ValidationError type, `Validate()` method with intelligent auto-detection
- `validate_core.go`: Core business rules (BR-*, BR-CO-*, BR-DEC-*)
  - `validateCore()`: BR-1 through BR-65
  - `validateCalculations()`: BR-CO-* calculation rules
  - `validateDecimals()`: BR-DEC-* decimal precision rules
- `validate_vat_standard.go`: Standard rated VAT validations (BR-S-1 to BR-S-10)
- `validate_vat_reverse.go`: Reverse charge VAT validations (BR-AE-1 to BR-AE-10)
- `validate_vat_exempt.go`: Exempt from VAT validations (BR-E-1 to BR-E-10)
- `validate_vat_zero.go`: Zero rated VAT validations (BR-Z-1 to BR-Z-10)
- `validate_vat_export.go`: Export outside EU validations (BR-G-1 to BR-G-10)
- `validate_vat_ic.go`: Intra-community supply validations (BR-IC-1 to BR-IC-12)
- `validate_vat_igic.go`: IGIC (Canary Islands) validations (BR-IG-1 to BR-IG-10)
- `validate_vat_ipsi.go`: IPSI (Ceuta/Melilla) validations (BR-IP-1 to BR-IP-10)
- `validate_vat_notsubject.go`: Not subject to VAT validations (BR-O-1 to BR-O-14)
- `validate_peppol.go`: PEPPOL BIS Billing 3.0 validations (PEPPOL-EN16931-R*)

Each validation file contains a single method (e.g., `validateVATStandard()`) with comprehensive documentation explaining:
- The tax category purpose and requirements
- All business rules implemented in that file
- Field references (BT-/BG-) per EN 16931 specification

**Validation API:**
- Public: `Invoice.Validate() error` - Intelligent validation with auto-detection
  - Always validates EN 16931 core rules
  - Auto-detects PEPPOL based on specification identifier (BT-24)
  - Auto-detects country rules based on seller location (future: DK, IT, NL, NO, SE)
  - Returns `ValidationError` if violations exist, nil if valid
- Private: `Invoice.violations` field - use `Validate()` accessor
- Automatically runs during parsing; call explicitly when building invoices programmatically

**Auto-Detection:**
The `Validate()` method uses intelligent auto-detection:
- `isPEPPOL()`: Detects PEPPOL BIS Billing 3.0 from URN
- Country helpers: `isDanish()`, `isItalian()`, `isDutch()`, `isNorwegian()`, `isSwedish()`
- Eliminates need for separate validation methods like the deprecated `ValidatePEPPOL()`

**Business Rules (`rules/` package)**
- 203 rules auto-generated from EN 16931 schematron (v1.3.14.1)
- Source: [ConnectingEurope/eInvoicing-EN16931](https://github.com/ConnectingEurope/eInvoicing-EN16931)
- Regenerate: `cd rules && go generate`
- Details: [cmd/genrules/README.md](cmd/genrules/README.md)

Package structure:
- `types.go`: Rule struct (manual)
- `custom.go`: Custom rules and aliases (manual)
- `en16931.go`: Generated rule constants (auto-generated)
- `generate.go`: go:generate directive (manual)

Rule naming: `BR-01` → `rules.BR1`, `BR-S-08` → `rules.BRS8`, `BR-CO-14` → `rules.BRCO14`

**Writing (`writer.go`)**
- `Invoice.Write(io.Writer)`: Outputs ZUGFeRD/Factur-X XML
- Uses `github.com/beevik/etree` for XML generation
- Profile-aware: outputs fields based on `ProfileLevel()` method
- Level constants: `levelMinimum`=1, `levelBasicWL`=2, `levelBasic`=3, `levelEN16931`=4, `levelExtended`=5
- Helper functions: `formatPercent()`, `addTimeUDT()`, `addTimeQDT()`

### Key Design Patterns

**Profile Hierarchy**
Profiles are ordered by inclusiveness: Extended (5) > EN16931/PEPPOL/XRechnung (4) > Basic (3) > BasicWL (2) > Minimum (1).
- `ProfileLevel()` returns int 0-5 based on URN
- `MeetsProfileLevel(minLevel int)` replaces old enum comparisons
- `is(minLevel int, inv)` writer helper checks profile level
- Higher profiles include all fields from lower profiles
- Single source of truth: `GuidelineSpecifiedDocumentContextParameter` URN string

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
