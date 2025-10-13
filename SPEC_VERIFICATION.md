# EN 16931 Specification Verification for PR #100

This document verifies that the fixes in PR #100 are compliant with the EN 16931 European standard for electronic invoicing and its CII (Cross Industry Invoice) syntax binding.

## Summary

✅ **All fixes are spec-compliant** - The implementations correctly follow the EN 16931-1:2017 semantic model and EN 16931-3-3:2017 CII syntax binding.

## Field-by-Field Verification

### 1. BT-16: Despatch Advice Reference ✅

**EN 16931 Definition:**
- **Business Term**: BT-16 "Despatch advice reference"
- **Description**: An identifier of a referenced despatch advice
- **Cardinality**: 0..1 (optional, at most once)
- **Profile Level**: BASIC WL and above
- **Purpose**: Supplier's reference to a despatch advice that may be used to reconcile the invoice with goods despatched and received

**CII Syntax Mapping:**
```xml
rsm:CrossIndustryInvoice/
  rsm:SupplyChainTradeTransaction/
    ram:ApplicableHeaderTradeDelivery/
      ram:DespatchAdviceReferencedDocument/
        ram:IssuerAssignedID
```

**Fix Applied:**
- ✅ Parser: Changed from `ram:DespatchAdviceReferencedDocument` (reads entire element) to `ram:DespatchAdviceReferencedDocument/ram:IssuerAssignedID` (reads ID child)
- ✅ Writer: Moved output from `ApplicableHeaderTradeAgreement` to correct location `ApplicableHeaderTradeDelivery`
- ✅ Creates proper XML structure: `<ram:DespatchAdviceReferencedDocument><ram:IssuerAssignedID>value</ram:IssuerAssignedID></ram:DespatchAdviceReferencedDocument>`

**Verification Source**: ConnectingEurope/eInvoicing-EN16931 GitHub repository, ZUGFeRD 2.2 Implementation Guide

---

### 2. BT-15: Receiving Advice Reference ✅

**EN 16931 Definition:**
- **Business Term**: BT-15 "Receiving advice reference"
- **Description**: An identifier of a referenced receiving advice
- **Cardinality**: 0..1 (optional, at most once)
- **Profile Level**: EN 16931 and above
- **Maximum Length**: 250 characters (per CIUS-ES-FACE)
- **Purpose**: Links invoice to goods receipt documentation

**CII Syntax Mapping:**
```xml
rsm:CrossIndustryInvoice/
  rsm:SupplyChainTradeTransaction/
    ram:ApplicableHeaderTradeDelivery/
      ram:ReceivingAdviceReferencedDocument/
        ram:IssuerAssignedID
```

**UBL Syntax Mapping (for reference):**
```xml
/cac:ReceiptDocumentReference/cbc:ID
```

**Fix Applied:**
- ✅ Parser: Added parsing with XPath `ram:ReceivingAdviceReferencedDocument/ram:IssuerAssignedID`
- ✅ Writer: Added output in `ApplicableHeaderTradeDelivery` section
- ✅ Creates proper XML structure parallel to BT-16

**Verification Source**: EN 16931 semantic model, CIUS-ES-FACE specification, Peppol BIS Billing 3.0

---

### 3. BT-83: Remittance Information (Payment Reference) ✅

**EN 16931 Definition:**
- **Business Term**: BT-83 "Remittance information"
- **Description**: A textual value used to establish a link between the payment and the Invoice, issued by the Seller
- **Cardinality**: 0..1 (optional, at most once)
- **Profile Level**: BASIC WL and above
- **German Term**: "Verwendungszweck" (remittance purpose)
- **Purpose**: Specifies desired reference information for payment transfer if different from invoice number

**CII Syntax Mapping:**
```xml
rsm:CrossIndustryInvoice/
  rsm:SupplyChainTradeTransaction/
    ram:ApplicableHeaderTradeSettlement/
      ram:PaymentReference
```

**Fix Applied:**
- ✅ Parser: Added parsing with XPath `ram:PaymentReference`
- ✅ Writer: Added output in `ApplicableHeaderTradeSettlement` section
- ✅ Correct location within payment settlement information

**Verification Source**: ZUGFeRD 2.2/2.3.2 Technical Specification, Factur-X 1.07.2, ZUGFeRD Implementation Guide

---

### 4. BT-89: Direct Debit Mandate Reference Identifier ✅

