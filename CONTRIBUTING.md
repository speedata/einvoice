# Contributing to einvoice

Thank you for your interest in contributing! This document provides guidelines for contributing to the einvoice library.

## Prerequisites

- **Go 1.24 or later** (check with `go version`)
- **golangci-lint** for linting (optional but recommended)
  ```bash
  # Install golangci-lint
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **xmllint** for fixture validation (optional)

## Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR-USERNAME/einvoice.git
   cd einvoice
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Verify your setup**
   ```bash
   go test ./...
   ```

   If all tests pass, you're ready to contribute!

## Development Workflow

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

### Coverage Requirements

- **Target coverage: 80%** ✅ **Achieved: 84.8%** (Phase 1 implementation)
- **Baseline (before Phase 1): 63.2%**
- **Improvement: +21.6 percentage points**

Check your coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

### Code Formatting and Linting

```bash
# Format code (always run before committing)
gofmt -s -w .

# Run linter
golangci-lint run --timeout=5m

# Vet code
go vet ./...
```

## Project Structure

```
einvoice/
├── *.go              # Core library code (parser, writer, validation)
├── cmd/              # Command-line tools
│   ├── einvoice/     # CLI validator tool
│   ├── gencodelists/ # Code list generator
│   └── genrules/     # Business rule generator
├── pkg/
│   └── codelists/    # EN 16931 code lists
├── rules/            # Business rule definitions (auto-generated)
├── testdata/         # Test fixtures (see testdata/README.md)
└── .github/          # CI/CD workflows
```

For detailed architecture and design patterns, see [CLAUDE.md](CLAUDE.md).

## Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (enforced in CI)
- Write clear, descriptive variable names
- Add comments for exported functions and complex logic
- Follow patterns established in existing code
- Use EN 16931 field references (BT-*, BG-*) in comments for traceability

## Testing

### Writing Tests

- Place tests in `*_test.go` files alongside the code they test
- Use table-driven tests for multiple scenarios
- Use subtests with `t.Run()` for clarity
- Mock external dependencies when needed

### Test Fixtures

Test fixtures are organized by profile and format in `testdata/`. See [testdata/README.md](testdata/README.md) for:
- Directory structure and organization
- Fixture sources and provenance
- How to add new fixtures
- Usage patterns in tests

### Example Test

```go
func TestParseInvoice(t *testing.T) {
    tests := []struct {
        name    string
        file    string
        wantErr bool
    }{
        {
            name:    "valid EN 16931 invoice",
            file:    "testdata/cii/en16931/CII_example1.xml",
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inv, err := ParseXMLFile(tt.file)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseXMLFile() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && inv == nil {
                t.Error("ParseXMLFile() returned nil invoice")
            }
        })
    }
}
```

### Table-Driven Tests

Table-driven tests are the preferred pattern for testing multiple scenarios. They improve readability, reduce code duplication, and make it easy to add new test cases.

**Basic Structure:**
```go
func TestUpdateTotals_BRCORules(t *testing.T) {
    tests := []struct {
        name                  string
        lines                 []InvoiceLine
        allowancesCharges     []AllowanceCharge
        tradeTaxes            []TradeTax
        expectedLineTotal     decimal.Decimal
        expectedTaxBasisTotal decimal.Decimal
        expectedGrandTotal    decimal.Decimal
        expectedDuePayable    decimal.Decimal
    }{
        {
            name: "BR-CO-10: Simple line total",
            lines: []InvoiceLine{
                {NetAmount: decimal.NewFromFloat(100.00)},
                {NetAmount: decimal.NewFromFloat(50.00)},
            },
            expectedLineTotal: decimal.NewFromFloat(150.00),
            // ... other expectations
        },
        {
            name: "BR-CO-13: With allowances and charges",
            // ... test case data
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inv := &Invoice{
                InvoiceLines:        tt.lines,
                AllowancesCharges:   tt.allowancesCharges,
                ApplicableTradeTax:  tt.tradeTaxes,
            }

            inv.UpdateTotals()

            if !inv.LineTotal.Equal(tt.expectedLineTotal) {
                t.Errorf("LineTotal = %v, want %v", inv.LineTotal, tt.expectedLineTotal)
            }
            // ... more assertions
        })
    }
}
```

**Best Practices:**
- Use descriptive test names that explain the scenario (e.g., "BR-CO-10: Simple line total")
- Reference EN 16931 business rules (BR-*) in test names for traceability
- Use `t.Run()` to create subtests for each case
- Add `t.Parallel()` when tests are independent and can run concurrently
- Keep test cases focused on one aspect/rule per test function
- Use `t.Fatal()` instead of `t.Error()` when a nil pointer would cause panic

**Examples in codebase:**
- `calculate_test.go:TestUpdateTotals_BRCORules` - Tests BR-CO calculation rules
- `einvoice_test.go:TestProfileDetection` - Tests all profile URN detection
- `parser_ubl_test.go:TestUBLDateParsingInvalid` - Tests invalid date formats

### Benchmarks

Benchmarks measure performance and help detect regressions. All benchmarks use the Go 1.24+ `b.Loop()` pattern.

**Running benchmarks:**
```bash
# Run all benchmarks
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkParse -benchmem

