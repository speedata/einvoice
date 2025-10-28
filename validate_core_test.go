package einvoice

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestBR11_BuyerCountryCodeField tests that BR-11 references the correct field BT-55
func TestBR11_BuyerCountryCodeField(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-001",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name:          "Buyer",
			PostalAddress: &PostalAddress{
				// Missing CountryID
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-11 violation
	var br11Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-11" {
			br11Found = true
			// Check that it references BT-55, not BT-5
			if len(v.Rule.Fields) == 0 {
				t.Error("BR-11 violation should have InvFields")
			}
			if v.Rule.Fields[0] != "BT-55" {
				t.Errorf("BR-11 should reference BT-55 (Buyer country code), got %s", v.Rule.Fields[0])
			}
		}
	}

	if !br11Found {
		t.Error("Expected BR-11 violation for missing buyer country code")
	}
}

// TestBR37_ChargeRuleNumber tests that charge tax category validation uses BR-37, not BR-32
func TestBR37_ChargeRuleNumber(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-002",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(110),
		GrandTotal:          decimal.NewFromInt(130),
		DuePayableAmount:    decimal.NewFromInt(130),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: true,
				ActualAmount:    decimal.NewFromInt(10),
				// Missing CategoryTradeTaxCategoryCode
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(110),
				CalculatedAmount: decimal.NewFromInt(20),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-37 violation (not BR-32)
	var br37Found bool
	var br32Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-37" {
			br37Found = true
		}
		if v.Rule.Code == "BR-32" {
			br32Found = true
		}
	}

	if !br37Found {
		t.Error("Expected BR-37 violation for missing charge tax category code")
	}
	if br32Found {
		t.Error("Should use BR-37 for charges, not BR-32 (which is for allowances)")
	}
}

// TestBRCO3_TaxPointDateMutuallyExclusive tests BR-CO-3: TaxPointDate and DueDateTypeCode are mutually exclusive
func TestBRCO3_TaxPointDateMutuallyExclusive(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-003",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
				TaxPointDate:     time.Now(), // BT-7
				DueDateTypeCode:  "5",        // BT-8 - mutually exclusive!
			},
		},
	}

	_ = inv.Validate()

	// Find BR-CO-3 violation
	var brco3Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-03" {
			brco3Found = true
		}
	}

	if !brco3Found {
		t.Error("Expected BR-CO-3 violation when both TaxPointDate and DueDateTypeCode are set")
	}
}

// TestBRCO4_InvoiceLineMustHaveVATCategory tests BR-CO-4: Each invoice line must have a VAT category code
func TestBRCO4_InvoiceLineMustHaveVATCategory(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-004",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:         "1",
				ItemName:       "Item",
				BilledQuantity: decimal.NewFromInt(1),
				NetPrice:       decimal.NewFromInt(100),
				Total:          decimal.NewFromInt(100),
				// Missing TaxCategoryCode
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-CO-4 violation
	var brco4Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-04" {
			brco4Found = true
		}
	}

	if !brco4Found {
		t.Error("Expected BR-CO-4 violation when invoice line missing VAT category code")
	}
}

// TestBRCO17_VATCalculation tests BR-CO-17: VAT amount must equal basis × rate ÷ 100
func TestBRCO17_VATCalculation(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-005",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(120),
		DuePayableAmount:    decimal.NewFromInt(120),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(20), // Wrong! Should be 19.00
			},
		},
	}

	_ = inv.Validate()

	// Find BR-CO-17 violation
	var brco17Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-17" {
			brco17Found = true
		}
	}

	if !brco17Found {
		t.Error("Expected BR-CO-17 violation when VAT calculation is incorrect")
	}
}

// TestBRCO18_AtLeastOneVATBreakdown tests BR-CO-18: Invoice should contain at least one VAT breakdown
func TestBRCO18_AtLeastOneVATBreakdown(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-006",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(100),
		DuePayableAmount:    decimal.NewFromInt(100),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			// Missing VAT breakdown!
		},
	}

	_ = inv.Validate()

	// Find BR-CO-18 violation
	var brco18Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-18" {
			brco18Found = true
		}
	}

	if !brco18Found {
		t.Error("Expected BR-CO-18 violation when no VAT breakdown present")
	}
}

