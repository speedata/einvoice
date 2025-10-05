package einvoice_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice"
)

// TestWrite_PayeeTradeParty tests that PayeeTradeParty is written with correct XML structure
// This test verifies the fix for the critical bug where PayeeTradeParty element was missing
func TestWrite_PayeeTradeParty(t *testing.T) {
	t.Parallel()

	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")

	inv := einvoice.Invoice{
		InvoiceNumber:       "TEST-001",
		InvoiceTypeCode:     380,
		Profile:             einvoice.CProfileEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "EUR",
		Seller: einvoice.Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: einvoice.Party{
			Name: "Buyer Company",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		// BG-10: PayeeTradeParty - different from seller
		PayeeTradeParty: &einvoice.Party{
			Name:              "Payment Receiver Inc",
			VATaxRegistration: "DE789012",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Payee Street 1",
				City:         "Munich",
				PostcodeCode: "80331",
				CountryID:    "DE",
			},
		},
		InvoiceLines: []einvoice.InvoiceLine{
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

	inv := einvoice.Invoice{
		InvoiceNumber:       "MULTI-CURR-001",
		InvoiceTypeCode:     380,
		Profile:             einvoice.CProfileEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "USD",        // BT-5: Invoice in USD
		TaxCurrencyCode:     "EUR",        // BT-6: Tax accounting in EUR
		Seller: einvoice.Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: einvoice.Party{
			Name: "Buyer Company",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "New York",
				PostcodeCode: "10001",
				CountryID:    "US",
			},
		},
		InvoiceLines: []einvoice.InvoiceLine{
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

	inv := einvoice.Invoice{
		InvoiceNumber:       "SINGLE-CURR-001",
		InvoiceTypeCode:     380,
		Profile:             einvoice.CProfileEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "EUR",  // BT-5
		TaxCurrencyCode:     "EUR",  // BT-6: Same as invoice currency
		Seller: einvoice.Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: einvoice.Party{
			Name: "Buyer Company",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []einvoice.InvoiceLine{
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

	inv := einvoice.Invoice{
		InvoiceNumber:       "NO-TAX-CURR-001",
		InvoiceTypeCode:     380,
		Profile:             einvoice.CProfileEN16931,
		InvoiceDate:         fixedDate,
		InvoiceCurrencyCode: "EUR",
		// TaxCurrencyCode not set (empty string)
		Seller: einvoice.Party{
			Name:              "Seller Company",
			VATaxRegistration: "DE123456",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Seller Street 1",
				City:         "Berlin",
				PostcodeCode: "10115",
				CountryID:    "DE",
			},
		},
		Buyer: einvoice.Party{
			Name: "Buyer Company",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Buyer Street 1",
				City:         "Paris",
				PostcodeCode: "75001",
				CountryID:    "FR",
			},
		},
		InvoiceLines: []einvoice.InvoiceLine{
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
