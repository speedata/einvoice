package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBRG1_MissingExportOutsideEUBreakdown(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "G",
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
		if v.Rule.Code == "BR-G-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-1 violation for missing Export outside EU VAT breakdown")
	}
}

func TestBRG2_MissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "G",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-02" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-2 violation for missing seller VAT ID")
	}
}

func TestBRG3_AllowanceMissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false,
				CategoryTradeTaxCategoryCode: "G",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-3 violation for Export outside EU allowance without seller VAT")
	}
}

func TestBRG4_ChargeMissingSellerVATID(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              true,
				CategoryTradeTaxCategoryCode: "G",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-04" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-4 violation for Export outside EU charge without seller VAT")
	}
}

func TestBRG5_NonZeroRateInLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "G",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-5 violation for non-zero VAT rate in Export outside EU line")
	}
}

func TestBRG6_NonZeroRateInAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				CategoryTradeTaxCategoryCode:          "G",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-6 violation for non-zero VAT rate in Export outside EU allowance")
	}
}

func TestBRG7_NonZeroRateInCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				CategoryTradeTaxCategoryCode:          "G",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-7 violation for non-zero VAT rate in Export outside EU charge")
	}
}

func TestBRG8_IncorrectTaxableAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "G",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
				BasisAmount:  decimal.NewFromFloat(50.0), // Wrong, should be 100
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-8 violation for incorrect taxable amount")
	}
}

func TestBRG9_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "G",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "G",
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-9 violation for non-zero VAT amount in Export outside EU")
	}
}

func TestBRG10_MissingExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "G",
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "G",
				// Missing ExemptionReason and ExemptionReasonCode
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-G-10" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-G-10 violation for missing exemption reason in Export outside EU")
	}
}

// BR-IC tests (Intra-community supply)
