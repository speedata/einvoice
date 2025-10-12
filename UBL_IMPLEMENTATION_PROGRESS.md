# UBL Implementation Progress Tracker

**Issue**: #74 - Add UBL Support to Enable PEPPOL and Broader European Market Coverage
**Started**: 2025-10-12
**Target**: Phase 1 Parser Implementation

---

## Implementation Status Overview

| Phase | Status | Started | Completed | Notes |
|-------|--------|---------|-----------|-------|
| **Phase 1: UBL Parser** | 🚧 In Progress | 2025-10-12 | - | Starting implementation |
| Phase 2: UBL Writer | ⏸️ Not Started | - | - | Awaiting Phase 1 completion |
| Phase 3: Integration | ⏸️ Not Started | - | - | Awaiting Phase 2 completion |

---

## Phase 1: UBL Parser - Detailed Progress

**Goal**: Parse UBL 2.1 invoices into `Invoice` struct
**Estimated Effort**: 50 hours (~1-2 weeks)
**Status**: 🚧 In Progress

### Task Checklist

#### 1. Setup & Infrastructure
- [ ] Create `parser_ubl.go` file
- [ ] Add UBL namespace constants
- [ ] Implement `setupUBLNamespaces()` function
- [ ] Create helper function `getDecimalUBL()` (if needed beyond shared)
- [ ] Create helper function `parseTimeUBL()` (if needed for ISO dates)
- [ ] Document file structure and approach

**Status**: ⏸️ Not Started
**Estimated**: 2-3 hours
**Actual**: -

#### 2. Format Detection & Routing
- [ ] Modify `ParseReader()` in `parser.go` to detect UBL namespace
- [ ] Add case for `Invoice` namespace
- [ ] Add case for `CreditNote` namespace
- [ ] Test format detection with sample files

**Status**: ⏸️ Not Started
**Estimated**: 1-2 hours
**Actual**: -

#### 3. Header Parsing (BT-1 to BT-24)
- [ ] Implement `parseUBLHeader()` function
- [ ] BT-24: CustomizationID → GuidelineSpecifiedDocumentContextParameter
- [ ] BT-23: ProfileID → BPSpecifiedDocumentContextParameter
- [ ] BT-1: Invoice number → InvoiceNumber
- [ ] BT-2: Issue date → InvoiceDate
- [ ] BT-3: Invoice type code → InvoiceTypeCode
- [ ] BT-5: Document currency → InvoiceCurrencyCode
- [ ] BT-6: Tax currency → TaxCurrencyCode (optional)
- [ ] BT-10: Buyer reference → BuyerReference (optional)
- [ ] BT-13: Purchase order reference → BuyerOrderReferencedDocument
- [ ] BT-12: Contract reference → ContractReferencedDocument
- [ ] BG-1: Notes → Notes[]
- [ ] BG-3: Preceding invoices → InvoiceReferencedDocument[]
- [ ] BG-14: Invoice period → BillingSpecifiedPeriod (Start/End)
- [ ] Unit tests for header parsing

**Status**: ⏸️ Not Started
**Estimated**: 6 hours
**Actual**: -

#### 4. Party Parsing (BG-4, BG-7, BG-10, BG-11, BG-13)
- [ ] Implement `parseUBLParty()` helper function
- [ ] Parse endpoint ID (electronic address)
- [ ] Parse party identification (ID, GlobalID)
- [ ] Parse party name
- [ ] Parse postal address (street, city, postcode, country)
- [ ] Parse tax registration (VAT, FC)
- [ ] Parse legal entity information
- [ ] Parse contact information
- [ ] BG-4: AccountingSupplierParty → Seller
- [ ] BG-7: AccountingCustomerParty → Buyer
- [ ] BG-10: PayeeParty → PayeeTradeParty (optional)
- [ ] BG-11: TaxRepresentativeParty → SellerTaxRepresentativeTradeParty (optional)
- [ ] BG-13: Delivery → ShipTo + OccurrenceDateTime (optional)
- [ ] Unit tests for party parsing

**Status**: ⏸️ Not Started
**Estimated**: 8 hours
**Actual**: -

#### 5. Line Item Parsing (BG-25)
- [ ] Implement `parseUBLLines()` function
- [ ] BT-126: Line ID → LineID
- [ ] BT-153: Item name → ItemName
- [ ] BT-154: Item description → Description (optional)
- [ ] BT-155: Seller item ID → ArticleNumber
- [ ] BT-156: Buyer item ID → ArticleNumberBuyer
- [ ] BT-157: Standard item ID → GlobalID + GlobalIDType
- [ ] BT-158: Item classification → ProductClassification[]
- [ ] BT-159: Country of origin → OriginTradeCountry
- [ ] BG-32: Item attributes → Characteristics[]
- [ ] BT-129: Invoiced quantity → BilledQuantity
- [ ] BT-130: Quantity unit → BilledQuantityUnit
- [ ] BT-146: Item net price → NetPrice
- [ ] BT-148: Item gross price → GrossPrice (optional)
- [ ] BT-149: Price base quantity → BasisQuantity (optional)
- [ ] BG-27: Line allowances → InvoiceLineAllowances[]
- [ ] BG-28: Line charges → InvoiceLineCharges[]
- [ ] BT-151: Line VAT category → TaxCategoryCode + TaxTypeCode
- [ ] BT-152: Line VAT rate → TaxRateApplicablePercent
- [ ] BT-131: Line net amount → Total
- [ ] BG-26: Line period → BillingSpecifiedPeriod (optional)
- [ ] Track XML element presence (BR-24, BR-26 validation)
- [ ] Unit tests for line item parsing

