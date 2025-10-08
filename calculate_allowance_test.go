package einvoice

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// mustParseDate is a helper for tests
func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

// TestUpdateApplicableTradeTax_AllowanceWithoutLineItems tests the fix for
// the critical edge case where document-level allowances exist without
// corresponding line items, which previously created negative VAT basis amounts.
func TestUpdateApplicableTradeTax_AllowanceWithoutLineItems(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{}, // No line items!
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				ActualAmount:                          decimal.NewFromInt(100),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)

	// Should create a tax entry with ZERO basis (not negative)
	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 trade tax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]
	if tt.CategoryCode != "S" {
		t.Errorf("Expected category S, got %s", tt.CategoryCode)
	}

	// Critical: basis should be ZERO, not negative
	if !tt.BasisAmount.IsZero() {
		t.Errorf("Expected zero basis amount for allowance without line items, got %s (should not be negative!)",
			tt.BasisAmount.String())
	}

	// Calculated amount should also be zero (since basis is zero)
	if !tt.CalculatedAmount.IsZero() {
		t.Errorf("Expected zero calculated amount, got %s", tt.CalculatedAmount.String())
	}
}

// TestUpdateApplicableTradeTax_ChargeWithoutLineItems tests that document-level
// charges without line items correctly create a positive basis amount.
func TestUpdateApplicableTradeTax_ChargeWithoutLineItems(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{}, // No line items
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true, // Charge
				ActualAmount:                          decimal.NewFromInt(50),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)

	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 trade tax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]

	// Charge should create positive basis
	expectedBasis := decimal.NewFromInt(50)
	if !tt.BasisAmount.Equal(expectedBasis) {
		t.Errorf("Expected basis amount %s for standalone charge, got %s",
			expectedBasis.String(), tt.BasisAmount.String())
	}

	// Calculated VAT should be 50 * 19% = 9.50
	expectedVAT := decimal.RequireFromString("9.50")
	if !tt.CalculatedAmount.Equal(expectedVAT) {
		t.Errorf("Expected calculated amount %s, got %s",
			expectedVAT.String(), tt.CalculatedAmount.String())
	}
}

// TestUpdateApplicableTradeTax_AllowanceWithLineItems tests the normal case
// where allowances have corresponding line items.
func TestUpdateApplicableTradeTax_AllowanceWithLineItems(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				Total:                    decimal.NewFromInt(200),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				ActualAmount:                          decimal.NewFromInt(20),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)

	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 trade tax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]

	// Basis should be: line total - allowance = 200 - 20 = 180
	expectedBasis := decimal.NewFromInt(180)
	if !tt.BasisAmount.Equal(expectedBasis) {
		t.Errorf("Expected basis amount %s (line total - allowance), got %s",
			expectedBasis.String(), tt.BasisAmount.String())
	}

	// VAT should be 180 * 19% = 34.20
	expectedVAT := decimal.RequireFromString("34.20")
	if !tt.CalculatedAmount.Equal(expectedVAT) {
		t.Errorf("Expected calculated amount %s, got %s",
			expectedVAT.String(), tt.CalculatedAmount.String())
	}
}

// TestValidate_AllowanceWithoutLineItems tests that validation detects
// the edge case of document-level allowances without line items.
func TestValidate_AllowanceWithoutLineItems(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         mustParseDate("2024-01-15"),
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		LineTotal:        decimal.Zero,
		TaxBasisTotal:    decimal.Zero,
		TaxTotal:         decimal.Zero,
		GrandTotal:       decimal.Zero,
		DuePayableAmount: decimal.Zero,
		InvoiceLines:     []InvoiceLine{}, // No line items!
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				ActualAmount:                          decimal.NewFromInt(100),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for allowance without line items, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Should have a violation for allowance without line items
	violations := valErr.Violations()
	foundAllowanceViolation := false
	for _, v := range violations {
		if v.Rule.Code == rules.Check.Code {
			if contains(v.Text, "Document-level allowance") && contains(v.Text, "no corresponding invoice lines") {
				foundAllowanceViolation = true
				t.Logf("Found expected violation: %s", v.Text)
				break
			}
		}
	}

	if !foundAllowanceViolation {
		t.Errorf("Expected violation for document-level allowance without line items, but not found. Got %d violations", len(violations))
		for _, v := range violations {
			t.Logf("  - %s: %s", v.Rule.Code, v.Text)
		}
	}
}