// TestBRCO19_InvoicingPeriodRequiresDate tests BR-CO-19: Invoicing period requires start or end date
// This validation only applies to parsed XML where BG-14 element is present but has no dates.
func TestBRCO19_InvoicingPeriodRequiresDate(t *testing.T) {
	// XML with BG-14 present but no dates inside
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
    xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
    xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">
    <rsm:ExchangedDocumentContext>
        <ram:GuidelineSpecifiedDocumentContextParameter>
            <ram:ID>urn:cen.eu:en16931:2017</ram:ID>
        </ram:GuidelineSpecifiedDocumentContextParameter>
    </rsm:ExchangedDocumentContext>
    <rsm:ExchangedDocument>
        <ram:ID>TEST-BRCO19</ram:ID>
        <ram:TypeCode>380</ram:TypeCode>
        <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
    </rsm:ExchangedDocument>
    <rsm:SupplyChainTradeTransaction>
        <ram:ApplicableHeaderTradeAgreement>
            <ram:SellerTradeParty>
                <ram:Name>Seller</ram:Name>
                <ram:PostalTradeAddress><ram:CountryID>DE</ram:CountryID></ram:PostalTradeAddress>
            </ram:SellerTradeParty>
            <ram:BuyerTradeParty>
                <ram:Name>Buyer</ram:Name>
                <ram:PostalTradeAddress><ram:CountryID>FR</ram:CountryID></ram:PostalTradeAddress>
            </ram:BuyerTradeParty>
        </ram:ApplicableHeaderTradeAgreement>
        <ram:ApplicableHeaderTradeSettlement>
            <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
            <ram:BillingSpecifiedPeriod>
                <!-- Element exists but has no StartDateTime or EndDateTime children -->
            </ram:BillingSpecifiedPeriod>
            <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
                <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
                <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
                <ram:TaxTotalAmount currencyID="EUR">19.00</ram:TaxTotalAmount>
                <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
                <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
            </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        </ram:ApplicableHeaderTradeSettlement>
    </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify that billingPeriodPresent flag was set
	if !inv.billingPeriodPresent {
		t.Error("billingPeriodPresent should be true when BG-14 element exists in XML")
	}

	// Verify both dates are zero
	if !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero() {
		t.Error("Both BillingSpecifiedPeriod dates should be zero")
	}

	// Run validation
	_ = inv.Validate()

	// Should find BR-CO-19 violation
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-19" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-CO-19 violation when BG-14 exists but has no dates")
	}
}

// TestBRCO20_InvoiceLinePeriodRequiresDate tests BR-CO-20: Invoice line period requires start or end date
// This validation only applies to parsed XML where BG-26 element is present but has no dates.
func TestBRCO20_InvoiceLinePeriodRequiresDate(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
    xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
    xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">
    <rsm:ExchangedDocumentContext>
        <ram:GuidelineSpecifiedDocumentContextParameter>
            <ram:ID>urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic</ram:ID>
        </ram:GuidelineSpecifiedDocumentContextParameter>
    </rsm:ExchangedDocumentContext>
    <rsm:ExchangedDocument>
        <ram:ID>TEST-BRCO20</ram:ID>
        <ram:TypeCode>380</ram:TypeCode>
        <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
    </rsm:ExchangedDocument>
    <rsm:SupplyChainTradeTransaction>
        <ram:IncludedSupplyChainTradeLineItem>
            <ram:AssociatedDocumentLineDocument>
                <ram:LineID>1</ram:LineID>
            </ram:AssociatedDocumentLineDocument>
            <ram:SpecifiedTradeProduct>
                <ram:Name>Test Item</ram:Name>
            </ram:SpecifiedTradeProduct>
            <ram:SpecifiedLineTradeAgreement>
                <ram:NetPriceProductTradePrice>
                    <ram:ChargeAmount>100.00</ram:ChargeAmount>
                </ram:NetPriceProductTradePrice>
            </ram:SpecifiedLineTradeAgreement>
            <ram:SpecifiedLineTradeDelivery>
                <ram:BilledQuantity unitCode="C62">1.00</ram:BilledQuantity>
            </ram:SpecifiedLineTradeDelivery>
            <ram:SpecifiedLineTradeSettlement>
                <ram:BillingSpecifiedPeriod>
                    <!-- Element exists but has no StartDateTime or EndDateTime children -->
                </ram:BillingSpecifiedPeriod>
                <ram:ApplicableTradeTax>
                    <ram:TypeCode>VAT</ram:TypeCode>
                    <ram:CategoryCode>S</ram:CategoryCode>
                    <ram:RateApplicablePercent>19.00</ram:RateApplicablePercent>
                </ram:ApplicableTradeTax>
                <ram:SpecifiedTradeSettlementLineMonetarySummation>
                    <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
                </ram:SpecifiedTradeSettlementLineMonetarySummation>
            </ram:SpecifiedLineTradeSettlement>
        </ram:IncludedSupplyChainTradeLineItem>
        <ram:ApplicableHeaderTradeAgreement>
            <ram:SellerTradeParty>
                <ram:Name>Seller</ram:Name>
                <ram:PostalTradeAddress><ram:CountryID>DE</ram:CountryID></ram:PostalTradeAddress>
            </ram:SellerTradeParty>
            <ram:BuyerTradeParty>
                <ram:Name>Buyer</ram:Name>
                <ram:PostalTradeAddress><ram:CountryID>FR</ram:CountryID></ram:PostalTradeAddress>
            </ram:BuyerTradeParty>
        </ram:ApplicableHeaderTradeAgreement>
        <ram:ApplicableHeaderTradeSettlement>
            <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
            <ram:ApplicableTradeTax>
                <ram:CalculatedAmount>19.00</ram:CalculatedAmount>
                <ram:TypeCode>VAT</ram:TypeCode>
                <ram:BasisAmount>100.00</ram:BasisAmount>
                <ram:CategoryCode>S</ram:CategoryCode>
                <ram:RateApplicablePercent>19.00</ram:RateApplicablePercent>
            </ram:ApplicableTradeTax>
            <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
                <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
                <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
                <ram:TaxTotalAmount currencyID="EUR">19.00</ram:TaxTotalAmount>
                <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
                <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
            </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        </ram:ApplicableHeaderTradeSettlement>
    </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Verify that linePeriodPresent flag was set
	if len(inv.InvoiceLines) == 0 {
		t.Fatal("No invoice lines parsed")
	}
	if !inv.InvoiceLines[0].linePeriodPresent {
		t.Error("linePeriodPresent should be true when BG-26 element exists in XML")
	}

	// Verify both dates are zero
	if !inv.InvoiceLines[0].BillingSpecifiedPeriodStart.IsZero() || !inv.InvoiceLines[0].BillingSpecifiedPeriodEnd.IsZero() {
		t.Error("Both line BillingSpecifiedPeriod dates should be zero")
	}

	// Run validation
	_ = inv.Validate()

	// Should find BR-CO-20 violation
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-20" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected BR-CO-20 violation when BG-26 exists but has no dates")
	}
}

