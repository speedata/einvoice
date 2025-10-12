# Contributing to einvoice

Thank you for your interest in contributing to this EN 16931 electronic invoice library!

## Prerequisites

- **Go 1.24 or later**
- **golangci-lint** (optional but recommended)
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

## Quick Start

1. Fork and clone:
   ```bash
   git clone https://github.com/YOUR-USERNAME/einvoice.git
   cd einvoice
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Verify your setup:
   ```bash
   go test ./...
   ```

## Development Workflow

### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# With race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem

# Run fuzz tests
go test -fuzz=FuzzParseCII -fuzztime=30s
```

**Coverage target:** 80% (currently at 84.8%)

### Formatting and Linting

```bash
# Format code (required before committing)
gofmt -s -w .

# Run linter
golangci-lint run --timeout=5m

# Vet code
go vet ./...
```

## Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (enforced in CI)
- Use table-driven tests with `t.Run()` for subtests
- Reference EN 16931 business rules (BR-*, BG-*) in test names and comments
- See existing code for patterns

**Architecture details:** See [CLAUDE.md](CLAUDE.md) for design patterns and architecture.

**Test fixtures:** See [testdata/README.md](testdata/README.md) for fixture organization.

## Submitting Changes

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 2. Make Your Changes

- Write code following the style guide
- Add tests for new functionality
- Ensure tests pass and coverage is adequate
- Format and lint your code

### 3. Commit

Use clear, descriptive commit messages:

```bash
git commit -m "Add support for PEPPOL payment terms

- Implement parsePaymentTerms in UBL parser
- Add validation for BT-9 (payment due date)
- Add test fixtures from PEPPOL BIS
- Closes #123"
```

**Format:**
- First line: Brief summary (imperative mood, 50 chars max)
- Blank line
- Detailed explanation (if needed)
- Reference issues: `Closes #123` or `Fixes #456`

### 4. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Create a pull request with:
- Clear title describing the change
- Description explaining what and why
- Link to related issues

### 5. CI Requirements

All PRs must pass:
- Tests on Go 1.24 and 1.25 (Linux, macOS, Windows)
- golangci-lint checks
- Fuzz tests (30s each)
- Benchmark tests

Run locally before pushing:
```bash
go test -race -coverprofile=coverage.out ./...
golangci-lint run --timeout=5m
```

## Questions?

- **Bug reports:** Open an issue with reproduction steps
- **Feature requests:** Open an issue describing the use case
- **Questions:** Check existing issues or start a discussion

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
