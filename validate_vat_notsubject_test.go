package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestBRO1_ExactlyOneBreakdown tests BR-O-01: Must have exactly one VAT breakdown with category O
func TestBRO1_ExactlyOneBreakdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		breakdownsO   int
		breakdownsS   int
		expectViolation bool
	}{
		{"No category O breakdown", 0, 1, true},
		{"Exactly one category O breakdown", 1, 0, false},
		{"Two category O breakdowns", 2, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := Invoice{
				InvoiceLines: []InvoiceLine{
					{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
				},
				TradeTaxes: []TradeTax{},
			}

			for i := 0; i < tt.breakdownsO; i++ {
				inv.TradeTaxes = append(inv.TradeTaxes, TradeTax{
					CategoryCode:     "O",
					BasisAmount:      decimal.NewFromInt(100),
					CalculatedAmount: decimal.Zero,
					ExemptionReason:  "Not subject to VAT",
				})
			}
			for i := 0; i < tt.breakdownsS; i++ {
				inv.TradeTaxes = append(inv.TradeTaxes, TradeTax{
					CategoryCode: "S",
					BasisAmount:  decimal.NewFromInt(100),
				})
			}

			_ = inv.Validate()

			found := false
			for _, v := range inv.violations {
				if v.Rule.Code == "BR-O-01" {
					found = true
					break
				}
			}

			if found != tt.expectViolation {
				if tt.expectViolation {
					t.Errorf("Expected BR-O-01 violation for %s", tt.name)
				} else {
					t.Errorf("Unexpected BR-O-01 violation for %s", tt.name)
				}
			}
		})
	}
}

// TestBRO2_NoVATIdentifiers tests BR-O-02: Invoice lines with category O must NOT contain VAT identifiers
func TestBRO2_NoVATIdentifiers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		sellerVAT       string
		buyerVAT        string
		taxRepVAT       string
		expectViolation bool
	}{
		{"No VAT identifiers", "", "", "", false},
		{"Has seller VAT", "DE123", "", "", true},
		{"Has buyer VAT", "", "FR456", "", true},
		{"Has tax rep VAT", "", "", "GB789", true},
		{"Has all VAT identifiers", "DE123", "FR456", "GB789", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := Invoice{
				InvoiceLines: []InvoiceLine{
					{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
				},
				TradeTaxes: []TradeTax{
					{
						CategoryCode:     "O",
						BasisAmount:      decimal.NewFromInt(100),
						CalculatedAmount: decimal.Zero,
						ExemptionReason:  "Not subject to VAT",
					},
				},
				Seller: Party{VATaxRegistration: tt.sellerVAT},
				Buyer:  Party{VATaxRegistration: tt.buyerVAT},
			}

			if tt.taxRepVAT != "" {
				inv.SellerTaxRepresentativeTradeParty = &Party{VATaxRegistration: tt.taxRepVAT}
			}

			_ = inv.Validate()

			found := false
			for _, v := range inv.violations {
				if v.Rule.Code == "BR-O-02" {
					found = true
					break
				}
			}

			if found != tt.expectViolation {
				if tt.expectViolation {
					t.Errorf("Expected BR-O-02 violation for %s", tt.name)
				} else {
					t.Errorf("Unexpected BR-O-02 violation for %s", tt.name)
				}
			}
		})
	}
}

// TestBRO3_AllowanceNoVATIdentifiers tests BR-O-03: Allowances with category O must NOT contain VAT identifiers
func TestBRO3_AllowanceNoVATIdentifiers(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				CategoryTradeTaxCategoryCode:          "O",
				ActualAmount:                          decimal.NewFromInt(10),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(90),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
		Seller: Party{VATaxRegistration: "DE123"}, // Has VAT ID - should violate
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-03" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-03 violation when allowance with category O has VAT identifiers")
	}
}

// TestBRO4_ChargeNoVATIdentifiers tests BR-O-04: Charges with category O must NOT contain VAT identifiers
func TestBRO4_ChargeNoVATIdentifiers(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true, // Charge
				CategoryTradeTaxCategoryCode:          "O",
				ActualAmount:                          decimal.NewFromInt(10),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(110),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
		Buyer: Party{VATaxRegistration: "FR456"}, // Has VAT ID - should violate
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-04" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-04 violation when charge with category O has VAT identifiers")
	}
}

// TestBRO5_LineNoVATRate tests BR-O-05: Invoice lines with category O must NOT contain VAT rate
func TestBRO5_LineNoVATRate(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "O",
				TaxRateApplicablePercent: decimal.NewFromFloat(19.0), // Should NOT have rate
				Total:                    decimal.NewFromInt(100),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-05" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-05 violation for line with VAT rate in Not subject to VAT")
	}
}

// TestBRO6_AllowanceNoVATRate tests BR-O-06: Allowances with category O must NOT contain VAT rate
func TestBRO6_AllowanceNoVATRate(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				CategoryTradeTaxCategoryCode:          "O",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(10.0), // Should NOT have rate
				ActualAmount:                          decimal.NewFromInt(10),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(90),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-06" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-06 violation for allowance with VAT rate in Not subject to VAT")
	}
}

// TestBRO7_ChargeNoVATRate tests BR-O-07: Charges with category O must NOT contain VAT rate
func TestBRO7_ChargeNoVATRate(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true, // Charge
				CategoryTradeTaxCategoryCode:          "O",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(5.0), // Should NOT have rate
				ActualAmount:                          decimal.NewFromInt(5),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(105),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-07" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-07 violation for charge with VAT rate in Not subject to VAT")
	}
}

