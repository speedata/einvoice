package einvoice

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name       string
		violations []SemanticError
		want       string
	}{
		{
			name:       "no violations",
			violations: []SemanticError{},
			want:       "validation failed with no violations",
		},
		{
			name: "single violation",
			violations: []SemanticError{
				{Rule: rules.BR1, Text: "Invoice number is required"},
			},
			want: "validation failed: BR-01 - Invoice number is required",
		},
		{
			name: "multiple violations",
			violations: []SemanticError{
				{Rule: rules.BR1, Text: "Invoice number is required"},
				{Rule: rules.BR2, Text: "Invoice date is required"},
				{Rule: rules.BR3, Text: "Currency is required"},
			},
			want: "validation failed with 3 violations (first: BR-01 - Invoice number is required)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{violations: tt.violations}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Violations(t *testing.T) {
	t.Run("returns copy of violations", func(t *testing.T) {
		original := []SemanticError{
			{Rule: rules.BR1, Text: "Test violation"},
		}
		e := &ValidationError{violations: original}

		// Get violations
		violations := e.Violations()

		// Verify content
		if len(violations) != 1 {
			t.Errorf("Violations() returned %d violations, want 1", len(violations))
		}
		if violations[0].Rule.Code != "BR-01" {
			t.Errorf("Violations()[0].Rule.Code = %v, want BR-1", violations[0].Rule.Code)
		}

		// Modify the returned slice - should not affect internal state
		violations[0].Rule = rules.BR2

		// Verify internal state unchanged
		if e.violations[0].Rule.Code != "BR-01" {
			t.Errorf("Internal violations were modified, want BR-1, got %v", e.violations[0].Rule.Code)
		}
	})

	t.Run("returns nil for nil violations", func(t *testing.T) {
		e := &ValidationError{violations: nil}
		violations := e.Violations()
		if violations != nil {
			t.Errorf("Violations() = %v, want nil", violations)
		}
	})

	t.Run("returns empty slice for empty violations", func(t *testing.T) {
		e := &ValidationError{violations: []SemanticError{}}
		violations := e.Violations()
		if violations == nil {
			t.Error("Violations() = nil, want empty slice")
		}
		if len(violations) != 0 {
			t.Errorf("Violations() length = %d, want 0", len(violations))
		}
	})
}

func TestValidationError_Count(t *testing.T) {
	tests := []struct {
		name       string
		violations []SemanticError
		want       int
	}{
		{
			name:       "no violations",
			violations: []SemanticError{},
			want:       0,
		},
		{
			name: "one violation",
			violations: []SemanticError{
				{Rule: rules.BR1, Text: "Test"},
			},
			want: 1,
		},
		{
			name: "multiple violations",
			violations: []SemanticError{
				{Rule: rules.BR1, Text: "Test 1"},
				{Rule: rules.BR2, Text: "Test 2"},
				{Rule: rules.BR3, Text: "Test 3"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{violations: tt.violations}
			if got := e.Count(); got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_HasRule(t *testing.T) {
	violations := []SemanticError{
		{Rule: rules.BR1, Text: "Test 1"},
		{Rule: rules.BRS8, Text: "Test 2"},
		{Rule: rules.BRCO10, Text: "Test 3"},
	}
	e := &ValidationError{violations: violations}

	t.Run("HasRule with Rule constants", func(t *testing.T) {
		tests := []struct {
			name string
			rule rules.Rule
			want bool
		}{
			{
				name: "rule exists - rules.BR1",
				rule: rules.BR1,
				want: true,
			},
			{
				name: "rule exists - rules.BRS8",
				rule: rules.BRS8,
				want: true,
			},
			{
				name: "rule exists - rules.BRCO10",
				rule: rules.BRCO10,
				want: true,
			},
			{
				name: "rule does not exist",
				rule: rules.BR2,
				want: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := e.HasRule(tt.rule); got != tt.want {
					t.Errorf("HasRule(%v) = %v, want %v", tt.rule.Code, got, tt.want)
				}
			})
		}
	})

	t.Run("HasRuleCode with string codes", func(t *testing.T) {
		tests := []struct {
			name string
			code string
			want bool
		}{
			{
				name: "rule exists - BR-1",
				code: "BR-01",
				want: true,
			},
			{
				name: "rule exists - BR-S-8",
				code: "BR-S-08",
				want: true,
			},
			{
				name: "rule exists - BR-CO-10",
				code: "BR-CO-10",
				want: true,
			},
			{
				name: "rule does not exist",
				code: "BR-99",
				want: false,
			},
			{
				name: "empty rule",
				code: "",
				want: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := e.HasRuleCode(tt.code); got != tt.want {
					t.Errorf("HasRuleCode(%v) = %v, want %v", tt.code, got, tt.want)
				}
			})
		}
	})
}

func TestValidationError_AsError(t *testing.T) {
	t.Run("can be used with errors.As", func(t *testing.T) {
		originalErr := &ValidationError{
			violations: []SemanticError{
				{Rule: rules.BR1, Text: "Test violation"},
			},
		}

		var err error = originalErr

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Error("errors.As failed to extract ValidationError")
		}

		if valErr.Count() != 1 {
			t.Errorf("Count() = %d, want 1", valErr.Count())
		}

		if !valErr.HasRule(rules.BR1) {
			t.Error("HasRule(rules.BR1) = false, want true")
		}

		if !valErr.HasRuleCode("BR-01") {
			t.Error("HasRuleCode(BR-1) = false, want true")
		}
	})
}

func TestValidationError_Warnings(t *testing.T) {
	t.Run("returns copy of warnings", func(t *testing.T) {
		original := []SemanticError{
			{Rule: rules.BRDE21, Text: "Test warning"},
		}
		e := &ValidationError{warnings: original}

		// Get warnings
		warnings := e.Warnings()

		// Verify content
		if len(warnings) != 1 {
			t.Errorf("Warnings() returned %d warnings, want 1", len(warnings))
		}
		if warnings[0].Rule.Code != "BR-DE-21" {
			t.Errorf("Warnings()[0].Rule.Code = %v, want BR-DE-21", warnings[0].Rule.Code)
		}

		// Modify the returned slice - should not affect internal state
		warnings[0].Rule = rules.BR2

		// Verify internal state unchanged
		if e.warnings[0].Rule.Code != "BR-DE-21" {
			t.Errorf("Internal warnings were modified, want BR-DE-21, got %v", e.warnings[0].Rule.Code)
		}
	})

	t.Run("returns nil for nil warnings", func(t *testing.T) {
		e := &ValidationError{warnings: nil}
		warnings := e.Warnings()
		if warnings != nil {
			t.Errorf("Warnings() = %v, want nil", warnings)
		}
	})

	t.Run("returns empty slice for empty warnings", func(t *testing.T) {
		e := &ValidationError{warnings: []SemanticError{}}
		warnings := e.Warnings()
		if warnings == nil {
			t.Error("Warnings() = nil, want empty slice")
		}
		if len(warnings) != 0 {
			t.Errorf("Warnings() length = %d, want 0", len(warnings))
		}
	})
}

func TestValidationError_WarningCount(t *testing.T) {
	tests := []struct {
		name     string
		warnings []SemanticError
		want     int
	}{
		{
			name:     "no warnings",
			warnings: []SemanticError{},
			want:     0,
		},
		{
			name: "one warning",
			warnings: []SemanticError{
				{Rule: rules.BRDE21, Text: "Test"},
			},
			want: 1,
		},
		{
			name: "multiple warnings",
			warnings: []SemanticError{
				{Rule: rules.BRDE21, Text: "Test 1"},
				{Rule: rules.BRDE17, Text: "Test 2"},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{warnings: tt.warnings}
			if got := e.WarningCount(); got != tt.want {
				t.Errorf("WarningCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_HasWarnings(t *testing.T) {
	t.Run("returns true when warnings exist", func(t *testing.T) {
		e := &ValidationError{warnings: []SemanticError{{Rule: rules.BRDE21, Text: "Test"}}}
		if !e.HasWarnings() {
			t.Error("HasWarnings() = false, want true")
		}
	})

	t.Run("returns false when no warnings", func(t *testing.T) {
		e := &ValidationError{warnings: []SemanticError{}}
		if e.HasWarnings() {
			t.Error("HasWarnings() = true, want false")
		}
	})

	t.Run("returns false for nil warnings", func(t *testing.T) {
		e := &ValidationError{warnings: nil}
		if e.HasWarnings() {
			t.Error("HasWarnings() = true, want false")
		}
	})
}

func TestValidationError_BothViolationsAndWarnings(t *testing.T) {
	e := &ValidationError{
		violations: []SemanticError{
			{Rule: rules.BR1, Text: "Error 1"},
			{Rule: rules.BR2, Text: "Error 2"},
		},
		warnings: []SemanticError{
			{Rule: rules.BRDE21, Text: "Warning 1"},
		},
	}

	if e.Count() != 2 {
		t.Errorf("Count() = %d, want 2", e.Count())
	}

	if e.WarningCount() != 1 {
		t.Errorf("WarningCount() = %d, want 1", e.WarningCount())
	}

	if !e.HasWarnings() {
		t.Error("HasWarnings() = false, want true")
	}

	if !e.HasRule(rules.BR1) {
		t.Error("HasRule(rules.BR1) = false, want true")
	}

	// Verify Error() message doesn't include warnings count
	errMsg := e.Error()
	if errMsg != "validation failed with 2 violations (first: BR-01 - Error 1)" {
		t.Errorf("Error() = %q, want proper violation-only message", errMsg)
	}
}

func TestInvoice_Warnings(t *testing.T) {
	t.Run("returns copy of warnings", func(t *testing.T) {
		inv := &Invoice{
			warnings: []SemanticError{
				{Rule: rules.BRDE21, Text: "Test warning"},
			},
		}

		warnings := inv.Warnings()

		if len(warnings) != 1 {
			t.Errorf("Warnings() returned %d warnings, want 1", len(warnings))
		}

		// Modify the returned slice
		warnings[0].Rule = rules.BR2

		// Verify internal state unchanged
		if inv.warnings[0].Rule.Code != "BR-DE-21" {
			t.Errorf("Internal warnings were modified")
		}
	})

	t.Run("returns nil for nil warnings", func(t *testing.T) {
		inv := &Invoice{warnings: nil}
		if inv.Warnings() != nil {
			t.Error("Warnings() should return nil for nil warnings")
		}
	})
}

func TestInvoice_HasWarnings(t *testing.T) {
	t.Run("returns true when warnings exist", func(t *testing.T) {
		inv := &Invoice{warnings: []SemanticError{{Rule: rules.BRDE21, Text: "Test"}}}
		if !inv.HasWarnings() {
			t.Error("HasWarnings() = false, want true")
		}
	})

	t.Run("returns false when no warnings", func(t *testing.T) {
		inv := &Invoice{warnings: []SemanticError{}}
		if inv.HasWarnings() {
			t.Error("HasWarnings() = true, want false")
		}
	})
}

func TestValidate_WarningsDoNotFailValidation(t *testing.T) {
	// Create a valid invoice that should pass validation
	// but receive a warning for BR-DE-21 (German seller not using XRechnung)
	inv := createValidGermanInvoice()

	// Use pure EN 16931 profile (not XRechnung) - should trigger BR-DE-21 warning
	inv.GuidelineSpecifiedDocumentContextParameter = SpecEN16931

	err := inv.Validate()

	// Validation should pass (no error)
	if err != nil {
		t.Errorf("Validate() returned error: %v, want nil", err)
	}

	// But warnings should exist
	if !inv.HasWarnings() {
		t.Error("HasWarnings() = false, want true")
	}

	warnings := inv.Warnings()
	if len(warnings) != 1 {
		t.Errorf("Warnings() returned %d warnings, want 1", len(warnings))
	}

	if warnings[0].Rule.Code != "BR-DE-21" {
		t.Errorf("Warnings()[0].Rule.Code = %s, want BR-DE-21", warnings[0].Rule.Code)
	}
}

func TestValidate_WarningsIncludedInValidationError(t *testing.T) {
	// Create an invoice that has both errors AND should trigger warnings
	inv := createMinimalInvoice()

	// Set German seller to trigger BR-DE-21 warning
	inv.Seller.PostalAddress = &PostalAddress{CountryID: "DE", City: "Berlin", PostcodeCode: "10115"}

	// Use pure EN 16931 profile (not XRechnung)
	inv.GuidelineSpecifiedDocumentContextParameter = SpecEN16931

	// Remove required field to cause a violation
	inv.InvoiceNumber = "" // This should trigger BR-02

	err := inv.Validate()

	// Should have validation error
	if err == nil {
		t.Fatal("Validate() returned nil, want error")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatal("Error is not a ValidationError")
	}

	// Check violations exist
	if valErr.Count() == 0 {
		t.Error("Count() = 0, want at least 1 violation")
	}

	// Check warnings are included in ValidationError
	if valErr.WarningCount() == 0 {
		t.Error("WarningCount() = 0, want at least 1 warning")
	}

	// Verify BR-DE-21 warning is present
	foundBRDE21 := false
	for _, w := range valErr.Warnings() {
		if w.Rule.Code == "BR-DE-21" {
			foundBRDE21 = true
			break
		}
	}
	if !foundBRDE21 {
		t.Error("BR-DE-21 warning not found in ValidationError.Warnings()")
	}
}

func TestValidate_ClearsWarningsOnRevalidation(t *testing.T) {
	// Create invoice with German seller
	inv := createValidGermanInvoice()
	inv.GuidelineSpecifiedDocumentContextParameter = SpecEN16931

	// First validation - should have warning
	_ = inv.Validate()
	if !inv.HasWarnings() {
		t.Error("First Validate(): HasWarnings() = false, want true")
	}

	// Change to XRechnung profile - warning should no longer apply
	inv.GuidelineSpecifiedDocumentContextParameter = SpecXRechnung30

	// Second validation - warnings should be cleared
	_ = inv.Validate()

	// Warnings should be cleared (XRechnung invoices don't get BR-DE-21 warning)
	if inv.HasWarnings() {
		t.Errorf("Second Validate(): HasWarnings() = true, want false (got %v)", inv.Warnings())
	}
}

// createMinimalInvoice creates a minimal invoice for testing
func createMinimalInvoice() *Invoice {
	return &Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		InvoiceDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{DueDate: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)},
		},
		Seller: Party{
			Name: "Seller Company",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
			VATaxRegistration: "FR12345678901",
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				Total:                    decimal.NewFromInt(100),
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				ItemName:                 "Test Item",
				NetPrice:                 decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				TypeCode:         "VAT",
				CategoryCode:     "S",
				BasisAmount:      decimal.NewFromInt(100),
				Percent:          decimal.NewFromInt(19),
				CalculatedAmount: decimal.NewFromFloat(19),
			},
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		TaxTotal:         decimal.NewFromFloat(19),
		GrandTotal:       decimal.NewFromFloat(119),
		DuePayableAmount: decimal.NewFromFloat(119),
	}
}

// createValidGermanInvoice creates a valid invoice with a German seller
func createValidGermanInvoice() *Invoice {
	inv := createMinimalInvoice()
	inv.Seller.PostalAddress = &PostalAddress{
		CountryID:    "DE",
		City:         "Berlin",
		PostcodeCode: "10115",
	}
	return inv
}

// Benchmark tests for validation performance

// BenchmarkValidate benchmarks validation performance across different profile types
func BenchmarkValidate(b *testing.B) {
	benchmarks := []struct {
		name string
		file string
	}{
		{"EN16931_CII", "testdata/cii/en16931/zugferd_2p3_EN16931_1.xml"},
		{"PEPPOL_UBL", "testdata/ubl/invoice/UBL-Invoice-2.1-Example.xml"},
		{"Extended_CII", "testdata/cii/extended/zugferd-extended-rechnung.xml"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data, err := os.ReadFile(bm.file)
			if err != nil {
				b.Skipf("File not found: %s", bm.file)
			}

			inv, err := ParseReader(bytes.NewReader(data))
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				err := inv.Validate()
				_ = err // Validation may or may not return errors
			}
		})
	}
}

// FuzzValidate fuzzes the validation logic to ensure it never panics
func FuzzValidate(f *testing.F) {
	// Seed corpus with valid invoices that should validate successfully
	seeds := []string{
		"testdata/cii/minimum/zugferd-minimum-rechnung.xml",
		"testdata/cii/en16931/zugferd_2p3_EN16931_1.xml",
		"testdata/ubl/invoice/UBL-Invoice-2.1-Example.xml",
	}

	for _, seed := range seeds {
		data, err := os.ReadFile(seed)
		if err == nil {
			f.Add(data)
		}
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parse the data first (may fail, which is fine)
		inv, err := ParseReader(bytes.NewReader(data))
		if err != nil {
			return // Skip invalid XML
		}

		// Validation should never panic, even with invalid/malformed data
		err = inv.Validate()
		_ = err // Validation errors are expected for malformed invoices
	})
}
