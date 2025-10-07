package einvoice

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// mustParseTime is a helper function that parses a date string in YYYYMMDD format
func mustParseTime(dateStr string) time.Time {
	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		panic(err)
	}
	return t
}

// TestBR46_AllowsZeroCalculatedAmount verifies that BR-46 accepts zero VAT amounts
// for exempt categories (fix for critical bug)
func TestBR46_AllowsZeroCalculatedAmount(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "TEST-001",
		InvoiceDate:         mustParseTime("20240101"),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromFloat(100),
		TaxBasisTotal:       decimal.NewFromFloat(100),
		TaxTotal:            decimal.Zero, // Zero is valid
		GrandTotal:          decimal.NewFromFloat(100),
		DuePayableAmount:    decimal.NewFromFloat(100),
		Seller: Party{
			Name:              "Test Seller",
			VATaxRegistration: "DE123456789",
			PostalAddress:     &PostalAddress{CountryID: "DE"},
		},
		Buyer: Party{
			Name:          "Test Buyer",
			PostalAddress: &PostalAddress{CountryID: "DE"},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "EA",
				NetPrice:                 decimal.NewFromFloat(100),
				Total:                    decimal.NewFromFloat(100),
				TaxCategoryCode:          "E", // Exempt from VAT
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.Zero, // Zero amount for exempt category
				BasisAmount:      decimal.NewFromFloat(100),
				Typ:              "VAT",
				CategoryCode:     "E",
				Percent:          decimal.Zero, // Zero rate for exempt category
				ExemptionReason:  "Exempt from VAT",
			},
		},
	}

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRule("BR-46") {
			t.Errorf("BR-46 should not fail on zero CalculatedAmount for exempt categories")
		}
	}
}

// TestBR48_AllowsZeroPercent verifies that BR-48 accepts zero VAT rates
// for exempt categories (fix for critical bug)
func TestBR48_AllowsZeroPercent(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "TEST-002",
		InvoiceDate:         mustParseTime("20240101"),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromFloat(100),
		TaxBasisTotal:       decimal.NewFromFloat(100),
		TaxTotal:            decimal.Zero,
		GrandTotal:          decimal.NewFromFloat(100),
		DuePayableAmount:    decimal.NewFromFloat(100),
		Seller: Party{
			Name:              "Test Seller",
			VATaxRegistration: "DE123456789",
			PostalAddress:     &PostalAddress{CountryID: "DE"},
		},
		Buyer: Party{
			Name:          "Test Buyer",
			PostalAddress: &PostalAddress{CountryID: "DE"},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "EA",
				NetPrice:                 decimal.NewFromFloat(100),
				Total:                    decimal.NewFromFloat(100),
				TaxCategoryCode:          "AE", // Reverse charge
				TaxRateApplicablePercent: decimal.Zero,
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.Zero,
				BasisAmount:      decimal.NewFromFloat(100),
				Typ:              "VAT",
				CategoryCode:     "AE",
				Percent:          decimal.Zero, // Zero rate is required for reverse charge
				ExemptionReason:  "Reverse charge",
			},
		},
	}

	// Add buyer VAT ID for reverse charge
	inv.Buyer.VATaxRegistration = "FR987654321"

	err := inv.Validate()
	if err != nil {
		valErr, ok := err.(*ValidationError)
		if ok && valErr.HasRule("BR-48") {
			t.Errorf("BR-48 should not fail on zero Percent for exempt categories")
		}
	}
}

// TestBRCO11_ValidatesAllowanceTotal verifies that allowance totals are validated
func TestBRCO11_ValidatesAllowanceTotal(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "TEST-003",
		InvoiceDate:         mustParseTime("20240101"),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromFloat(1000),
		AllowanceTotal:      decimal.NewFromFloat(999), // Wrong! Should be 150
		ChargeTotal:         decimal.Zero,
		TaxBasisTotal:       decimal.NewFromFloat(850),
		TaxTotal:            decimal.NewFromFloat(161.5),
		GrandTotal:          decimal.NewFromFloat(1011.5),
		DuePayableAmount:    decimal.NewFromFloat(1011.5),
		Seller: Party{
			Name:              "Test Seller",
			VATaxRegistration: "DE123456789",
			PostalAddress:     &PostalAddress{CountryID: "DE"},
		},
		Buyer: Party{
			Name:          "Test Buyer",
			PostalAddress: &PostalAddress{CountryID: "DE"},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "EA",
				NetPrice:                 decimal.NewFromFloat(1000),
				Total:                    decimal.NewFromFloat(1000),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromFloat(100),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
				Reason:                                "Discount",
			},
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromFloat(50),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
				Reason:                                "Early payment",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromFloat(161.5),
				BasisAmount:      decimal.NewFromFloat(850),
				Typ:              "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
			},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for incorrect AllowanceTotal")
	}

	valErr := err.(*ValidationError)
	if !valErr.HasRule("BR-CO-11") {
		t.Errorf("Expected BR-CO-11 violation for incorrect allowance total")
	}
}