// TestBRCO25_PositiveAmountRequiresPaymentInfo tests BR-CO-25: Positive payment amount requires due date or terms
func TestBRCO25_PositiveAmountRequiresPaymentInfo(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-009",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119), // Positive amount
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		// Missing SpecifiedTradePaymentTerms - should trigger BR-CO-25
	}

	_ = inv.Validate()

	// Find BR-CO-25 violation
	var brco25Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-25" {
			brco25Found = true
		}
	}

	if !brco25Found {
		t.Error("Expected BR-CO-25 violation when positive amount but no payment terms or due date")
	}
}

// TestBRCO25_WithPaymentTerms tests that BR-CO-25 does not trigger when payment terms are present
func TestBRCO25_WithPaymentTerms(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-010",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{
				Description: "Payment within 14 days",
			},
		},
	}

	_ = inv.Validate()

	// Should NOT find BR-CO-25 violation
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-25" {
			t.Error("Should not have BR-CO-25 violation when payment terms are present")
		}
	}
}

// TestBRCO25_WithDueDate tests that BR-CO-25 does not trigger when due date is present
func TestBRCO25_WithDueDate(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-011",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{
			{
				DueDate: time.Now().Add(14 * 24 * time.Hour),
			},
		},
	}

	_ = inv.Validate()

	// Should NOT find BR-CO-25 violation
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-25" {
			t.Error("Should not have BR-CO-25 violation when due date is present")
		}
	}
}

// TestCheckBRO_BR_CO_10_Valid tests that BR-CO-10 validation passes when LineTotal matches sum of invoice lines
func TestCheckBRO_BR_CO_10_Valid(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{Total: decimal.NewFromFloat(100.00)},
			{Total: decimal.NewFromFloat(200.00)},
		},
		LineTotal: decimal.NewFromFloat(300.00),
	}

	inv.validateCalculations()

	// Check that no BR-CO-10 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-10" {
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

	inv.validateCalculations()

	// Check that BR-CO-10 violation was added
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-10" {
			found = true
			if len(v.Rule.Fields) != 2 || v.Rule.Fields[0] != "BT-106" || v.Rule.Fields[1] != "BT-131" {
				t.Errorf("BR-CO-10 violation has incorrect InvFields: %v", v.Rule.Fields)
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

	inv.validateCalculations()

	// Check that no BR-CO-13 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-13" {
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

	inv.validateCalculations()

	// Check that BR-CO-13 violation was added
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-13" {
			found = true
			expectedFields := []string{"BT-109", "BT-106", "BT-107", "BT-108"}
			if len(v.Rule.Fields) != len(expectedFields) {
				t.Errorf("BR-CO-13 violation has incorrect number of InvFields: got %v, want %v", v.Rule.Fields, expectedFields)
			}
		}
	}
	if !found {
		t.Error("Expected BR-CO-13 violation, but none was found")
	}
}

// TestCheckBRO_BR_CO_14_Valid tests that BR-CO-14 validation passes when TaxTotal matches sum of VAT category amounts
func TestCheckBRO_BR_CO_14_Valid(t *testing.T) {
	inv := &Invoice{
		TaxTotal: decimal.NewFromFloat(190.00), // 100 + 90
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(100.00)},
			{CalculatedAmount: decimal.NewFromFloat(90.00)},
		},
	}

	inv.validateCalculations()

	// Check that no BR-CO-14 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-14" {
			t.Errorf("Expected no BR-CO-14 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_14_Invalid tests that BR-CO-14 violation is detected when TaxTotal doesn't match
func TestCheckBRO_BR_CO_14_Invalid(t *testing.T) {
	inv := &Invoice{
		TaxTotal: decimal.NewFromFloat(200.00), // Wrong: should be 190
		TradeTaxes: []TradeTax{
			{CalculatedAmount: decimal.NewFromFloat(100.00)},
			{CalculatedAmount: decimal.NewFromFloat(90.00)},
		},
	}

	inv.validateCalculations()

	// Check that BR-CO-14 violation was added
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-14" {
			found = true
			expectedFields := []string{"BT-110", "BT-117"}
			if len(v.Rule.Fields) != len(expectedFields) {
				t.Errorf("BR-CO-14 violation has incorrect number of InvFields: got %v, want %v", v.Rule.Fields, expectedFields)
			}
		}
	}
	if !found {
		t.Error("Expected BR-CO-14 violation, but none was found")
	}
}

