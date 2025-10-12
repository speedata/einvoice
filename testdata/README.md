# Test Fixtures

This directory contains organized test fixtures for the einvoice library. Fixtures are organized by **profile** and **format** to support targeted testing.

## Directory Structure

```
testdata/
├── README.md              # This file
├── SOURCES.md            # Provenance tracking for all fixtures
├── cii/                  # Cross Industry Invoice (ZUGFeRD/Factur-X) fixtures
│   ├── minimum/          # Minimum profile (Level 1) - 2 files
│   ├── basicwl/          # Basic WL profile (Level 2) - 2 files
│   ├── basic/            # Basic profile (Level 3) - 4 files
│   ├── en16931/          # EN 16931 profile (Level 4) - 16 files
│   ├── extended/         # Extended profile (Level 5) - 6 files
│   └── xrechnung/        # XRechnung profile (Level 4) - 4 files
├── ubl/                  # Universal Business Language 2.1 fixtures
│   ├── invoice/          # UBL Invoice documents - 12 files
│   └── creditnote/       # UBL CreditNote documents - 3 files
├── peppol/               # PEPPOL BIS Billing 3.0 fixtures
│   ├── valid/            # Valid PEPPOL invoices - 11 files
│   └── invalid/          # Invalid examples for negative testing
└── negative/             # Negative test cases
    ├── malformed/        # Malformed XML - 2 files
    └── missing_fields/   # Missing required fields
```

**Total: 62 test fixtures** (organized from official test suites and standards)

## Fixture Organization

### Why Profile-Based Organization?

The einvoice library is **profile-aware**:
- Parser auto-detects profiles from `GuidelineSpecifiedDocumentContextParameter` (BT-24)
- Writer outputs different fields based on `ProfileLevel()`
- Validation enforces profile-specific rules

Organizing by profile makes tests simpler:
```go
// Easy: Get all EN 16931 fixtures
fixtures, _ := filepath.Glob("testdata/cii/en16931/*.xml")

// Without organization: Parse each file to determine profile
for _, file := range allFixtures {
    inv, _ := ParseXMLFile(file)
    if inv.IsEN16931() {
        // test it
    }
}
```

The directory path encodes profile metadata, eliminating the need to parse files just to categorize them.

## Profile Levels

CII profiles are hierarchical (higher includes lower):

| Level | Profile    | URN Pattern                                                  |
|-------|------------|--------------------------------------------------------------|
| 1     | Minimum    | `urn:factur-x.eu:1p0:minimum`                                |
| 2     | BasicWL    | `urn:factur-x.eu:1p0:basicwl`                                |
| 3     | Basic      | `urn:cen.eu:en16931:2017#compliant#...factur-x.eu:1p0:basic`|
| 4     | EN 16931   | `urn:cen.eu:en16931:2017`                                    |
| 4     | XRechnung  | `urn:cen.eu:en16931:2017#compliant#...xrechnung_3.0`         |
| 4     | PEPPOL     | `urn:cen.eu:en16931:2017#compliant#...peppol.eu:2017:poacc` |
| 5     | Extended   | `urn:cen.eu:en16931:2017#conformant#...factur-x.eu:1p0:ext` |

## Test Usage Patterns

### Fixture-Based Tests

```go
func TestParseEN16931Fixtures(t *testing.T) {
    fixtures, err := filepath.Glob("testdata/cii/en16931/*.xml")
    require.NoError(t, err)

    for _, fixture := range fixtures {
        t.Run(filepath.Base(fixture), func(t *testing.T) {
            inv, err := einvoice.ParseXMLFile(fixture)
            require.NoError(t, err)
            assert.True(t, inv.IsEN16931())
        })
    }
}
```

### Round-Trip Tests

```go
func TestRoundTripUBL(t *testing.T) {
    fixtures, _ := filepath.Glob("testdata/ubl/invoice/*.xml")

    for _, fixture := range fixtures {
        t.Run(filepath.Base(fixture), func(t *testing.T) {
            // Parse original
            inv1, err := einvoice.ParseXMLFile(fixture)
            require.NoError(t, err)

            // Write to buffer
            var buf bytes.Buffer
            err = inv1.Write(&buf)
            require.NoError(t, err)

            // Parse written output
            inv2, err := einvoice.ParseReader(&buf)
            require.NoError(t, err)

            // Compare
            assert.Equal(t, inv1.InvoiceNumber, inv2.InvoiceNumber)
            assert.Equal(t, inv1.InvoiceTypeCode, inv2.InvoiceTypeCode)
            // ... more assertions
        })
    }
}
```

