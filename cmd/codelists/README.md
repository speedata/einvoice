# Code Lists

This package provides human-readable descriptions for standard code lists used in electronic invoicing.

## Overview

Electronic invoices use standardized codes from various standards (UNTDID, UNECE, etc.). While these codes are machine-readable, they're not user-friendly. This package translates codes into human-readable descriptions.

## Supported Code Lists

### UNTDID 1001 - Document Type

Provides descriptions for invoice document type codes (e.g., "380" → "Standard Invoice").

**Usage:**
```go
import "github.com/speedata/einvoice/cmd/codelists"

description := codelists.DocumentType("380")
// Returns: "Standard Invoice"

description := codelists.DocumentType("999")
// Returns: "Unknown" (for codes not in the list)
```

### UNECE Recommendation 20 - Unit Codes

Provides descriptions for unit of measure codes (e.g., "XPP" → "package", "C62" → "one").

**Usage:**
```go
description := codelists.UnitCode("XPP")
// Returns: "package"

description := codelists.UnitCode("UNKNOWN")
// Returns: "UNKNOWN" (returns the code itself if not found)
```

## Code Generation

The code lists are **generated** from official sources, not checked into the repository. This follows the same pattern as the `rules` package.

### Generating Code Lists

To regenerate the code lists from upstream sources:

```bash
cd cmd/codelists
go generate
```

This runs `cmd/gencodelists` which:
1. Fetches UNTDID 1001 document types from [invopop/gobl](https://github.com/invopop/gobl) (Apache 2.0)
2. Fetches UNECE Rec 20 unit codes from [datasets/unece-units-of-measure](https://github.com/datasets/unece-units-of-measure)
3. Adds custom ZUGFeRD-specific codes (e.g., XPP)
4. Generates `generated.go` with Go maps

**Note:** The generated file is ~60KB of Go code, which is version-controlled. The source data files (JSON/CSV) are NOT checked in.

## Data Sources

- **Document Types (UNTDID 1001)**: EN16931 code list via invopop/gobl
- **Unit Codes (UNECE Rec 20)**: Official UNECE CSV from GitHub datasets

## Future Enhancements

Planned additions include:
- Payment means codes (UNTDID 4461)
- Tax category codes (UNTDID 5305)
- Multi-language support (see [#30](https://github.com/speedata/einvoice/issues/30))

## Updating Code Lists

When upstream sources are updated:

1. Run `go generate` in this directory
2. Review the generated changes
3. Commit the updated `generated.go`

The generator automatically filters deprecated codes and ensures consistency.
