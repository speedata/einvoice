package einvoice

import (
	"testing"

	"strings"
	"github.com/shopspring/decimal"
)

func TestBRIG1_MissingSellerVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "L",
			},
		},
		// Seller VAT missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-1 violation for missing seller VAT in IGIC")
	}
}

func TestBRIG5_TaxableAmountMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "L",
				Total:           decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "L",
				BasisAmount:  decimal.NewFromFloat(90.0), // Should be 100.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-5 violation for taxable amount mismatch in IGIC")
	}
}

func TestBRIG5_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for rounding bug in IGIC category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "L",
				Total:           decimal.NewFromFloat(75.334),
			},
			{
				TaxCategoryCode: "L",
				Total:           decimal.NewFromFloat(24.667),
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"L": "IGIC"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-AF-05") {
			t.Error("BR-AF-05 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRIG6_VATAmountMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "L",
				BasisAmount:      decimal.NewFromFloat(100.0),
				Percent:          decimal.NewFromFloat(7.0),
				CalculatedAmount: decimal.NewFromFloat(10.0), // Should be 7.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-6 violation for VAT amount mismatch in IGIC")
	}
}

func TestBRIG7_TaxableAmountByRateMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "L",
				TaxRateApplicablePercent: decimal.NewFromFloat(7.0),
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "L",
				Percent:      decimal.NewFromFloat(7.0),
				BasisAmount:  decimal.NewFromFloat(80.0), // Should be 100.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-7 violation for taxable amount by rate mismatch in IGIC")
	}
}

func TestBRIG7_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for per-rate rounding bug in IGIC category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "L",
				TaxRateApplicablePercent: decimal.NewFromFloat(7.0),
				Total:                    decimal.NewFromFloat(33.334),
			},
			{
				TaxCategoryCode:          "L",
				TaxRateApplicablePercent: decimal.NewFromFloat(7.0),
				Total:                    decimal.NewFromFloat(66.667),
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"L": "IGIC"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-AF-07") {
			t.Error("BR-AF-07 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRIG8_VATAmountByRateMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "L",
				BasisAmount:      decimal.NewFromFloat(100.0),
				Percent:          decimal.NewFromFloat(7.0),
				CalculatedAmount: decimal.NewFromFloat(10.0), // Should be 7.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-8 violation for VAT amount by rate mismatch in IGIC")
	}
}

func TestBRIG9_HasExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:    "L",
				ExemptionReason: "Should not be present", // Must not have
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-9 violation for having exemption reason in IGIC")
	}
}

func TestBRIG10_MissingSellerTaxID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "L",
			},
		},
		// Seller VAT and tax registration missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-10" && strings.Contains(v.Text, "seller") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-10 violation for missing seller tax ID in IGIC")
	}
}

func TestBRIG10_HasBuyerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "L",
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
		Buyer:  Party{VATaxRegistration: "DE456"}, // Must not have
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-10" && strings.Contains(v.Text, "buyer") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AF-10 violation for having buyer VAT ID in IGIC")
	}
}

func TestBRIG10_ValidIGIC(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "L",
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
		// Buyer without VAT ID is OK
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AF-10" {
			t.Errorf("Should not have BR-AF-10 violation when seller has VAT ID and buyer has no VAT ID: %v", v)
		}
	}
}

// BR-IP tests (IPSI - Ceuta/Melilla)
