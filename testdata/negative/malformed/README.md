# Malformed XML Fixtures

This directory is for **malformed XML** test fixtures used for parser robustness testing.

**Purpose**: Ensure parser handles malformed XML gracefully with clear error messages.

**Examples to add**:
- Unclosed tags: `<Invoice><ID>123</Invoice>`
- Mismatched tags: `<Invoice></invoice>`
- Invalid encoding
- Missing XML declaration
- Broken CDATA sections
- Invalid namespace declarations
- Truncated files

**Expected behavior**: Parser should return clear error indicating XML parsing failure, not crash or hang.

**Status**: Awaiting creation of malformed test cases.