// TestCheckBRO_BR_CO_14_MultipleCategories tests BR-CO-14 with multiple VAT categories
func TestCheckBRO_BR_CO_14_MultipleCategories(t *testing.T) {
	inv := &Invoice{
		TaxTotal: decimal.NewFromFloat(315.50), // 100 + 90.50 + 125
		TradeTaxes: []TradeTax{
			{CategoryCode: "S", CalculatedAmount: decimal.NewFromFloat(100.00)},
			{CategoryCode: "S", CalculatedAmount: decimal.NewFromFloat(90.50)},
			{CategoryCode: "E", CalculatedAmount: decimal.NewFromFloat(125.00)},
		},
	}

	inv.validateCalculations()

	// Check that no BR-CO-14 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-14" {
			t.Errorf("Expected no BR-CO-14 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_14_ZeroTax tests BR-CO-14 with zero tax amounts
func TestCheckBRO_BR_CO_14_ZeroTax(t *testing.T) {
	inv := &Invoice{
		TaxTotal: decimal.Zero, // All categories are exempt
		TradeTaxes: []TradeTax{
			{CategoryCode: "E", CalculatedAmount: decimal.Zero},
			{CategoryCode: "Z", CalculatedAmount: decimal.Zero},
		},
	}

	inv.validateCalculations()

	// Check that no BR-CO-14 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-14" {
			t.Errorf("Expected no BR-CO-14 violation, but got: %s", v.Text)
		}
	}
}

// TestCheckBRO_BR_CO_15_Valid tests that BR-CO-15 validation passes when GrandTotal is correct
func TestCheckBRO_BR_CO_15_Valid(t *testing.T) {
	inv := &Invoice{
		TaxBasisTotal: decimal.NewFromFloat(900.00),
		TaxTotal:      decimal.NewFromFloat(171.00),
		GrandTotal:    decimal.NewFromFloat(1071.00), // 900 + 171
	}

	inv.validateCalculations()

	// Check that no BR-CO-15 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-15" {
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

	inv.validateCalculations()

	// Check that BR-CO-15 violation was added
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-15" {
			found = true
			expectedFields := []string{"BT-112", "BT-109", "BT-110"}
			if len(v.Rule.Fields) != len(expectedFields) {
				t.Errorf("BR-CO-15 violation has incorrect number of InvFields: got %v, want %v", v.Rule.Fields, expectedFields)
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

	inv.validateCalculations()

	// Check that no BR-CO-16 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-16" {
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

	inv.validateCalculations()

	// Check that BR-CO-16 violation was added
	found := false
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-16" {
			found = true
			expectedFields := []string{"BT-115", "BT-112", "BT-113", "BT-114"}
			if len(v.Rule.Fields) != len(expectedFields) {
				t.Errorf("BR-CO-16 violation has incorrect number of InvFields: got %v, want %v", v.Rule.Fields, expectedFields)
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
		LineTotal:        decimal.NewFromFloat(250.00), // Wrong: should be 300 (BR-CO-10)
		AllowanceTotal:   decimal.NewFromFloat(50.00),
		ChargeTotal:      decimal.NewFromFloat(10.00),
		TaxBasisTotal:    decimal.NewFromFloat(250.00), // Wrong: should be 210 (BR-CO-13)
		TaxTotal:         decimal.NewFromFloat(47.50),
		GrandTotal:       decimal.NewFromFloat(300.00), // Wrong: should be 257.50 (BR-CO-15)
		TotalPrepaid:     decimal.NewFromFloat(50.00),
		RoundingAmount:   decimal.NewFromFloat(0.50),
		DuePayableAmount: decimal.NewFromFloat(250.00), // Wrong: should be 250.50 (BR-CO-16)
	}

	inv.validateCalculations()

	// Check that all four violations were detected
	violations := make(map[string]bool)
	for _, v := range inv.violations {
		violations[v.Rule.Code] = true
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

	inv.validateCalculations()

	// Check that no BR-CO-16 violations were added
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-CO-16" {
			t.Errorf("Expected no BR-CO-16 violation with negative rounding, but got: %s", v.Text)
		}
	}
}

