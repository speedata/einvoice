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

- **Minimum coverage: 80%** (enforced in CI)
- **Current coverage: 63.2%**
- **Priority gaps:** UBL writer functions

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

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkParse -benchmem ./...
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
├── scripts/          # Development helper scripts
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

## Continuous Integration

All pull requests must pass:
- ✅ Tests on Go 1.24 and 1.25
- ✅ 80% code coverage threshold
- ✅ golangci-lint checks

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