// TestValidate_ChargeWithoutLineItems tests that standalone charges are allowed
// (they don't require line items).
func TestValidate_ChargeWithoutLineItems(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         mustParseDate("2024-01-15"),
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		LineTotal:        decimal.Zero,
		TaxBasisTotal:    decimal.NewFromInt(50),
		TaxTotal:         decimal.RequireFromString("9.50"),
		GrandTotal:       decimal.RequireFromString("59.50"),
		DuePayableAmount: decimal.RequireFromString("59.50"),
		InvoiceLines:     []InvoiceLine{}, // No line items
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true, // Charge (OK without line items)
				ActualAmount:                          decimal.NewFromInt(50),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(50),
				CalculatedAmount: decimal.RequireFromString("9.50"),
				Typ:              "VAT",
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{Description: "Due on receipt"},
		},
	}

	err := inv.Validate()

	// Should NOT have violation for standalone charge
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			for _, v := range valErr.Violations() {
				if contains(v.Text, "Document-level") && contains(v.Text, "no corresponding invoice lines") {
					t.Errorf("Unexpected violation for standalone charge: %s", v.Text)
				}
			}
		}
	}
}

// TestUpdateApplicableTradeTax_MixedCategoriesAllowances tests a complex scenario
// with both allowances and charges across multiple categories.
func TestUpdateApplicableTradeTax_MixedCategoriesAllowances(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				Total:                    decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
			{
				Total:                    decimal.NewFromInt(50),
				TaxCategoryCode:          "E",
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				// Allowance for category with line items (OK)
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromInt(10),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
			},
			{
				// Charge for category without line items (creates standalone - OK)
				ChargeIndicator:                       true,
				ActualAmount:                          decimal.NewFromInt(20),
				CategoryTradeTaxCategoryCode:          "Z",
				CategoryTradeTaxRateApplicablePercent: decimal.Zero,
			},
			{
				// Allowance for category without line items (creates zero basis)
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromInt(5),
				CategoryTradeTaxCategoryCode:          "G",
				CategoryTradeTaxRateApplicablePercent: decimal.Zero,
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)

	// Should have 4 tax entries: S, E, Z, G
	if len(inv.TradeTaxes) != 4 {
		t.Fatalf("Expected 4 trade tax entries, got %d", len(inv.TradeTaxes))
	}

	// Find each category and verify basis
	for _, tt := range inv.TradeTaxes {
		switch tt.CategoryCode {
		case "S":
			// 100 (line) - 10 (allowance) = 90
			expected := decimal.NewFromInt(90)
			if !tt.BasisAmount.Equal(expected) {
				t.Errorf("Category S: expected basis %s, got %s", expected.String(), tt.BasisAmount.String())
			}
		case "E":
			// 50 (line only, no allowances/charges)
			expected := decimal.NewFromInt(50)
			if !tt.BasisAmount.Equal(expected) {
				t.Errorf("Category E: expected basis %s, got %s", expected.String(), tt.BasisAmount.String())
			}
		case "Z":
			// 20 (standalone charge)
			expected := decimal.NewFromInt(20)
			if !tt.BasisAmount.Equal(expected) {
				t.Errorf("Category Z: expected basis %s, got %s", expected.String(), tt.BasisAmount.String())
			}
		case "G":
			// 0 (allowance without line items - fixed to zero instead of -5)
			if !tt.BasisAmount.IsZero() {
				t.Errorf("Category G: expected zero basis for allowance without lines, got %s", tt.BasisAmount.String())
			}
		default:
			t.Errorf("Unexpected category: %s", tt.CategoryCode)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