**EN 16931 Definition:**
- **Business Term**: BT-89 "Mandate reference identifier"
- **Description**: Unique identifier assigned by the Payee for referencing the direct debit authorization
- **German Term**: "Mandatsreferenznummer"
- **Cardinality**: 0..1 (optional, at most once)
- **Profile Level**: BASIC WL and above
- **Purpose**: Required for SEPA direct debit processing when direct debit is specified as payment method

**CII Syntax Mapping:**
```xml
rsm:CrossIndustryInvoice/
  rsm:SupplyChainTradeTransaction/
    ram:ApplicableHeaderTradeSettlement/
      ram:SpecifiedTradePaymentTerms/
        ram:DirectDebitMandateID
```

**Related Fields:**
- BT-90: Bank assigned creditor identifier (PayeeTradeParty)
- BT-91: Debited account identifier (IBAN)

**Fix Applied:**
- ✅ Writer: Added output within `SpecifiedTradePaymentTerms` section
- ✅ Correct XML element name: `ram:DirectDebitMandateID`
- ✅ Proper nesting within payment terms

**Verification Source**: ZUGFeRD 2.2 Implementation Guide, EN 16931 semantic model

---

### 5. BT-140/BT-141: Invoice Line Allowance/Charge Calculation Percent ✅

**EN 16931 Definition:**
- **Business Term**: BT-140 "Invoice line allowance percentage" / BT-141 "Invoice line charge percentage"
- **Description**: Percentage that may be used, in conjunction with the invoice line allowance/charge base amount, to calculate the invoice line allowance/charge amount
- **Cardinality**: 0..1 (optional)
- **Profile Level**: BASIC and above

**CII Syntax Mapping:**
```xml
ram:IncludedSupplyChainTradeLineItem/
  ram:SpecifiedLineTradeSettlement/
    ram:SpecifiedTradeAllowanceCharge/
      ram:CalculationPercent
```

**Fix Applied:**
- ✅ Writer: Added `CalculationPercent` output for both line allowances (BG-27) and line charges (BG-28)
- ✅ Conditional output: Only writes if value is non-zero
- ✅ Uses `formatPercent()` helper for proper formatting

**Note**: The writer already correctly handled document-level allowance/charge `CalculationPercent` - only line-level was missing.

**Verification Source**: EN 16931 semantic model, ZUGFeRD CII data structure

---

### 6. BT-72: Actual Delivery Date (Profile Level Fix) ✅

**EN 16931 Definition:**
- **Business Term**: BT-72 "Actual delivery date" / "OccurrenceDateTime"
- **Description**: Date on which the supply of goods or services was made or completed
- **Format**: YYYY-MM-DD / YYYYMMDD (CII format 102)
- **Cardinality**: 0..1 (optional)
- **Profile Level**: **BASIC WL and above** (not BASIC)

**CII Syntax Mapping:**
```xml
rsm:CrossIndustryInvoice/
  rsm:SupplyChainTradeTransaction/
    ram:ApplicableHeaderTradeDelivery/
      ram:ActualDeliverySupplyChainEvent/
        ram:OccurrenceDateTime/
          udt:DateTimeString[@format="102"]
```

**Fix Applied:**
- ✅ Writer: Changed profile check from `is(levelBasic, inv)` to `is(levelBasicWL, inv)`
- ✅ Now correctly outputs for BASIC WL, BASIC, EN 16931, EXTENDED, and XRECHNUNG profiles
- ✅ Excludes MINIMUM profile (as per spec)

**Verification Source**: ZUGFeRD profile level specifications, Factur-X 1.07.2

---

### 7. BG-8: Buyer Postal Address (Profile Level Fix) ✅

**EN 16931 Definition:**
- **Business Group**: BG-8 "BUYER POSTAL ADDRESS"
- **Contains**: BT-50 through BT-55 (address lines, postal code, city, country, subdivision)
- **Cardinality**: 0..1 (optional)
- **Profile Level**: **BASIC WL and above for buyer** (MINIMUM has no buyer address)
- **Note**: Seller address is mandatory at all profiles

**CII Syntax Mapping:**
```xml
ram:BuyerTradeParty/
  ram:PostalTradeAddress/
    ram:PostcodeCode        (BT-53)
    ram:LineOne             (BT-50)
    ram:LineTwo             (BT-51)
    ram:LineThree           (BT-52)
    ram:CityName            (BT-52)
    ram:CountryID           (BT-55)
    ram:CountrySubDivisionName (BT-54)
```

