# UBL Support Options for einvoice Library

**Document Version:** 1.0
**Date:** 2025-10-12
**Status:** Analysis & Recommendations

## Executive Summary

This document analyzes options for adding Universal Business Language (UBL) support to the einvoice library, which currently supports only Cross Industry Invoice (CII) format. The analysis concludes that adding UBL support is:

- **Technically Feasible**: The current architecture is well-suited for dual-format support
- **Architecturally Sound**: Minimal changes needed, zero breaking changes
- **High Value**: Enables PEPPOL network support and broader European market coverage
- **Reasonable Effort**: Estimated 2-3 weeks of focused development

**Recommended Approach**: Incremental implementation with separate parser_ubl.go and writer_ubl.go files, maintaining format-agnostic core model and validation.

---

## Table of Contents

1. [Background](#background)
2. [Current Architecture Analysis](#current-architecture-analysis)
3. [UBL vs CII Comparison](#ubl-vs-cii-comparison)
4. [Implementation Options](#implementation-options)
5. [Recommended Architecture](#recommended-architecture)
6. [Technical Specifications](#technical-specifications)
7. [Implementation Phases](#implementation-phases)
8. [Testing Strategy](#testing-strategy)
9. [Effort Estimation](#effort-estimation)
10. [Risks and Mitigations](#risks-and-mitigations)
11. [Benefits and Use Cases](#benefits-and-use-cases)
12. [Resources Required](#resources-required)
13. [Appendix](#appendix)

---

## Background

### What is UBL?

Universal Business Language (UBL) is an OASIS standard XML format for electronic business documents, including invoices. UBL 2.1 is one of the two primary syntaxes for implementing the European Standard EN 16931 for electronic invoicing.

### Why Add UBL Support?

1. **PEPPOL Network**: PEPPOL BIS Billing 3.0 mandates UBL format
2. **European Markets**: Several countries prefer or mandate UBL (Italy, Netherlands, Nordic countries)
3. **Interoperability**: Cross-border invoicing often requires supporting both CII and UBL
4. **Market Coverage**: ZUGFeRD/Factur-X (CII) dominates Germany/France, UBL dominates elsewhere

### Current State

The library currently:
- ✅ Supports CII (ZUGFeRD/Factur-X) format
- ✅ Has UBL enum value in CodeSchemaType (prepared for future support)
- ✅ Returns error when attempting to write UBL format
- ❌ Cannot parse UBL invoices
- ❌ Cannot write UBL invoices

---

## Current Architecture Analysis

### Strengths for Dual-Format Support

The existing architecture is remarkably well-suited for adding UBL support:

#### 1. Format-Agnostic Data Model
```go
type Invoice struct {
    InvoiceNumber    string          // BT-1 (works for both formats)
    Seller           Party           // BG-4 (works for both formats)
    Buyer            Party           // BG-7 (works for both formats)
    InvoiceLines     []InvoiceLine   // BG-25 (works for both formats)
    TradeTaxes       []TradeTax      // BG-23 (works for both formats)
    SchemaType       CodeSchemaType  // Distinguishes CII vs UBL
    // ... all fields map to EN 16931 BT/BG references
}
```

The `Invoice` struct uses EN 16931 semantic model (BT/BG fields), not format-specific XML elements. This means:
- ✅ Same struct works for both CII and UBL
- ✅ No model changes needed
- ✅ Validation is format-independent

#### 2. Separated Parsing Logic
```go
// parser.go:514
func ParseReader(r io.Reader) (*Invoice, error) {
    ctx, err := cxpath.NewFromReader(r)
    rootns := ctx.Root().Eval("namespace-uri()").String()

    switch rootns {
    case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
        inv, err = parseCII(cii)  // CII-specific parsing
    default:
        return nil, fmt.Errorf("unknown root element namespace: %s", rootns)
    }
}
```

Format detection already exists. Adding UBL is just adding another case statement.

#### 3. Separated Writing Logic
```go
// writer.go:554
func (inv *Invoice) Write(w io.Writer) error {
    switch inv.SchemaType {
    case UBL:
        return fmt.Errorf("unknown schema UBL %w", ErrUnsupportedSchema)
    case CII, SchemaTypeUnknown:
        return writeCII(inv, w)  // CII-specific writing
    }
}
```

Writer already checks SchemaType. Just need to implement `writeUBL()`.

#### 4. Format-Independent Validation

All validation files (`validate_*.go`) operate on the `Invoice` struct using BT/BG field references:

```go
// validate_core.go
func (inv *Invoice) validateCore() []SemanticError {
    if inv.InvoiceNumber == "" {  // BT-1 check, not XML element check
        errors = append(errors, SemanticError{Rule: rules.BR1, ...})
    }
    // ... all validation works for both formats
}
```

This means:
- ✅ All 203 business rules work for both formats
- ✅ PEPPOL validation works for both formats
- ✅ VAT category validation works for both formats
- ✅ No validation changes needed

### Architecture Summary

**Excellent separation of concerns:**
- Model layer: Format-agnostic ✅
- Parsing layer: Format-specific, isolated
- Writing layer: Format-specific, isolated
- Validation layer: Format-agnostic ✅
- Business rules: Format-agnostic ✅

---

## UBL vs CII Comparison

### XML Structure Differences

#### CII Structure (ZUGFeRD/Factur-X)
```xml
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter>
      <ram:ID>urn:cen.eu:en16931:2017</ram:ID>
    </ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument>
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:SellerTradeParty>
        <ram:Name>Seller Corp</ram:Name>
      </ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>
```

#### UBL Structure
```xml
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:CustomizationID>urn:cen.eu:en16931:2017</cbc:CustomizationID>
  <cbc:ID>INV-001</cbc:ID>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cac:AccountingSupplierParty>
    <cac:Party>
      <cac:PartyName>
        <cbc:Name>Seller Corp</cbc:Name>
      </cac:PartyName>
    </cac:Party>
  </cac:AccountingSupplierParty>
</Invoice>
```

### Key Differences

| Aspect | CII | UBL |
|--------|-----|-----|
| **Namespaces** | rsm, ram, udt, qdt | cac, cbc, base Invoice namespace |
| **Root Element** | `CrossIndustryInvoice` | `Invoice` |
| **Nesting** | Moderate (Germanic naming) | Deeper (more hierarchical) |
| **Profile ID** | `GuidelineSpecifiedDocumentContextParameter` | `CustomizationID` |
| **Seller** | `SellerTradeParty` | `AccountingSupplierParty/Party` |
| **Buyer** | `BuyerTradeParty` | `AccountingCustomerParty/Party` |
| **Line Items** | `IncludedSupplyChainTradeLineItem` | `InvoiceLine` |
| **Tax** | `ApplicableTradeTax` | `TaxTotal/TaxSubtotal` |
| **Verbosity** | Medium | Higher (more elements) |
| **File Size** | Smaller | Larger (20-30% typically) |

### Semantic Equivalence

Despite structural differences, both formats implement the same EN 16931 semantic model:

| EN 16931 Reference | CII XPath | UBL XPath |
|-------------------|-----------|-----------|
| BT-1 Invoice number | `/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:ID` | `/Invoice/cbc:ID` |
| BT-2 Invoice date | `/rsm:ExchangedDocument/ram:IssueDateTime/udt:DateTimeString` | `/Invoice/cbc:IssueDate` |
| BT-5 Invoice currency | `/ram:ApplicableHeaderTradeSettlement/ram:InvoiceCurrencyCode` | `/Invoice/cbc:DocumentCurrencyCode` |
| BG-4 Seller | `/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty` | `/Invoice/cac:AccountingSupplierParty/cac:Party` |
| BG-7 Buyer | `/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty` | `/Invoice/cac:AccountingCustomerParty/cac:Party` |
| BG-25 Invoice lines | `/ram:IncludedSupplyChainTradeLineItem` | `/Invoice/cac:InvoiceLine` |

This 1:1 mapping means the `Invoice` struct can represent both formats without modification.

---

## Implementation Options

### Option 1: Minimal Symmetric Approach

**Description**: Add UBL parsing/writing functions directly to existing parser.go and writer.go files.

**Structure**:
```
parser.go
  ├─ parseCII()    [existing]
  └─ parseUBL()    [new]

writer.go
  ├─ writeCII()    [existing]
  └─ writeUBL()    [new]
```

**Pros**:
- Minimal file changes
- All parsing logic in one file
- Simple to understand

**Cons**:
- Large files (parser.go would be 1000+ lines)
- Mixing two different formats in same file
- Harder to maintain separately
- Git conflicts more likely

**Verdict**: ❌ Not recommended - violates Single Responsibility Principle

### Option 2: Separate File Structure (RECOMMENDED)

**Description**: Create separate files for UBL parsing and writing, keep CII files unchanged.

**Structure**:
```
parser.go        [existing, CII only]
parser_ubl.go    [new, UBL parsing]
writer.go        [existing, CII only]
writer_ubl.go    [new, UBL writing]
model.go         [unchanged]
validate_*.go    [unchanged]
```

**Implementation**:
```go
// parser.go - Modified ParseReader() only
func ParseReader(r io.Reader) (*Invoice, error) {
    ctx, err := cxpath.NewFromReader(r)
    rootns := ctx.Root().Eval("namespace-uri()").String()

    switch rootns {
    case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
        return parseCII(ctx)
    case "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2":
        return parseUBL(ctx)  // Defined in parser_ubl.go
    default:
        return nil, fmt.Errorf("unknown format: %s", rootns)
    }
}

// parser_ubl.go - New file
func parseUBL(ctx *cxpath.Context) (*Invoice, error) {
    inv := &Invoice{SchemaType: UBL}
    // UBL-specific parsing logic
    return inv, nil
}

// writer.go - Modified Write() only
func (inv *Invoice) Write(w io.Writer) error {
    switch inv.SchemaType {
    case CII, SchemaTypeUnknown:
        return writeCII(inv, w)
    case UBL:
        return writeUBL(inv, w)  // Defined in writer_ubl.go
    }
}

// writer_ubl.go - New file
func writeUBL(inv *Invoice, w io.Writer) error {
    // UBL-specific writing logic
    return nil
}
```

**Pros**:
- ✅ Clean separation of concerns
- ✅ Each format has its own file
- ✅ Easy to maintain independently
- ✅ Follows Go conventions (feature-specific files)
- ✅ No breaking changes to existing code
- ✅ Easy to test separately

**Cons**:
- Two more files to manage (minor)
- Need to coordinate ParseReader() modifications

**Verdict**: ✅ **RECOMMENDED** - Best balance of organization and simplicity

### Option 3: Abstract Interface Approach

**Description**: Define Parser and Writer interfaces, implement format-specific versions.

**Structure**:
```go
// parser.go
type Parser interface {
    Parse(ctx *cxpath.Context) (*Invoice, error)
}

type CIIParser struct{}
func (p CIIParser) Parse(ctx *cxpath.Context) (*Invoice, error) { ... }

type UBLParser struct{}
func (p UBLParser) Parse(ctx *cxpath.Context) (*Invoice, error) { ... }

func ParseReader(r io.Reader) (*Invoice, error) {
    ctx := setupXPath(r)
    parser := detectParser(ctx)  // Factory pattern
    return parser.Parse(ctx)
}

// writer.go
type Writer interface {
    Write(inv *Invoice, w io.Writer) error
}

type CIIWriter struct{}
type UBLWriter struct{}
```

**Pros**:
- Most flexible and extensible
- Easy to add more formats (EDIFACT, JSON, etc.)
- Testable with mocks
- Professional architecture

**Cons**:
- Over-engineering for just 2 formats
- More complex than needed
- Significant refactoring required
- Breaking changes for internal APIs
- More boilerplate code

**Verdict**: ⚠️ Future consideration - Good for 3+ formats, overkill for 2

### Option 4: Hybrid Approach

**Description**: Separate files (Option 2) with shared helper functions.

**Structure**:
```
parser.go           [CII parsing]
parser_ubl.go       [UBL parsing]
parser_common.go    [shared helpers]
writer.go           [CII writing]
writer_ubl.go       [UBL writing]
writer_common.go    [shared helpers]
```

**Pros**:
- Reduces code duplication
- Shared utilities (date parsing, decimal conversion, etc.)
- Clean separation

**Cons**:
- More files to manage
- Need to identify truly common functions
- May not have much shared code

**Verdict**: ⚠️ Evaluate during implementation - Add common files only if duplication emerges

---

## Recommended Architecture

### Selected Approach: Option 2 (Separate File Structure)

Based on the analysis, **Option 2** provides the best balance of simplicity, maintainability, and alignment with Go best practices.

### File Structure

```
einvoice/
├── model.go                  [unchanged - format-agnostic]
├── parser.go                 [minimal changes - add UBL case]
├── parser_ubl.go             [new - UBL parsing logic]
├── writer.go                 [minimal changes - add UBL case]
├── writer_ubl.go             [new - UBL writing logic]
├── calculate.go              [unchanged]
├── validation.go             [unchanged]
├── validate_*.go             [unchanged - all validators]
├── rules/                    [unchanged]
├── parser_test.go            [existing CII tests]
├── parser_ubl_test.go        [new - UBL parsing tests]
├── writer_test.go            [existing CII tests]
└── writer_ubl_test.go        [new - UBL writing tests]
```

### Implementation Details

#### 1. Format Detection (parser.go)

```go
func ParseReader(r io.Reader) (*Invoice, error) {
    ctx, err := cxpath.NewFromReader(r)
    if err != nil {
        return nil, fmt.Errorf("cannot read from reader: %w", err)
    }

    rootns := ctx.Root().Eval("namespace-uri()").String()

    var inv *Invoice
    switch rootns {
    case "":
        return nil, fmt.Errorf("empty root element namespace")

    // CII format detection
    case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
        setupCIINamespaces(ctx)
        inv, err = parseCII(ctx)
        if err != nil {
            return nil, err
        }
        inv.SchemaType = CII

    // UBL format detection
    case "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
         "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2":
        setupUBLNamespaces(ctx)
        inv, err = parseUBL(ctx)
        if err != nil {
            return nil, err
        }
        inv.SchemaType = UBL

    default:
        return nil, fmt.Errorf("unknown root element namespace: %s", rootns)
    }

    return inv, nil
}
```

#### 2. UBL Parser Structure (parser_ubl.go)

```go
package einvoice

import (
    "fmt"
    "github.com/shopspring/decimal"
    "github.com/speedata/cxpath"
    "time"
)

// setupUBLNamespaces registers UBL 2.1 namespaces
func setupUBLNamespaces(ctx *cxpath.Context) {
    ctx.SetNamespace("inv", "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2")
    ctx.SetNamespace("cn", "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2")
    ctx.SetNamespace("cac", "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2")
    ctx.SetNamespace("cbc", "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2")
}

// parseUBL parses a UBL 2.1 Invoice or CreditNote
func parseUBL(ctx *cxpath.Context) (*Invoice, error) {
    inv := &Invoice{SchemaType: UBL}

    // Determine document type (Invoice vs CreditNote)
    root := ctx.Root()
    localName := root.Eval("local-name()").String()

    // Set namespace prefix based on document type
    prefix := "inv:"
    if localName == "CreditNote" {
        prefix = "cn:"
    }

    // Parse header fields (BT-1, BT-2, BT-3, BT-5, etc.)
    if err := parseUBLHeader(root, inv, prefix); err != nil {
        return nil, err
    }

    // Parse parties (BG-4, BG-7, BG-10, BG-11, BG-13)
    if err := parseUBLParties(root, inv, prefix); err != nil {
        return nil, err
    }

    // Parse line items (BG-25)
    if err := parseUBLLines(root, inv, prefix); err != nil {
        return nil, err
    }

    // Parse tax breakdown (BG-23)
    if err := parseUBLTaxTotal(root, inv, prefix); err != nil {
        return nil, err
    }

    // Parse monetary totals (BT-106 through BT-115)
    if err := parseUBLMonetarySummation(root, inv, prefix); err != nil {
        return nil, err
    }

    // Parse payment means (BG-16, BG-17, BG-18, BG-19)
    if err := parseUBLPaymentMeans(root, inv, prefix); err != nil {
        return nil, err
    }

    // Parse allowances/charges (BG-20, BG-21)
    if err := parseUBLAllowanceCharge(root, inv, prefix); err != nil {
        return nil, err
    }

    return inv, nil
}

func parseUBLHeader(root *cxpath.Context, inv *Invoice, prefix string) error {
    // BT-24: CustomizationID (maps to GuidelineSpecifiedDocumentContextParameter)
    inv.GuidelineSpecifiedDocumentContextParameter = root.Eval(prefix + "cbc:CustomizationID").String()

    // BT-23: ProfileID
    inv.BPSpecifiedDocumentContextParameter = root.Eval(prefix + "cbc:ProfileID").String()

    // BT-1: Invoice number
    inv.InvoiceNumber = root.Eval(prefix + "cbc:ID").String()

    // BT-3: Invoice type code
    inv.InvoiceTypeCode = CodeDocument(root.Eval(prefix + "cbc:InvoiceTypeCode").Int())

    // BT-2: Invoice date
    dateStr := root.Eval(prefix + "cbc:IssueDate").String()
    if dateStr != "" {
        t, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            return fmt.Errorf("invalid issue date: %w", err)
        }
        inv.InvoiceDate = t
    }

    // BT-5: Invoice currency
    inv.InvoiceCurrencyCode = root.Eval(prefix + "cbc:DocumentCurrencyCode").String()

    // BT-6: Tax currency (optional)
    inv.TaxCurrencyCode = root.Eval(prefix + "cbc:TaxCurrencyCode").String()

    // BT-10: Buyer reference (optional)
    inv.BuyerReference = root.Eval(prefix + "cbc:BuyerReference").String()

    // BT-13: Purchase order reference
    inv.BuyerOrderReferencedDocument = root.Eval(prefix + "cac:OrderReference/cbc:ID").String()

    // BT-12: Contract reference
    inv.ContractReferencedDocument = root.Eval(prefix + "cac:ContractDocumentReference/cbc:ID").String()

    // BG-1: Notes
    for note := range root.Each(prefix + "cbc:Note") {
        inv.Notes = append(inv.Notes, Note{
            Text: note.String(),
            // UBL doesn't have subject code in Note element
        })
    }

    // BG-3: Preceding invoice references
    for ref := range root.Each(prefix + "cac:BillingReference/cac:InvoiceDocumentReference") {
        refDoc := ReferencedDocument{
            ID: ref.Eval("cbc:ID").String(),
        }
        dateStr := ref.Eval("cbc:IssueDate").String()
        if dateStr != "" {
            t, _ := time.Parse("2006-01-02", dateStr)
            refDoc.Date = t
        }
        inv.InvoiceReferencedDocument = append(inv.InvoiceReferencedDocument, refDoc)
    }

    // BG-14: Billing period
    if root.Eval(fmt.Sprintf("count(%scac:InvoicePeriod)", prefix)).Int() > 0 {
        startStr := root.Eval(prefix + "cac:InvoicePeriod/cbc:StartDate").String()
        if startStr != "" {
            t, _ := time.Parse("2006-01-02", startStr)
            inv.BillingSpecifiedPeriodStart = t
        }

        endStr := root.Eval(prefix + "cac:InvoicePeriod/cbc:EndDate").String()
        if endStr != "" {
            t, _ := time.Parse("2006-01-02", endStr)
            inv.BillingSpecifiedPeriodEnd = t
        }
    }

    return nil
}

func parseUBLParty(ctx *cxpath.Context, partyPath string) Party {
    party := Party{}

    // Electronic address (BT-34, BT-49)
    party.URIUniversalCommunication = ctx.Eval(partyPath + "/cbc:EndpointID").String()
    party.URIUniversalCommunicationScheme = ctx.Eval(partyPath + "/cbc:EndpointID/@schemeID").String()

    // Party identification (BT-29, BT-46, BT-60, BT-71)
    for id := range ctx.Each(partyPath + "/cac:PartyIdentification") {
        idValue := id.Eval("cbc:ID").String()
        idScheme := id.Eval("cbc:ID/@schemeID").String()

        if idScheme != "" {
            party.GlobalID = append(party.GlobalID, GlobalID{
                ID:     idValue,
                Scheme: idScheme,
            })
        } else {
            party.ID = append(party.ID, idValue)
        }
    }

    // Party name (BT-27, BT-44, BT-59, BT-70)
    party.Name = ctx.Eval(partyPath + "/cac:PartyName/cbc:Name").String()
    if party.Name == "" {
        // Fallback to PartyLegalEntity/RegistrationName
        party.Name = ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:RegistrationName").String()
    }

    // Postal address (BG-5, BG-8, BG-12, BG-15)
    if ctx.Eval(fmt.Sprintf("count(%s/cac:PostalAddress)", partyPath)).Int() > 0 {
        postalAddr := &PostalAddress{
            Line1:                  ctx.Eval(partyPath + "/cac:PostalAddress/cbc:StreetName").String(),
            Line2:                  ctx.Eval(partyPath + "/cac:PostalAddress/cbc:AdditionalStreetName").String(),
            Line3:                  ctx.Eval(partyPath + "/cac:PostalAddress/cac:AddressLine/cbc:Line").String(),
            City:                   ctx.Eval(partyPath + "/cac:PostalAddress/cbc:CityName").String(),
            PostcodeCode:           ctx.Eval(partyPath + "/cac:PostalAddress/cbc:PostalZone").String(),
            CountrySubDivisionName: ctx.Eval(partyPath + "/cac:PostalAddress/cbc:CountrySubentity").String(),
            CountryID:              ctx.Eval(partyPath + "/cac:PostalAddress/cac:Country/cbc:IdentificationCode").String(),
        }
        party.PostalAddress = postalAddr
    }

    // Legal organization (BT-30, BT-47, BT-61)
    if ctx.Eval(fmt.Sprintf("count(%s/cac:PartyLegalEntity)", partyPath)).Int() > 0 {
        legalOrg := &SpecifiedLegalOrganization{
            ID:                  ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:CompanyID").String(),
            Scheme:              ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:CompanyID/@schemeID").String(),
            TradingBusinessName: ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:RegistrationName").String(),
        }
        party.SpecifiedLegalOrganization = legalOrg
    }

    // Tax registration (BT-31, BT-32, BT-48, BT-63)
    for taxScheme := range ctx.Each(partyPath + "/cac:PartyTaxScheme") {
        taxID := taxScheme.Eval("cbc:CompanyID").String()
        scheme := taxScheme.Eval("cac:TaxScheme/cbc:ID").String()

        if scheme == "VAT" {
            party.VATaxRegistration = taxID
        } else if scheme == "FC" {
            party.FCTaxRegistration = taxID
        }
    }

    // Contact (BG-6, BG-9)
    for contact := range ctx.Each(partyPath + "/cac:Contact") {
        dtc := DefinedTradeContact{
            PersonName:  contact.Eval("cbc:Name").String(),
            PhoneNumber: contact.Eval("cbc:Telephone").String(),
            EMail:       contact.Eval("cbc:ElectronicMail").String(),
        }
        party.DefinedTradeContact = append(party.DefinedTradeContact, dtc)
    }

    return party
}

func parseUBLParties(root *cxpath.Context, inv *Invoice, prefix string) error {
    // BG-4: Seller (AccountingSupplierParty)
    inv.Seller = parseUBLParty(root, prefix+"cac:AccountingSupplierParty/cac:Party")

    // BG-7: Buyer (AccountingCustomerParty)
    inv.Buyer = parseUBLParty(root, prefix+"cac:AccountingCustomerParty/cac:Party")

    // BG-10: Payee (optional)
    if root.Eval(fmt.Sprintf("count(%scac:PayeeParty)", prefix)).Int() > 0 {
        payee := parseUBLParty(root, prefix+"cac:PayeeParty")
        inv.PayeeTradeParty = &payee
    }

    // BG-11: Seller tax representative (optional)
    if root.Eval(fmt.Sprintf("count(%scac:TaxRepresentativeParty)", prefix)).Int() > 0 {
        taxRep := parseUBLParty(root, prefix+"cac:TaxRepresentativeParty")
        inv.SellerTaxRepresentativeTradeParty = &taxRep
    }

    // BG-13: Delivery information (optional)
    if root.Eval(fmt.Sprintf("count(%scac:Delivery)", prefix)).Int() > 0 {
        // BT-72: Actual delivery date
        dateStr := root.Eval(prefix + "cac:Delivery/cbc:ActualDeliveryDate").String()
        if dateStr != "" {
            t, _ := time.Parse("2006-01-02", dateStr)
            inv.OccurrenceDateTime = t
        }

        // Delivery location party
        if root.Eval(fmt.Sprintf("count(%scac:Delivery/cac:DeliveryLocation/cac:Address)", prefix)).Int() > 0 {
            shipTo := parseUBLParty(root, prefix+"cac:Delivery/cac:DeliveryParty")

            // If DeliveryParty is empty, create party with just address
            if shipTo.Name == "" {
                shipTo.Name = root.Eval(prefix + "cac:Delivery/cac:DeliveryParty/cac:PartyName/cbc:Name").String()
            }

            // Get delivery address
            if root.Eval(fmt.Sprintf("count(%scac:Delivery/cac:DeliveryLocation/cac:Address)", prefix)).Int() > 0 {
                postalAddr := &PostalAddress{
                    Line1:                  root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:StreetName").String(),
                    Line2:                  root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:AdditionalStreetName").String(),
                    Line3:                  root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cac:AddressLine/cbc:Line").String(),
                    City:                   root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:CityName").String(),
                    PostcodeCode:           root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:PostalZone").String(),
                    CountrySubDivisionName: root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:CountrySubentity").String(),
                    CountryID:              root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cac:Country/cbc:IdentificationCode").String(),
                }
                shipTo.PostalAddress = postalAddr
            }

            inv.ShipTo = &shipTo
        }
    }

    return nil
}

// Additional parse functions would follow similar patterns...
// parseUBLLines(), parseUBLTaxTotal(), parseUBLMonetarySummation(), etc.
```

#### 3. UBL Writer Structure (writer_ubl.go)

```go
package einvoice

import (
    "fmt"
    "io"
    "github.com/beevik/etree"
    "github.com/shopspring/decimal"
)

// writeUBL generates UBL 2.1 XML from Invoice struct
func writeUBL(inv *Invoice, w io.Writer) error {
    doc := etree.NewDocument()

    // Determine root element (Invoice vs CreditNote)
    rootName := "Invoice"
    if inv.InvoiceTypeCode == 381 || inv.InvoiceTypeCode == 383 {
        rootName = "CreditNote"
    }

    root := doc.CreateElement(rootName)

    // Set namespaces
    root.CreateAttr("xmlns", "urn:oasis:names:specification:ubl:schema:xsd:"+rootName+"-2")
    root.CreateAttr("xmlns:cac", "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2")
    root.CreateAttr("xmlns:cbc", "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2")

    // Write header elements
    writeUBLHeader(root, inv)

    // Write parties
    writeUBLParties(root, inv)

    // Write delivery information
    writeUBLDelivery(root, inv)

    // Write payment means
    writeUBLPaymentMeans(root, inv)

    // Write payment terms
    writeUBLPaymentTerms(root, inv)

    // Write allowances and charges
    writeUBLAllowanceCharge(root, inv)

    // Write tax total
    writeUBLTaxTotal(root, inv)

    // Write monetary totals
    writeUBLMonetarySummation(root, inv)

    // Write invoice lines
    writeUBLLines(root, inv)

    doc.Indent(2)
    if _, err := doc.WriteTo(w); err != nil {
        return fmt.Errorf("write UBL: %w", err)
    }

    return nil
}

func writeUBLHeader(root *etree.Element, inv *Invoice) {
    // BT-24: CustomizationID
    if inv.GuidelineSpecifiedDocumentContextParameter != "" {
        root.CreateElement("cbc:CustomizationID").SetText(inv.GuidelineSpecifiedDocumentContextParameter)
    }

    // BT-23: ProfileID
    if inv.BPSpecifiedDocumentContextParameter != "" {
        root.CreateElement("cbc:ProfileID").SetText(inv.BPSpecifiedDocumentContextParameter)
    }

    // BT-1: Invoice number
    root.CreateElement("cbc:ID").SetText(inv.InvoiceNumber)

    // BT-2: Invoice date
    root.CreateElement("cbc:IssueDate").SetText(inv.InvoiceDate.Format("2006-01-02"))

    // BT-3: Invoice type code
    root.CreateElement("cbc:InvoiceTypeCode").SetText(inv.InvoiceTypeCode.String())

    // BG-1: Notes
    for _, note := range inv.Notes {
        noteElt := root.CreateElement("cbc:Note")
        noteElt.SetText(note.Text)
    }

    // BT-5: Document currency
    root.CreateElement("cbc:DocumentCurrencyCode").SetText(inv.InvoiceCurrencyCode)

    // BT-6: Tax currency (optional)
    if inv.TaxCurrencyCode != "" && inv.TaxCurrencyCode != inv.InvoiceCurrencyCode {
        root.CreateElement("cbc:TaxCurrencyCode").SetText(inv.TaxCurrencyCode)
    }

    // BT-10: Buyer reference
    if inv.BuyerReference != "" {
        root.CreateElement("cbc:BuyerReference").SetText(inv.BuyerReference)
    }

    // BG-14: Invoice period
    if !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero() {
        period := root.CreateElement("cac:InvoicePeriod")
        if !inv.BillingSpecifiedPeriodStart.IsZero() {
            period.CreateElement("cbc:StartDate").SetText(inv.BillingSpecifiedPeriodStart.Format("2006-01-02"))
        }
        if !inv.BillingSpecifiedPeriodEnd.IsZero() {
            period.CreateElement("cbc:EndDate").SetText(inv.BillingSpecifiedPeriodEnd.Format("2006-01-02"))
        }
    }

    // BT-13: Purchase order reference
    if inv.BuyerOrderReferencedDocument != "" {
        orderRef := root.CreateElement("cac:OrderReference")
        orderRef.CreateElement("cbc:ID").SetText(inv.BuyerOrderReferencedDocument)
    }

    // BG-3: Preceding invoice references
    for _, refDoc := range inv.InvoiceReferencedDocument {
        billingRef := root.CreateElement("cac:BillingReference")
        invDocRef := billingRef.CreateElement("cac:InvoiceDocumentReference")
        invDocRef.CreateElement("cbc:ID").SetText(refDoc.ID)
        if !refDoc.Date.IsZero() {
            invDocRef.CreateElement("cbc:IssueDate").SetText(refDoc.Date.Format("2006-01-02"))
        }
    }

    // BT-12: Contract reference
    if inv.ContractReferencedDocument != "" {
        contractRef := root.CreateElement("cac:ContractDocumentReference")
        contractRef.CreateElement("cbc:ID").SetText(inv.ContractReferencedDocument)
    }

    // BG-24: Additional documents
    for _, doc := range inv.AdditionalReferencedDocument {
        addDoc := root.CreateElement("cac:AdditionalDocumentReference")
        addDoc.CreateElement("cbc:ID").SetText(doc.IssuerAssignedID)
        if doc.TypeCode != "" {
            addDoc.CreateElement("cbc:DocumentTypeCode").SetText(doc.TypeCode)
        }
        if doc.Name != "" {
            addDoc.CreateElement("cbc:DocumentDescription").SetText(doc.Name)
        }
        if len(doc.AttachmentBinaryObject) > 0 {
            attach := addDoc.CreateElement("cac:Attachment")
            binary := attach.CreateElement("cbc:EmbeddedDocumentBinaryObject")
            binary.CreateAttr("mimeCode", doc.AttachmentMimeCode)
            binary.CreateAttr("filename", doc.AttachmentFilename)
            // binary.SetText(base64.StdEncoding.EncodeToString(doc.AttachmentBinaryObject))
        }
    }
}

func writeUBLParty(parent *etree.Element, party Party, elementName string) {
    partyElt := parent.CreateElement(elementName)
    partyNode := partyElt.CreateElement("cac:Party")

    // Endpoint ID
    if party.URIUniversalCommunication != "" {
        endpoint := partyNode.CreateElement("cbc:EndpointID")
        endpoint.CreateAttr("schemeID", party.URIUniversalCommunicationScheme)
        endpoint.SetText(party.URIUniversalCommunication)
    }

    // Party identifications
    for _, id := range party.ID {
        partyID := partyNode.CreateElement("cac:PartyIdentification")
        partyID.CreateElement("cbc:ID").SetText(id)
    }

    for _, gid := range party.GlobalID {
        partyID := partyNode.CreateElement("cac:PartyIdentification")
        idElt := partyID.CreateElement("cbc:ID")
        idElt.CreateAttr("schemeID", gid.Scheme)
        idElt.SetText(gid.ID)
    }

    // Party name
    if party.Name != "" {
        partyName := partyNode.CreateElement("cac:PartyName")
        partyName.CreateElement("cbc:Name").SetText(party.Name)
    }

    // Postal address
    if party.PostalAddress != nil {
        addr := partyNode.CreateElement("cac:PostalAddress")

        if party.PostalAddress.Line1 != "" {
            addr.CreateElement("cbc:StreetName").SetText(party.PostalAddress.Line1)
        }
        if party.PostalAddress.Line2 != "" {
            addr.CreateElement("cbc:AdditionalStreetName").SetText(party.PostalAddress.Line2)
        }
        if party.PostalAddress.Line3 != "" {
            addressLine := addr.CreateElement("cac:AddressLine")
            addressLine.CreateElement("cbc:Line").SetText(party.PostalAddress.Line3)
        }
        if party.PostalAddress.City != "" {
            addr.CreateElement("cbc:CityName").SetText(party.PostalAddress.City)
        }
        if party.PostalAddress.PostcodeCode != "" {
            addr.CreateElement("cbc:PostalZone").SetText(party.PostalAddress.PostcodeCode)
        }
        if party.PostalAddress.CountrySubDivisionName != "" {
            addr.CreateElement("cbc:CountrySubentity").SetText(party.PostalAddress.CountrySubDivisionName)
        }
        if party.PostalAddress.CountryID != "" {
            country := addr.CreateElement("cac:Country")
            country.CreateElement("cbc:IdentificationCode").SetText(party.PostalAddress.CountryID)
        }
    }

    // Tax scheme
    if party.VATaxRegistration != "" {
        taxScheme := partyNode.CreateElement("cac:PartyTaxScheme")
        taxScheme.CreateElement("cbc:CompanyID").SetText(party.VATaxRegistration)
        scheme := taxScheme.CreateElement("cac:TaxScheme")
        scheme.CreateElement("cbc:ID").SetText("VAT")
    }

    if party.FCTaxRegistration != "" {
        taxScheme := partyNode.CreateElement("cac:PartyTaxScheme")
        taxScheme.CreateElement("cbc:CompanyID").SetText(party.FCTaxRegistration)
        scheme := taxScheme.CreateElement("cac:TaxScheme")
        scheme.CreateElement("cbc:ID").SetText("FC")
    }

    // Legal entity
    if party.SpecifiedLegalOrganization != nil {
        legalEntity := partyNode.CreateElement("cac:PartyLegalEntity")
        legalEntity.CreateElement("cbc:RegistrationName").SetText(party.Name)

        if party.SpecifiedLegalOrganization.ID != "" {
            companyID := legalEntity.CreateElement("cbc:CompanyID")
            if party.SpecifiedLegalOrganization.Scheme != "" {
                companyID.CreateAttr("schemeID", party.SpecifiedLegalOrganization.Scheme)
            }
            companyID.SetText(party.SpecifiedLegalOrganization.ID)
        }
    }

    // Contacts
    for _, contact := range party.DefinedTradeContact {
        contactElt := partyNode.CreateElement("cac:Contact")
        if contact.PersonName != "" {
            contactElt.CreateElement("cbc:Name").SetText(contact.PersonName)
        }
        if contact.PhoneNumber != "" {
            contactElt.CreateElement("cbc:Telephone").SetText(contact.PhoneNumber)
        }
        if contact.EMail != "" {
            contactElt.CreateElement("cbc:ElectronicMail").SetText(contact.EMail)
        }
    }
}

func writeUBLParties(root *etree.Element, inv *Invoice) {
    // BG-4: Accounting supplier party (Seller)
    writeUBLParty(root, inv.Seller, "cac:AccountingSupplierParty")

    // BG-7: Accounting customer party (Buyer)
    writeUBLParty(root, inv.Buyer, "cac:AccountingCustomerParty")

    // BG-10: Payee party (optional)
    if inv.PayeeTradeParty != nil {
        writeUBLParty(root, *inv.PayeeTradeParty, "cac:PayeeParty")
    }

    // BG-11: Tax representative party (optional)
    if inv.SellerTaxRepresentativeTradeParty != nil {
        writeUBLParty(root, *inv.SellerTaxRepresentativeTradeParty, "cac:TaxRepresentativeParty")
    }
}

// Additional write functions would follow...
// writeUBLDelivery(), writeUBLPaymentMeans(), writeUBLTaxTotal(), etc.
```

### Backward Compatibility

No breaking changes:
- ✅ Existing CII parsing/writing unchanged
- ✅ All existing tests pass
- ✅ Public API unchanged
- ✅ Format auto-detection transparent to users
- ✅ Validation works identically

---

## Technical Specifications

### Supported UBL Elements

#### Core Elements (EN 16931 Mandatory)
- ✅ cbc:CustomizationID (BT-24)
- ✅ cbc:ID (BT-1)
- ✅ cbc:IssueDate (BT-2)
- ✅ cbc:InvoiceTypeCode (BT-3)
- ✅ cbc:DocumentCurrencyCode (BT-5)
- ✅ cac:AccountingSupplierParty (BG-4)
- ✅ cac:AccountingCustomerParty (BG-7)
- ✅ cac:InvoiceLine (BG-25)
- ✅ cac:TaxTotal (BG-23)
- ✅ cac:LegalMonetaryTotal (BT-106 through BT-115)

#### Optional Elements
- ✅ cbc:TaxCurrencyCode (BT-6)
- ✅ cbc:BuyerReference (BT-10)
- ✅ cac:InvoicePeriod (BG-14)
- ✅ cac:OrderReference (BT-13)
- ✅ cac:BillingReference (BG-3)
- ✅ cac:PaymentMeans (BG-16)
- ✅ cac:PaymentTerms (BT-20)
- ✅ cac:AllowanceCharge (BG-20, BG-21)
- ✅ cac:PayeeParty (BG-10)
- ✅ cac:TaxRepresentativeParty (BG-11)
- ✅ cac:Delivery (BG-13)

### UBL Profiles Supported

1. **EN 16931 Core UBL**
   - CustomizationID: `urn:cen.eu:en16931:2017`
   - All EN 16931 business rules apply

2. **PEPPOL BIS Billing 3.0**
   - CustomizationID: `urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0`
   - Additional PEPPOL validation rules apply
   - Existing validate_peppol.go works unchanged

3. **Future: Country-Specific**
   - Netherlands SI-UBL (NLCIUS)
   - Norway EHF
   - Denmark Nemhandel
   - Sweden Svefaktura (phased out, PEPPOL now)

### Date Format Handling

**CII**: Uses format code "102" (YYYYMMDD)
```xml
<udt:DateTimeString format="102">20231231</udt:DateTimeString>
```

**UBL**: Uses ISO 8601 date format (YYYY-MM-DD)
```xml
<cbc:IssueDate>2023-12-31</cbc:IssueDate>
```

**Implementation**:
```go
// Parsing
ciiDate, _ := time.Parse("20060102", "20231231")      // CII
ublDate, _ := time.Parse("2006-01-02", "2023-12-31")  // UBL

// Writing
ciiStr := date.Format("20060102")      // CII
ublStr := date.Format("2006-01-02")    // UBL
```

### Namespace Handling

```go
// CII Namespaces
ctx.SetNamespace("rsm", "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100")
ctx.SetNamespace("ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
ctx.SetNamespace("udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")
ctx.SetNamespace("qdt", "urn:un:unece:uncefact:data:standard:QualifiedDataType:100")

// UBL Namespaces
ctx.SetNamespace("inv", "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2")
ctx.SetNamespace("cn", "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2")
ctx.SetNamespace("cac", "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2")
ctx.SetNamespace("cbc", "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2")
```

### Credit Note Support

Both Invoice and CreditNote documents map to the same `Invoice` struct:

```go
// Detection
localName := root.Eval("local-name()").String()
switch localName {
case "Invoice":
    // Invoice type codes: 380, 384, 389, 751
case "CreditNote":
    // Credit note type codes: 381, 383, 396, 532
}

// Writing
rootName := "Invoice"
if inv.InvoiceTypeCode == 381 || inv.InvoiceTypeCode == 383 {
    rootName = "CreditNote"
}
```

---

## Implementation Phases

### Phase 1: UBL Parser (Read Support)

**Goal**: Parse UBL invoices into Invoice struct

**Tasks**:
1. Create parser_ubl.go file
2. Implement parseUBL() with namespace setup
3. Implement header parsing (BT-1 through BT-24)
4. Implement party parsing (BG-4, BG-7, BG-10, BG-11, BG-13)
5. Implement line item parsing (BG-25)
6. Implement tax parsing (BG-23)
7. Implement monetary summation (BT-106 through BT-115)
8. Implement payment means (BG-16, BG-17, BG-18, BG-19)
9. Implement allowance/charge (BG-20, BG-21)
10. Modify ParseReader() to detect UBL
11. Add parseUBL tests with sample files

**Deliverables**:
- parser_ubl.go (~500-700 lines)
- parser_ubl_test.go (~300-400 lines)
- Test data: 5-10 sample UBL files

**Success Criteria**:
- ✅ Parse valid PEPPOL BIS Billing 3.0 invoice
- ✅ Parse EN 16931 UBL invoice
- ✅ All BT/BG fields correctly mapped
- ✅ Validation works on parsed invoices
- ✅ Profile detection works (isPEPPOL(), IsEN16931())

**Estimated Duration**: 1-2 weeks

### Phase 2: UBL Writer (Write Support)

**Goal**: Generate UBL XML from Invoice struct

**Tasks**:
1. Create writer_ubl.go file
2. Implement writeUBL() with root element setup
3. Implement header writing (BT-1 through BT-24)
4. Implement party writing (BG-4, BG-7, BG-10, BG-11, BG-13)
5. Implement line item writing (BG-25)
6. Implement tax writing (BG-23)
7. Implement monetary summation (BT-106 through BT-115)
8. Implement payment means (BG-16, BG-17, BG-18, BG-19)
9. Implement allowance/charge (BG-20, BG-21)
10. Modify Write() method to call writeUBL()
11. Profile-aware output (only include elements for profile level)
12. Add writeUBL tests

**Deliverables**:
- writer_ubl.go (~600-800 lines)
- writer_ubl_test.go (~400-500 lines)

**Success Criteria**:
- ✅ Generate valid PEPPOL BIS Billing 3.0 XML
- ✅ Generate EN 16931 UBL XML
- ✅ Generated XML validates against UBL 2.1 XSD
- ✅ Profile-aware output works
- ✅ Roundtrip test passes: parse → write → parse → compare

**Estimated Duration**: 1-2 weeks

### Phase 3: Integration & Polish

**Goal**: Complete end-to-end testing and documentation

**Tasks**:
1. Roundtrip testing (CII → UBL, UBL → CII)
2. Format conversion helper functions
3. Real-world sample testing
4. Performance benchmarking
5. Update CLAUDE.md with UBL info
6. Update README.md with UBL examples
7. Create migration guide
8. Add godoc comments
9. Example code for common use cases

**Deliverables**:
- Updated documentation
- Conversion examples
- Benchmark results
- Migration guide

**Success Criteria**:
- ✅ Parse 20+ real-world UBL invoices
- ✅ Generate valid UBL accepted by PEPPOL test tools
- ✅ Format conversion works both directions
- ✅ Performance acceptable (<10% slowdown for CII)
- ✅ Documentation complete

**Estimated Duration**: 3-5 days

### Phase 4: Extended Features (Optional)

**Goal**: Advanced UBL features and country-specific support

**Tasks**:
1. Italian FatturaPA support
2. Netherlands SI-UBL (NLCIUS) support
3. Norwegian EHF support
4. Extended profile support
5. Additional document types (Orders, Reminders, etc.)

**Estimated Duration**: Ongoing, per feature

---

## Testing Strategy

### Unit Tests

#### Parser Tests (parser_ubl_test.go)

```go
func TestParseUBL_MinimalInvoice(t *testing.T) {
    // Test parsing minimal valid UBL invoice
}

func TestParseUBL_PEPPOLBISBilling(t *testing.T) {
    // Test parsing PEPPOL BIS Billing 3.0 invoice
}

func TestParseUBL_EN16931(t *testing.T) {
    // Test parsing EN 16931 UBL invoice
}

func TestParseUBL_AllFields(t *testing.T) {
    // Test parsing invoice with all optional fields
}

func TestParseUBL_CreditNote(t *testing.T) {
    // Test parsing UBL credit note
}

func TestParseUBL_MultipleLines(t *testing.T) {
    // Test parsing invoice with many line items
}

func TestParseUBL_ComplexTax(t *testing.T) {
    // Test parsing with multiple VAT rates
}
```

#### Writer Tests (writer_ubl_test.go)

```go
func TestWriteUBL_MinimalInvoice(t *testing.T) {
    // Test generating minimal valid UBL
}

func TestWriteUBL_PEPPOLCompliant(t *testing.T) {
    // Test generating PEPPOL-compliant UBL
}

func TestWriteUBL_ProfileAware(t *testing.T) {
    // Test that profile level controls output
}

func TestWriteUBL_ValidatesAgainstXSD(t *testing.T) {
    // Test that generated XML validates against UBL 2.1 XSD
}
```

### Integration Tests

```go
func TestRoundtrip_UBL(t *testing.T) {
    // Parse UBL → Write UBL → Parse UBL → Compare
    original := parseTestFile("testdata/sample.xml")

    var buf bytes.Buffer
    original.Write(&buf)

    reparsed, _ := ParseReader(&buf)

    assert.Equal(t, original, reparsed)
}

func TestConversion_CII_to_UBL(t *testing.T) {
    // Parse CII → Change SchemaType → Write UBL → Validate
    ciiInv := parseTestFile("testdata/zugferd.xml")

    ciiInv.SchemaType = UBL
    var buf bytes.Buffer
    ciiInv.Write(&buf)

    ublInv, _ := ParseReader(&buf)

    assert.Equal(t, ciiInv.InvoiceNumber, ublInv.InvoiceNumber)
    // ... compare all critical fields
}

func TestConversion_UBL_to_CII(t *testing.T) {
    // Parse UBL → Change SchemaType → Write CII → Validate
}
```

### Validation Tests

```go
func TestValidation_UBL_PEPPOL(t *testing.T) {
    // Parse PEPPOL UBL → Validate → Check PEPPOL rules
    inv := parseTestFile("testdata/peppol.xml")

    err := inv.Validate()

    assert.NoError(t, err)
    assert.True(t, inv.isPEPPOL())
}

func TestValidation_UBL_EN16931(t *testing.T) {
    // Parse EN 16931 UBL → Validate → Check core rules
}
```

### Test Data Sources

1. **PEPPOL Test Files**
   - https://github.com/OpenPEPPOL/peppol-bis-invoice-3
   - Official PEPPOL BIS Billing 3.0 examples

2. **EN 16931 Test Files**
   - https://github.com/ConnectingEurope/eInvoicing-EN16931
   - Official EN 16931 validation artifacts

3. **UBL Samples**
   - http://docs.oasis-open.org/ubl/os-UBL-2.1/
   - Official UBL 2.1 example files

4. **Real-World Samples**
   - Collect samples from production systems (anonymized)

### Performance Tests

```go
func BenchmarkParseUBL(b *testing.B) {
    data := loadTestFile("testdata/large.xml")
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ParseReader(bytes.NewReader(data))
    }
}

func BenchmarkWriteUBL(b *testing.B) {
    inv := createLargeInvoice()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        var buf bytes.Buffer
        inv.Write(&buf)
    }
}

func BenchmarkParseCII_vs_UBL(b *testing.B) {
    // Compare parsing performance
}
```

---

## Effort Estimation

### Detailed Breakdown

| Phase | Task | Estimated Hours |
|-------|------|----------------|
| **Phase 1: UBL Parser** | | |
| | Project setup & file structure | 2 |
| | Namespace setup & detection | 2 |
| | Header parsing (BT-1 to BT-24) | 6 |
| | Party parsing (BG-4, BG-7) | 6 |
| | Optional parties (BG-10, BG-11, BG-13) | 4 |
| | Line item parsing (BG-25) | 6 |
| | Tax parsing (BG-23) | 4 |
| | Monetary summation | 3 |
| | Payment means & terms | 4 |
| | Allowances & charges | 3 |
| | Testing & debugging | 10 |
| | **Phase 1 Subtotal** | **50 hours** |
| **Phase 2: UBL Writer** | | |
| | File structure & setup | 2 |
| | Root element & namespaces | 2 |
| | Header writing | 5 |
| | Party writing | 6 |
| | Line item writing | 6 |
| | Tax writing | 4 |
| | Monetary summation | 3 |
| | Payment means & terms | 4 |
| | Allowances & charges | 3 |
| | Profile-aware output | 3 |
| | Testing & debugging | 10 |
| | **Phase 2 Subtotal** | **48 hours** |
| **Phase 3: Integration** | | |
| | Roundtrip testing | 6 |
| | Format conversion helpers | 4 |
| | Real-world sample testing | 4 |
| | Performance benchmarking | 2 |
| | Documentation updates | 4 |
| | Example code | 3 |
| | Code review & cleanup | 3 |
| | **Phase 3 Subtotal** | **26 hours** |
| **Total Estimated Effort** | | **124 hours** |

### Timeline Estimate

**Single Developer**:
- Phase 1: 1.5-2 weeks (50 hours)
- Phase 2: 1-1.5 weeks (48 hours)
- Phase 3: 3-4 days (26 hours)
- **Total: 3-4 weeks** (at 30-40 hours/week)

**With Code Review**:
- Add 20% for review cycles and revisions
- **Total: 4-5 weeks**

---

## Risks and Mitigations

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **UBL complexity underestimated** | High | Medium | Start with Phase 1, reassess after parser complete |
| **XPath performance issues** | Medium | Low | Benchmark early, optimize hot paths if needed |
| **Namespace handling edge cases** | Medium | Medium | Extensive testing with varied real-world files |
| **Date format conversions** | Low | Low | Thorough testing, use standard library |
| **Breaking changes to existing code** | High | Very Low | Comprehensive regression testing |

### Process Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Scope creep** | Medium | Medium | Stick to EN 16931 + PEPPOL initially, defer country variants |
| **Insufficient test data** | Medium | Low | Source test files early from official repositories |
| **Documentation lag** | Low | Medium | Write docs alongside code, not after |

### Business Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Low adoption** | Medium | Low | UBL is widely needed in PEPPOL network |
| **Maintenance burden** | Medium | Medium | Good test coverage, clear architecture |

---

## Benefits and Use Cases

### Key Benefits

1. **PEPPOL Network Support**
   - Access to pan-European e-invoicing network
   - Mandatory for B2G in many countries
   - 400,000+ organizations use PEPPOL

2. **Broader Market Coverage**
   - ZUGFeRD/CII: Germany, France, Austria
   - UBL: Netherlands, Norway, Denmark, Sweden, Italy, Spain
   - Together: Complete European coverage

3. **Format Conversion**
   - Convert between CII and UBL programmatically
   - Bridge between different standards
   - Useful for ERP systems and integrators

4. **Single Library**
   - One dependency for all EN 16931 needs
   - Consistent API across formats
   - Unified validation

5. **Future-Proof**
   - EN 16931 is the EU standard
   - Both syntaxes officially supported
   - Foundation for future formats

### Use Cases

#### Use Case 1: ERP System Supporting Multiple Markets

**Scenario**: German ERP vendor wants to expand to Netherlands

**Current**: Supports ZUGFeRD (CII) only

**With UBL Support**:
```go
// Receive invoice from Dutch supplier (UBL)
inv, _ := einvoice.ParseXMLFile("dutch_supplier.xml")  // Auto-detects UBL

// Validate (works for both formats)
if err := inv.Validate(); err != nil {
    log.Fatal(err)
}

// Process invoice...

// Send invoice to German customer (CII)
inv.SchemaType = einvoice.CII
inv.Write(germanCustomerFile)
```

**Result**: Support both markets with single integration

#### Use Case 2: PEPPOL Access Point

**Scenario**: Service provider offers PEPPOL access point

**Requirement**: Must receive/send UBL invoices per PEPPOL spec

**With UBL Support**:
```go
// Receive PEPPOL invoice
inv, _ := einvoice.ParseReader(peppolMessage)

// Validate PEPPOL rules
if err := inv.Validate(); err != nil {
    return peppolError(err)
}

// Convert to internal format or forward as UBL
```

**Result**: Complete PEPPOL compliance

#### Use Case 3: Government Portal

**Scenario**: Government portal receives invoices from vendors

**Requirement**: Accept both CII and UBL formats

**With UBL Support**:
```go
// Unified intake regardless of format
inv, err := einvoice.ParseReader(uploadedFile)
if err != nil {
    return fmt.Errorf("invalid invoice format")
}

// Same validation for all
if err := inv.Validate(); err != nil {
    return fmt.Errorf("invoice validation failed: %w", err)
}

// Process...
storeInvoice(inv)
```

**Result**: Flexible vendor onboarding

#### Use Case 4: Format Migration

**Scenario**: Company migrating from CII to UBL

**Challenge**: Historical data in CII, new system uses UBL

**With UBL Support**:
```go
// Batch conversion utility
func convertArchive() {
    files := findCIIInvoices("archive/")

    for _, file := range files {
        inv, _ := einvoice.ParseXMLFile(file)
        inv.SchemaType = einvoice.UBL

        outFile := strings.Replace(file, ".cii.xml", ".ubl.xml", 1)
        w, _ := os.Create(outFile)
        inv.Write(w)
    }
}
```

**Result**: Seamless migration

---

## Resources Required

### Development Resources

**No New Dependencies**:
- ✅ github.com/speedata/cxpath (already used)
- ✅ github.com/beevik/etree (already used)
- ✅ github.com/shopspring/decimal (already used)

### Documentation Resources

**Specifications**:
- [EN 16931-1:2017](https://standards.cen.eu/dyn/www/f?p=204:110:0::::FSP_PROJECT:60602&cs=1B0F862919A7304F13AE9F46A5A5A1D73) - European Standard
- [UBL 2.1 Specification](http://docs.oasis-open.org/ubl/os-UBL-2.1/) - OASIS
- [PEPPOL BIS Billing 3.0](https://docs.peppol.eu/poacc/billing/3.0/) - PEPPOL

**Syntax Mappings**:
- [EN 16931 UBL Syntax Mapping](https://standards.cen.eu/dyn/www/f?p=204:110:0::::FSP_PROJECT,FSP_ORG_ID:60602,481830&cs=1920F04222A2D6D6A05F21A0BBC6F8850) - Official mapping
- [CII to UBL Comparison](https://ec.europa.eu/cefdigital/wiki/display/CEFDIGITAL/Conformance+testing) - EU Digital

### Test Data Sources

**Official Test Suites**:
1. PEPPOL: https://github.com/OpenPEPPOL/peppol-bis-invoice-3
2. EN 16931: https://github.com/ConnectingEurope/eInvoicing-EN16931
3. UBL Examples: http://docs.oasis-open.org/ubl/os-UBL-2.1/

**Validation Tools**:
1. PEPPOL Validator: https://peppol-test-validator.com/
2. UBL XSD Schemas: Download from OASIS
3. Schematron Rules: From EN 16931 repository

### Community Resources

**Existing Go Projects**:
- Limited mature UBL libraries in Go
- Opportunity to lead in this space

**Reference Implementations**:
- Mustangproject (Java): https://github.com/ZUGFeRD/mustangproject
- Python libraries: factur-x, peppol-validator

---

## Appendix

### A. Sample XPath Mappings

#### Invoice Number (BT-1)

```xpath
# CII
/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:ID

# UBL
/Invoice/cbc:ID
```

#### Seller Name (BT-27)

```xpath
# CII
/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/
  ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:Name

# UBL
/Invoice/cac:AccountingSupplierParty/cac:Party/cac:PartyName/cbc:Name
```

#### Line Item Net Amount (BT-131)

```xpath
# CII
/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/
  ram:IncludedSupplyChainTradeLineItem/
  ram:SpecifiedLineTradeSettlement/
  ram:SpecifiedTradeSettlementLineMonetarySummation/
  ram:LineTotalAmount

# UBL
/Invoice/cac:InvoiceLine/cbc:LineExtensionAmount
```

### B. Profile URN Reference

#### CII Profiles

| Profile | URN |
|---------|-----|
| Minimum | `urn:factur-x.eu:1p0:minimum` |
| Basic WL | `urn:factur-x.eu:1p0:basicwl` |
| Basic | `urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic` |
| EN 16931 | `urn:cen.eu:en16931:2017` |
| Extended | `urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended` |
| XRechnung 3.0 | `urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0` |
| PEPPOL BIS | `urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0` |

#### UBL Profiles

| Profile | CustomizationID |
|---------|----------------|
| EN 16931 | `urn:cen.eu:en16931:2017` |
| PEPPOL BIS | `urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0` |
| SI-UBL (NL) | `urn:cen.eu:en16931:2017#compliant#urn:fdc:nen.nl:nlcius:v1.0` |

### C. Code Examples

#### Example 1: Parse Any Format

```go
package main

import (
    "fmt"
    "github.com/speedata/einvoice"
)

func main() {
    // Auto-detects format (CII or UBL)
    inv, err := einvoice.ParseXMLFile("invoice.xml")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Format: %s\n", inv.SchemaType)
    fmt.Printf("Invoice Number: %s\n", inv.InvoiceNumber)
    fmt.Printf("Seller: %s\n", inv.Seller.Name)
    fmt.Printf("Total: %s %s\n",
        inv.DuePayableAmount.StringFixed(2),
        inv.InvoiceCurrencyCode)
}
```

#### Example 2: Convert CII to UBL

```go
package main

import (
    "os"
    "github.com/speedata/einvoice"
)

func main() {
    // Parse CII invoice
    inv, _ := einvoice.ParseXMLFile("zugferd.xml")

    // Change format to UBL
    inv.SchemaType = einvoice.UBL

    // Write as UBL
    out, _ := os.Create("converted.xml")
    defer out.Close()

    inv.Write(out)
}
```

#### Example 3: Create and Write UBL

```go
package main

import (
    "os"
    "time"
    "github.com/speedata/einvoice"
    "github.com/shopspring/decimal"
)

func main() {
    inv := &einvoice.Invoice{
        SchemaType:      einvoice.UBL,
        GuidelineSpecifiedDocumentContextParameter: einvoice.SpecPEPPOLBISBilling,
        InvoiceNumber:   "INV-2023-001",
        InvoiceTypeCode: 380,
        InvoiceDate:     time.Now(),
        InvoiceCurrencyCode: "EUR",

        Seller: einvoice.Party{
            Name: "My Company",
            PostalAddress: &einvoice.PostalAddress{
                Line1:     "123 Main St",
                City:      "Amsterdam",
                PostcodeCode: "1000 AA",
                CountryID: "NL",
            },
            VATaxRegistration: "NL123456789B01",
        },

        Buyer: einvoice.Party{
            Name: "Customer Corp",
            PostalAddress: &einvoice.PostalAddress{
                Line1:     "456 Customer Ave",
                City:      "Rotterdam",
                PostcodeCode: "3000 AA",
                CountryID: "NL",
            },
        },

        InvoiceLines: []einvoice.InvoiceLine{
            {
                LineID:            "1",
                ItemName:          "Consulting Services",
                BilledQuantity:    decimal.NewFromInt(10),
                BilledQuantityUnit: "HUR",
                NetPrice:          decimal.NewFromInt(100),
                Total:             decimal.NewFromInt(1000),
                TaxCategoryCode:   "S",
                TaxRateApplicablePercent: decimal.NewFromInt(21),
            },
        },
    }

    // Calculate totals
    inv.UpdateApplicableTradeTax("")
    inv.UpdateTotals()

    // Validate
    if err := inv.Validate(); err != nil {
        panic(err)
    }

    // Write UBL
    out, _ := os.Create("invoice.xml")
    defer out.Close()

    inv.Write(out)
}
```

### D. Glossary

- **BG**: Business Group - Collection of related business terms in EN 16931
- **BT**: Business Term - Individual data element in EN 16931
- **CII**: Cross Industry Invoice - UN/CEFACT standard used by ZUGFeRD/Factur-X
- **EN 16931**: European Standard for electronic invoicing semantic data model
- **PEPPOL**: Pan-European Public Procurement Online - e-procurement network
- **UBL**: Universal Business Language - OASIS standard XML format
- **URN**: Uniform Resource Name - Used for specification identifiers
- **XPath**: XML Path Language - Query language for XML
- **ZUGFeRD**: Zentraler User Guide des Forums elektronische Rechnung Deutschland

---

## Conclusion

Adding UBL support to the einvoice library is:

✅ **Technically Feasible** - Current architecture is well-suited
✅ **Architecturally Sound** - Clean separation, no breaking changes
✅ **High Value** - Enables PEPPOL and broader EU market
✅ **Reasonable Effort** - 3-4 weeks focused development
✅ **Well-Defined** - Clear phases and deliverables
✅ **Low Risk** - Comprehensive testing strategy

**Recommendation**: Proceed with Option 2 (Separate File Structure) implementation approach, starting with Phase 1 (UBL Parser).

The library would become the most comprehensive Go solution for EN 16931 electronic invoicing, supporting both major syntaxes with a unified, validated, and well-tested codebase.

---

**Document End**
