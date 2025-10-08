package einvoice

import (
	"testing"

	"strings"
	"github.com/shopspring/decimal"
)

func TestBRIP1_MissingSellerVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "M",
			},
		},
		// Seller VAT missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-1 violation for missing seller VAT in IPSI")
	}
}

func TestBRIP5_TaxableAmountMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "M",
				Total:           decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "M",
				BasisAmount:  decimal.NewFromFloat(90.0), // Should be 100.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-5 violation for taxable amount mismatch in IPSI")
	}
}

func TestBRIP5_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for rounding bug in IPSI category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "M",
				Total:           decimal.NewFromFloat(88.889),
			},
			{
				TaxCategoryCode: "M",
				Total:           decimal.NewFromFloat(11.112),
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"M": "IPSI"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-AG-05") {
			t.Error("BR-AG-05 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRIP6_VATAmountMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "M",
				BasisAmount:      decimal.NewFromFloat(100.0),
				Percent:          decimal.NewFromFloat(10.0),
				CalculatedAmount: decimal.NewFromFloat(15.0), // Should be 10.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-6 violation for VAT amount mismatch in IPSI")
	}
}

func TestBRIP7_TaxableAmountByRateMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "M",
				TaxRateApplicablePercent: decimal.NewFromFloat(10.0),
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "M",
				Percent:      decimal.NewFromFloat(10.0),
				BasisAmount:  decimal.NewFromFloat(80.0), // Should be 100.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-7 violation for taxable amount by rate mismatch in IPSI")
	}
}

func TestBRIP7_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for per-rate rounding bug in IPSI category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "M",
				TaxRateApplicablePercent: decimal.NewFromFloat(10.0),
				Total:                    decimal.NewFromFloat(44.445),
			},
			{
				TaxCategoryCode:          "M",
				TaxRateApplicablePercent: decimal.NewFromFloat(10.0),
				Total:                    decimal.NewFromFloat(55.556),
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"M": "IPSI"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-AG-07") {
			t.Error("BR-AG-07 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRIP8_VATAmountByRateMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "M",
				BasisAmount:      decimal.NewFromFloat(100.0),
				Percent:          decimal.NewFromFloat(10.0),
				CalculatedAmount: decimal.NewFromFloat(15.0), // Should be 10.0
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-8 violation for VAT amount by rate mismatch in IPSI")
	}
}

func TestBRIP9_HasExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:    "M",
				ExemptionReason: "Should not be present", // Must not have
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-9 violation for having exemption reason in IPSI")
	}
}

func TestBRIP10_MissingSellerTaxID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "M",
			},
		},
		// Seller VAT and tax registration missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-10" && strings.Contains(v.Text, "seller") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-10 violation for missing seller tax ID in IPSI")
	}
}

func TestBRIP10_HasBuyerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "M",
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
		Buyer:  Party{VATaxRegistration: "DE456"}, // Must not have
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-10" && strings.Contains(v.Text, "buyer") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AG-10 violation for having buyer VAT ID in IPSI")
	}
}

func TestBRIP10_ValidIPSI(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "M",
			},
		},
		Seller: Party{VATaxRegistration: "ES123"},
		// Buyer without VAT ID is OK
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AG-10" {
			t.Errorf("Should not have BR-AG-10 violation when seller has VAT ID and buyer has no VAT ID: %v", v)
		}
	}
}

// BR-O tests (Not subject to VAT)
