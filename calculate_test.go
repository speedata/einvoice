package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestUpdateTotals_BasicCalculation tests the basic calculation of totals from invoice lines
// according to BR-CO-10: LineTotal should be the sum of all invoice line net amounts
func TestUpdateTotals_BasicCalculation(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
			{Total: decimal.NewFromFloat(200.00)},
			{Total: decimal.NewFromFloat(50.50)},
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(19.00)},
			{CalculatedAmount: decimal.NewFromFloat(38.00)},
		},
	}

	inv.UpdateTotals()

	// BR-CO-10: LineTotal = Sum of invoice line totals
	expectedLineTotal := decimal.NewFromFloat(350.50)
	if !inv.LineTotal.Equal(expectedLineTotal) {
		t.Errorf("LineTotal = %s, want %s (BR-CO-10)", inv.LineTotal, expectedLineTotal)
	}

	// TaxTotal = Sum of calculated tax amounts
	expectedTaxTotal := decimal.NewFromFloat(57.00)
	if !inv.TaxTotal.Equal(expectedTaxTotal) {
		t.Errorf("TaxTotal = %s, want %s", inv.TaxTotal, expectedTaxTotal)
	}

	// BR-CO-13: TaxBasisTotal = LineTotal (no allowances/charges)
	if !inv.TaxBasisTotal.Equal(expectedLineTotal) {
		t.Errorf("TaxBasisTotal = %s, want %s (BR-CO-13)", inv.TaxBasisTotal, expectedLineTotal)
	}

	// BR-CO-15: GrandTotal = TaxBasisTotal + TaxTotal
	expectedGrandTotal := decimal.NewFromFloat(407.50)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		t.Errorf("GrandTotal = %s, want %s (BR-CO-15)", inv.GrandTotal, expectedGrandTotal)
	}

	// BR-CO-16: DuePayableAmount = GrandTotal (no prepaid/rounding)
	if !inv.DuePayableAmount.Equal(expectedGrandTotal) {
		t.Errorf("DuePayableAmount = %s, want %s (BR-CO-16)", inv.DuePayableAmount, expectedGrandTotal)
	}
}

// TestUpdateTotals_WithAllowancesAndCharges tests that document-level allowances
// and charges are correctly applied according to BR-CO-13
func TestUpdateTotals_WithAllowancesAndCharges(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(1000.00)},
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(171.00)}, // 19% of 900
		},
		// Provide actual allowances/charges instead of manually setting totals
		// UpdateTotals() will calculate AllowanceTotal and ChargeTotal from these
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(100.00),
			},
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(50.00),
			},
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromFloat(50.00),
			},
		},
	}

	inv.UpdateTotals()

	// BR-CO-10: LineTotal = Sum of invoice lines
	expectedLineTotal := decimal.NewFromFloat(1000.00)
	if !inv.LineTotal.Equal(expectedLineTotal) {
		t.Errorf("LineTotal = %s, want %s (BR-CO-10)", inv.LineTotal, expectedLineTotal)
	}

	// BR-CO-13: TaxBasisTotal = LineTotal - AllowanceTotal + ChargeTotal
	// = 1000 - 150 + 50 = 900
	expectedTaxBasisTotal := decimal.NewFromFloat(900.00)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		t.Errorf("TaxBasisTotal = %s, want %s (BR-CO-13: LineTotal - AllowanceTotal + ChargeTotal)",
			inv.TaxBasisTotal, expectedTaxBasisTotal)
	}

	// BR-CO-15: GrandTotal = TaxBasisTotal + TaxTotal
	// = 900 + 171 = 1071
	expectedGrandTotal := decimal.NewFromFloat(1071.00)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		t.Errorf("GrandTotal = %s, want %s (BR-CO-15)", inv.GrandTotal, expectedGrandTotal)
	}
}