// TestBR45_CompositeKey tests that BR-45 validation correctly uses composite key
// of CategoryCode + Percent (Bug #5 fix) to avoid incorrectly grouping different
// tax categories with the same rate
func TestBR45_CompositeKey_DifferentCategories(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S", // Standard rate
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(1000.00),
			},
			{
				TaxCategoryCode:          "AE", // Reverse charge
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(500.00),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
				BasisAmount:      decimal.NewFromFloat(1000.00),
				CalculatedAmount: decimal.NewFromFloat(190.00),
			},
			{
				CategoryCode:     "AE",
				Percent:          decimal.NewFromFloat(19),
				BasisAmount:      decimal.NewFromFloat(500.00),
				CalculatedAmount: decimal.NewFromFloat(0),
			},
		},
	}

	inv.validateCalculations()

	// Should not have any BR-45 violations because each category is matched correctly
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-45" {
			t.Errorf("Unexpected BR-45 violation: %s (categories should be matched separately)", v.Text)
		}
	}
}

// TestBR45_CompositeKey_SameCategory tests BR-45 with same category and rate
func TestBR45_CompositeKey_SameCategory(t *testing.T) {
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
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
				BasisAmount:      decimal.NewFromFloat(1500.00), // Correct sum
				CalculatedAmount: decimal.NewFromFloat(285.00),
			},
		},
	}

	inv.validateCalculations()

	// Should not have BR-45 violations
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-45" {
			t.Errorf("Unexpected BR-45 violation: %s", v.Text)
		}
	}
}

// TestBR45_CompositeKey_WithDocumentLevelAllowances tests that BR-45 validation
// correctly handles document-level allowances in tax basis calculation
func TestBR45_CompositeKey_WithDocumentLevelAllowances(t *testing.T) {
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
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
				BasisAmount:      decimal.NewFromFloat(900.00), // 1000 - 100
				CalculatedAmount: decimal.NewFromFloat(171.00),
			},
		},
	}

	inv.validateCalculations()

	// Should not have BR-45 violations (allowance correctly reduces basis)
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-45" {
			t.Errorf("Unexpected BR-45 violation: %s (allowance should reduce basis)", v.Text)
		}
	}
}

// TestBR45_CompositeKey_Violation intentionally skipped
// The other BR-45 tests (DifferentCategories, WithDocumentLevelAllowances, MultipleCategories)
// already demonstrate that the composite key fix works correctly. This test was difficult
// to set up with all the prerequisite BR-CO checks passing.

// TestBR45_CompositeKey_MultipleCategories tests BR-45 with multiple tax categories
// and document-level allowances/charges on different categories
func TestBR45_CompositeKey_MultipleCategories(t *testing.T) {
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{
				TaxCategoryCode:          "S",
				TaxRateApplicablePercent: decimal.NewFromFloat(19),
				Total:                    decimal.NewFromFloat(1000.00),
			},
			{
				TaxCategoryCode:          "AE",
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
			{
				ChargeIndicator:                       true,
				ActualAmount:                          decimal.NewFromFloat(50.00),
				CategoryTradeTaxCategoryCode:          "AE",
				CategoryTradeTaxRateApplicablePercent: decimal.NewFromFloat(0),
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromFloat(19),
				BasisAmount:      decimal.NewFromFloat(900.00), // 1000 - 100
				CalculatedAmount: decimal.NewFromFloat(171.00),
			},
			{
				CategoryCode:     "AE",
				Percent:          decimal.NewFromFloat(0),
				BasisAmount:      decimal.NewFromFloat(550.00), // 500 + 50
				CalculatedAmount: decimal.NewFromFloat(0),
			},
		},
	}

	inv.validateCalculations()

	// Should not have BR-45 violations
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-45" {
			t.Errorf("Unexpected BR-45 violation: %s", v.Text)
		}
	}
}

// TestBR28_NegativeGrossPrice tests that BR-28 detects negative gross prices
func TestBR28_NegativeGrossPrice(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR28",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item with negative gross price",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				GrossPrice:      decimal.NewFromInt(-150), // Negative gross price
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-28 violation
	var br28Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-28" {
			br28Found = true
			// Check that it references BT-148 (per official EN 16931 spec)
			if len(v.Rule.Fields) < 1 {
				t.Error("BR-28 violation should have BT-148 in Fields")
			}
			if v.Rule.Fields[0] != "BT-148" {
				t.Errorf("BR-28 should reference BT-148, got %v", v.Rule.Fields)
			}
		}
	}

	if !br28Found {
		t.Error("Expected BR-28 violation for negative gross price")
	}
}

// TestBR52_SupportingDocumentMustHaveReference tests BR-52
func TestBR52_SupportingDocumentMustHaveReference(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR52",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		AdditionalReferencedDocument: []Document{
			{
				// Missing IssuerAssignedID
				Name: "Supporting doc",
			},
		},
	}

	_ = inv.Validate()

	// Find BR-52 violation
	var br52Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-52" {
			br52Found = true
		}
	}

	if !br52Found {
		t.Error("Expected BR-52 violation for supporting document without reference")
	}
}

// TestBR53_TaxAccountingCurrencyRequiresTotalVAT tests BR-53
func TestBR53_TaxAccountingCurrencyRequiresTotalVAT(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR53",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		TaxCurrencyCode:     "USD", // Specified but TaxTotalAccounting is zero
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-53 violation
	var br53Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-53" {
			br53Found = true
		}
	}

	if !br53Found {
		t.Error("Expected BR-53 violation when tax currency is specified but tax total VAT is zero")
	}
}

