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
		AllowanceTotal: decimal.NewFromFloat(150.00), // BT-107
		ChargeTotal:    decimal.NewFromFloat(50.00),  // BT-108
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
		AllowanceTotal: decimal.NewFromFloat(250.00), // Document discount
		ChargeTotal:    decimal.NewFromFloat(50.00),  // Shipping charge
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
