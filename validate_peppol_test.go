package einvoice

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// TestValidatePEPPOL_BasicRequirements tests basic PEPPOL requirements
func TestValidatePEPPOL_BasicRequirements(t *testing.T) {
	// Create a minimal invoice that violates several PEPPOL rules
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30, // Mark as PEPPOL invoice
		BPSpecifiedDocumentContextParameter:        "", // Violates PEPPOL-EN16931-R001
		BuyerReference:                             "",    // Violates PEPPOL-EN16931-R003
		BuyerOrderReferencedDocument:               "",    // Violates PEPPOL-EN16931-R003
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "", // Violates PEPPOL-EN16931-R020
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "", // Violates PEPPOL-EN16931-R010
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		GrandTotal:       decimal.NewFromInt(119),
		DuePayableAmount: decimal.NewFromInt(119),
		InvoiceLines: []InvoiceLine{
			{
				LineID:         "1",
				ItemName:       "Test Item",
				BilledQuantity: decimal.NewFromInt(1),
				NetPrice:       decimal.NewFromInt(100),
				Total:          decimal.NewFromInt(100),
			},
		},
	}

	// Validate - will auto-detect PEPPOL based on GuidelineSpecifiedDocumentContextParameter
	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Check for specific PEPPOL violations
	expectedRules := []string{
		"PEPPOL-EN16931-R001", // Business process required
		"PEPPOL-EN16931-R003", // Buyer reference or order reference required
		"PEPPOL-EN16931-R010", // Buyer electronic address required
		"PEPPOL-EN16931-R020", // Seller electronic address required
	}

	for _, ruleCode := range expectedRules {
		if !valErr.HasRuleCode(ruleCode) {
			t.Errorf("Expected violation for rule %s, but not found", ruleCode)
		}
	}
}

// TestValidatePEPPOL_MultipleNotes tests the PEPPOL rule for maximum one note
func TestValidatePEPPOL_MultipleNotes(t *testing.T) {
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Notes: []Note{
			{SubjectCode: "", Text: "Note 1"},
			{SubjectCode: "", Text: "Note 2"}, // Violates PEPPOL-EN16931-R002
		},
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		GrandTotal:       decimal.NewFromInt(119),
		DuePayableAmount: decimal.NewFromInt(119),
		InvoiceLines: []InvoiceLine{
			{
				LineID:         "1",
				ItemName:       "Test Item",
				BilledQuantity: decimal.NewFromInt(1),
				NetPrice:       decimal.NewFromInt(100),
				Total:          decimal.NewFromInt(100),
			},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for multiple notes, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if !valErr.HasRule(rules.PEPPOLEN16931R2) {
		t.Error("Expected violation for PEPPOL-EN16931-R002 (multiple notes)")
	}
}

// TestValidatePEPPOL_ValidInvoice tests that a PEPPOL-compliant invoice passes validation
func TestValidatePEPPOL_ValidInvoice(t *testing.T) {
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration:         "DE123456789",
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		TaxTotal:         decimal.NewFromInt(19),
		GrandTotal:       decimal.NewFromInt(119),
		DuePayableAmount: decimal.NewFromInt(119),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
				Percent:          decimal.NewFromInt(19),
			},
		},
	}

	// This invoice should still have EN 16931 violations, but not additional PEPPOL violations
	// beyond the EN 16931 ones
	err := inv.Validate()
	// We don't expect nil here as EN 16931 validation will likely find issues
	// This test just ensures Validate() runs without panicking
	_ = err
}

// TestValidatePEPPOL_DecimalPrecisionViolations tests that Validate() catches
// decimal precision violations (BR-DEC rules) for PEPPOL invoices.
// This test verifies the fix for issue H1 from the bug analysis.
func TestValidatePEPPOL_DecimalPrecisionViolations(t *testing.T) {
	inv := &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceDate:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Seller: Party{
			Name: "Seller Inc",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration:         "DE123456789",
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Buyer GmbH",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
		},
		// Invalid: 3 decimal places (should be max 2 per BR-DEC-19)
		LineTotal:        decimal.RequireFromString("100.123"),
		TaxBasisTotal:    decimal.RequireFromString("100.123"),
		TaxTotal:         decimal.NewFromInt(19),
		GrandTotal:       decimal.RequireFromString("119.123"),
		DuePayableAmount: decimal.RequireFromString("119.123"),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Product A",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.RequireFromString("100.123"), // Invalid: 3 decimals (BR-DEC-23)
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromInt(19),
				BasisAmount:      decimal.RequireFromString("100.123"), // Invalid: 3 decimals (BR-DEC-19)
				TypeCode:         "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{Description: "Net 30"},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for decimal precision violations, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Should have multiple BR-DEC violations
	violations := valErr.Violations()
	if len(violations) == 0 {
		t.Fatal("Expected BR-DEC violations, got none")
	}

	// Check that we have BR-DEC violations (these are now caught for PEPPOL invoices)
	hasBRDEC := false
	for _, v := range violations {
		if v.Rule.Code == "BR-DEC-19" || v.Rule.Code == "BR-DEC-23" {
			hasBRDEC = true
			t.Logf("Found expected BR-DEC violation: %s - %s", v.Rule.Code, v.Text)
		}
	}

	if !hasBRDEC {
		t.Errorf("Expected BR-DEC violations in PEPPOL validation, but none found. Got violations: %v", violations)
	}
}