// TestUpdateTotals_WithRoundingAmount tests that rounding amount is correctly
// applied according to BR-CO-16
func TestUpdateTotals_WithRoundingAmount(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(19.00)},
		},
		TotalPrepaid:   decimal.NewFromFloat(50.00),  // BT-113
		RoundingAmount: decimal.NewFromFloat(-0.14),  // BT-114
	}

	inv.UpdateTotals()

	// BR-CO-16: DuePayableAmount = GrandTotal - TotalPrepaid + RoundingAmount
	// = 119.00 - 50.00 + (-0.14) = 68.86
	expectedDuePayableAmount := decimal.NewFromFloat(68.86)
	if !inv.DuePayableAmount.Equal(expectedDuePayableAmount) {
		t.Errorf("DuePayableAmount = %s, want %s (BR-CO-16: GrandTotal - TotalPrepaid + RoundingAmount)",
			inv.DuePayableAmount, expectedDuePayableAmount)
	}
}

// TestUpdateTotals_Idempotent tests that calling UpdateTotals() multiple times
// produces the same result (fixes the accumulation bug)
func TestUpdateTotals_Idempotent(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
			{Total: decimal.NewFromFloat(200.00)},
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(57.00)},
		},
	}

	// Call UpdateTotals() first time
	inv.UpdateTotals()
	firstLineTotal := inv.LineTotal
	firstTaxTotal := inv.TaxTotal
	firstGrandTotal := inv.GrandTotal

	// Call UpdateTotals() second time
	inv.UpdateTotals()

	// Results should be identical
	if !inv.LineTotal.Equal(firstLineTotal) {
		t.Errorf("Second call: LineTotal = %s, want %s (should be idempotent)",
			inv.LineTotal, firstLineTotal)
	}

	if !inv.TaxTotal.Equal(firstTaxTotal) {
		t.Errorf("Second call: TaxTotal = %s, want %s (should be idempotent)",
			inv.TaxTotal, firstTaxTotal)
	}

	if !inv.GrandTotal.Equal(firstGrandTotal) {
		t.Errorf("Second call: GrandTotal = %s, want %s (should be idempotent)",
			inv.GrandTotal, firstGrandTotal)
	}

	// Call UpdateTotals() third time to be sure
	inv.UpdateTotals()

	if !inv.LineTotal.Equal(firstLineTotal) {
		t.Errorf("Third call: LineTotal = %s, want %s (should be idempotent)",
			inv.LineTotal, firstLineTotal)
	}
}

// TestUpdateTotals_ComprehensiveScenario tests all business rules together
// with a realistic invoice scenario
func TestUpdateTotals_ComprehensiveScenario(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(500.00)},  // Line 1
			{Total: decimal.NewFromFloat(300.00)},  // Line 2
			{Total: decimal.NewFromFloat(200.00)},  // Line 3
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(152.00)}, // 19% VAT on 800 basis
		},
		// Provide actual allowances/charges instead of manually setting totals
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(150.00), // Document discount
			},
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(100.00), // Additional discount
			},
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromFloat(50.00), // Shipping charge
			},
		},
		TotalPrepaid:   decimal.NewFromFloat(100.00), // Partial payment
		RoundingAmount: decimal.NewFromFloat(0.05),   // Rounding
	}

	inv.UpdateTotals()

	// BR-CO-10: LineTotal = 500 + 300 + 200 = 1000
	expectedLineTotal := decimal.NewFromFloat(1000.00)
	if !inv.LineTotal.Equal(expectedLineTotal) {
		t.Errorf("LineTotal = %s, want %s (BR-CO-10)", inv.LineTotal, expectedLineTotal)
	}

	// TaxTotal = 152 (from TradeTaxes)
	expectedTaxTotal := decimal.NewFromFloat(152.00)
	if !inv.TaxTotal.Equal(expectedTaxTotal) {
		t.Errorf("TaxTotal = %s, want %s", inv.TaxTotal, expectedTaxTotal)
	}

	// BR-CO-13: TaxBasisTotal = 1000 - 250 + 50 = 800
	expectedTaxBasisTotal := decimal.NewFromFloat(800.00)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		t.Errorf("TaxBasisTotal = %s, want %s (BR-CO-13)", inv.TaxBasisTotal, expectedTaxBasisTotal)
	}

	// BR-CO-15: GrandTotal = 800 + 152 = 952
	expectedGrandTotal := decimal.NewFromFloat(952.00)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		t.Errorf("GrandTotal = %s, want %s (BR-CO-15)", inv.GrandTotal, expectedGrandTotal)
	}

	// BR-CO-16: DuePayableAmount = 952 - 100 + 0.05 = 852.05
	expectedDuePayableAmount := decimal.NewFromFloat(852.05)
	if !inv.DuePayableAmount.Equal(expectedDuePayableAmount) {
		t.Errorf("DuePayableAmount = %s, want %s (BR-CO-16)", inv.DuePayableAmount, expectedDuePayableAmount)
	}
}

