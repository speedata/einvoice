# Test Fixture Sources

This document tracks the provenance of all test fixtures in this directory. Fixtures are copied from upstream repositories and organized by profile/format for testing convenience.

## Last Updated

**Date**: 2025-10-13

## Upstream Repositories

### EN 16931 Test Suite

- **Repository**: https://github.com/ConnectingEurope/eInvoicing-EN16931
- **Commit**: `a99371b18e1e924f4b5eaa75ffa83cdbc150aefd`
- **Date**: 2025-10-09
- **Purpose**: Official EN 16931 European e-invoicing standard test files (CII and UBL formats)

### ZUGFeRD 2.3.3 Official Examples

- **Source**: https://www.ferd-net.de/download-zugferd
- **Package**: ZUGFeRD 2.3.3 EN (ZF233_EN_01)
- **Date**: May 2024
- **Purpose**: Official FeRD test files for all ZUGFeRD/Factur-X profiles (Minimum, BasicWL, Basic, EN16931, Extended, XRechnung)

### horstoeko/zugferd Test Files

- **Repository**: https://github.com/horstoeko/zugferd
- **Purpose**: Additional ZUGFeRD test files (Basic, Extended) and invalid test cases for negative testing
- **License**: MIT

### UBL 2.1 OASIS Examples

- **Source**: https://docs.oasis-open.org/ubl/os-UBL-2.2/xml/
- **Repository**: https://github.com/Tradeshift/tradeshift-ubl-examples
- **Purpose**: Official OASIS UBL 2.1 Invoice and CreditNote examples
- **License**: OASIS open standard

### PEPPOL BIS Billing 3.0 Test Suite

- **Repository**: https://github.com/OpenPEPPOL/peppol-bis-invoice-3
- **Commit**: `78d7f7dfa223f39f8ebd8d35c127f2690e646322`
- **Date**: 2025-05-29
- **Purpose**: PEPPOL BIS Billing 3.0 validation examples and test files

## File Mappings

### CII (Cross Industry Invoice) Format

#### `cii/minimum/` (2 files)
Source: ZUGFeRD 2.3.3 official package
- `zugferd-minimum-buchungshilfe.xml` - Minimum profile accounting aid example
- `zugferd-minimum-rechnung.xml` - Minimum profile invoice example

#### `cii/basicwl/` (2 files)
Source: ZUGFeRD 2.3.3 official package
- `zugferd-basicwl-buchungshilfe.xml` - Basic WL profile accounting aid example
- `zugferd-basicwl-einfach.xml` - Basic WL profile simple example

#### `cii/basic/` (4 files)
Sources:
- ZUGFeRD 2.3.3 official package (3 files)
- horstoeko/zugferd repository (1 file)

Files:
- `zugferd-basic-1.xml` - Basic profile from horstoeko test suite
- `zugferd-basic-einfach.xml` - Simple basic invoice
- `zugferd-basic-rechnungskorrektur.xml` - Invoice correction
- `zugferd-basic-taxifahrt.xml` - Taxi ride invoice

#### `cii/en16931/` (16 files)
Sources:
- EN 16931 test suite (10 files)
- ZUGFeRD 2.3.3 official package (6 files)

Files from EN 16931:
- `CII_example1.xml` - Comprehensive EN 16931 example
- `CII_example2.xml` - EN 16931 with multiple line items
- `CII_example3.xml` - EN 16931 simplified
- `CII_example4.xml` - EN 16931 with allowances
- `CII_example5.xml` - EN 16931 with charges
- `CII_example6.xml` - EN 16931 minimal
- `CII_example7.xml` - EN 16931 with delivery
- `CII_example8.xml` - EN 16931 with payment terms
- `CII_example9.xml` - EN 16931 with references
- `zugferd_2p0_EN16931_1_Teilrechnung.xml` - Partial invoice example

Files from ZUGFeRD 2.3.3:
- `zugferd-en16931-einfach.xml` - Simple EN16931 invoice
- `zugferd-en16931-gutschrift.xml` - Credit note
- `zugferd-en16931-intra-community.xml` - Intra-community supply (IC VAT category)
- `zugferd-en16931-payee.xml` - Invoice with different payee party
- `zugferd-en16931-rabatte.xml` - Invoice with discounts/allowances
- `zugferd-en16931-rechnungskorrektur.xml` - Invoice correction

#### `cii/extended/` (6 files)
Sources:
- ZUGFeRD 2.3.3 official package (4 files)
- horstoeko/zugferd repository (2 files)

Files:
- `zugferd-extended-1.xml` - Extended profile from horstoeko test suite
- `zugferd-extended-2.xml` - Extended profile variant from horstoeko
- `zugferd-extended-fremdwaehrung.xml` - Foreign currency invoice
- `zugferd-extended-intra-community-multi.xml` - Intra-community with multiple orders
- `zugferd-extended-rechnungskorrektur.xml` - Invoice correction
- `zugferd-extended-warenrechnung.xml` - Goods invoice