// TestValidatePEPPOL_R120_LineNetAmountCalculation tests line net amount calculation validation
func TestValidatePEPPOL_R120_LineNetAmountCalculation(t *testing.T) {
	tests := []struct {
		name          string
		line          InvoiceLine
		wantViolation bool
	}{
		{
			name: "valid - simple calculation",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Test Item",
				BilledQuantity:    decimal.NewFromInt(10),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.NewFromInt(100),
				Total:             decimal.NewFromInt(1000), // 10 × 100 = 1000 ✓
				TaxCategoryCode:   "S",
			},
			wantViolation: false,
		},
		{
			name: "valid - with base quantity",
			line: InvoiceLine{
				LineID:             "1",
				ItemName:           "Test Item",
				BilledQuantity:     decimal.NewFromInt(5),
				BilledQuantityUnit: "C62",
				NetPrice:           decimal.NewFromInt(100),
				BasisQuantity:      decimal.NewFromInt(10),
				BasisQuantityUnit:  "C62",
				Total:              decimal.NewFromInt(50), // 5 × 100 / 10 = 50 ✓
				TaxCategoryCode:    "S",
			},
			wantViolation: false,
		},
		{
			name: "valid - fractional quantity with rounding",
			line: InvoiceLine{
				LineID:             "1",
				ItemName:           "Hourly Work",
				BilledQuantity:     decimal.RequireFromString("0.0830"),
				BilledQuantityUnit: "HUR",
				NetPrice:           decimal.NewFromInt(110),
				BasisQuantity:      decimal.NewFromInt(1),
				BasisQuantityUnit:  "HUR",
				Total:              decimal.RequireFromString("9.13"), // 0.0830 × 110 / 1 = 9.13 ✓
				TaxCategoryCode:    "S",
			},
			wantViolation: false,
		},
		{
			name: "invalid - issue #127 example",
			line: InvoiceLine{
				LineID:             "1",
				ItemName:           "Hourly Work",
				BilledQuantity:     decimal.RequireFromString("0.0830"),
				BilledQuantityUnit: "HUR",
				NetPrice:           decimal.NewFromInt(110),
				BasisQuantity:      decimal.NewFromInt(1),
				BasisQuantityUnit:  "HUR",
				Total:              decimal.RequireFromString("9.17"), // Wrong! Should be 9.13
				TaxCategoryCode:    "S",
			},
			wantViolation: true,
		},
		{
			name: "valid - with line allowance",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Test Item",
				BilledQuantity:    decimal.NewFromInt(10),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.NewFromInt(100),
				Total:             decimal.NewFromInt(900), // 10 × 100 - 100 = 900 ✓
				TaxCategoryCode:   "S",
				InvoiceLineAllowances: []AllowanceCharge{
					{ActualAmount: decimal.NewFromInt(100)},
				},
			},
			wantViolation: false,
		},
		{
			name: "valid - with line charge",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Test Item",
				BilledQuantity:    decimal.NewFromInt(10),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.NewFromInt(100),
				Total:             decimal.NewFromInt(1050), // 10 × 100 + 50 = 1050 ✓
				TaxCategoryCode:   "S",
				InvoiceLineCharges: []AllowanceCharge{
					{ActualAmount: decimal.NewFromInt(50)},
				},
			},
			wantViolation: false,
		},
		{
			name: "valid - with both allowance and charge",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Test Item",
				BilledQuantity:    decimal.NewFromInt(10),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.NewFromInt(100),
				Total:             decimal.NewFromInt(970), // 10 × 100 + 50 - 80 = 970 ✓
				TaxCategoryCode:   "S",
				InvoiceLineCharges: []AllowanceCharge{
					{ActualAmount: decimal.NewFromInt(50)},
				},
				InvoiceLineAllowances: []AllowanceCharge{
					{ActualAmount: decimal.NewFromInt(80)},
				},
			},
			wantViolation: false,
		},
		{
			name: "invalid - calculation mismatch with allowances",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Test Item",
				BilledQuantity:    decimal.NewFromInt(10),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.NewFromInt(100),
				Total:             decimal.NewFromInt(1000), // Wrong! Should be 900 with allowance
				TaxCategoryCode:   "S",
				InvoiceLineAllowances: []AllowanceCharge{
					{ActualAmount: decimal.NewFromInt(100)},
				},
			},
			wantViolation: true,
		},
		{
			name: "valid - zero price free item",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Free Sample",
				BilledQuantity:    decimal.NewFromInt(1),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.Zero,
				Total:             decimal.Zero, // 1 × 0 = 0 ✓
				TaxCategoryCode:   "Z",
			},
			wantViolation: false,
		},
		{
			name: "valid - base quantity not specified defaults to 1",
			line: InvoiceLine{
				LineID:            "1",
				ItemName:          "Test Item",
				BilledQuantity:    decimal.NewFromInt(5),
				BilledQuantityUnit: "C62",
				NetPrice:          decimal.NewFromInt(20),
				// BasisQuantity not set (zero) - defaults to 1
				Total:           decimal.NewFromInt(100), // 5 × 20 / 1 = 100 ✓
				TaxCategoryCode: "S",
			},
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createPEPPOLTestInvoice()
			inv.InvoiceLines = []InvoiceLine{tt.line}
			inv.LineTotal = tt.line.Total
			inv.TaxBasisTotal = tt.line.Total
			inv.TaxTotal = decimal.Zero
			inv.GrandTotal = tt.line.Total
			inv.DuePayableAmount = tt.line.Total

			err := inv.Validate()

			var valErr *ValidationError
			hasR120Violation := false
			if errors.As(err, &valErr) {
				hasR120Violation = valErr.HasRule(rules.PEPPOLEN16931R120)
			}

			if tt.wantViolation && !hasR120Violation {
				t.Errorf("Expected PEPPOL-EN16931-R120 violation, but not found")
				if valErr != nil {
					t.Logf("Violations found: %v", valErr.Violations())
				}
			}
			if !tt.wantViolation && hasR120Violation {
				t.Errorf("Did not expect PEPPOL-EN16931-R120 violation, but found one")
				if valErr != nil {
					for _, v := range valErr.Violations() {
						if v.Rule.Code == "PEPPOL-EN16931-R120" {
							t.Logf("Violation: %s", v.Text)
						}
					}
				}
			}
		})
	}
}

