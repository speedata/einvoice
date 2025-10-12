# Test Fixtures

Test fixtures for the einvoice library, organized by profile and format for targeted testing.

See [SOURCES.md](SOURCES.md) for provenance tracking, license information, and file counts.

## Directory Structure

```
testdata/
├── README.md              # This file
├── SOURCES.md            # Provenance and license tracking
├── cii/                  # Cross Industry Invoice (ZUGFeRD/Factur-X)
│   ├── minimum/          # Minimum profile (Level 1)
│   ├── basicwl/          # Basic WL profile (Level 2)
│   ├── basic/            # Basic profile (Level 3)
│   ├── en16931/          # EN 16931 profile (Level 4)
│   ├── extended/         # Extended profile (Level 5)
│   └── xrechnung/        # XRechnung profile (Level 4)
├── ubl/                  # Universal Business Language 2.1
│   ├── invoice/          # Invoice documents
│   └── creditnote/       # CreditNote documents
├── peppol/               # PEPPOL BIS Billing 3.0
│   ├── valid/            # Valid PEPPOL invoices
│   └── invalid/          # Invalid examples (future)
└── negative/             # Negative test cases
    ├── malformed/        # Malformed XML
    └── missing_fields/   # Missing required fields (future)
```

## Why Profile-Based Organization?

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

## Coverage Gaps

Areas needing more test fixtures (UBL writer functions with low coverage):

| Function                    | Coverage | Priority | Needed Fixtures |
|-----------------------------|----------|----------|-----------------|
| writeUBLLineAllowanceCharge | 0.0%     | HIGH     | Line-level allowances/charges |
| writeUBLPaymentMeans        | 4.2%     | HIGH     | Credit transfer, direct debit, card, SEPA |
| writeUBLPaymentTerms        | 14.3%    | HIGH     | Payment terms with various conditions |
| writeUBLLineItem            | 37.1%    | MEDIUM   | Line item variations |
| writeUBLHeader              | 40.7%    | MEDIUM   | Header field variations |
| writeUBLParties             | 43.8%    | MEDIUM   | Complex party structures |
| writeUBLParty               | 46.6%    | MEDIUM   | Party field combinations |
| writeUBLLinePrice           | 50.0%    | MEDIUM   | Price variations |
| writeUBLLines               | 55.2%    | MEDIUM   | Multiple line scenarios |

## Adding Custom Fixtures

You can add custom fixtures alongside official ones:

1. Place in appropriate directory (by profile/format)
2. Use descriptive filenames: `custom-<description>.xml`
3. Document in XML comment header:
   ```xml
   <!-- Custom test fixture for testing <specific scenario> -->
   ```

## Status

- ✅ All ZUGFeRD profiles (Minimum, BasicWL, Basic, EN16931, Extended, XRechnung) have official test fixtures
- ✅ UBL 2.1 Invoice and CreditNote formats covered
- ✅ PEPPOL BIS Billing 3.0 validation examples included
- ⚠️  UBL writer functions need more coverage (see table above)