**Status**: ⏸️ Not Started
**Estimated**: 8 hours
**Actual**: -

#### 6. Tax Parsing (BG-23)
- [ ] Implement `parseUBLTaxTotal()` function
- [ ] Parse TaxTotal/TaxSubtotal elements
- [ ] BT-116: VAT category taxable amount → BasisAmount
- [ ] BT-117: VAT category tax amount → CalculatedAmount
- [ ] BT-118: VAT category code → CategoryCode
- [ ] BT-118-0: Tax type → Typ (should be "VAT")
- [ ] BT-119: VAT category rate → Percent
- [ ] BT-120: VAT exemption reason text → ExemptionReason (optional)
- [ ] BT-121: VAT exemption reason code → ExemptionReasonCode (optional)
- [ ] Aggregate multiple TaxSubtotal → TradeTaxes[]
- [ ] BT-110: Total tax amount → TaxTotal
- [ ] BT-111: Tax total in accounting currency → TaxTotalVAT (optional)
- [ ] Unit tests for tax parsing

**Status**: ⏸️ Not Started
**Estimated**: 5 hours
**Actual**: -

#### 7. Monetary Summation (BT-106 to BT-115)
- [ ] Implement `parseUBLMonetarySummation()` function
- [ ] BT-106: Sum of line amounts → LineTotal
- [ ] BT-107: Sum of allowances → AllowanceTotal
- [ ] BT-108: Sum of charges → ChargeTotal
- [ ] BT-109: Tax basis amount → TaxBasisTotal
- [ ] BT-110: Tax total → TaxTotal (verify consistency)
- [ ] BT-112: Invoice total with VAT → GrandTotal
- [ ] BT-113: Paid amount → TotalPrepaid
- [ ] BT-114: Rounding amount → RoundingAmount (optional)
- [ ] BT-115: Amount due → DuePayableAmount
- [ ] Track XML element presence (BR-12 to BR-15 validation)
- [ ] Unit tests for monetary summation

**Status**: ⏸️ Not Started
**Estimated**: 3 hours
**Actual**: -

#### 8. Payment Means & Terms (BG-16, BG-17, BG-18, BG-19)
- [ ] Implement `parseUBLPaymentMeans()` function
- [ ] BT-81: Payment means code → TypeCode
- [ ] BT-82: Payment means text → Information (optional)
- [ ] BT-83: Remittance information → PaymentReference (optional)
- [ ] BG-17: Credit transfer → PayeePartyCreditorFinancialAccount*
- [ ] BT-84: Account ID (IBAN/other) → PayeePartyCreditorFinancialAccountIBAN
- [ ] BT-85: Account name → PayeePartyCreditorFinancialAccountName (optional)
- [ ] BT-86: Bank ID (BIC) → PayeeSpecifiedCreditorFinancialInstitutionBIC (optional)
- [ ] BG-18: Payment card → ApplicableTradeSettlementFinancialCard*
- [ ] BT-87: Card PAN → ApplicableTradeSettlementFinancialCardID
- [ ] BT-88: Cardholder name → ApplicableTradeSettlementFinancialCardCardholderName (optional)
- [ ] BG-19: Direct debit → PayerPartyDebtorFinancialAccount*
- [ ] BT-91: Debited account ID → PayerPartyDebtorFinancialAccountIBAN
- [ ] Handle multiple payment means
- [ ] Unit tests for payment means

**Status**: ⏸️ Not Started
**Estimated**: 4 hours
**Actual**: -

#### 9. Payment Terms (BT-20, BT-9)
- [ ] Implement `parseUBLPaymentTerms()` function
- [ ] BT-20: Payment terms text → SpecifiedTradePaymentTerms[].Description
- [ ] BT-9: Payment due date → SpecifiedTradePaymentTerms[].DueDate
- [ ] BT-89: Direct debit mandate ID → SpecifiedTradePaymentTerms[].DirectDebitMandateID (if present)
- [ ] Handle multiple payment terms
- [ ] Unit tests for payment terms

**Status**: ⏸️ Not Started
**Estimated**: 2 hours
**Actual**: -

#### 10. Allowances & Charges (BG-20, BG-21)
- [ ] Implement `parseUBLAllowanceCharge()` function
- [ ] Distinguish allowances (BG-20) vs charges (BG-21) by ChargeIndicator
- [ ] BT-92/BT-99: Amount → ActualAmount
- [ ] BT-93/BT-100: Base amount → BasisAmount (optional)
- [ ] BT-94/BT-101: Percentage → CalculationPercent (optional)
- [ ] BT-95/BT-102: VAT category → CategoryTradeTaxCategoryCode
- [ ] BT-96/BT-103: VAT rate → CategoryTradeTaxRateApplicablePercent
- [ ] BT-97/BT-104: Reason text → Reason (optional)
- [ ] BT-98/BT-105: Reason code → ReasonCode (optional)
- [ ] Aggregate to SpecifiedTradeAllowanceCharge[]
- [ ] Unit tests for allowances/charges