// TestBRO8_TaxableAmountCalculation tests BR-O-08: VAT category taxable amount calculation
func TestBRO8_TaxableAmountCalculation(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromFloat(100.0)},
			{TaxCategoryCode: "O", Total: decimal.NewFromFloat(50.0)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false, // Allowance
				CategoryTradeTaxCategoryCode: "O",
				ActualAmount:                 decimal.NewFromFloat(10.0),
			},
			{
				ChargeIndicator:              true, // Charge
				CategoryTradeTaxCategoryCode: "O",
				ActualAmount:                 decimal.NewFromFloat(5.0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromFloat(100.0), // Should be 150 - 10 + 5 = 145
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-08" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-08 violation for taxable amount mismatch in Not subject to VAT")
	}
}

// TestBRO8_TaxableAmountRoundingPrecision tests BR-O-08 with rounding
func TestBRO8_TaxableAmountRoundingPrecision(t *testing.T) {
	t.Parallel()

	// Regression test for rounding bug in Not subject to VAT category
	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode: "O",
				Total:           decimal.NewFromFloat(123.456),
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"O": "Not subject to VAT"})

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRuleCode("BR-O-08") {
			t.Error("BR-O-08 false positive: rounding precision not handled correctly")
		}
	}
}

// TestBRO9_VATAmountMustBeZero tests BR-O-09: VAT amount must be 0
func TestBRO9_VATAmountMustBeZero(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromFloat(19.0), // Should be 0
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-09" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-09 violation for non-zero VAT amount in Not subject to VAT")
	}
}

// TestBRO10_MustHaveExemptionReason tests BR-O-10: Must have exemption reason
func TestBRO10_MustHaveExemptionReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		reason          string
		reasonCode      string
		expectViolation bool
	}{
		{"Has reason text", "Not subject to VAT", "", false},
		{"Has reason code", "", "VATEX-EU-O", false},
		{"Has both", "Not subject to VAT", "VATEX-EU-O", false},
		{"Has neither", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := Invoice{
				InvoiceLines: []InvoiceLine{
					{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
				},
				TradeTaxes: []TradeTax{
					{
						CategoryCode:        "O",
						BasisAmount:         decimal.NewFromInt(100),
						CalculatedAmount:    decimal.Zero,
						ExemptionReason:     tt.reason,
						ExemptionReasonCode: tt.reasonCode,
					},
				},
			}

			_ = inv.Validate()

			found := false
			for _, v := range inv.violations {
				if v.Rule.Code == "BR-O-10" {
					found = true
					break
				}
			}

			if found != tt.expectViolation {
				if tt.expectViolation {
					t.Errorf("Expected BR-O-10 violation for %s", tt.name)
				} else {
					t.Errorf("Unexpected BR-O-10 violation for %s", tt.name)
				}
			}
		})
	}
}

// TestBRO11_OnlyOneCategoryAllowed tests BR-O-11: Cannot mix category O with other categories
func TestBRO11_OnlyOneCategoryAllowed(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
			{
				CategoryCode:     "S",
				BasisAmount:      decimal.NewFromInt(50),
				CalculatedAmount: decimal.NewFromFloat(9.5),
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-11" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-11 violation when mixing category O with other categories")
	}
}

// TestBRO12_AllLinesMustBeCategoryO tests BR-O-12: All invoice lines must be category O
func TestBRO12_AllLinesMustBeCategoryO(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
			{TaxCategoryCode: "S", Total: decimal.NewFromInt(50)}, // Wrong category
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-12" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-12 violation when lines have mixed categories")
	}
}

// TestBRO13_AllAllowancesMustBeCategoryO tests BR-O-13: All allowances must be category O
func TestBRO13_AllAllowancesMustBeCategoryO(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              false,
				CategoryTradeTaxCategoryCode: "S", // Wrong category
				ActualAmount:                 decimal.NewFromInt(10),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-13" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-13 violation when allowance has wrong category")
	}
}

// TestBRO14_AllChargesMustBeCategoryO tests BR-O-14: All charges must be category O
func TestBRO14_AllChargesMustBeCategoryO(t *testing.T) {
	t.Parallel()

	inv := Invoice{
		InvoiceLines: []InvoiceLine{
			{TaxCategoryCode: "O", Total: decimal.NewFromInt(100)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:              true,
				CategoryTradeTaxCategoryCode: "E", // Wrong category
				ActualAmount:                 decimal.NewFromInt(5),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
	}

	_ = inv.Validate()

	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-O-14" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-O-14 violation when charge has wrong category")
	}
}

// TestBRO_ValidNotSubjectToVAT tests a valid "Not subject to VAT" invoice
func TestBRO_ValidNotSubjectToVAT(t *testing.T) {
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
				CategoryCode:     "O",
				BasisAmount:      decimal.NewFromFloat(100.0),
				CalculatedAmount: decimal.Zero,
				ExemptionReason:  "Not subject to VAT",
			},
		},
		// No VAT identifiers - this is correct for category O
	}

	_ = inv.Validate()

	brORules := []string{"BR-O-01", "BR-O-02", "BR-O-03", "BR-O-04", "BR-O-05", "BR-O-06", "BR-O-07", "BR-O-08", "BR-O-09", "BR-O-10", "BR-O-11", "BR-O-12", "BR-O-13", "BR-O-14"}
	for _, v := range inv.violations {
		for _, rule := range brORules {
			if v.Rule.Code == rule {
				t.Errorf("Should not have %s violation for valid Not subject to VAT invoice: %v", rule, v)
			}
		}
	}
}