// TestBRCO12_ValidatesChargeTotal verifies that charge totals are validated
func TestBRCO12_ValidatesChargeTotal(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "TEST-004",
		InvoiceDate:         mustParseTime("20240101"),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromFloat(1000),
		AllowanceTotal:      decimal.Zero,
		ChargeTotal:         decimal.NewFromFloat(999), // Wrong! Should be 50
		TaxBasisTotal:       decimal.NewFromFloat(1050),
		TaxTotal:            decimal.NewFromFloat(199.5),
		GrandTotal:          decimal.NewFromFloat(1249.5),
		DuePayableAmount:    decimal.NewFromFloat(1249.5),
		Seller: Party{
			Name:              "Test Seller",
			VATaxRegistration: "DE123456789",
			PostalAddress:     &PostalAddress{CountryID: "DE"},
		},
		Buyer: Party{
			Name:          "Test Buyer",
			PostalAddress: &PostalAddress{CountryID: "DE"},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "EA",
				NetPrice:                 decimal.NewFromFloat(1000),
				Total:                    decimal.NewFromFloat(1000),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       true,
				ActualAmount:                          decimal.NewFromFloat(50),
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(19),
				Reason:                                "Shipping",
				ReasonCode:                            1,
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromFloat(199.5),
				BasisAmount:      decimal.NewFromFloat(1050),
				Typ:              "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
			},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for incorrect ChargeTotal")
	}

	valErr := err.(*ValidationError)
	if !valErr.HasRule("BR-CO-12") {
		t.Errorf("Expected BR-CO-12 violation for incorrect charge total")
	}
}

// TestUpdateAllowancesAndCharges verifies the new calculation function
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

	inv.UpdateAllowancesAndCharges()

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

	inv.UpdateAllowancesAndCharges()
	first := inv.AllowanceTotal

	inv.UpdateAllowancesAndCharges()
	second := inv.AllowanceTotal

	if !first.Equal(second) {
		t.Errorf("UpdateAllowancesAndCharges not idempotent: first=%s, second=%s", first, second)
	}
}

// TestBR20_ErrorMessage verifies the corrected error message
func TestBR20_ErrorMessage(t *testing.T) {
	inv := &Invoice{
		Profile:             CProfileEN16931,
		InvoiceNumber:       "TEST-005",
		InvoiceDate:         mustParseTime("20240101"),
		InvoiceTypeCode:     380,
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromFloat(100),
		TaxBasisTotal:       decimal.NewFromFloat(100),
		TaxTotal:            decimal.NewFromFloat(19),
		GrandTotal:          decimal.NewFromFloat(119),
		DuePayableAmount:    decimal.NewFromFloat(119),
		Seller: Party{
			Name:              "Test Seller",
			VATaxRegistration: "DE123456789",
			PostalAddress:     &PostalAddress{CountryID: "DE"},
		},
		SellerTaxRepresentativeTradeParty: &Party{
			Name:              "Tax Rep",
			VATaxRegistration: "FR123456789",
			PostalAddress:     &PostalAddress{}, // Missing CountryID
		},
		Buyer: Party{
			Name:          "Test Buyer",
			PostalAddress: &PostalAddress{CountryID: "DE"},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "EA",
				NetPrice:                 decimal.NewFromFloat(100),
				Total:                    decimal.NewFromFloat(100),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromFloat(19),
				BasisAmount:      decimal.NewFromFloat(100),
				Typ:              "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
			},
		},
	}

	err := inv.Validate()
	if err == nil {
		t.Fatal("Expected validation error for missing country code")
	}

	valErr := err.(*ValidationError)
	if !valErr.HasRule("BR-20") {
		t.Fatalf("Expected BR-20 violation")
	}

	// Check that the error message mentions country code
	for _, v := range valErr.Violations() {
		if v.Rule == "BR-20" {
			if v.Text != "Tax representative postal address missing country code" {
				t.Errorf("BR-20 error message = %q, want 'Tax representative postal address missing country code'", v.Text)
			}
		}
	}
}
