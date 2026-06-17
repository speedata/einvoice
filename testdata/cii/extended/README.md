# Extended Profile Fixtures

This directory is for CII (Cross Industry Invoice) test fixtures with Extended profile (Level 5).

**Profile URN**: `urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended`

**Characteristics**:
- Most comprehensive profile
- All EN 16931 fields plus extensions
- Additional business scenarios
- Complex party structures
- Advanced logistics and payment information

## Sub invoice line fixtures (ZUGFeRD 2.5 / Factur-X 1.09)

Official examples exercising sub invoice lines (chapter 7.6.2). Lines form a
hierarchy via `ram:ParentLineID` (BT-X-304) and carry a subtype in
`ram:LineStatusReasonCode` (BT-X-8): `DETAIL`, `GROUP` (subtotal container) or
`INFORMATION` (informational breakdown). Only detail lines contribute to the
totals and VAT breakdown. Covered by `validate_subinvoicelines_test.go`.

- `zf25-subline-group-hardware.xml` - GROUP subtotal containers (X17)
- `zf25-subline-group-bundle.xml` - multi-level GROUP bundle/kit (X18)
- `zf25-subline-information.xml` - INFORMATION breakdown of a detail line (X06)
- `zf25-subline-nested.xml` - deeply nested GROUP hierarchy (X03)

Add further custom fixtures or contribute official examples as they become available.