// TestUpdateTotals_ZeroValues tests that UpdateTotals works correctly with zero values
func TestUpdateTotals_ZeroValues(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{},
		TradeTaxes:   []TradeTax{},
	}

	inv.UpdateTotals()

	if !inv.LineTotal.IsZero() {
		t.Errorf("LineTotal = %s, want 0", inv.LineTotal)
	}

	if !inv.TaxTotal.IsZero() {
		t.Errorf("TaxTotal = %s, want 0", inv.TaxTotal)
	}

	if !inv.TaxBasisTotal.IsZero() {
		t.Errorf("TaxBasisTotal = %s, want 0", inv.TaxBasisTotal)
	}

	if !inv.GrandTotal.IsZero() {
		t.Errorf("GrandTotal = %s, want 0", inv.GrandTotal)
	}

	if !inv.DuePayableAmount.IsZero() {
		t.Errorf("DuePayableAmount = %s, want 0", inv.DuePayableAmount)
	}
}

// TestUpdateApplicableTradeTax_WithDocumentLevelAllowances tests that document-level
// allowances are correctly subtracted from the tax basis amount (Bug #4 fix)
func TestUpdateApplicableTradeTax_WithDocumentLevelAllowances(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(1000.00),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				ActualAmount:                          decimal.NewFromFloat(100.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{})

	// Should have one tax entry for category S with 19%
	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 TradeTax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]

	// BR-S-8: Basis amount should be line total minus allowance
	// = 1000 - 100 = 900
	expectedBasisAmount := decimal.NewFromFloat(900.00)
	if !tt.BasisAmount.Equal(expectedBasisAmount) {
		t.Errorf("BasisAmount = %s, want %s (should subtract allowance)", tt.BasisAmount, expectedBasisAmount)
	}

	// Tax should be 19% of 900 = 171
	expectedTax := decimal.NewFromFloat(171.00)
	if !tt.CalculatedAmount.Equal(expectedTax) {
		t.Errorf("CalculatedAmount = %s, want %s", tt.CalculatedAmount, expectedTax)
	}

	if tt.CategoryCode != "S" {
		t.Errorf("CategoryCode = %s, want S", tt.CategoryCode)
	}
}

// TestUpdateApplicableTradeTax_WithDocumentLevelCharges tests that document-level
// charges are correctly added to the tax basis amount (Bug #4 fix)
func TestUpdateApplicableTradeTax_WithDocumentLevelCharges(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(1000.00),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true, // Charge
				ActualAmount:                          decimal.NewFromFloat(50.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{})

	// Should have one tax entry for category S with 19%
	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 TradeTax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]

	// BR-S-8: Basis amount should be line total plus charge
	// = 1000 + 50 = 1050
	expectedBasisAmount := decimal.NewFromFloat(1050.00)
	if !tt.BasisAmount.Equal(expectedBasisAmount) {
		t.Errorf("BasisAmount = %s, want %s (should add charge)", tt.BasisAmount, expectedBasisAmount)
	}

	// Tax should be 19% of 1050 = 199.50
	expectedTax := decimal.NewFromFloat(199.50)
	if !tt.CalculatedAmount.Equal(expectedTax) {
		t.Errorf("CalculatedAmount = %s, want %s", tt.CalculatedAmount, expectedTax)
	}
}