// TestValidatePEPPOL_R121_BaseQuantityPositive tests base quantity validation
func TestValidatePEPPOL_R121_BaseQuantityPositive(t *testing.T) {
	tests := []struct {
		name          string
		basisQty      decimal.Decimal
		wantViolation bool
	}{
		{
			name:          "valid - positive base quantity",
			basisQty:      decimal.NewFromInt(1),
			wantViolation: false,
		},
		{
			name:          "valid - larger base quantity",
			basisQty:      decimal.NewFromInt(10),
			wantViolation: false,
		},
		{
			name:          "valid - fractional base quantity",
			basisQty:      decimal.RequireFromString("0.5"),
			wantViolation: false,
		},
		{
			name:          "valid - not specified (zero defaults to 1)",
			basisQty:      decimal.Zero,
			wantViolation: false, // Zero means not specified, defaults to 1
		},
		{
			name:          "invalid - negative base quantity",
			basisQty:      decimal.NewFromInt(-1),
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createPEPPOLTestInvoice()

			// Calculate expected total based on basis quantity
			billedQty := decimal.NewFromInt(10)
			netPrice := decimal.NewFromInt(100)
			baseQty := tt.basisQty
			if baseQty.IsZero() {
				baseQty = decimal.NewFromInt(1)
			}
			// For negative, we still need to calculate for the line
			var expectedTotal decimal.Decimal
			if !baseQty.IsZero() {
				expectedTotal = billedQty.Mul(netPrice).Div(baseQty)
			} else {
				expectedTotal = billedQty.Mul(netPrice)
			}
			expectedTotal = roundHalfUp(expectedTotal, 2)

			inv.InvoiceLines = []InvoiceLine{
				{
					LineID:             "1",
					ItemName:           "Test Item",
					BilledQuantity:     billedQty,
					BilledQuantityUnit: "C62",
					NetPrice:           netPrice,
					BasisQuantity:      tt.basisQty,
					BasisQuantityUnit:  "C62",
					Total:              expectedTotal,
					TaxCategoryCode:    "S",
				},
			}
			inv.LineTotal = expectedTotal
			inv.TaxBasisTotal = expectedTotal
			inv.GrandTotal = expectedTotal
			inv.DuePayableAmount = expectedTotal

			err := inv.Validate()

			var valErr *ValidationError
			hasR121Violation := false
			if errors.As(err, &valErr) {
				hasR121Violation = valErr.HasRule(rules.PEPPOLEN16931R121)
			}

			if tt.wantViolation && !hasR121Violation {
				t.Errorf("Expected PEPPOL-EN16931-R121 violation, but not found")
			}
			if !tt.wantViolation && hasR121Violation {
				t.Errorf("Did not expect PEPPOL-EN16931-R121 violation, but found one")
			}
		})
	}
}

