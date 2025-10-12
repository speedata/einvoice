package einvoice

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestWriteUBL_BasicInvoice tests writing a basic UBL invoice
func TestWriteUBL_BasicInvoice(t *testing.T) {
	inv := &Invoice{
		SchemaType:                                 UBL,
		GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0",
		BPSpecifiedDocumentContextParameter:        "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
		InvoiceNumber:                              "INV-001",
		InvoiceTypeCode:                            380,
		InvoiceDate:                                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceCurrencyCode:                        "EUR",
		Seller: Party{
			Name: "Seller Inc",
			PostalAddress: &PostalAddress{
				CountryID:    "DE",
				PostcodeCode: "10115",
				City:         "Berlin",
				Line1:        "Main Street 1",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Buyer GmbH",
			PostalAddress: &PostalAddress{
				CountryID:    "FR",
				PostcodeCode: "75001",
				City:         "Paris",
				Line1:        "Rue de la Paix 1",
			},
			VATaxRegistration: "FR987654321",
		},
		LineTotal:        decimal.NewFromInt(100),
		TaxBasisTotal:    decimal.NewFromInt(100),
		TaxTotal:         decimal.NewFromInt(19),
		GrandTotal:       decimal.NewFromInt(119),
		DuePayableAmount: decimal.NewFromInt(119),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Product A",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxTypeCode:              "VAT",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				TypeCode:         "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
	}

	var buf bytes.Buffer
	err := inv.Write(&buf)
	if err != nil {
		t.Fatalf("Failed to write invoice: %v", err)
	}

	xmlOutput := buf.String()

	// Verify basic structure
	requiredElements := []string{
		"<Invoice",
		"xmlns=\"urn:oasis:names:specification:ubl:schema:xsd:Invoice-2\"",
		"xmlns:cac=\"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2\"",
		"xmlns:cbc=\"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2\"",
		"<cbc:CustomizationID>urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0</cbc:CustomizationID>",
		"<cbc:ID>INV-001</cbc:ID>",
		"<cbc:IssueDate>2024-01-15</cbc:IssueDate>",
		"<cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>",
		"<cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>",
		"<cac:AccountingSupplierParty>",
		"<cac:AccountingCustomerParty>",
		"<cac:TaxTotal>",
		"<cac:LegalMonetaryTotal>",
		"<cac:InvoiceLine>",
	}

	for _, required := range requiredElements {
		if !strings.Contains(xmlOutput, required) {
			t.Errorf("Expected element not found in XML: %s", required)
		}
	}

	// Verify seller information
	if !strings.Contains(xmlOutput, "<cbc:CompanyID>DE123456789</cbc:CompanyID>") {
		t.Error("Seller VAT ID not found in XML")
	}

	// Verify buyer information
	if !strings.Contains(xmlOutput, "<cbc:CompanyID>FR987654321</cbc:CompanyID>") {
		t.Error("Buyer VAT ID not found in XML")
	}
}

// TestWriteUBL_CreditNote tests writing a UBL CreditNote (type code 381)
func TestWriteUBL_CreditNote(t *testing.T) {
	inv := &Invoice{
		SchemaType:                                 UBL,
		GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017",
		InvoiceNumber:                              "CN-001",
		InvoiceTypeCode:                            381, // Credit Note
		InvoiceDate:                                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceCurrencyCode:                        "EUR",
		Seller: Party{
			Name: "Seller Inc",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
				City:      "Berlin",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Buyer GmbH",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
				City:      "Paris",
			},
			VATaxRegistration: "FR987654321",
		},
		LineTotal:        decimal.NewFromInt(50),
		TaxBasisTotal:    decimal.NewFromInt(50),
		TaxTotal:         decimal.RequireFromString("9.50"),
		GrandTotal:       decimal.RequireFromString("59.50"),
		DuePayableAmount: decimal.RequireFromString("59.50"),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Refund for Product A",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(50),
				Total:                    decimal.NewFromInt(50),
				TaxCategoryCode:          "S",
				TaxTypeCode:              "VAT",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.RequireFromString("9.50"),
				BasisAmount:      decimal.NewFromInt(50),
				TypeCode:         "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
	}

	var buf bytes.Buffer
	err := inv.Write(&buf)
	if err != nil {
		t.Fatalf("Failed to write credit note: %v", err)
	}

	xmlOutput := buf.String()

	// Verify it's a CreditNote, not an Invoice
	if !strings.Contains(xmlOutput, "<CreditNote") {
		t.Error("Expected <CreditNote> root element for type code 381")
	}

	if strings.Contains(xmlOutput, "<Invoice") {
		t.Error("Should not contain <Invoice> root element for CreditNote")
	}

	// Verify namespace
	if !strings.Contains(xmlOutput, "xmlns=\"urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2\"") {
		t.Error("Expected CreditNote namespace")
	}
}

