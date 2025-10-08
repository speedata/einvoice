package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBRO1_MissingBothTaxIDs(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
			},
		},
		// Both seller and buyer tax IDs missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-1 violation for missing both seller and buyer tax IDs in Not subject to VAT")
	}
}

func TestBRO1_HasSellerTaxID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-01" {
			t.Error("Should not have BR-O-1 violation when seller has tax ID")
		}
	}
}

func TestBRO1_HasBuyerTaxID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
			},
		},
		Buyer: Party{VATaxRegistration: "DE456"},
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-01" {
			t.Error("Should not have BR-O-1 violation when buyer has tax ID")
		}
	}
}

func TestBRO2_LineMissingSellerTaxID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
			},
		},
		Buyer: Party{VATaxRegistration: "DE456"}, // Has buyer but not seller
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-02" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-2 violation for line missing seller tax ID in Not subject to VAT")
	}
}

func TestBRO3_LineMissingVATBreakdown(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
			},
		},
		// No VAT breakdown with category O
		Seller: Party{VATaxRegistration: "DE123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-3 violation for line without corresponding VAT breakdown")
	}
}

func TestBRO6_NonZeroRateInLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "O",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-6 violation for non-zero VAT rate in Not subject to VAT line")
	}
}

func TestBRO9_TaxableAmountMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
				Total:           decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "O",
				BasisAmount:  decimal.NewFromFloat(90.0), // Should be 100.0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-9 violation for taxable amount mismatch in Not subject to VAT")
	}
}

func TestBRO9_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for rounding bug in Not subject to VAT category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
				Total:           decimal.NewFromFloat(123.456),
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"O": "Not subject to VAT"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-O-09") {
			t.Error("BR-O-09 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRO10_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for per-rate rounding bug in Not subject to VAT category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "O",
				TaxRateApplicablePercent: decimal.NewFromFloat(0.0),
				Total:                    decimal.NewFromFloat(25.003),
			},
			{
				TaxCategoryCode:          "O",
				TaxRateApplicablePercent: decimal.NewFromFloat(0.0),
				Total:                    decimal.NewFromFloat(74.998),
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"O": "Not subject to VAT"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-O-10") {
			t.Error("BR-O-10 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRO11_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-11" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-11 violation for non-zero VAT amount in Not subject to VAT")
	}
}

func TestBRO13_MissingExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "O",
				// ExemptionReason and ExemptionReasonCode missing
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-13" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-13 violation for missing exemption reason in Not subject to VAT")
	}
}

func TestBRO14_MultipleOCategories(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "O",
			},
			{
				CategoryCode: "O",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-14" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-O-14 violation for multiple Not subject to VAT categories")
	}
}

func TestBRO_ValidNotSubjectToVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "O",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				Percent:          decimal.Zero,
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
	}

	_ = inv.Validate()

	brORules := []string{"BR-O-01", "BR-O-02", "BR-O-03", "BR-O-06", "BR-O-09", "BR-O-11", "BR-O-13", "BR-O-14"}
	for _, v := range inv.violations {
		for _, rule := range brORules {
			if v.Rule.Code == rule {
				t.Errorf("Should not have %s violation for valid Not subject to VAT invoice: %v", rule, v)
			}
		}
	}
}
