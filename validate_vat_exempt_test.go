package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBRE1_MissingExemptBreakdown(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "E",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "S", // Wrong category
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-1 violation for missing Exempt from VAT breakdown")
	}
}

func TestBRE2_MissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "E",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-02" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-2 violation for missing seller VAT ID")
	}
}

func TestBRE3_AllowanceMissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false,
				CategoryTradeTaxCategoryCode: "E",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-3 violation for Exempt allowance without seller VAT")
	}
}

func TestBRE4_ChargeMissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              true,
				CategoryTradeTaxCategoryCode: "E",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-04" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-4 violation for Exempt charge without seller VAT")
	}
}

func TestBRE5_NonZeroRateInLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "E",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-5 violation for non-zero VAT rate in Exempt line")
	}
}

func TestBRE6_NonZeroRateInAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				CategoryTradeTaxCategoryCode:          "E",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-6 violation for non-zero VAT rate in Exempt allowance")
	}
}

func TestBRE7_NonZeroRateInCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				CategoryTradeTaxCategoryCode:          "E",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-7 violation for non-zero VAT rate in Exempt charge")
	}
}

func TestBRE8_IncorrectTaxableAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "E",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
				BasisAmount:  decimal.NewFromFloat(50.0), // Wrong, should be 100
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-8 violation for incorrect taxable amount")
	}
}

func TestBRE9_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "E",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "E",
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-9 violation for non-zero VAT amount in Exempt")
	}
}

func TestBRE10_MissingExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "E",
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "E",
				// Missing ExemptionReason and ExemptionReasonCode
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-E-10" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-E-10 violation for missing exemption reason in Exempt")
	}
}
