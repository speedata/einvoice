# Test Data for PDF Support

This directory contains test fixtures for ZUGFeRD/Factur-X PDF functionality.

## Adding Test PDF Files

To enable PDF parsing tests, add ZUGFeRD/Factur-X PDF files to this directory.

### Where to Get Test PDFs

1. **ZUGFeRD Corpus** (recommended):
   - https://github.com/ZUGFeRD/corpus
   - Contains real-world samples and test cases
   - Includes various profiles (EN 16931, Basic, Extended, etc.)

2. **Factur-X Samples**:
   - https://fnfe-mpe.org/factur-x/
   - Official Factur-X test files

3. **Mustang Project**:
   - https://www.mustangproject.org/
   - Open-source ZUGFeRD/Factur-X library with samples

### Requirements

Test PDF files should:
- Be valid ZUGFeRD/Factur-X PDFs (PDF/A-3 with embedded XML)
- Contain an embedded invoice XML file named one of:
  - `factur-x.xml` (Factur-X standard)
  - `ZUGFeRD-invoice.xml` (ZUGFeRD 2.x)
  - `zugferd-invoice.xml` (ZUGFeRD 1.x)
  - `xrechnung.xml` (XRechnung)
  - or any `.xml` file as fallback

### Running Tests

Once PDF files are added to this directory:

```bash
# Run all tests including PDF tests
go test -v ./cmd/einvoice

# Run only PDF-related tests
go test -v ./cmd/einvoice -run PDF
```

Tests will automatically detect and use any `.pdf` files in this directory.

## Current Status

No PDF test files are currently included in the repository.

PDF parsing tests are skipped when no test files are present - this is expected behavior. The implementation has been tested manually with real ZUGFeRD PDFs.

To contribute test PDFs, ensure they are properly licensed for distribution in an open-source project.