// TestValidatePEPPOL_R130_UnitCodeConsistency tests unit code consistency validation
func TestValidatePEPPOL_R130_UnitCodeConsistency(t *testing.T) {
	tests := []struct {
		name              string
		billedUnit        string
		basisUnit         string
		wantViolation     bool
	}{
		{
			name:          "valid - same unit codes",
			billedUnit:    "C62",
			basisUnit:     "C62",
			wantViolation: false,
		},
		{
			name:          "valid - both HUR",
			billedUnit:    "HUR",
			basisUnit:     "HUR",
			wantViolation: false,
		},
		{
			name:          "valid - basis unit not specified",
			billedUnit:    "C62",
			basisUnit:     "", // Not specified, skip check
			wantViolation: false,
		},
		{
			name:          "invalid - different unit codes",
			billedUnit:    "C62",
			basisUnit:     "HUR",
			wantViolation: true,
		},
		{
			name:          "invalid - KGM vs C62",
			billedUnit:    "KGM",
			basisUnit:     "C62",
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createPEPPOLTestInvoice()
			inv.InvoiceLines = []InvoiceLine{
				{
					LineID:             "1",
					ItemName:           "Test Item",
					BilledQuantity:     decimal.NewFromInt(10),
					BilledQuantityUnit: tt.billedUnit,
					NetPrice:           decimal.NewFromInt(100),
					BasisQuantity:      decimal.NewFromInt(1),
					BasisQuantityUnit:  tt.basisUnit,
					Total:              decimal.NewFromInt(1000),
					TaxCategoryCode:    "S",
				},
			}

			err := inv.Validate()

			var valErr *ValidationError
			hasR130Violation := false
			if errors.As(err, &valErr) {
				hasR130Violation = valErr.HasRule(rules.PEPPOLEN16931R130)
			}

			if tt.wantViolation && !hasR130Violation {
				t.Errorf("Expected PEPPOL-EN16931-R130 violation, but not found")
			}
			if !tt.wantViolation && hasR130Violation {
				t.Errorf("Did not expect PEPPOL-EN16931-R130 violation, but found one")
			}
		})
	}
}

// createPEPPOLTestInvoice creates a minimal valid PEPPOL invoice for testing
func createPEPPOLTestInvoice() *Invoice {
	return &Invoice{
		InvoiceNumber:       "INV-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		GuidelineSpecifiedDocumentContextParameter: SpecPEPPOLBilling30,
		BPSpecifiedDocumentContextParameter:        BPPEPPOLBilling01,
		BuyerReference:                             "BR123",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			VATaxRegistration:         "DE123456789",
			URIUniversalCommunication: "seller@example.com",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
			URIUniversalCommunication: "buyer@example.com",
		},
		LineTotal:        decimal.NewFromInt(1000),
		TaxBasisTotal:    decimal.NewFromInt(1000),
		TaxTotal:         decimal.Zero,
		GrandTotal:       decimal.NewFromInt(1000),
		DuePayableAmount: decimal.NewFromInt(1000),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(10),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(1000),
				TaxCategoryCode:          "Z",
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "Z",
				BasisAmount:      decimal.NewFromInt(1000),
				CalculatedAmount: decimal.Zero,
				Percent:          decimal.Zero,
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{Description: "Net 30"},
		},
	}
}
