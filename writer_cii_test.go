package einvoice

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestWrite_PayeeTradeParty tests that PayeeTradeParty is written with correct XML structure
// This test verifies the fix for the critical bug where PayeeTradeParty element was missing
func TestWrite_PayeeTradeParty(t *testing.T) {
	t.Parallel()

	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "TEST-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		// BG-10: PayeeTradeParty - different from seller
		PayeeTradeParty: &Party{
			Name:              "Payment Receiver Inc",
			VATaxRegistration: "DE789012",
			PostalAddress: &PostalAddress{
				Line1:        "Payee Street 1",
				City:         "Munich",
				PostcodeCode: "80331",
				CountryID:    "DE",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Verify that PayeeTradeParty element exists
	if !strings.Contains(xmlOutput, "<ram:PayeeTradeParty>") {
		t.Error("Expected <ram:PayeeTradeParty> element to be present")
	}

	// Verify PayeeTradeParty contains the correct name
	if !strings.Contains(xmlOutput, "<ram:Name>Payment Receiver Inc</ram:Name>") {
		t.Error("Expected PayeeTradeParty to contain correct name")
	}

	// Verify PayeeTradeParty contains postal address
	if !strings.Contains(xmlOutput, "<ram:PostcodeCode>80331</ram:PostcodeCode>") {
		t.Error("Expected PayeeTradeParty to contain postal address")
	}

	// Verify PayeeTradeParty contains VAT registration
	if !strings.Contains(xmlOutput, "DE789012") {
		t.Error("Expected PayeeTradeParty to contain VAT registration")
	}

	// Verify the structure: PayeeTradeParty should be inside ApplicableHeaderTradeSettlement
	// and should have proper child elements, not be a sibling
	payeeStartIdx := strings.Index(xmlOutput, "<ram:PayeeTradeParty>")
	payeeEndIdx := strings.Index(xmlOutput, "</ram:PayeeTradeParty>")

	if payeeStartIdx == -1 || payeeEndIdx == -1 {
		t.Fatal("PayeeTradeParty tags not found")
	}

	// Extract the PayeeTradeParty section
	payeeSection := xmlOutput[payeeStartIdx:payeeEndIdx]

	// Verify it contains child elements (not just closing tag)
	if !strings.Contains(payeeSection, "<ram:Name>") {
		t.Error("PayeeTradeParty should contain child elements like <ram:Name>")
	}
}

// TestWrite_MultiCurrencyTaxTotal tests that BT-111 (TaxTotalAmount in accounting currency)
// is written when TaxCurrencyCode differs from InvoiceCurrencyCode
func TestWrite_MultiCurrencyTaxTotal(t *testing.T) {
	t.Parallel()

	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "MULTI-CURR-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "USD", // BT-5: Invoice in USD
		TaxCurrencyCode:     "EUR", // BT-6: Tax accounting in EUR
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "New York",
				PostcodeCode: "10001",
				CountryID:    "US",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
		// BT-110: Tax total in invoice currency (USD)
		TaxTotal:         decimal.NewFromInt(19),
		TaxTotalCurrency: "USD",
		// BT-111: Tax total in accounting currency (EUR) - at exchange rate
		TaxTotalVAT:         decimal.NewFromFloat(17.50),
		TaxTotalVATCurrency: "EUR",
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Count TaxTotalAmount elements
	taxTotalCount := strings.Count(xmlOutput, "<ram:TaxTotalAmount")
	if taxTotalCount != 2 {
		t.Errorf("Expected 2 TaxTotalAmount elements (BT-110 and BT-111), got %d", taxTotalCount)
	}

	// Verify BT-110: TaxTotalAmount in invoice currency (USD)
	if !strings.Contains(xmlOutput, `currencyID="USD"`) {
		t.Error("Expected TaxTotalAmount with currencyID='USD' (BT-110)")
	}
	if !strings.Contains(xmlOutput, `currencyID="USD">19.00<`) {
		t.Error("Expected TaxTotalAmount in USD to be 19.00")
	}

	// Verify BT-111: TaxTotalAmount in accounting currency (EUR)
	if !strings.Contains(xmlOutput, `currencyID="EUR"`) {
		t.Error("Expected TaxTotalAmount with currencyID='EUR' (BT-111)")
	}
	if !strings.Contains(xmlOutput, `currencyID="EUR">17.50<`) {
		t.Error("Expected TaxTotalAmount in EUR to be 17.50")
	}
}

// TestWrite_SingleCurrencyTaxTotal tests that only ONE TaxTotalAmount is written
// when invoice and tax currencies are the same
func TestWrite_SingleCurrencyTaxTotal(t *testing.T) {
	t.Parallel()

	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "SINGLE-CURR-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "EUR", // BT-5
		TaxCurrencyCode:     "EUR", // BT-6: Same as invoice currency
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
		TaxTotal: decimal.NewFromInt(19),
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Count TaxTotalAmount elements - should be only 1 when currencies match
	taxTotalCount := strings.Count(xmlOutput, "<ram:TaxTotalAmount")
	if taxTotalCount != 1 {
		t.Errorf("Expected 1 TaxTotalAmount element when currencies match, got %d", taxTotalCount)
	}

	// Verify it has EUR currency
	if !strings.Contains(xmlOutput, `currencyID="EUR">19.00<`) {
		t.Error("Expected single TaxTotalAmount in EUR to be 19.00")
	}
}

// TestWrite_NoTaxCurrencyCode tests backward compatibility when TaxCurrencyCode is not set
func TestWrite_NoTaxCurrencyCode(t *testing.T) {
	t.Parallel()

	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "NO-TAX-CURR-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "EUR",
		// TaxCurrencyCode not set (empty string)
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
		TaxTotal: decimal.NewFromInt(19),
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Should have only 1 TaxTotalAmount
	taxTotalCount := strings.Count(xmlOutput, "<ram:TaxTotalAmount")
	if taxTotalCount != 1 {
		t.Errorf("Expected 1 TaxTotalAmount element when TaxCurrencyCode not set, got %d", taxTotalCount)
	}
}

// TestWrite_BillingPeriod_OnlyEndDate tests that document-level billing period
// is written correctly when only end date is provided (Bug #3)
// BR-CO-19 states: "either start date OR end date OR both must be filled"
func TestWrite_BillingPeriod_OnlyEndDate(t *testing.T) {
	t.Parallel()

	invoiceDate, _ := time.Parse("02.01.2006", "31.12.2025")
	periodEnd, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "PERIOD-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         invoiceDate,
		InvoiceCurrencyCode: "EUR",
		// BG-14: Billing period with ONLY end date (no start date)
		BillingSpecifiedPeriodStart: time.Time{}, // Zero value
		BillingSpecifiedPeriodEnd:   periodEnd,
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// BUG: The billing period should be written even with only end date
	// The current code only writes if start date is non-zero
	if !strings.Contains(xmlOutput, "<ram:BillingSpecifiedPeriod>") {
		t.Error("BillingSpecifiedPeriod element should be written when end date is provided (BR-CO-19)")
	}

	// Should have EndDateTime element
	if !strings.Contains(xmlOutput, "<ram:EndDateTime>") {
		t.Error("EndDateTime should be present in BillingSpecifiedPeriod")
	}

	// Should NOT write "00010101" for zero start date
	if strings.Contains(xmlOutput, "00010101") {
		t.Error("Should not write zero date value (00010101) for missing start date")
	}

	// Should NOT have StartDateTime element when start date is zero
	periodStartIdx := strings.Index(xmlOutput, "<ram:BillingSpecifiedPeriod>")
	periodEndIdx := strings.Index(xmlOutput, "</ram:BillingSpecifiedPeriod>")
	if periodStartIdx != -1 && periodEndIdx != -1 {
		periodSection := xmlOutput[periodStartIdx:periodEndIdx]
		if strings.Contains(periodSection, "<ram:StartDateTime>") {
			t.Error("StartDateTime should NOT be written when start date is zero")
		}
	}
}

// TestWrite_BillingPeriod_OnlyStartDate tests that document-level billing period
// is written correctly when only start date is provided (Bug #3)
func TestWrite_BillingPeriod_OnlyStartDate(t *testing.T) {
	t.Parallel()

	invoiceDate, _ := time.Parse("02.01.2006", "31.12.2025")
	periodStart, _ := time.Parse("02.01.2006", "01.12.2025")

	inv := Invoice{
		InvoiceNumber:   "PERIOD-002",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         invoiceDate,
		InvoiceCurrencyCode: "EUR",
		// BG-14: Billing period with ONLY start date (no end date)
		BillingSpecifiedPeriodStart: periodStart,
		BillingSpecifiedPeriodEnd:   time.Time{}, // Zero value
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Should have BillingSpecifiedPeriod element
	if !strings.Contains(xmlOutput, "<ram:BillingSpecifiedPeriod>") {
		t.Error("BillingSpecifiedPeriod element should be written when start date is provided")
	}

	// Should have StartDateTime element
	if !strings.Contains(xmlOutput, "<ram:StartDateTime>") {
		t.Error("StartDateTime should be present in BillingSpecifiedPeriod")
	}

	// Should NOT write "00010101" for zero end date
	if strings.Contains(xmlOutput, "00010101") {
		t.Error("Should not write zero date value (00010101) for missing end date")
	}

	// Should NOT have EndDateTime element when end date is zero
	periodStartIdx := strings.Index(xmlOutput, "<ram:BillingSpecifiedPeriod>")
	periodEndIdx := strings.Index(xmlOutput, "</ram:BillingSpecifiedPeriod>")
	if periodStartIdx != -1 && periodEndIdx != -1 {
		periodSection := xmlOutput[periodStartIdx:periodEndIdx]
		if strings.Contains(periodSection, "<ram:EndDateTime>") {
			t.Error("EndDateTime should NOT be written when end date is zero")
		}
	}
}

// TestWrite_BillingPeriod_BothDates tests that document-level billing period
// is written correctly when both start and end dates are provided
func TestWrite_BillingPeriod_BothDates(t *testing.T) {
	t.Parallel()

	invoiceDate, _ := time.Parse("02.01.2006", "31.12.2025")
	periodStart, _ := time.Parse("02.01.2006", "01.12.2025")
	periodEnd, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "PERIOD-003",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         invoiceDate,
		InvoiceCurrencyCode: "EUR",
		// BG-14: Billing period with both dates
		BillingSpecifiedPeriodStart: periodStart,
		BillingSpecifiedPeriodEnd:   periodEnd,
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Should have BillingSpecifiedPeriod element
	if !strings.Contains(xmlOutput, "<ram:BillingSpecifiedPeriod>") {
		t.Error("BillingSpecifiedPeriod element should be written when both dates are provided")
	}

	// Should have both StartDateTime and EndDateTime elements
	if !strings.Contains(xmlOutput, "<ram:StartDateTime>") {
		t.Error("StartDateTime should be present in BillingSpecifiedPeriod")
	}
	if !strings.Contains(xmlOutput, "<ram:EndDateTime>") {
		t.Error("EndDateTime should be present in BillingSpecifiedPeriod")
	}

	// Extract the dates and verify they're correct
	if !strings.Contains(xmlOutput, "20251201") { // Start date: 01.12.2025
		t.Error("StartDateTime should contain correct date value (20251201)")
	}
	if !strings.Contains(xmlOutput, "20251231") { // End date: 31.12.2025
		t.Error("EndDateTime should contain correct date value (20251231)")
	}
}

// TestWrite_InvoiceLineAllowances tests that invoice line allowances (BG-27)
// are written correctly (Bug #1)
func TestWrite_InvoiceLineAllowances(t *testing.T) {
	t.Parallel()

	invoiceDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "LINE-ALLOW-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         invoiceDate,
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(90),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(90),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
				// BG-27: Invoice line allowances
				InvoiceLineAllowances: []AllowanceCharge{
					{
						ChargeIndicator:                       false,
						BasisAmount:                           decimal.NewFromInt(100),
						ActualAmount:                          decimal.NewFromInt(10),
						Reason:                                "Volume discount",
						CategoryTradeTaxType:                  "VAT",
						CategoryTradeTaxCategoryCode:          "S",
						CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
					},
				},
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// BUG: Invoice line allowances should be written as SpecifiedTradeAllowanceCharge
	// under SpecifiedLineTradeSettlement
	if !strings.Contains(xmlOutput, "<ram:SpecifiedLineTradeSettlement>") {
		t.Error("SpecifiedLineTradeSettlement should be present")
	}

	// Should contain SpecifiedTradeAllowanceCharge within SpecifiedLineTradeSettlement
	settlementStart := strings.Index(xmlOutput, "<ram:SpecifiedLineTradeSettlement>")
	settlementEnd := strings.Index(xmlOutput, "</ram:SpecifiedLineTradeSettlement>")

	if settlementStart == -1 || settlementEnd == -1 {
		t.Fatal("Could not find SpecifiedLineTradeSettlement section")
	}

	settlementSection := xmlOutput[settlementStart:settlementEnd]

	if !strings.Contains(settlementSection, "<ram:SpecifiedTradeAllowanceCharge>") {
		t.Error("SpecifiedTradeAllowanceCharge should be present in SpecifiedLineTradeSettlement for invoice line allowances (BG-27)")
	}

	// Should have ChargeIndicator = false for allowance
	if !strings.Contains(settlementSection, "<ram:ChargeIndicator>") {
		t.Error("ChargeIndicator should be present in line allowance")
	}
	if !strings.Contains(settlementSection, "false") {
		t.Error("ChargeIndicator should be false for allowance")
	}

	// Should have ActualAmount
	if !strings.Contains(settlementSection, "<ram:ActualAmount>10") {
		t.Error("ActualAmount should be 10 for the allowance")
	}

	// Should have Reason
	if !strings.Contains(settlementSection, "Volume discount") {
		t.Error("Allowance reason should be present")
	}
}

// TestWrite_InvoiceLineCharges tests that invoice line charges (BG-28)
// are written correctly (Bug #1)
func TestWrite_InvoiceLineCharges(t *testing.T) {
	t.Parallel()

	invoiceDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := Invoice{
		InvoiceNumber:   "LINE-CHARGE-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         invoiceDate,
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Item",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(105),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(105),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
				// BG-28: Invoice line charges
				InvoiceLineCharges: []AllowanceCharge{
					{
						ChargeIndicator:                       true,
						ActualAmount:                          decimal.NewFromInt(5),
						Reason:                                "Handling fee",
						CategoryTradeTaxType:                  "VAT",
						CategoryTradeTaxCategoryCode:          "S",
						CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
					},
				},
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// Extract SpecifiedLineTradeSettlement section
	settlementStart := strings.Index(xmlOutput, "<ram:SpecifiedLineTradeSettlement>")
	settlementEnd := strings.Index(xmlOutput, "</ram:SpecifiedLineTradeSettlement>")

	if settlementStart == -1 || settlementEnd == -1 {
		t.Fatal("Could not find SpecifiedLineTradeSettlement section")
	}

	settlementSection := xmlOutput[settlementStart:settlementEnd]

	if !strings.Contains(settlementSection, "<ram:SpecifiedTradeAllowanceCharge>") {
		t.Error("SpecifiedTradeAllowanceCharge should be present in SpecifiedLineTradeSettlement for invoice line charges (BG-28)")
	}

	// Should have ChargeIndicator = true for charge
	if !strings.Contains(settlementSection, "true") {
		t.Error("ChargeIndicator should be true for charge")
	}

	// Should have ActualAmount
	if !strings.Contains(settlementSection, "<ram:ActualAmount>5") {
		t.Error("ActualAmount should be 5 for the charge")
	}

	// Should have Reason
	if !strings.Contains(settlementSection, "Handling fee") {
		t.Error("Charge reason should be present")
	}
}

// TestWrite_InvoiceLineBillingPeriod tests that invoice line billing period
// is written correctly (Bug #2)
func TestWrite_InvoiceLineBillingPeriod(t *testing.T) {
	t.Parallel()

	invoiceDate, _ := time.Parse("02.01.2006", "31.12.2025")
	linePeriodStart, _ := time.Parse("02.01.2006", "01.11.2025")
	linePeriodEnd, _ := time.Parse("02.01.2006", "30.11.2025")

	inv := Invoice{
		InvoiceNumber:   "LINE-PERIOD-001",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         invoiceDate,
		InvoiceCurrencyCode: "EUR",
		Seller: Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer Company",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Monthly Service",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(100),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
				// BT-134, BT-135: Invoice line billing period
				BillingSpecifiedPeriodStart: linePeriodStart,
				BillingSpecifiedPeriodEnd:   linePeriodEnd,
			},
		},
	}

	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	xmlOutput := buf.String()

	// BUG: Invoice line billing period should be written
	// Extract SpecifiedLineTradeSettlement section
	settlementStart := strings.Index(xmlOutput, "<ram:SpecifiedLineTradeSettlement>")
	settlementEnd := strings.Index(xmlOutput, "</ram:SpecifiedLineTradeSettlement>")

	if settlementStart == -1 || settlementEnd == -1 {
		t.Fatal("Could not find SpecifiedLineTradeSettlement section")
	}

	settlementSection := xmlOutput[settlementStart:settlementEnd]

	if !strings.Contains(settlementSection, "<ram:BillingSpecifiedPeriod>") {
		t.Error("BillingSpecifiedPeriod should be present in invoice line (BG-26)")
	}

	// Should have both StartDateTime and EndDateTime
	if !strings.Contains(settlementSection, "<ram:StartDateTime>") {
		t.Error("StartDateTime should be present in line billing period")
	}
	if !strings.Contains(settlementSection, "<ram:EndDateTime>") {
		t.Error("EndDateTime should be present in line billing period")
	}

	// Verify dates are correct
	if !strings.Contains(settlementSection, "20251101") {
		t.Error("StartDateTime should contain 20251101")
	}
	if !strings.Contains(settlementSection, "20251130") {
		t.Error("EndDateTime should contain 20251130")
	}
}

// TestFormatPercent tests the formatPercent function with various percentage values
func TestFormatPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    decimal.Decimal
		expected string
	}{
		{
			name:     "integer percentage",
			input:    decimal.NewFromInt(19),
			expected: "19",
		},
		{
			name:     "decimal percentage with trailing zeros",
			input:    decimal.NewFromFloat(19.5000),
			expected: "19.5",
		},
		{
			name:     "zero percentage",
			input:    decimal.NewFromInt(0),
			expected: "0",
		},
		{
			name:     "high precision percentage",
			input:    decimal.NewFromFloat(7.7),
			expected: "7.7",
		},
		{
			name:     "percentage with four decimals",
			input:    decimal.NewFromFloat(19.2345),
			expected: "19.2345",
		},
		{
			name:     "percentage with two decimals",
			input:    decimal.NewFromFloat(5.25),
			expected: "5.25",
		},
		{
			name:     "small percentage",
			input:    decimal.NewFromFloat(0.5),
			expected: "0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPercent(tt.input)
			if result != tt.expected {
				t.Errorf("formatPercent(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// BenchmarkWriteCII benchmarks CII XML writing performance
func BenchmarkWriteCII(b *testing.B) {
	// Load a sample invoice to write
	data, err := os.ReadFile("testdata/cii/en16931/zugferd_2p3_EN16931_1.xml")
	if err != nil {
		b.Skipf("Test fixture not found: %v", err)
	}

	inv, err := ParseReader(bytes.NewReader(data))
	if err != nil {
		b.Fatalf("Failed to parse fixture: %v", err)
	}

	// Pre-allocate buffer outside the benchmark loop
	var buf bytes.Buffer

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		buf.Reset()
		err := inv.Write(&buf)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(buf.Len()))
	}
}
