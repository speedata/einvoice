package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestCheckBRO_BR_CO_10_Valid tests that BR-CO-10 validation passes when LineTotal matches sum of invoice lines
func TestCheckBRO_BR_CO_10_Valid(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
			{Total: decimal.NewFromFloat(200.00)},
		},
		LineTotal: decimal.NewFromFloat(300.00),
	}

	inv.checkBRO()

	// Check that no BR-CO-10 violations were added
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-10" {
			t.Errorf("Expected no BR-CO-10 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_10_Invalid tests that BR-CO-10 violation is detected when LineTotal doesn't match
func TestCheckBRO_BR_CO_10_Invalid(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
			{Total: decimal.NewFromFloat(200.00)},
		},
		LineTotal: decimal.NewFromFloat(250.00), // Wrong value
	}

	inv.checkBRO()

	// Check that BR-CO-10 violation was added
	found := false
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-10" {
			found = true
			if len(v.InvFields) != 2 || v.InvFields[0] != "BT-106" || v.InvFields[1] != "BT-131" {
				t.Errorf("BR-CO-10 violation has incorrect InvFields: %v", v.InvFields)
			}
		}
	}
	if !found {
		t.Error("Expected BR-CO-10 violation, but none was found")
	}
}

// TestCheckBRO_BR_CO_13_Valid tests that BR-CO-13 validation passes when TaxBasisTotal is correct
func TestCheckBRO_BR_CO_13_Valid(t *testing.T) {
	inv := &Invoice{
		LineTotal:      decimal.NewFromFloat(1000.00),
		AllowanceTotal: decimal.NewFromFloat(150.00),
		ChargeTotal:    decimal.NewFromFloat(50.00),
		TaxBasisTotal:  decimal.NewFromFloat(900.00), // 1000 - 150 + 50
	}

	inv.checkBRO()

	// Check that no BR-CO-13 violations were added
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-13" {
			t.Errorf("Expected no BR-CO-13 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_13_Invalid tests that BR-CO-13 violation is detected when TaxBasisTotal is wrong
func TestCheckBRO_BR_CO_13_Invalid(t *testing.T) {
	inv := &Invoice{
		LineTotal:      decimal.NewFromFloat(1000.00),
		AllowanceTotal: decimal.NewFromFloat(150.00),
		ChargeTotal:    decimal.NewFromFloat(50.00),
		TaxBasisTotal:  decimal.NewFromFloat(1000.00), // Wrong: should be 900
	}

	inv.checkBRO()

	// Check that BR-CO-13 violation was added
	found := false
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-13" {
			found = true
			expectedFields := []string{"BT-109", "BT-106", "BT-107", "BT-108"}
			if len(v.InvFields) != len(expectedFields) {
				t.Errorf("BR-CO-13 violation has incorrect number of InvFields: got %v, want %v", v.InvFields, expectedFields)
			}
		}
	}
	if !found {
		t.Error("Expected BR-CO-13 violation, but none was found")
	}
}

// TestCheckBRO_BR_CO_15_Valid tests that BR-CO-15 validation passes when GrandTotal is correct
func TestCheckBRO_BR_CO_15_Valid(t *testing.T) {
	inv := &Invoice{
		TaxBasisTotal: decimal.NewFromFloat(900.00),
		TaxTotal:      decimal.NewFromFloat(171.00),
		GrandTotal:    decimal.NewFromFloat(1071.00), // 900 + 171
	}

	inv.checkBRO()

	// Check that no BR-CO-15 violations were added
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-15" {
			t.Errorf("Expected no BR-CO-15 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_15_Invalid tests that BR-CO-15 violation is detected when GrandTotal is wrong
func TestCheckBRO_BR_CO_15_Invalid(t *testing.T) {
	inv := &Invoice{
		TaxBasisTotal: decimal.NewFromFloat(900.00),
		TaxTotal:      decimal.NewFromFloat(171.00),
		GrandTotal:    decimal.NewFromFloat(1000.00), // Wrong: should be 1071
	}

	inv.checkBRO()

	// Check that BR-CO-15 violation was added
	found := false
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-15" {
			found = true
			expectedFields := []string{"BT-112", "BT-109", "BT-110"}
			if len(v.InvFields) != len(expectedFields) {
				t.Errorf("BR-CO-15 violation has incorrect number of InvFields: got %v, want %v", v.InvFields, expectedFields)
			}
		}
	}
	if !found {
		t.Error("Expected BR-CO-15 violation, but none was found")
	}
}

