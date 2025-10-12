# Missing Required Fields Fixtures

This directory is for test fixtures with **missing required fields** used for validation testing.

**Purpose**: Test that validation correctly detects missing mandatory fields per EN 16931.

**Examples to add**:
- Missing invoice number (BT-1) - violates BR-2
- Missing invoice date (BT-2) - violates BR-2
- Missing buyer reference (BT-10) - violates BR-16
- Missing seller VAT (BT-31) - violates BR-CO-9
- Missing payment means (BG-16) - violates BR-49
- Missing line totals

**Expected behavior**: Validation should fail with specific business rule violations (BR-*).

**Status**: Awaiting creation of negative test cases.

Each file should be named to indicate what field is missing:
- `missing-BT-1-invoice-number.xml`
- `missing-BT-2-invoice-date.xml`
- etc.
