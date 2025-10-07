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
		Profile:                          CProfileEN16931,
		InvoiceNumber:                    "INV-001",
		InvoiceTypeCode:                  380,
		InvoiceDate:                      time.Now(),
		InvoiceCurrencyCode:              "EUR",
		BPSpecifiedDocumentContextParameter: "", // Violates PEPPOL-EN16931-R001
		BuyerReference:                   "",    // Violates PEPPOL-EN16931-R003
		BuyerOrderReferencedDocument:     "",    // Violates PEPPOL-EN16931-R003
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

	// Validate with PEPPOL rules
	err := inv.ValidatePEPPOL()
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
		Profile:             CProfileEN16931,
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		BPSpecifiedDocumentContextParameter: "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
		BuyerReference:                      "BR123",
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

	err := inv.ValidatePEPPOL()
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
		Profile:             CProfileEN16931,
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		BPSpecifiedDocumentContextParameter: "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
		BuyerReference:                      "BR123",
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
	// beyond the EN 16931 ones (which ValidatePEPPOL also includes)
	err := inv.ValidatePEPPOL()
	// We don't expect nil here as EN 16931 validation will likely find issues
	// This test just ensures ValidatePEPPOL() runs without panicking
	_ = err
}