// TestUpdateApplicableTradeTax_MixedAllowancesAndCharges tests a realistic scenario
// with both allowances and charges on the same tax category
func TestUpdateApplicableTradeTax_MixedAllowancesAndCharges(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(1000.00),
			},
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(500.00),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				ActualAmount:                          decimal.NewFromFloat(150.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
			{
				ChargeIndicator:                       true, // Charge (shipping)
				ActualAmount:                          decimal.NewFromFloat(50.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{})

	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 TradeTax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]

	// Basis = (1000 + 500) - 150 + 50 = 1400
	expectedBasisAmount := decimal.NewFromFloat(1400.00)
	if !tt.BasisAmount.Equal(expectedBasisAmount) {
		t.Errorf("BasisAmount = %s, want %s (lines - allowance + charge)", tt.BasisAmount, expectedBasisAmount)
	}

	// Tax = 19% of 1400 = 266
	expectedTax := decimal.NewFromFloat(266.00)
	if !tt.CalculatedAmount.Equal(expectedTax) {
		t.Errorf("CalculatedAmount = %s, want %s", tt.CalculatedAmount, expectedTax)
	}
}

// TestUpdateApplicableTradeTax_MultipleCategories tests handling of multiple
// tax categories with document-level allowances/charges
func TestUpdateApplicableTradeTax_MultipleCategories(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(1000.00),
			},
			{
				TaxCategoryCode:          "AE", // Reverse charge
				TaxRateApplicablePercent: decimal.NewFromFloat(0),
				Total:                    decimal.NewFromFloat(500.00),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromFloat(100.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"AE": "Reverse charge"})

	// Should have two tax entries
	if len(inv.TradeTaxes) != 2 {
		t.Fatalf("Expected 2 TradeTax entries, got %d", len(inv.TradeTaxes))
	}

	// Find category S and AE
	var ttS, ttAE *TradeTax
	for i := range inv.TradeTaxes {
		switch inv.TradeTaxes[i].CategoryCode {
		case "S":
			ttS = &inv.TradeTaxes[i]
		case "AE":
			ttAE = &inv.TradeTaxes[i]
		}
	}

	if ttS == nil {
		t.Fatal("Category S not found in TradeTaxes")
	}
	if ttAE == nil {
		t.Fatal("Category AE not found in TradeTaxes")
	}

	// Category S: 1000 - 100 = 900
	expectedBasisS := decimal.NewFromFloat(900.00)
	if !ttS.BasisAmount.Equal(expectedBasisS) {
		t.Errorf("Category S BasisAmount = %s, want %s", ttS.BasisAmount, expectedBasisS)
	}

	// Category AE: 500 (no allowances/charges for this category)
	expectedBasisAE := decimal.NewFromFloat(500.00)
	if !ttAE.BasisAmount.Equal(expectedBasisAE) {
		t.Errorf("Category AE BasisAmount = %s, want %s", ttAE.BasisAmount, expectedBasisAE)
	}

	// Check exemption reason
	if ttAE.ExemptionReason != "Reverse charge" {
		t.Errorf("ExemptionReason = %q, want %q", ttAE.ExemptionReason, "Reverse charge")
	}
}

// TestUpdateApplicableTradeTax_AllowanceOnlyCategory tests that allowances
// can create a tax category even without invoice lines (edge case)
func TestUpdateApplicableTradeTax_AllowanceOnlyCategory(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromFloat(50.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{})

	// Should create a tax entry for the allowance
	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 TradeTax entry, got %d", len(inv.TradeTaxes))
	}

	tt := inv.TradeTaxes[0]

	// Basis = -50 (allowance with no lines)
	expectedBasis := decimal.NewFromFloat(-50.00)
	if !tt.BasisAmount.Equal(expectedBasis) {
		t.Errorf("BasisAmount = %s, want %s", tt.BasisAmount, expectedBasis)
	}

	// Tax = 19% of -50 = -9.50
	expectedTax := decimal.NewFromFloat(-9.50)
	if !tt.CalculatedAmount.Equal(expectedTax) {
		t.Errorf("CalculatedAmount = %s, want %s", tt.CalculatedAmount, expectedTax)
	}
}