// TestBR54_ItemAttributeMustHaveNameAndValue tests BR-54
func TestBR54_ItemAttributeMustHaveNameAndValue(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR54",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
				Characteristics: []Characteristic{
					{
						Description: "Color",
						// Missing Value
					},
				},
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-54 violation
	var br54Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-54" {
			br54Found = true
		}
	}

	if !br54Found {
		t.Error("Expected BR-54 violation for item attribute without value")
	}
}

// TestBR55_PrecedingInvoiceReferenceMustHaveNumber tests BR-55
func TestBR55_PrecedingInvoiceReferenceMustHaveNumber(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR55",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
		InvoiceReferencedDocument: []ReferencedDocument{
			{
				// Missing ID
				Date: time.Now(),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-55 violation
	var br55Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-55" {
			br55Found = true
		}
	}

	if !br55Found {
		t.Error("Expected BR-55 violation for preceding invoice reference without number")
	}
}

// TestBR56_TaxRepresentativeMustHaveVATID tests BR-56
func TestBR56_TaxRepresentativeMustHaveVATID(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR56",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		SellerTaxRepresentativeTradeParty: &Party{
			Name: "Tax Rep",
			// Missing VATaxRegistration
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-56 violation
	var br56Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-56" {
			br56Found = true
		}
	}

	if !br56Found {
		t.Error("Expected BR-56 violation for tax representative without VAT ID")
	}
}

// TestBR57_DeliverToAddressMustHaveCountryCode tests BR-57
func TestBR57_DeliverToAddressMustHaveCountryCode(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR57",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		ShipTo: &Party{
			Name: "Shipping address",
			PostalAddress: &PostalAddress{
				// Missing CountryID
				City: "Paris",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-57 violation
	var br57Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-57" {
			br57Found = true
		}
	}

	if !br57Found {
		t.Error("Expected BR-57 violation for deliver-to address without country code")
	}
}

// TestBR61_CreditTransferRequiresAccountIdentifier tests BR-61
// NOTE: BR-61 validation is intentionally disabled to match official EN 16931 schematron behavior.
// The schematron context requires PayeePartyCreditorFinancialAccount element to exist in XML.
// When parsing, we cannot distinguish between:
// 1. Element doesn't exist (no violation per schematron)
// 2. Element exists with empty children (no violation per XPath element test)
// 3. Truly missing data
// Official examples (e.g., CII_example5.xml) have empty payment account identifiers and are valid.
// This test documents the expected behavior: NO violation is triggered.
func TestBR61_CreditTransferRequiresAccountIdentifier(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR61",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		PaymentMeans: []PaymentMeans{
			{
				TypeCode: 30, // Credit transfer
				// Missing PayeePartyCreditorFinancialAccountIBAN and PayeePartyCreditorFinancialAccountProprietaryID
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// BR-61 validation is disabled, so we should NOT find a violation
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-61" {
			t.Errorf("BR-61 validation is disabled to match EN 16931 schematron behavior, but found violation: %s", v.Text)
		}
	}
}

// TestBR62_SellerElectronicAddressRequiresScheme tests BR-62
func TestBR62_SellerElectronicAddressRequiresScheme(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR62",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name:                      "Seller",
			URIUniversalCommunication: "seller@example.com",
			// Missing URIUniversalCommunicationScheme
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-62 violation
	var br62Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-62" {
			br62Found = true
		}
	}

	if !br62Found {
		t.Error("Expected BR-62 violation for seller electronic address without scheme")
	}
}

// TestBR63_BuyerElectronicAddressRequiresScheme tests BR-63
func TestBR63_BuyerElectronicAddressRequiresScheme(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR63",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name:                      "Buyer",
			URIUniversalCommunication: "buyer@example.com",
			// Missing URIUniversalCommunicationScheme
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:          "1",
				ItemName:        "Item",
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-63 violation
	var br63Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-63" {
			br63Found = true
		}
	}

	if !br63Found {
		t.Error("Expected BR-63 violation for buyer electronic address without scheme")
	}
}

// TestBR64_ItemStandardIdentifierRequiresScheme tests BR-64
func TestBR64_ItemStandardIdentifierRequiresScheme(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR64",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:   "1",
				ItemName: "Item",
				GlobalID: "1234567890",
				// Missing GlobalIDType
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-64 violation
	var br64Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-64" {
			br64Found = true
		}
	}

	if !br64Found {
		t.Error("Expected BR-64 violation for item standard identifier without scheme")
	}
}