# Run with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof

# Compare before/after performance
go test -bench=. -benchmem > old.txt
# ... make changes ...
go test -bench=. -benchmem > new.txt
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

**Performance targets (baseline from Phase 1):**
- **Parse CII (EN16931)**: ~325-450μs per operation
- **Parse CII (Minimum)**: ~325-400μs per operation
- **Parse UBL (Invoice)**: ~2.2ms per operation
- **Write CII**: ~134μs @ 117 MB/s
- **Write UBL**: ~170-200μs per operation
- **Validate**: ~15μs per operation
- **Calculate (UpdateTotals)**: ~7.6μs per operation
- **Round-trip (parse→write→parse)**: 625μs (CII) to 4.3ms (UBL)

**Writing benchmarks:**
```go
func BenchmarkParseCII(b *testing.B) {
    data, err := os.ReadFile("testdata/cii/en16931/CII_example1.xml")
    if err != nil {
        b.Fatal(err)
    }

    b.SetBytes(int64(len(data)))  // Track throughput
    b.ReportAllocs()               // Report memory allocations
    b.ResetTimer()                 // Reset timer after setup

    for b.Loop() {  // Go 1.24+ pattern
        _, err := ParseReader(bytes.NewReader(data))
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Benchmark best practices:**
- Use `b.Loop()` instead of manual `for i := 0; i < b.N; i++` (Go 1.24+)
- Call `b.ResetTimer()` after expensive setup
- Use `b.SetBytes()` for operations with I/O to track throughput (MB/s)
- Use `b.ReportAllocs()` to track memory allocations
- Create sub-benchmarks with `b.Run()` for different scenarios
- Use real fixtures from `testdata/` directory

**Examples in codebase:**
- `parser_cii_test.go:BenchmarkParseCII` - Benchmarks all CII profiles
- `writer_cii_test.go:BenchmarkWriteCII` - Benchmarks CII writing
- `validation_test.go:BenchmarkValidate` - Benchmarks validation with different rule counts
- `einvoice_test.go:BenchmarkRoundTrip` - Benchmarks full round-trip cycle

### Fuzz Testing

Fuzz testing uses Go 1.18+ native fuzzing to find crashes and edge cases.

**Running fuzz tests:**
```bash
# Run for 30 seconds
go test -fuzz=FuzzParseCII -fuzztime=30s

# Run for 5 minutes
go test -fuzz=FuzzRoundTrip -fuzztime=5m

# Run specific number of iterations
go test -fuzz=FuzzParseCII -fuzztime=100000x
```

**Available fuzz tests:**
- `parser_cii_test.go:FuzzParseCII` - CII parser
- `parser_ubl_test.go:FuzzParseUBL` - UBL parser
- `validation_test.go:FuzzValidate` - Validation logic
- `einvoice_test.go:FuzzRoundTrip` - Full round-trip cycle

**Results:**
All fuzz tests passed with no crashes found (Phase 1: 109K-578K executions each)

## Continuous Integration

All pull requests must pass:
- ✅ Tests on Go 1.24 and 1.25 (3 OS: ubuntu, macos, windows)
- ✅ golangci-lint checks

Coverage is tracked but not currently enforced. Goal: 80%.

Run the same checks locally before pushing:
```bash
# Run all tests with coverage
go test -race -coverprofile=coverage.out ./...

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Run linter
golangci-lint run --timeout=5m
```

## Submitting Changes

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 2. Make Your Changes

- Write code following the style guide
- Add tests for new functionality
- Update documentation if needed
- Ensure tests pass and coverage is adequate

### 3. Commit Your Changes

Use clear, descriptive commit messages:
```bash
git add .
git commit -m "Add support for UBL payment terms parsing

- Implement writeUBLPaymentTerms function
- Add tests with PEPPOL fixtures
- Closes #123"
```

**Commit message format:**
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed explanation (if needed)
- Reference issues with "Closes #123" or "Fixes #456"

### 4. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub with:
- **Clear title** describing the change
- **Description** explaining what and why
- **Link to related issues** (if any)
- **Test results** or screenshots (if applicable)

### 5. Code Review

- Address reviewer feedback promptly
- Keep discussions focused and respectful
- Update your branch as needed
- Once approved, maintainers will merge your PR

## Questions or Issues?

- **Bug reports:** Open an issue with reproduction steps
- **Feature requests:** Open an issue describing the use case
- **Questions:** Check existing issues or start a discussion

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to einvoice! 🎉