// TestUpdateAllowancesAndCharges verifies the internal calculation function
func TestUpdateAllowancesAndCharges(t *testing.T) {
	inv := &Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(100),
			},
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(50),
			},
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromFloat(30),
			},
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromFloat(20),
			},
		},
	}

	inv.updateAllowancesAndCharges()

	expectedAllowance := decimal.NewFromFloat(150)
	expectedCharge := decimal.NewFromFloat(50)

	if !inv.AllowanceTotal.Equal(expectedAllowance) {
		t.Errorf("AllowanceTotal = %s, want %s", inv.AllowanceTotal, expectedAllowance)
	}

	if !inv.ChargeTotal.Equal(expectedCharge) {
		t.Errorf("ChargeTotal = %s, want %s", inv.ChargeTotal, expectedCharge)
	}
}

// TestUpdateAllowancesAndCharges_Idempotent verifies idempotent behavior
func TestUpdateAllowancesAndCharges_Idempotent(t *testing.T) {
	inv := &Invoice{
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(100),
			},
		},
	}

	inv.updateAllowancesAndCharges()
	first := inv.AllowanceTotal

	inv.updateAllowancesAndCharges()
	second := inv.AllowanceTotal

	if !first.Equal(second) {
		t.Errorf("updateAllowancesAndCharges not idempotent: first=%s, second=%s", first, second)
	}
}

// TestUpdateTotals_AutomaticallyCalculatesAllowancesAndCharges verifies that
// UpdateTotals() automatically calculates AllowanceTotal and ChargeTotal
// from SpecifiedTradeAllowanceCharge, making the API more intuitive
func TestUpdateTotals_AutomaticallyCalculatesAllowancesAndCharges(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(1000.00)},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(100.00),
				Reason:          "Volume discount",
			},
			{
				ChargeIndicator: false,
				ActualAmount:    decimal.NewFromFloat(50.00),
				Reason:          "Early payment",
			},
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromFloat(30.00),
				Reason:          "Shipping",
			},
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(166.60)}, // 19% of 880
		},
	}

	// User only needs to call UpdateTotals() - it handles everything
	inv.UpdateTotals()

	// Verify AllowanceTotal was calculated automatically
	expectedAllowanceTotal := decimal.NewFromFloat(150.00) // 100 + 50
	if !inv.AllowanceTotal.Equal(expectedAllowanceTotal) {
		t.Errorf("AllowanceTotal = %s, want %s (should be calculated automatically)",
			inv.AllowanceTotal, expectedAllowanceTotal)
	}

	// Verify ChargeTotal was calculated automatically
	expectedChargeTotal := decimal.NewFromFloat(30.00)
	if !inv.ChargeTotal.Equal(expectedChargeTotal) {
		t.Errorf("ChargeTotal = %s, want %s (should be calculated automatically)",
			inv.ChargeTotal, expectedChargeTotal)
	}

	// Verify TaxBasisTotal uses the calculated allowances/charges
	// LineTotal (1000) - AllowanceTotal (150) + ChargeTotal (30) = 880
	expectedTaxBasisTotal := decimal.NewFromFloat(880.00)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		t.Errorf("TaxBasisTotal = %s, want %s (should use calculated allowances/charges)",
			inv.TaxBasisTotal, expectedTaxBasisTotal)
	}

	// Verify GrandTotal = TaxBasisTotal + TaxTotal = 880 + 166.60 = 1046.60
	expectedGrandTotal := decimal.NewFromFloat(1046.60)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		t.Errorf("GrandTotal = %s, want %s", inv.GrandTotal, expectedGrandTotal)
	}
}

