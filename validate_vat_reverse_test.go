package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBRAE1_MissingReverseChargeBreakdown(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "AE",
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
		if v.Rule.Code == "BR-AE-01" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-1 violation for missing Reverse charge VAT breakdown")
	}
}

func TestBRAE2_MissingVATIDs(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "AE",
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123", // Has seller VAT
		},
		Buyer: Party{
			VATaxRegistration: "", // Missing buyer VAT
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-02" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-2 violation for missing buyer VAT ID")
	}
}

func TestBRAE3_AllowanceMissingVATIDs(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false,
				CategoryTradeTaxCategoryCode: "AE",
			},
		},
		Seller: Party{
			VATaxRegistration: "", // Missing
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-3 violation for Reverse charge allowance without seller VAT")
	}
}

func TestBRAE4_ChargeMissingVATIDs(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              true,
				CategoryTradeTaxCategoryCode: "AE",
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "", // Missing
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-04" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-4 violation for Reverse charge charge without buyer VAT")
	}
}

func TestBRAE5_NonZeroRateInLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "AE",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-5 violation for non-zero VAT rate in Reverse charge line")
	}
}

func TestBRAE6_NonZeroRateInAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				CategoryTradeTaxCategoryCode:          "AE",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-6 violation for non-zero VAT rate in Reverse charge allowance")
	}
}

func TestBRAE7_NonZeroRateInCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				CategoryTradeTaxCategoryCode:          "AE",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-7 violation for non-zero VAT rate in Reverse charge charge")
	}
}

func TestBRAE8_IncorrectTaxableAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "AE",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
				BasisAmount:  decimal.NewFromFloat(50.0), // Wrong, should be 100
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-8 violation for incorrect taxable amount")
	}
}

func TestBRAE9_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "AE",
				TaxRateApplicablePercent: decimal.Zero,
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "AE",
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-9 violation for non-zero VAT amount in Reverse charge")
	}
}

func TestBRAE10_MissingExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "AE",
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		Seller: Party{
			VATaxRegistration: "DE123",
		},
		Buyer: Party{
			VATaxRegistration: "FR456",
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "AE",
				// Missing ExemptionReason and ExemptionReasonCode
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-AE-10" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-AE-10 violation for missing exemption reason in Reverse charge")
	}
}