// TestCheckBRO_BR_CO_16_Valid tests that BR-CO-16 validation passes when DuePayableAmount is correct
func TestCheckBRO_BR_CO_16_Valid(t *testing.T) {
	inv := &Invoice{
		GrandTotal:       decimal.NewFromFloat(1071.00),
		TotalPrepaid:     decimal.NewFromFloat(100.00),
		RoundingAmount:   decimal.NewFromFloat(0.05),
		DuePayableAmount: decimal.NewFromFloat(971.05), // 1071 - 100 + 0.05
	}

	inv.checkBRO()

	// Check that no BR-CO-16 violations were added
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-16" {
			t.Errorf("Expected no BR-CO-16 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_16_Invalid tests that BR-CO-16 violation is detected when DuePayableAmount is wrong
func TestCheckBRO_BR_CO_16_Invalid(t *testing.T) {
	inv := &Invoice{
		GrandTotal:       decimal.NewFromFloat(1071.00),
		TotalPrepaid:     decimal.NewFromFloat(100.00),
		RoundingAmount:   decimal.NewFromFloat(0.05),
		DuePayableAmount: decimal.NewFromFloat(971.00), // Wrong: should be 971.05
	}

	inv.checkBRO()

	// Check that BR-CO-16 violation was added
	found := false
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-16" {
			found = true
			expectedFields := []string{"BT-115", "BT-112", "BT-113", "BT-114"}
			if len(v.InvFields) != len(expectedFields) {
				t.Errorf("BR-CO-16 violation has incorrect number of InvFields: got %v, want %v", v.InvFields, expectedFields)
			}
		}
	}
	if !found {
		t.Error("Expected BR-CO-16 violation, but none was found")
	}
}

// TestCheckBRO_MultipleViolations tests detection of multiple violations at once
func TestCheckBRO_MultipleViolations(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
			{Total: decimal.NewFromFloat(200.00)},
		},
		LineTotal:        decimal.NewFromFloat(250.00),  // Wrong: should be 300 (BR-CO-10)
		AllowanceTotal:   decimal.NewFromFloat(50.00),
		ChargeTotal:      decimal.NewFromFloat(10.00),
		TaxBasisTotal:    decimal.NewFromFloat(250.00),  // Wrong: should be 210 (BR-CO-13)
		TaxTotal:         decimal.NewFromFloat(47.50),
		GrandTotal:       decimal.NewFromFloat(300.00),  // Wrong: should be 257.50 (BR-CO-15)
		TotalPrepaid:     decimal.NewFromFloat(50.00),
		RoundingAmount:   decimal.NewFromFloat(0.50),
		DuePayableAmount: decimal.NewFromFloat(250.00),  // Wrong: should be 250.50 (BR-CO-16)
	}

	inv.checkBRO()

	// Check that all four violations were detected
	violations := make(map[string]bool)
	for _, v := range inv.Violations {
		violations[v.Rule] = true
	}

	expectedViolations := []string{"BR-CO-10", "BR-CO-13", "BR-CO-15", "BR-CO-16"}
	for _, rule := range expectedViolations {
		if !violations[rule] {
			t.Errorf("Expected %s violation, but it was not found", rule)
		}
	}
}

// TestCheckBRO_WithNegativeRounding tests BR-CO-16 with negative rounding amount
func TestCheckBRO_BR_CO_16_NegativeRounding(t *testing.T) {
	inv := &Invoice{
		GrandTotal:       decimal.NewFromFloat(119.00),
		TotalPrepaid:     decimal.NewFromFloat(50.00),
		RoundingAmount:   decimal.NewFromFloat(-0.14),
		DuePayableAmount: decimal.NewFromFloat(68.86), // 119 - 50 + (-0.14)
	}

	inv.checkBRO()

	// Check that no BR-CO-16 violations were added
	for _, v := range inv.Violations {
		if v.Rule == "BR-CO-16" {
			t.Errorf("Expected no BR-CO-16 violation with negative rounding, but got: %s", v.Text)
		}
	}
}