### Validation Tests

```go
func TestValidatePEPPOL(t *testing.T) {
    fixtures, _ := filepath.Glob("testdata/peppol/valid/*.xml")

    for _, fixture := range fixtures {
        t.Run(filepath.Base(fixture), func(t *testing.T) {
            inv, err := einvoice.ParseXMLFile(fixture)
            require.NoError(t, err)

            err = inv.Validate()
            if err != nil {
                validErr := err.(*einvoice.ValidationError)
                t.Logf("Violations: %v", validErr.Violations())
            }
        })
    }
}
```

## Coverage Goals

### Current Coverage (Baseline)

- **Total: 63.2%** (target: 80%)
- Main package: 73.9%
- cmd/einvoice: 28.5%
- pkg/codelists: 100.0%

### Coverage Gaps

Areas needing more tests (UBL writer has most gaps):

| Function                    | Coverage | Priority |
|-----------------------------|----------|----------|
| writeUBLLineAllowanceCharge | 0.0%     | HIGH     |
| writeUBLPaymentMeans        | 4.2%     | HIGH     |
| writeUBLPaymentTerms        | 14.3%    | HIGH     |
| writeUBLLineItem            | 37.1%    | MEDIUM   |
| writeUBLHeader              | 40.7%    | MEDIUM   |
| writeUBLParties             | 43.8%    | MEDIUM   |
| writeUBLParty               | 46.6%    | MEDIUM   |
| writeUBLLinePrice           | 50.0%    | MEDIUM   |
| writeUBLLines               | 55.2%    | MEDIUM   |

**Action**: More round-trip tests with UBL fixtures covering:
- Payment means (credit transfer, direct debit, card, SEPA)
- Payment terms
- Line-level allowances/charges
- Complex party structures
- Line item variations

## Fixture Sources

All fixtures are sourced from official standards and test suites:

- **ZUGFeRD 2.3.3 Official Examples**: https://www.ferd-net.de/download-zugferd
  - All profiles: Minimum (2), BasicWL (2), Basic (3), EN16931 (6), Extended (4), XRechnung (3)
  - Total: 20 files covering complete ZUGFeRD/Factur-X profile range

- **EN 16931 Test Suite**: https://github.com/ConnectingEurope/eInvoicing-EN16931
  - CII examples (10 files in `cii/en16931/`, 1 file in `cii/xrechnung/`)
  - UBL examples (10 invoices + 1 credit note in `ubl/`)

- **UBL 2.1 OASIS Specification**: https://docs.oasis-open.org/ubl/
  - Official UBL 2.1 Invoice and CreditNote examples (2 files)

- **PEPPOL BIS Billing 3.0**: https://github.com/OpenPEPPOL/peppol-bis-invoice-3
  - Base examples, VAT categories, national examples (11 files in `peppol/valid/`)

- **horstoeko/zugferd**: https://github.com/horstoeko/zugferd
  - Additional test files (3 files) and invalid examples for negative testing (2 files)

See [SOURCES.md](SOURCES.md) for detailed provenance tracking with commit hashes.

## Updating Fixtures

Fixtures update rarely. See [SOURCES.md](SOURCES.md) for the manual update process when upstream test suites are updated.

## Adding Custom Fixtures

You can add custom fixtures alongside official ones:

1. Place in appropriate directory (by profile/format)
2. Use descriptive filenames: `custom-<description>.xml`
3. Document in a comment at top of file:
   ```xml
   <!-- Custom test fixture for testing <specific scenario> -->
   ```

## Testing Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

## Notes

- **Profile coverage**: All ZUGFeRD profiles (Minimum, BasicWL, Basic, EN16931, Extended, XRechnung) now have official test fixtures
- **Negative tests**: `negative/malformed/` directory contains invalid XML examples for error handling tests
- **Parser robustness**: Parser should handle fixtures gracefully even if they have validation violations
- **Format auto-detection**: Parser automatically detects CII vs UBL format
- **Total fixtures**: 62 XML files across all profiles, formats, and document types
