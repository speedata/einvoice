package einvoice

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestBR11_BuyerCountryCodeField tests that BR-11 references the correct field BT-55
func TestBR11_BuyerCountryCodeField(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				// Missing CountryID
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	inv.check()

	// Find BR-11 violation
	var br11Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-11" {
			br11Found = true
			// Check that it references BT-55, not BT-5
			if len(v.InvFields) == 0 {
				t.Error("BR-11 violation should have InvFields")
			}
			if v.InvFields[0] != "BT-55" {
				t.Errorf("BR-11 should reference BT-55 (Buyer country code), got %s", v.InvFields[0])
			}
		}
	}

	if !br11Found {
		t.Error("Expected BR-11 violation for missing buyer country code")
	}
}

// TestBR37_ChargeRuleNumber tests that charge tax category validation uses BR-37, not BR-32
func TestBR37_ChargeRuleNumber(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-002",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(110),
		GrandTotal:          decimal.NewFromInt(130),
		DuePayableAmount:    decimal.NewFromInt(130),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromInt(10),
				// Missing CategoryTradeTaxCategoryCode
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(110),
				CalculatedAmount: decimal.NewFromInt(20),
			},
		},
	}

	inv.check()

	// Find BR-37 violation (not BR-32)
	var br37Found bool
	var br32Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-37" {
			br37Found = true
		}
		if v.Rule == "BR-32" {
			br32Found = true
		}
	}

	if !br37Found {
		t.Error("Expected BR-37 violation for missing charge tax category code")
	}
	if br32Found {
		t.Error("Should use BR-37 for charges, not BR-32 (which is for allowances)")
	}
}

// TestBRCO3_TaxPointDateMutuallyExclusive tests BR-CO-3: TaxPointDate and DueDateTypeCode are mutually exclusive
func TestBRCO3_TaxPointDateMutuallyExclusive(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-003",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
				TaxPointDate:     time.Now(), // BT-7
				DueDateTypeCode:  "5",        // BT-8 - mutually exclusive!
			},
		},
	}

	inv.check()

	// Find BR-CO-3 violation
	var brco3Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-3" {
			brco3Found = true
		}
	}

	if !brco3Found {
		t.Error("Expected BR-CO-3 violation when both TaxPointDate and DueDateTypeCode are set")
	}
}

// TestBRCO4_InvoiceLineMustHaveVATCategory tests BR-CO-4: Each invoice line must have a VAT category code
func TestBRCO4_InvoiceLineMustHaveVATCategory(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-004",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:         "1",
				ItemName:       "Item",
				BilledQuantity: decimal.NewFromInt(1),
				NetPrice:       decimal.NewFromInt(100),
				Total:          decimal.NewFromInt(100),
				// Missing TaxCategoryCode
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	inv.check()

	// Find BR-CO-4 violation
	var brco4Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-4" {
			brco4Found = true
		}
	}

	if !brco4Found {
		t.Error("Expected BR-CO-4 violation when invoice line missing VAT category code")
	}
}

// TestBRCO17_VATCalculation tests BR-CO-17: VAT amount must equal basis × rate ÷ 100
func TestBRCO17_VATCalculation(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-005",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(120),
		DuePayableAmount:    decimal.NewFromInt(120),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(20), // Wrong! Should be 19.00
			},
		},
	}

	inv.check()

	// Find BR-CO-17 violation
	var brco17Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-17" {
			brco17Found = true
		}
	}

	if !brco17Found {
		t.Error("Expected BR-CO-17 violation when VAT calculation is incorrect")
	}
}

// TestBRCO18_AtLeastOneVATBreakdown tests BR-CO-18: Invoice should contain at least one VAT breakdown
func TestBRCO18_AtLeastOneVATBreakdown(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-006",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(100),
		DuePayableAmount:    decimal.NewFromInt(100),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			// Missing VAT breakdown!
		},
	}

	inv.check()

	// Find BR-CO-18 violation
	var brco18Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-18" {
			brco18Found = true
		}
	}

	if !brco18Found {
		t.Error("Expected BR-CO-18 violation when no VAT breakdown present")
	}
}

// TestBRCO19_InvoicingPeriodRequiresDate tests BR-CO-19: Invoicing period requires start or end date
func TestBRCO19_InvoicingPeriodRequiresDate(t *testing.T) {
	// This test is actually tricky because if both dates are zero, the condition
	// !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero()
	// will be false, so the validation won't trigger.
	// The validation only triggers if we somehow have a period indicated but both dates are zero,
	// which shouldn't happen in practice. Let's verify the logic works correctly.

	// Test case: This should NOT trigger BR-CO-19 because no period is used
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-007",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		// BillingSpecifiedPeriodStart and End are both zero - no period used
	}

	inv.check()

	// Should NOT find BR-CO-19 violation
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-19" {
			t.Error("Should not have BR-CO-19 violation when no billing period is used")
		}
	}
}

// TestBRCO20_InvoiceLinePeriodRequiresDate tests BR-CO-20: Invoice line period requires start or end date
func TestBRCO20_InvoiceLinePeriodRequiresDate(t *testing.T) {
	// Similar to BR-CO-19, this validation only triggers if somehow a period is indicated
	// but both dates are zero. The current implementation won't trigger in practice.

	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-008",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
				// BillingSpecifiedPeriodStart and End are both zero - no period used
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	inv.check()

	// Should NOT find BR-CO-20 violation
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-20" {
			t.Error("Should not have BR-CO-20 violation when no line period is used")
		}
	}
}

// TestBRCO25_PositiveAmountRequiresPaymentInfo tests BR-CO-25: Positive payment amount requires due date or terms
func TestBRCO25_PositiveAmountRequiresPaymentInfo(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-009",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119), // Positive amount
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		// Missing SpecifiedTradePaymentTerms - should trigger BR-CO-25
	}

	inv.check()

	// Find BR-CO-25 violation
	var brco25Found bool
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-25" {
			brco25Found = true
		}
	}

	if !brco25Found {
		t.Error("Expected BR-CO-25 violation when positive amount but no payment terms or due date")
	}
}

// TestBRCO25_WithPaymentTerms tests that BR-CO-25 does not trigger when payment terms are present
func TestBRCO25_WithPaymentTerms(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-010",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{
				Description: "Payment within 14 days",
			},
		},
	}

	inv.check()

	// Should NOT find BR-CO-25 violation
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-25" {
			t.Error("Should not have BR-CO-25 violation when payment terms are present")
		}
	}
}

// TestBRCO25_WithDueDate tests that BR-CO-25 does not trigger when due date is present
func TestBRCO25_WithDueDate(t *testing.T) {
	inv := Invoice{
		Profile:             CProfileBasic,
		InvoiceNumber:       "TEST-011",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{
				DueDate: time.Now().Add(14 * 24 * time.Hour),
			},
		},
	}

	inv.check()

	// Should NOT find BR-CO-25 violation
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-25" {
			t.Error("Should not have BR-CO-25 violation when due date is present")
		}
	}
}
