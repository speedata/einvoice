package einvoice

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestWriter_NoEmptyElements tests that the writer does not create empty XML elements,
// which violates PEPPOL-EN16931-R008: "Document MUST not contain empty elements"
func TestWriter_NoEmptyElements(t *testing.T) {
	// Create an invoice with empty optional fields
	inv := &Invoice{
		SchemaType:                                 CII,
		GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017",
		InvoiceNumber:                              "INV-001",
		InvoiceTypeCode:                            380,
		InvoiceDate:                                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		InvoiceCurrencyCode:                        "EUR",
		Seller: Party{
			Name: "Seller Inc",
			PostalAddress: &PostalAddress{
				CountryID:    "DE",
				PostcodeCode: "", // Empty - should not create element
				City:         "Berlin",
			},
			VATaxRegistration: "DE123456789",
		},
		Buyer: Party{
			Name: "Buyer GmbH",
			PostalAddress: &PostalAddress{
				CountryID:    "FR",
				PostcodeCode: "", // Empty - should not create element
				City:         "Paris",
			},
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
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				Typ:              "VAT",
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromInt(10),
				BasisAmount:                           decimal.NewFromInt(110),
				Reason:                                "", // Empty - should not create element
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
				CategoryTradeTaxType:                  "VAT",
			},
		},
		PaymentMeans: []PaymentMeans{
			{
				TypeCode: 30,
				ApplicableTradeSettlementFinancialCardID:              "1234",
				ApplicableTradeSettlementFinancialCardCardholderName: "", // Empty - should not create element
			},
		},
	}

	// Write to buffer
	var buf bytes.Buffer
	err := inv.Write(&buf)
	if err != nil {
		t.Fatalf("Failed to write invoice: %v", err)
	}

	xmlOutput := buf.String()

	// Check for empty elements
	emptyElementPatterns := []string{
		"<ram:PostcodeCode></ram:PostcodeCode>",
		"<ram:PostcodeCode/>",
		"<ram:Reason></ram:Reason>",
		"<ram:Reason/>",
		"<ram:CardholderName></ram:CardholderName>",
		"<ram:CardholderName/>",
	}

	for _, pattern := range emptyElementPatterns {
		if strings.Contains(xmlOutput, pattern) {
			t.Errorf("Found empty element in XML output: %s (violates PEPPOL-EN16931-R008)", pattern)
		}
	}

	// Verify that non-empty elements ARE present
	requiredElements := []string{
		"<ram:CityName>Berlin</ram:CityName>",
		"<ram:CityName>Paris</ram:CityName>",
		"<ram:ID>1234</ram:ID>", // Card ID
	}

	for _, required := range requiredElements {
		if !strings.Contains(xmlOutput, required) {
			t.Errorf("Expected element not found in XML: %s", required)
		}
	}
}

// TestWriter_PostcodeOptional tests that postcode elements are only created when non-empty
func TestWriter_PostcodeOptional(t *testing.T) {
	tests := []struct {
		name         string
		postcodeCode string
		shouldExist  bool
	}{
		{
			name:         "With postcode",
			postcodeCode: "10115",
			shouldExist:  true,
		},
		{
			name:         "Empty postcode",
			postcodeCode: "",
			shouldExist:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invoice{
				SchemaType:                                 CII,
				GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017",
				InvoiceNumber:                              "INV-001",
				InvoiceTypeCode:                            380,
				InvoiceDate:                                time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				InvoiceCurrencyCode:                        "EUR",
				Seller: Party{
					Name: "Seller Inc",
					PostalAddress: &PostalAddress{
						CountryID:    "DE",
						PostcodeCode: tt.postcodeCode,
						City:         "Berlin",
					},
					VATaxRegistration: "DE123456789",
				},
				Buyer: Party{
					Name: "Buyer GmbH",
					PostalAddress: &PostalAddress{
						CountryID: "FR",
						City:      "Paris",
					},
				},
				LineTotal:        decimal.NewFromInt(100),
				TaxBasisTotal:    decimal.NewFromInt(100),
				TaxTotal:         decimal.NewFromInt(19),
				GrandTotal:       decimal.NewFromInt(119),
				DuePayableAmount: decimal.NewFromInt(119),
				InvoiceLines: []InvoiceLine{
					{
						LineID:                   "1",
						ItemName:                 "Product",
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
						CalculatedAmount: decimal.NewFromInt(19),
						BasisAmount:      decimal.NewFromInt(100),
						Typ:              "VAT",
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
			containsPostcode := strings.Contains(xmlOutput, "<ram:PostcodeCode>")

			if tt.shouldExist && !containsPostcode {
				t.Errorf("Expected <ram:PostcodeCode> element, but not found")
			}
			if !tt.shouldExist && containsPostcode {
				t.Errorf("Did not expect <ram:PostcodeCode> element, but found one (violates PEPPOL-EN16931-R008)")
			}
		})
	}
}

// TestWriter_AllowanceReasonOptional tests that reason elements are only created when non-empty
func TestWriter_AllowanceReasonOptional(t *testing.T) {
	inv := &Invoice{
		SchemaType:                                 CII,
		GuidelineSpecifiedDocumentContextParameter: "urn:cen.eu:en16931:2017",
		InvoiceNumber:                              "INV-001",
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
		},
		LineTotal:        decimal.NewFromInt(90),
		TaxBasisTotal:    decimal.NewFromInt(90),
		TaxTotal:         decimal.RequireFromString("17.10"),
		GrandTotal:       decimal.RequireFromString("107.10"),
		DuePayableAmount: decimal.RequireFromString("107.10"),
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Product",
				BilledQuantity:           decimal.NewFromInt(1),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				Total:                    decimal.NewFromInt(100),
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromInt(19),
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator:                       false,
				ActualAmount:                          decimal.NewFromInt(10),
				BasisAmount:                           decimal.Zero,
				Reason:                                "", // Empty - should NOT create element
				CategoryTradeTaxCategoryCode:          "S",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromInt(19),
				CategoryTradeTaxType:                  "VAT",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CalculatedAmount: decimal.RequireFromString("17.10"),
				BasisAmount:      decimal.NewFromInt(90),
				Typ:              "VAT",
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

	// Should NOT contain empty Reason element
	if strings.Contains(xmlOutput, "<ram:Reason></ram:Reason>") || strings.Contains(xmlOutput, "<ram:Reason/>") {
		t.Error("Found empty <ram:Reason> element (violates PEPPOL-EN16931-R008)")
	}

	// Should contain the allowance
	if !strings.Contains(xmlOutput, "<ram:SpecifiedTradeAllowanceCharge>") {
		t.Error("Expected <ram:SpecifiedTradeAllowanceCharge> element, but not found")
	}
}