**Fix Applied:**
- ✅ Writer: Changed comment and profile check from `levelBasic` to `levelBasicWL`
- ✅ Logic: `if partyType == CSellerParty || is(levelBasicWL, inv)` - seller always gets address, buyer only at BASIC WL+
- ✅ Correctly handles party type differentiation

**Profile Breakdown:**
- MINIMUM: Seller address only (no buyer postal address)
- BASIC WL: Both seller and buyer addresses
- BASIC/EN16931/EXTENDED/XRECHNUNG: Both addresses

**Verification Source**: ZUGFeRD profile specifications, EN 16931 BASIC WL profile definition

---

## Profile Level Summary

The ZUGFeRD/Factur-X profiles are hierarchical:

| Profile | Level | Buyer Address | BT-72 | BT-16/15 | BT-83 | BT-89 |
|---------|-------|---------------|-------|----------|-------|-------|
| MINIMUM | 1 | ❌ No | ❌ No | ❌ No | ❌ No | ❌ No |
| BASIC WL | 2 | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| BASIC | 3 | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| EN 16931 | 4 | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| EXTENDED | 5 | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| XRECHNUNG | 4 | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |

**Key Insight**: The fixes correctly changed two profile checks from BASIC (level 3) to BASIC WL (level 2), which is the minimum level where buyer postal addresses and actual delivery dates are supported.

---

## Validation Against Official Sources

### Primary Sources Consulted:

1. **EN 16931-1:2017** - Electronic invoicing — Part 1: Semantic data model
2. **EN 16931-3-3:2017** - Syntax binding for UN/CEFACT Cross Industry Invoice (CII)
3. **ZUGFeRD 2.2/2.3.2 Implementation Guide** (FeRD - Forum elektronische Rechnung Deutschland)
4. **Factur-X 1.07.2 Specification** (FNFE-MPE)
5. **ConnectingEurope/eInvoicing-EN16931** - Official validation artifacts
6. **Peppol BIS Billing 3.0** - Pan-European specification
7. **XRechnung 3.0** - German CIUS of EN 16931

### Validation Methods:

- ✅ Verified XPath expressions against official CII schematron rules
- ✅ Confirmed profile level requirements against ZUGFeRD/Factur-X specifications
- ✅ Validated XML structure against CII D16B schema
- ✅ Cross-referenced with UBL syntax mappings for consistency
- ✅ Checked field locations match semantic model group assignments

---

## Test Coverage Validation

The fixes ensure correct round-trip behavior for:

### Test Fixture: `testdata/cii/en16931/CII_example5.xml`
- **Profile**: EN 16931 (COMFORT)
- **Tests**: BT-16 (DespatchAdviceReferencedDocument), BT-15 (ReceivingAdviceReferencedDocument), BT-83 (PaymentReference), BT-89 (DirectDebitMandateID), BT-140/141 (line allowance/charge percentages)
- **Result**: ✅ PASS - No round-trip data loss

### Test Fixture: `testdata/cii/basicwl/zugferd-basicwl-buchungshilfe.xml`
- **Profile**: BASIC WL
- **Tests**: BG-8 (buyer postal address), BT-72 (OccurrenceDateTime) at BASIC WL level
- **Result**: ✅ PASS - No round-trip data loss

---

## Conclusion

**All fixes in PR #100 are fully compliant with EN 16931-1:2017 and EN 16931-3-3:2017.**

The implementations:
1. Use correct XPath expressions for CII syntax
2. Place elements in proper XML structure locations
3. Apply correct profile level checks per ZUGFeRD/Factur-X specifications
4. Follow EN 16931 semantic model business term definitions
5. Maintain cardinality constraints (all fields are 0..1)
6. Support round-trip Parse → Write → Parse cycles without data loss

The fixes address critical data loss issues affecting invoices with delivery documentation, payment references, and direct debit mandates - all essential for legal compliance and automated invoice processing in European e-invoicing systems.

---

## References

- [EN 16931 Official Page](https://ec.europa.eu/digital-building-blocks/sites/display/DIGITAL/EN+16931+compliance)
- [ZUGFeRD.org](https://www.ferd-net.de/standards/zugferd-2.3.2/index.html)
- [Factur-X Specification](https://fnfe-mpe.org/factur-x/)
- [ConnectingEurope eInvoicing Repository](https://github.com/ConnectingEurope/eInvoicing-EN16931)
- [Peppol BIS Billing 3.0](https://docs.peppol.eu/poacc/billing/3.0/)
- [XRechnung Standard](https://xeinkauf.de/xrechnung/)
