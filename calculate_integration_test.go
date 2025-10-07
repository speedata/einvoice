package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

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
