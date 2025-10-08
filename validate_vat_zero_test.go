package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBRZ1_MissingZeroRatedBreakdown(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "Z",
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
		if v.Rule.Code == "BR-Z-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-1 violation for missing Zero rated VAT breakdown")
	}
}

func TestBRZ2_MissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "Z",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-02" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-2 violation for missing seller VAT ID")
	}
}

func TestBRZ3_AllowanceMissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false,
				CategoryTradeTaxCategoryCode: "Z",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-3 violation for Zero rated allowance without seller VAT")
	}
}

func TestBRZ4_ChargeMissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              true,
				CategoryTradeTaxCategoryCode: "Z",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-04" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-4 violation for Zero rated charge without seller VAT")
	}
}

func TestBRZ5_NonZeroRateInLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "Z",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-5 violation for non-zero VAT rate in Zero rated line")
	}
}

func TestBRZ6_NonZeroRateInAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				CategoryTradeTaxCategoryCode:          "Z",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-6 violation for non-zero VAT rate in Zero rated allowance")
	}
}

func TestBRZ7_NonZeroRateInCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				CategoryTradeTaxCategoryCode:          "Z",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-7 violation for non-zero VAT rate in Zero rated charge")
	}
}

func TestBRZ8_IncorrectTaxableAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "Z",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "Z",
				BasisAmount:  decimal.NewFromFloat(50.0), // Wrong, should be 100
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-8 violation for incorrect taxable amount")
	}
}

func TestBRZ9_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "Z",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "Z",
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-9 violation for non-zero VAT amount in Zero rated")
	}
}

func TestBRZ10_ExemptionReasonPresent(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "Z",
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:    "Z",
				ExemptionReason: "Some reason", // Should not be present
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-Z-10" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-Z-10 violation for exemption reason in Zero rated")
	}
}