#### `cii/xrechnung/` (4 files)
Sources:
- EN 16931 test suite (1 file)
- ZUGFeRD 2.3.3 official package (3 files)

Files:
- `XRechnung-O.xml` - XRechnung profile example (EN 16931)
- `zugferd-xrechnung-betriebskosten.xml` - Operating costs invoice
- `zugferd-xrechnung-einfach.xml` - Simple XRechnung invoice
- `zugferd-xrechnung-elektron.xml` - Electronic invoice example

### UBL 2.1 Format

#### `ubl/invoice/` (12 files)
Sources:
- EN 16931 test suite (10 files)
- PEPPOL test suite (1 file)
- OASIS UBL 2.1 specification (1 file)

Files from EN 16931:
- `ubl-tc434-example1.xml` - Comprehensive UBL invoice
- `ubl-tc434-example2.xml` - UBL with multiple VAT categories
- `ubl-tc434-example3.xml` - UBL simplified
- `ubl-tc434-example4.xml` - UBL with allowances/charges
- `ubl-tc434-example5.xml` - UBL with delivery
- `ubl-tc434-example6.xml` - UBL minimal
- `ubl-tc434-example7.xml` - UBL with payment means
- `ubl-tc434-example8.xml` - UBL with references
- `ubl-tc434-example9.xml` - UBL with tax representative
- `ubl-tc434-example10.xml` - UBL with project reference

File from PEPPOL:
- `peppol-ubl-invoice-complete.xml` - Comprehensive PEPPOL syntax example

File from OASIS:
- `UBL-Invoice-2.1-Example.xml` - Official OASIS UBL 2.1 Invoice example

#### `ubl/creditnote/` (3 files)
Sources:
- EN 16931 test suite (1 file)
- PEPPOL test suite (1 file)
- OASIS UBL 2.1 specification (1 file)

Files:
- `ubl-tc434-creditnote1.xml` - EN 16931 credit note example
- `peppol-ubl-creditnote-complete.xml` - PEPPOL credit note syntax example
- `UBL-CreditNote-2.1-Example.xml` - Official OASIS UBL 2.1 CreditNote example

### PEPPOL BIS Billing 3.0

#### `peppol/valid/` (11 files)
Sources:
- `peppol-testsuite/rules/examples/` (8 files)
- `peppol-testsuite/rules/national-examples/GR/` (2 files)
- `peppol-testsuite/rules/national-examples/NO/` (1 file)

General examples:
- `base-example.xml` - Base PEPPOL invoice example
- `Allowance-example.xml` - Document-level allowances
- `base-creditnote-correction.xml` - Credit note for correction
- `base-negative-inv-correction.xml` - Negative invoice correction

VAT category examples:
- `Vat-category-S.xml` - Standard rated (S)
- `vat-category-E.xml` - Exempt from VAT (E)
- `vat-category-O.xml` - Not subject to VAT (O)
- `vat-category-Z.xml` - Zero rated (Z)

National examples:
- `GR-base-example-correct.xml` - Greek PEPPOL example
- `GR-base-example-TaxRepresentative.xml` - Greek with tax representative
- `Norwegian-example-1.xml` - Norwegian PEPPOL example

### Negative Test Cases

#### `negative/malformed/` (2 files)
Source: horstoeko/zugferd test suite
- `zugferd-invalid-1.xml` - Invalid ZUGFeRD XML structure
- `zugferd-invalid-2.xml` - Malformed ZUGFeRD invoice

## Organization Strategy

Fixtures are organized by **profile level** and **document type** rather than by validation rule. This organization directly supports our testing requirements:

1. **Profile-based testing**: Directory path encodes profile metadata
2. **Format-based testing**: Separate CII and UBL directories
3. **Validation testing**: PEPPOL directory contains BIS-compliant examples
4. **Simplicity**: Test code can use simple glob patterns like `filepath.Glob("testdata/cii/en16931/*.xml")`

## Updating Fixtures

Fixtures update rarely. When upstream test suites are updated:

1. Clone/pull the upstream repositories:
   ```bash
   git clone https://github.com/ConnectingEurope/eInvoicing-EN16931.git
   git clone https://github.com/OpenPEPPOL/peppol-bis-invoice-3.git
   ```

2. Copy relevant files to organized directories (see file mappings above)

3. Update this SOURCES.md with new commit hashes and dates:
   ```bash
   cd eInvoicing-EN16931
   git log -1 --format="%H %ci"
   ```

## Notes

- **Profile detection**: CII profiles are determined by `GuidelineSpecifiedDocumentContextParameter` (BT-24) URN
- **Profile coverage**: All ZUGFeRD profiles now have official test files (Minimum, BasicWL, Basic, EN16931, Extended, XRechnung)
- **Custom fixtures**: Additional custom-created fixtures may be added alongside official ones
- **Negative tests**: `negative/malformed/` contains invalid XML examples for error handling tests
- **Total fixtures**: 62 XML test files across all profiles and formats