**Status**: ⏸️ Not Started
**Estimated**: 3 hours
**Actual**: -

#### 11. Additional Documents (BG-24)
- [ ] Implement parsing for AdditionalDocumentReference
- [ ] BT-122: Supporting document ID → IssuerAssignedID
- [ ] BT-123: Document description → Name
- [ ] BT-124: External document URL → URIID (optional)
- [ ] BT-125: Attached document → AttachmentBinaryObject + metadata
- [ ] Document type code → TypeCode
- [ ] Handle multiple additional documents
- [ ] Unit tests for additional documents

**Status**: ⏸️ Not Started
**Estimated**: 2 hours
**Actual**: -

#### 12. Integration & Main Parser
- [ ] Implement main `parseUBL()` function
- [ ] Detect document type (Invoice vs CreditNote)
- [ ] Call all sub-parsers in correct order
- [ ] Set SchemaType = UBL
- [ ] Handle parsing errors gracefully
- [ ] Return complete Invoice struct
- [ ] Integration tests with complete invoices

**Status**: ⏸️ Not Started
**Estimated**: 3 hours
**Actual**: -

#### 13. Testing & Validation
- [ ] Create test data directory with UBL samples
- [ ] Test: Minimal valid UBL invoice
- [ ] Test: PEPPOL BIS Billing 3.0 invoice
- [ ] Test: EN 16931 UBL invoice
- [ ] Test: UBL credit note
- [ ] Test: Invoice with all optional fields
- [ ] Test: Multiple line items, multiple tax rates
- [ ] Test: All payment means types
- [ ] Test: Profile detection (isPEPPOL(), IsEN16931())
- [ ] Test: Validation works on parsed UBL
- [ ] Test: Edge cases (missing optionals, empty values)
- [ ] Test: Error handling (malformed XML, wrong namespace)
- [ ] Download and test real PEPPOL samples
- [ ] Compare with CII parsing for same invoice semantically
- [ ] Performance benchmarking vs CII parsing

**Status**: ⏸️ Not Started
**Estimated**: 10 hours
**Actual**: -

---

## Code Quality Checklist

- [ ] All functions have clear godoc comments
- [ ] XPath expressions are clear and maintainable
- [ ] Error handling is consistent with CII parser
- [ ] No code duplication (use helpers where appropriate)
- [ ] Follow existing code style and conventions
- [ ] All exported functions documented
- [ ] Edge cases handled gracefully
- [ ] Test coverage > 80%

---

## Testing Resources

### Test Data Sources
- [ ] Downloaded: PEPPOL BIS Billing 3.0 samples
- [ ] Downloaded: EN 16931 UBL test files
- [ ] Downloaded: UBL 2.1 example files
- [ ] Created: Minimal test invoice
- [ ] Created: Maximum test invoice (all fields)

### Test Files Location
```
testdata/
├── ubl/
│   ├── minimal.xml              [Basic valid UBL]
│   ├── peppol_bis_billing.xml   [PEPPOL sample]
│   ├── en16931.xml              [EN 16931 sample]
│   ├── creditnote.xml           [Credit note]
│   ├── maximal.xml              [All optional fields]
│   └── invalid/                 [Invalid samples for error testing]
```

---

## Issues & Blockers

| Date | Issue | Impact | Status | Resolution |
|------|-------|--------|--------|------------|
| - | - | - | - | - |

---

## Notes & Learnings

### 2025-10-12 - Initial Setup
- Created progress tracking document
- Beginning parser implementation
- Following Option 2 architecture (separate files)
- Target: Clean, maintainable code mirroring CII parser style

---

## XPath Quick Reference

### UBL Namespaces
```go
inv: urn:oasis:names:specification:ubl:schema:xsd:Invoice-2
cn:  urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2
cac: urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2
cbc: urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2
```

### Common XPath Patterns
```xpath
# Invoice number
/Invoice/cbc:ID

# Seller party
/Invoice/cac:AccountingSupplierParty/cac:Party

# Line items
/Invoice/cac:InvoiceLine

# Tax subtotals
/Invoice/cac:TaxTotal/cac:TaxSubtotal
```

---

## Phase 1 Completion Criteria

✅ All checklist items above completed
✅ All tests passing
✅ Code reviewed and cleaned up
✅ Documentation updated
✅ No regressions in existing CII functionality
✅ Performance acceptable (<20% slower than CII)

**Sign-off**: _______________ Date: _______________

---

## Next Steps After Phase 1

1. Review Phase 1 implementation
2. Address any feedback
3. Begin Phase 2: UBL Writer
4. Update this document with Phase 2 progress

---

**Last Updated**: 2025-10-12
**Status**: Phase 1 Started