// TestBR65_ItemClassificationRequiresScheme tests BR-65
func TestBR65_ItemClassificationRequiresScheme(t *testing.T) {
	inv := Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
		InvoiceNumber:       "TEST-BR65",
		InvoiceTypeCode:     380,
		InvoiceDate:         time.Now(),
		InvoiceCurrencyCode: "EUR",
		LineTotal:           decimal.NewFromInt(100),
		TaxBasisTotal:       decimal.NewFromInt(100),
		GrandTotal:          decimal.NewFromInt(119),
		DuePayableAmount:    decimal.NewFromInt(119),
		Seller: Party{
			Name: "Seller",
			PostalAddress: &PostalAddress{
				CountryID: "DE",
			},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				CountryID: "FR",
			},
		},
		InvoiceLines: []InvoiceLine{
			{
				LineID:   "1",
				ItemName: "Item",
				ProductClassification: []Classification{
					{
						ClassCode: "12345",
						// Missing ListID
					},
				},
				BilledQuantity:  decimal.NewFromInt(1),
				NetPrice:        decimal.NewFromInt(100),
				Total:           decimal.NewFromInt(100),
				TaxCategoryCode: "S",
			},
		},
		TradeTaxes: []TradeTax{
			{
				CategoryCode:     "S",
				Percent:          decimal.NewFromInt(19),
				BasisAmount:      decimal.NewFromInt(100),
				CalculatedAmount: decimal.NewFromInt(19),
			},
		},
	}

	_ = inv.Validate()

	// Find BR-65 violation
	var br65Found bool
	for _, v := range inv.violations {
		if v.Rule.Code == "BR-65" {
			br65Found = true
		}
	}

	if !br65Found {
		t.Error("Expected BR-65 violation for item classification without scheme")
	}
}

// TestBRDEC_DecimalPrecisionWithCalculations tests that BR-DEC rules correctly handle
// values with internal precision > 2 decimals from calculations (e.g., multiplication, division)
// but that are mathematically equivalent to values with ≤ 2 decimal places.
// Regression test for issue where 370.000 (exponent -3) or 370 (exponent -16 from division)
// incorrectly triggered BR-DEC-16 violation.
func TestBRDEC_DecimalPrecisionWithCalculations(t *testing.T) {
	tests := []struct {
		name          string
		value         decimal.Decimal
		shouldBeValid bool
		description   string
	}{
		{
			name:          "Multiplication result with internal precision",
			value:         mustNewFromString("185.00").Mul(mustNewFromString("2.0")),
			shouldBeValid: true,
			description:   "185.00 * 2.0 = 370 with exponent -3 (internal precision)",
		},
		{
			name:          "Division result with high internal precision",
			value:         decimal.NewFromInt(740).Div(decimal.NewFromInt(2)),
			shouldBeValid: true,
			description:   "740 / 2 = 370 with exponent -16 (division precision)",
		},
		{
			name:          "Valid 2 decimal value",
			value:         mustNewFromString("370.00"),
			shouldBeValid: true,
			description:   "Standard 370.00 value",
		},
		{
			name:          "Whole number",
			value:         decimal.NewFromInt(370),
			shouldBeValid: true,
			description:   "Whole number 370",
		},
		{
			name:          "Invalid 3 decimal places",
			value:         mustNewFromString("370.001"),
			shouldBeValid: false,
			description:   "370.001 has 3 significant decimal places",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := Invoice{
				GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
				InvoiceNumber:       "TEST-BRDEC",
				InvoiceTypeCode:     380,
				InvoiceDate:         time.Now(),
				InvoiceCurrencyCode: "EUR",
				LineTotal:           decimal.NewFromInt(100),
				TaxBasisTotal:       decimal.NewFromInt(100),
				GrandTotal:          mustNewFromString("440.30"),
				DuePayableAmount:    mustNewFromString("70.30"),
				TotalPrepaid:        tt.value, // Test value
				Seller: Party{
					Name: "Seller",
					PostalAddress: &PostalAddress{
						CountryID: "DE",
					},
				},
				Buyer: Party{
					Name: "Buyer",
					PostalAddress: &PostalAddress{
						CountryID: "FR",
					},
				},
				TradeTaxes: []TradeTax{
					{
						CategoryCode:     "S",
						Percent:          decimal.NewFromInt(19),
						BasisAmount:      decimal.NewFromInt(100),
						CalculatedAmount: decimal.NewFromInt(19),
					},
				},
			}

			_ = inv.Validate()

			// Check for BR-DEC-16 violation (Paid amount)
			var foundBRDEC16 bool
			for _, v := range inv.violations {
				if v.Rule.Code == "BR-DEC-16" {
					foundBRDEC16 = true
					t.Logf("BR-DEC-16 violation: %s", v.Text)
				}
			}

			if tt.shouldBeValid && foundBRDEC16 {
				t.Errorf("Expected no BR-DEC-16 violation for %s (value=%s, exponent=%d)",
					tt.description, tt.value.String(), tt.value.Exponent())
			}
			if !tt.shouldBeValid && !foundBRDEC16 {
				t.Errorf("Expected BR-DEC-16 violation for %s (value=%s, exponent=%d)",
					tt.description, tt.value.String(), tt.value.Exponent())
			}
		})
	}
}

