package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBRS1_MissingStandardRatedBreakdown(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z", // Wrong category
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-01" {
			found = true
			// Per official EN 16931 spec, BR-S-1 references multiple fields
			// Check that key fields BG-23 and BT-118 are included
			hasBG23 := false
			hasBT118 := false
			for _, field := range v.Rule.Fields {
				if field == "BG-23" {
					hasBG23 = true
				}
				if field == "BT-118" {
					hasBT118 = true
				}
			}
			if !hasBG23 || !hasBT118 {
				t.Errorf("BR-S-1 should include BG-23 and BT-118, got %v", v.Rule.Fields)
			}
		}
	}

	if !found {
		t.Error("Expected BR-S-1 violation for missing Standard rated VAT breakdown")
	}
}

func TestBRS2_MissingSellerVATForStandardLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "S",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
			FCTaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-02" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-2 violation for missing seller VAT identifier")
	}
}

func TestBRS3_MissingSellerVATForStandardAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false, // Allowance
				CategoryTradeTaxCategoryCode: "S",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-3 violation for Standard rated allowance without seller VAT")
	}
}

func TestBRS4_MissingSellerVATForStandardCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              true, // Charge
				CategoryTradeTaxCategoryCode: "S",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-04" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-4 violation for Standard rated charge without seller VAT")
	}
}

func TestBRS5_ZeroRateInStandardLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.Zero, // Should be > 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123456789",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-5 violation for zero VAT rate in Standard rated line")
	}
}

func TestBRS6_ZeroRateInStandardAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.Zero, // Should be > 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123456789",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-6 violation for zero VAT rate in Standard rated allowance")
	}
}

func TestBRS7_ZeroRateInStandardCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.Zero, // Should be > 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123456789",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-7 violation for zero VAT rate in Standard rated charge")
	}
}

func TestBRS8_IncorrectTaxableAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0),
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123456789",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S",
				Percent:      decimal.NewFromFloat(19.0),
				BasisAmount:  decimal.NewFromFloat(50.0), // Wrong, should be 100
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-8 violation for incorrect taxable amount")
	}
}

func TestBRS9_IncorrectVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0),
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123456789",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19.0),
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.NewFromFloat(10.0), // Wrong, should be 19.00
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-9 violation for incorrect VAT amount")
	}
}

func TestBRS10_ExemptionReasonInStandardRated(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123456789",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:        "S",
				Percent:             decimal.NewFromFloat(19.0),
				ExemptionReason:     "Some reason", // Should not be present
				ExemptionReasonCode: "VATEX-EU-O",  // Should not be present
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-S-10" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-S-10 violation for exemption reason in Standard rated")
	}
}
