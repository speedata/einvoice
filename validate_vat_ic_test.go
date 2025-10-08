package einvoice

import (
	"testing"

	"time"
	"strings"
	"github.com/shopspring/decimal"
)

func TestBRIC1_MissingSellerVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
			},
		},
		Buyer: Party{VATaxRegistration: "DE456"},
		// Seller VAT missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-01" && strings.Contains(v.Text, "seller") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-1 violation for missing seller VAT in Intra-community supply")
	}
}

func TestBRIC1_MissingBuyerVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		// Buyer VAT and legal ID missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-01" && strings.Contains(v.Text, "buyer") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-1 violation for missing buyer VAT in Intra-community supply")
	}
}

func TestBRIC1_BuyerLegalIDAccepted(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{SpecifiedLegalOrganization: &SpecifiedLegalOrganization{ID: "LEGAL123"}},
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-01" && strings.Contains(v.Text, "buyer") {
			t.Error("Should not have BR-IC-1 violation when buyer has legal registration ID")
		}
	}
}

func TestBRIC2_LineMissingSellerVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
			},
		},
		Buyer: Party{VATaxRegistration: "DE456"},
		// Seller VAT missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-02" && strings.Contains(v.Text, "seller") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-2 violation for missing seller VAT in Intra-community supply line")
	}
}

func TestBRIC2_LineMissingBuyerVAT(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		// Buyer VAT missing
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-02" && strings.Contains(v.Text, "buyer") {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-2 violation for missing buyer VAT in Intra-community supply line")
	}
}

func TestBRIC3_NonZeroRateInLine(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "K",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-03" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-3 violation for non-zero VAT rate in Intra-community supply line")
	}
}

func TestBRIC4_NonZeroRateInAllowance(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				CategoryTradeTaxCategoryCode:          "K",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-04" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-4 violation for non-zero VAT rate in Intra-community supply allowance")
	}
}

func TestBRIC5_NonZeroRateInCharge(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				CategoryTradeTaxCategoryCode:          "K",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-05" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-5 violation for non-zero VAT rate in Intra-community supply charge")
	}
}

func TestBRIC6_TaxableAmountMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
				Total:           decimal.NewFromFloat(100.0),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false,
				CategoryTradeTaxCategoryCode: "K",
				ActualAmount:                 decimal.NewFromFloat(10.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
				BasisAmount:  decimal.NewFromFloat(100.0), // Should be 90.0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-06" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-6 violation for taxable amount mismatch in Intra-community supply")
	}
}

func TestBRIC6_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for rounding bug: line totals that sum to a value
	// requiring rounding (e.g., 50.004 + 49.997 = 100.001 rounds to 100.00)
	// should not trigger false violations when properly rounded
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "K",
				Total:           decimal.NewFromFloat(50.004),
			},
			{
				TaxCategoryCode: "K",
				Total:           decimal.NewFromFloat(49.997),
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"K": "Intracommunity supply"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-IC-06") {
			t.Error("BR-IC-06 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRIC7_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "K",
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-07" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-7 violation for non-zero VAT amount in Intra-community supply")
	}
}

func TestBRIC8_TaxableAmountByRateMismatch(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "K",
				TaxRateApplicablePercent: decimal.NewFromFloat(0.0),
				Total:                    decimal.NewFromFloat(100.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
				Percent:      decimal.NewFromFloat(0.0),
				BasisAmount:  decimal.NewFromFloat(80.0), // Should be 100.0
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-08" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-8 violation for taxable amount by rate mismatch in Intra-community supply")
	}
}

func TestBRIC8_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for per-rate rounding bug
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "K",
				TaxRateApplicablePercent: decimal.NewFromFloat(0.0),
				Total:                    decimal.NewFromFloat(33.334),
			},
			{
				TaxCategoryCode:          "K",
				TaxRateApplicablePercent: decimal.NewFromFloat(0.0),
				Total:                    decimal.NewFromFloat(66.667),
			},
		},
		Seller: Party{VATaxRegistration: "DE123"},
		Buyer:  Party{VATaxRegistration: "DE456"},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"K": "Intracommunity supply"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-IC-08") {
			t.Error("BR-IC-08 false positive: rounding precision not handled correctly")
		}
	}
}

func TestBRIC9_NonZeroVATAmount(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "K",
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-09" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-9 violation for non-zero VAT amount in Intra-community supply")
	}
}

func TestBRIC10_MissingExemptionReason(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
				// ExemptionReason and ExemptionReasonCode missing
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-10" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-10 violation for missing exemption reason in Intra-community supply")
	}
}

func TestBRIC11_MissingDeliveryDate(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
			},
		},
		// OccurrenceDateTime, BillingSpecifiedPeriodStart, BillingSpecifiedPeriodEnd all zero
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-11" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-11 violation for missing delivery date in Intra-community supply")
	}
}

func TestBRIC11_HasDeliveryDate(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
			},
		},
		OccurrenceDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-11" {
			t.Error("Should not have BR-IC-11 violation when delivery date is present")
		}
	}
}

func TestBRIC12_MissingDeliverToCountry(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
			},
		},
		// ShipTo missing or has no CountryID
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-12" {
			found = true
		}
	}

	if !found {
		t.Error("Expected BR-IC-12 violation for missing deliver to country in Intra-community supply")
	}
}

func TestBRIC12_HasDeliverToCountry(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		TradeTaxes: []TradeTax{
			{
				CategoryCode: "K",
			},
		},
		ShipTo: &Party{
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
	}

	_ = inv.Validate()

	for _, v := range inv.violations {
		if v.Rule.Code == "BR-IC-12" {
			t.Error("Should not have BR-IC-12 violation when deliver to country is present")
		}
	}
}

// BR-IG tests (IGIC - Canary Islands)
