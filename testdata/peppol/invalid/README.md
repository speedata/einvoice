# Invalid PEPPOL Fixtures

This directory is for **invalid** PEPPOL BIS Billing 3.0 test fixtures used for negative testing.

**Purpose**: Test that validation correctly detects violations of PEPPOL rules.

**Examples to add**:
- Missing mandatory fields (BT-1, BT-2, etc.)
- Invalid code list values
- Incorrect VAT calculations
- Invalid party identifiers
- Schema violations
- Business rule violations (PEPPOL-EN16931-R*)

**Status**: Awaiting creation of negative test cases.

Each file should be named to indicate what violation it contains:
- `missing-invoice-number.xml`
- `invalid-vat-category.xml`
- `wrong-calculation.xml`
- etc.