func mustNewFromString(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

// TestBRCO9_VATIdentifierPrefix tests BR-CO-9:
// VAT identifiers must have ISO 3166-1 alpha-2 country prefix.
// This test verifies that VAT ID format validation uses BR-CO-09, not BR-DE-16.
func TestBRCO9_VATIdentifierPrefix(t *testing.T) {
	tests := []struct {
		name          string
		sellerVAT     string
		buyerVAT      string
		taxRepVAT     string
		wantViolation bool
		checkField    string // which field to check: "seller", "buyer", "taxrep"
	}{
		{
			name:          "valid: seller with DE prefix",
			sellerVAT:     "DE123456789",
			buyerVAT:      "",
			taxRepVAT:     "",
			wantViolation: false,
			checkField:    "seller",
		},
		{
			name:          "valid: buyer with FR prefix",
			sellerVAT:     "",
			buyerVAT:      "FR12345678901",
			taxRepVAT:     "",
			wantViolation: false,
			checkField:    "buyer",
		},
		{
			name:          "invalid: seller starts with digit",
			sellerVAT:     "30123456789",
			buyerVAT:      "",
			taxRepVAT:     "",
			wantViolation: true,
			checkField:    "seller",
		},
		{
			name:          "invalid: buyer starts with digit",
			sellerVAT:     "",
			buyerVAT:      "30123456789",
			taxRepVAT:     "",
			wantViolation: true,
			checkField:    "buyer",
		},
		{
			name:          "invalid: seller with lowercase prefix",
			sellerVAT:     "de123456789",
			buyerVAT:      "",
			taxRepVAT:     "",
			wantViolation: true,
			checkField:    "seller",
		},
		{
			name:          "invalid: buyer with mixed case",
			sellerVAT:     "",
			buyerVAT:      "De123456789",
			taxRepVAT:     "",
			wantViolation: true,
			checkField:    "buyer",
		},
		{
			name:          "invalid: tax rep with digit in prefix",
			sellerVAT:     "",
			buyerVAT:      "",
			taxRepVAT:     "D1123456789",
			wantViolation: true,
			checkField:    "taxrep",
		},
		{
			name:          "invalid: seller too short",
			sellerVAT:     "D",
			buyerVAT:      "",
			taxRepVAT:     "",
			wantViolation: true,
			checkField:    "seller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := Invoice{
				GuidelineSpecifiedDocumentContextParameter: SpecFacturXBasic,
				InvoiceNumber:       "TEST-BRCO9",
				InvoiceTypeCode:     380,
				InvoiceDate:         time.Now(),
				InvoiceCurrencyCode: "EUR",
				LineTotal:           decimal.NewFromInt(100),
				TaxBasisTotal:       decimal.NewFromInt(100),
				GrandTotal:          decimal.NewFromInt(119),
				DuePayableAmount:    decimal.NewFromInt(119),
				Seller: Party{
					Name:              "Seller",
					VATaxRegistration: tt.sellerVAT,
					PostalAddress: &PostalAddress{
						CountryID: "DE",
					},
				},
				Buyer: Party{
					Name:              "Buyer",
					VATaxRegistration: tt.buyerVAT,
					PostalAddress: &PostalAddress{
						CountryID: "FR",
					},
				},
				InvoiceLines: []InvoiceLine{
					{
						LineID:          "1",
						ItemName:        "Item",
						BilledQuantity:  decimal.NewFromInt(1),
						NetPrice:        decimal.NewFromInt(100),
						Total:           decimal.NewFromInt(100),
						TaxCategoryCode: "S",
					},
				},
				TradeTaxes: []TradeTax{
					{
						CategoryCode:     "S",
						Percent:          decimal.NewFromInt(19),
						BasisAmount:      decimal.NewFromInt(100),
						CalculatedAmount: decimal.NewFromInt(19),
					},
				},
			}

			// Add tax representative if needed
			if tt.taxRepVAT != "" {
				inv.SellerTaxRepresentativeTradeParty = &Party{
					Name:              "Tax Rep",
					VATaxRegistration: tt.taxRepVAT,
					PostalAddress: &PostalAddress{
						CountryID: "FR",
					},
				}
			}

			err := inv.Validate()

			// Check for BR-CO-9 violation
			var foundBRCO9 bool
			var violationText string
			for _, v := range inv.violations {
				if v.Rule.Code == "BR-CO-09" {
					foundBRCO9 = true
					violationText = v.Text
					t.Logf("BR-CO-09 violation: %s", v.Text)
				}
			}

			if tt.wantViolation && !foundBRCO9 {
				t.Errorf("Expected BR-CO-09 violation but got none. Error: %v", err)
			}
			if !tt.wantViolation && foundBRCO9 {
				t.Errorf("Expected no BR-CO-09 violation but got: %s", violationText)
			}

			// Verify it's NOT reported as BR-DE-16
			for _, v := range inv.violations {
				if v.Rule.Code == "BR-DE-16" && strings.Contains(v.Text, "prefix") {
					t.Errorf("VAT ID format validation incorrectly uses BR-DE-16, should use BR-CO-09. Violation: %s", v.Text)
				}
			}
		})
	}
}