// TestUpdateTotals_IntegrationWithCalculations tests the complete calculation flow
// that users would typically follow when building an invoice programmatically
func TestUpdateTotals_IntegrationWithCalculations(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				Total:                    decimal.NewFromFloat(500.00),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
			{
				Total:                    decimal.NewFromFloat(300.00),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromFloat(100.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
			{
				ChargeIndicator:                       true,
				ActualAmount:                          decimal.NewFromFloat(50.00),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
		TotalPrepaid:   decimal.NewFromFloat(200.00),
		RoundingAmount: decimal.NewFromFloat(-0.25),
	}

	// Step 1: Calculate VAT breakdown from lines and document-level allowances/charges
	inv.UpdateApplicableTradeTax(map[string]string{})

	// Verify VAT was calculated correctly
	// Basis: (500 + 300) - 100 + 50 = 750
	// VAT: 750 × 19% = 142.50
	if len(inv.TradeTaxes) != 1 {
		t.Fatalf("Expected 1 TradeTax, got %d", len(inv.TradeTaxes))
	}
	expectedVATBasis := decimal.NewFromFloat(750.00)
	if !inv.TradeTaxes[0].BasisAmount.Equal(expectedVATBasis) {
		t.Errorf("VAT BasisAmount = %s, want %s", inv.TradeTaxes[0].BasisAmount, expectedVATBasis)
	}
	expectedVAT := decimal.NewFromFloat(142.50)
	if !inv.TradeTaxes[0].CalculatedAmount.Equal(expectedVAT) {
		t.Errorf("VAT CalculatedAmount = %s, want %s", inv.TradeTaxes[0].CalculatedAmount, expectedVAT)
	}

	// Step 2: Calculate all totals (including allowances/charges)
	inv.UpdateTotals()

	// Verify all totals
	// LineTotal = 500 + 300 = 800
	expectedLineTotal := decimal.NewFromFloat(800.00)
	if !inv.LineTotal.Equal(expectedLineTotal) {
		t.Errorf("LineTotal = %s, want %s", inv.LineTotal, expectedLineTotal)
	}

	// AllowanceTotal = 100 (calculated automatically)
	expectedAllowanceTotal := decimal.NewFromFloat(100.00)
	if !inv.AllowanceTotal.Equal(expectedAllowanceTotal) {
		t.Errorf("AllowanceTotal = %s, want %s", inv.AllowanceTotal, expectedAllowanceTotal)
	}

	// ChargeTotal = 50 (calculated automatically)
	expectedChargeTotal := decimal.NewFromFloat(50.00)
	if !inv.ChargeTotal.Equal(expectedChargeTotal) {
		t.Errorf("ChargeTotal = %s, want %s", inv.ChargeTotal, expectedChargeTotal)
	}

	// TaxBasisTotal = 800 - 100 + 50 = 750
	expectedTaxBasisTotal := decimal.NewFromFloat(750.00)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		t.Errorf("TaxBasisTotal = %s, want %s", inv.TaxBasisTotal, expectedTaxBasisTotal)
	}

	// TaxTotal = 142.50
	expectedTaxTotal := decimal.NewFromFloat(142.50)
	if !inv.TaxTotal.Equal(expectedTaxTotal) {
		t.Errorf("TaxTotal = %s, want %s", inv.TaxTotal, expectedTaxTotal)
	}

	// GrandTotal = 750 + 142.50 = 892.50
	expectedGrandTotal := decimal.NewFromFloat(892.50)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		t.Errorf("GrandTotal = %s, want %s", inv.GrandTotal, expectedGrandTotal)
	}

	// DuePayableAmount = 892.50 - 200 + (-0.25) = 692.25
	expectedDuePayable := decimal.NewFromFloat(692.25)
	if !inv.DuePayableAmount.Equal(expectedDuePayable) {
		t.Errorf("DuePayableAmount = %s, want %s", inv.DuePayableAmount, expectedDuePayable)
	}
}

// TestUpdateTotals_NoAllowancesOrCharges verifies UpdateTotals works correctly
// when there are no document-level allowances or charges
func TestUpdateTotals_NoAllowancesOrCharges(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
		},
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(19.00)},
		},
		// No SpecifiedTradeAllowanceCharge
	}

	inv.UpdateTotals()

	// AllowanceTotal should be zero
	if !inv.AllowanceTotal.IsZero() {
		t.Errorf("AllowanceTotal = %s, want 0 (no allowances provided)", inv.AllowanceTotal)
	}

	// ChargeTotal should be zero
	if !inv.ChargeTotal.IsZero() {
		t.Errorf("ChargeTotal = %s, want 0 (no charges provided)", inv.ChargeTotal)
	}

	// TaxBasisTotal should equal LineTotal
	if !inv.TaxBasisTotal.Equal(inv.LineTotal) {
		t.Errorf("TaxBasisTotal = %s, want %s (LineTotal with no adjustments)",
			inv.TaxBasisTotal, inv.LineTotal)
	}
}