// TestWriteUBL_RoundTrip tests that we can write and then parse back a UBL invoice
func TestWriteUBL_RoundTrip(t *testing.T) {
	original := &Invoice{
		SchemaType:                                 UBL,
		GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017",
		InvoiceNumber:                              "INV-RoundTrip",
		InvoiceTypeCode:                            380,
		InvoiceDate:                                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceCurrencyCode:                        "EUR",
		BuyerReference:                             "REF-123",
		Seller: Party{
			Name: "Test Seller",
			PostalAddress: &PostalAddress{
				CountryID:    "DE",
				PostcodeCode: "10115",
				City:         "Berlin",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Test Buyer",
			PostalAddress: &PostalAddress{
				CountryID:    "FR",
				PostcodeCode: "75001",
				City:         "Paris",
			},
			VATaxRegistration: "FR987654321",
		},
		LineTotal:        decimal.NewFromInt(200),
		TaxBasisTotal:    decimal.NewFromInt(200),
		TaxTotal:         decimal.NewFromInt(38),
		GrandTotal:       decimal.NewFromInt(238),
		DuePayableAmount: decimal.NewFromInt(238),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Test Product",
				BilledQuantity:           decimal.NewFromInt(2),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(200),
				TaxCategoryCode:          "S",
				TaxTypeCode:              "VAT",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromInt(38),
				BasisAmount:      decimal.NewFromInt(200),
				TypeCode:         "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
	}

	// Write to buffer
	var buf bytes.Buffer
	err := original.Write(&buf)
	if err != nil {
		t.Fatalf("Failed to write invoice: %v", err)
	}

	// Parse back
	parsed, err := ParseReader(&buf)
	if err != nil {
		t.Fatalf("Failed to parse written invoice: %v", err)
	}

	// Verify key fields
	if parsed.SchemaType != UBL {
		t.Errorf("Expected SchemaType UBL, got %v", parsed.SchemaType)
	}

	if parsed.InvoiceNumber != original.InvoiceNumber {
		t.Errorf("InvoiceNumber mismatch: got %s, want %s", parsed.InvoiceNumber, original.InvoiceNumber)
	}

	if parsed.InvoiceTypeCode != original.InvoiceTypeCode {
		t.Errorf("InvoiceTypeCode mismatch: got %d, want %d", parsed.InvoiceTypeCode, original.InvoiceTypeCode)
	}

	if parsed.InvoiceCurrencyCode != original.InvoiceCurrencyCode {
		t.Errorf("InvoiceCurrencyCode mismatch: got %s, want %s", parsed.InvoiceCurrencyCode, original.InvoiceCurrencyCode)
	}

	if !parsed.GrandTotal.Equal(original.GrandTotal) {
		t.Errorf("GrandTotal mismatch: got %s, want %s", parsed.GrandTotal, original.GrandTotal)
	}

	if len(parsed.InvoiceLines) != len(original.InvoiceLines) {
		t.Errorf("InvoiceLines count mismatch: got %d, want %d", len(parsed.InvoiceLines), len(original.InvoiceLines))
	}

	if len(parsed.InvoiceLines) > 0 {
		if parsed.InvoiceLines[0].ItemName != original.InvoiceLines[0].ItemName {
			t.Errorf("Line ItemName mismatch: got %s, want %s", parsed.InvoiceLines[0].ItemName, original.InvoiceLines[0].ItemName)
		}
	}
}

// TestWriteUBL_WithAllowancesAndCharges tests writing UBL with document-level allowances and charges
func TestWriteUBL_WithAllowancesAndCharges(t *testing.T) {
	inv := &Invoice{
		SchemaType:                                 UBL,
		GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017",
		InvoiceNumber:                              "INV-AC-001",
		InvoiceTypeCode:                            380,
		InvoiceDate:                                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceCurrencyCode:                        "EUR",
		Seller: Party{
			Name: "Seller Inc",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
				City:      "Berlin",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Buyer GmbH",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
				City:      "Paris",
			},
			VATaxRegistration: "FR987654321",
		},
		LineTotal:        decimal.NewFromInt(100),
		AllowanceTotal:   decimal.NewFromInt(10),
		ChargeTotal:      decimal.NewFromInt(5),
		TaxBasisTotal:    decimal.NewFromInt(95),
		TaxTotal:         decimal.RequireFromString("18.05"),
		GrandTotal:       decimal.RequireFromString("113.05"),
		DuePayableAmount: decimal.RequireFromString("113.05"),
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false, // Allowance
				ActualAmount:                          decimal.NewFromInt(10),
				Reason:                                "Early payment discount",
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
				CategoryTradeTaxType:                  "VAT",
			},
			{
				ChargeIndicator:                       true, // Charge
				ActualAmount:                          decimal.NewFromInt(5),
				Reason:                                "Packaging fee",
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
				CategoryTradeTaxType:                  "VAT",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Product A",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxTypeCode:              "VAT",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.RequireFromString("18.05"),
				BasisAmount:      decimal.NewFromInt(95),
				TypeCode:         "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
	}

	var buf bytes.Buffer
	err := inv.Write(&buf)
	if err != nil {
		t.Fatalf("Failed to write invoice: %v", err)
	}

	xmlOutput := buf.String()

	// Verify allowance and charge elements
	if !strings.Contains(xmlOutput, "<cac:AllowanceCharge>") {
		t.Error("Expected <cac:AllowanceCharge> elements")
	}

	// Verify allowance (ChargeIndicator = false)
	if !strings.Contains(xmlOutput, "<cbc:ChargeIndicator>false</cbc:ChargeIndicator>") {
		t.Error("Expected allowance with ChargeIndicator false")
	}

	// Verify charge (ChargeIndicator = true)
	if !strings.Contains(xmlOutput, "<cbc:ChargeIndicator>true</cbc:ChargeIndicator>") {
		t.Error("Expected charge with ChargeIndicator true")
	}

	// Verify reasons
	if !strings.Contains(xmlOutput, "Early payment discount") {
		t.Error("Expected allowance reason in XML")
	}

	if !strings.Contains(xmlOutput, "Packaging fee") {
		t.Error("Expected charge reason in XML")
	}
}
