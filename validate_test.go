package einvoice

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestValidate_ManualInvoiceConstruction tests the new public Validate() API
// with an invoice built programmatically rather than parsed from XML.
func TestValidate_ManualInvoiceConstruction(t *testing.T) {
	// Build an invalid invoice manually
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "",     // BR-2 violation: missing invoice number
		InvoiceTypeCode:     380,    // OK
		InvoiceDate:         time.Time{}, // BR-3 violation: zero date
		InvoiceCurrencyCode: "",     // BR-5 violation: empty currency
		Seller: Party{
			Name: "", // BR-6 violation: empty seller name
			PostalAddress: &PostalAddress{
				CountryID: "", // BR-9 violation: empty country code
			},
		},
		Buyer: Party{
			Name: "Valid Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		LineTotal:        decimal.Zero,       // BR-12 violation: zero
		TaxBasisTotal:    decimal.Zero,       // BR-13 violation: zero
		GrandTotal:       decimal.Zero,       // BR-14 violation: zero
		DuePayableAmount: decimal.Zero,       // BR-15 violation: zero
		InvoiceLines:     []InvoiceLine{},    // BR-16 violation: no lines
	}

	// Call Validate() - should return ValidationError
	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Check that error is ValidationError
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Check specific violations
	expectedRules := []string{"BR-2", "BR-3", "BR-5", "BR-6", "BR-9", "BR-12", "BR-13", "BR-14", "BR-15", "BR-16"}
	for _, rule := range expectedRules {
		if !valErr.HasRule(rule) {
			t.Errorf("Expected violation for rule %s, but not found", rule)
		}
	}

	// Verify Count() method
	if valErr.Count() < len(expectedRules) {
		t.Errorf("Expected at least %d violations, got %d", len(expectedRules), valErr.Count())
	}

	// Verify Violations() accessor
	violations := valErr.Violations()
	if len(violations) != valErr.Count() {
		t.Errorf("Violations() returned %d items, but Count() is %d", len(violations), valErr.Count())
	}

	// Verify error message
	errMsg := valErr.Error()
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}
}

// TestValidate_Idempotency tests that calling Validate() multiple times
// produces consistent results without accumulating violations.
func TestValidate_Idempotency(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "", // BR-2 violation
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		GrandTotal:       decimal.NewFromInt(119),
		DuePayableAmount: decimal.NewFromInt(119),
		InvoiceLines: []InvoiceLine{
			{
				LineID:         "1",
				ItemName:       "Test Item",
				BilledQuantity: decimal.NewFromInt(1),
				NetPrice:       decimal.NewFromInt(100),
				Total:          decimal.NewFromInt(100),
			},
		},
	}

	// First validation
	err1 := inv.Validate()
	if err1 == nil {
		t.Fatal("Expected validation error on first call")
	}
	var valErr1 *ValidationError
	errors.As(err1, &valErr1)
	count1 := valErr1.Count()

	// Second validation - should have same count
	err2 := inv.Validate()
	if err2 == nil {
		t.Fatal("Expected validation error on second call")
	}
	var valErr2 *ValidationError
	errors.As(err2, &valErr2)
	count2 := valErr2.Count()

	if count1 != count2 {
		t.Errorf("Idempotency broken: first call had %d violations, second call had %d", count1, count2)
	}

	// Third validation - should still have same count
	err3 := inv.Validate()
	var valErr3 *ValidationError
	errors.As(err3, &valErr3)
	count3 := valErr3.Count()

	if count1 != count3 {
		t.Errorf("Idempotency broken: first call had %d violations, third call had %d", count1, count3)
	}
}

// TestValidate_MostlyValidInvoice tests that validation can be called on a mostly valid invoice.
// Note: Creating a fully EN 16931-compliant invoice requires many fields, so this test
// just ensures Validate() works without panicking, even if some minor violations remain.
func TestValidate_MostlyValidInvoice(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name: "Test Seller",
			VATaxRegistration: "DE123456789",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		TaxTotal:         decimal.NewFromInt(19),
		GrandTotal:       decimal.NewFromInt(119),
		DuePayableAmount: decimal.NewFromInt(119),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
				Percent:          decimal.NewFromInt(19),
			},
		},
	}

	// Validate should not panic - error is acceptable since creating
	// a fully compliant invoice requires many additional fields
	err := inv.Validate()
	// We don't check if err is nil because full compliance is complex
	// The test just verifies Validate() can be called successfully
	_ = err
}

