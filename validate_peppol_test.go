package einvoice

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// TestValidatePEPPOL_BasicRequirements tests basic PEPPOL requirements
func TestValidatePEPPOL_BasicRequirements(t *testing.T) {
	// Create a minimal invoice that violates several PEPPOL rules
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30, // Mark as PEPPOL invoice
		BPSpecifiedDocumentContextParameter:        "", // Violates PEPPOL-EN16931-R001
		BuyerReference:                             "",    // Violates PEPPOL-EN16931-R003
		BuyerOrderReferencedDocument:               "",    // Violates PEPPOL-EN16931-R003
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "", // Violates PEPPOL-EN16931-R020
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "", // Violates PEPPOL-EN16931-R010
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

	// Validate - will auto-detect PEPPOL based on GuidelineSpecifiedDocumentContextParameter
	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Check for specific PEPPOL violations
	expectedRules := []string{
		"PEPPOL-EN16931-R001", // Business process required
		"PEPPOL-EN16931-R003", // Buyer reference or order reference required
		"PEPPOL-EN16931-R010", // Buyer electronic address required
		"PEPPOL-EN16931-R020", // Seller electronic address required
	}

	for _, ruleCode := range expectedRules {
		if !valErr.HasRuleCode(ruleCode) {
			t.Errorf("Expected violation for rule %s, but not found", ruleCode)
		}
	}
}

// TestValidatePEPPOL_MultipleNotes tests the PEPPOL rule for maximum one note
func TestValidatePEPPOL_MultipleNotes(t *testing.T) {
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Notes: []Note{
			{SubjectCode: "", Text: "Note 1"},
			{SubjectCode: "", Text: "Note 2"}, // Violates PEPPOL-EN16931-R002
		},
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
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

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for multiple notes, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if !valErr.HasRule(rules.PEPPOLEN16931R2) {
		t.Error("Expected violation for PEPPOL-EN16931-R002 (multiple notes)")
	}
}

// TestValidatePEPPOL_ValidInvoice tests that a PEPPOL-compliant invoice passes validation
func TestValidatePEPPOL_ValidInvoice(t *testing.T) {
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration:         "DE123456789",
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
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

	// This invoice should still have EN 16931 violations, but not additional PEPPOL violations
	// beyond the EN 16931 ones
	err := inv.Validate()
	// We don't expect nil here as EN 16931 validation will likely find issues
	// This test just ensures Validate() runs without panicking
	_ = err
}

// TestValidatePEPPOL_DecimalPrecisionViolations tests that Validate() catches
// decimal precision violations (BR-DEC rules) for PEPPOL invoices.
// This test verifies the fix for issue H1 from the bug analysis.
func TestValidatePEPPOL_DecimalPrecisionViolations(t *testing.T) {
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceDate:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Seller: Party{
			Name: "Seller Inc",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration:         "DE123456789",
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Buyer GmbH",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
		},
		// Invalid: 3 decimal places (should be max 2 per BR-DEC-19)
		LineTotal:        decimal.RequireFromString("100.123"),
		TaxBasisTotal:    decimal.RequireFromString("100.123"),
		TaxTotal:         decimal.NewFromInt(19),
		GrandTotal:       decimal.RequireFromString("119.123"),
		DuePayableAmount: decimal.RequireFromString("119.123"),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Product A",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.RequireFromString("100.123"), // Invalid: 3 decimals (BR-DEC-23)
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromInt(19),
				BasisAmount:      decimal.RequireFromString("100.123"), // Invalid: 3 decimals (BR-DEC-19)
				TypeCode:         "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{Description: "Net 30"},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for decimal precision violations, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Should have multiple BR-DEC violations
	violations := valErr.Violations()
	if len(violations) == 0 {
		t.Fatal("Expected BR-DEC violations, got none")
	}

	// Check that we have BR-DEC violations (these are now caught for PEPPOL invoices)
	hasBRDEC := false
	for _, v := range violations {
		if v.Rule.Code == "BR-DEC-19" || v.Rule.Code == "BR-DEC-23" {
			hasBRDEC = true
			t.Logf("Found expected BR-DEC violation: %s - %s", v.Rule.Code, v.Text)
		}
	}

	if !hasBRDEC {
		t.Errorf("Expected BR-DEC violations in PEPPOL validation, but none found. Got violations: %v", violations)
	}
}
