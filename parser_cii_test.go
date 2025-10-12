package einvoice

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	t.Parallel()

	inv, err := ParseXMLFile("testdata/cii/en16931/zugferd_2p0_EN16931_1_Teilrechnung.xml")
	if err != nil {
		t.Fatal(err)
	}

	expected := Invoice{
		InvoiceNumber: "471102",
	}
	if got := inv.InvoiceNumber; got != expected.InvoiceNumber {
		t.Errorf("invoice number got %s, expected %s\n", got, expected.InvoiceNumber)
	}
}

func TestInvalidDecimalValue(t *testing.T) {
	t.Parallel()

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
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>INVALID</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	_, err := ParseReader(strings.NewReader(xml))
	if err == nil {
		t.Error("expected error for invalid decimal value, got nil")
	}
	if !strings.Contains(err.Error(), "invalid decimal value") {
		t.Errorf("expected error message to contain 'invalid decimal value', got: %v", err)
	}
}

func TestCountrySubDivisionNameParsing(t *testing.T) {
	t.Parallel()

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
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty>
        <ram:Name>Buyer Company</ram:Name>
        <ram:PostalTradeAddress>
          <ram:PostcodeCode>12345</ram:PostcodeCode>
          <ram:LineOne>123 Main St</ram:LineOne>
          <ram:CityName>Berlin</ram:CityName>
          <ram:CountryID>DE</ram:CountryID>
          <ram:CountrySubDivisionName>Brandenburg</ram:CountrySubDivisionName>
        </ram:PostalTradeAddress>
      </ram:BuyerTradeParty>
      <ram:SellerTradeParty>
        <ram:Name>Seller Company</ram:Name>
        <ram:PostalTradeAddress>
          <ram:PostcodeCode>54321</ram:PostcodeCode>
          <ram:LineOne>456 Oak Ave</ram:LineOne>
          <ram:CityName>Munich</ram:CityName>
          <ram:CountryID>DE</ram:CountryID>
          <ram:CountrySubDivisionName>Bavaria</ram:CountrySubDivisionName>
        </ram:PostalTradeAddress>
      </ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.Buyer.PostalAddress == nil {
		t.Fatal("buyer postal address is nil")
	}
	if got := inv.Buyer.PostalAddress.CountrySubDivisionName; got != "Brandenburg" {
		t.Errorf("buyer CountrySubDivisionName: got %q, want %q", got, "Brandenburg")
	}

	if inv.Seller.PostalAddress == nil {
		t.Fatal("seller postal address is nil")
	}
	if got := inv.Seller.PostalAddress.CountrySubDivisionName; got != "Bavaria" {
		t.Errorf("seller CountrySubDivisionName: got %q, want %q", got, "Bavaria")
	}
}

// TestCIIDateParsingValid tests that valid CII dates (Format 102: YYYYMMDD) parse successfully
func TestCIIDateParsingValid(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
                          xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"
                          xmlns:qdt="urn:un:unece:uncefact:data:standard:QualifiedDataType:100">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter>
      <ram:ID>urn:cen.eu:en16931:2017</ram:ID>
    </ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument>
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240315</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery>
      <ram:ActualDeliverySupplyChainEvent>
        <ram:OccurrenceDateTime><udt:DateTimeString format="102">20240310</udt:DateTimeString></ram:OccurrenceDateTime>
      </ram:ActualDeliverySupplyChainEvent>
    </ram:ApplicableHeaderTradeDelivery>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:BillingSpecifiedPeriod>
        <ram:StartDateTime><udt:DateTimeString format="102">20240201</udt:DateTimeString></ram:StartDateTime>
        <ram:EndDateTime><udt:DateTimeString format="102">20240229</udt:DateTimeString></ram:EndDateTime>
      </ram:BillingSpecifiedPeriod>
      <ram:InvoiceReferencedDocument>
        <ram:IssuerAssignedID>PREV-001</ram:IssuerAssignedID>
        <ram:FormattedIssueDateTime><qdt:DateTimeString format="102">20240101</qdt:DateTimeString></ram:FormattedIssueDateTime>
      </ram:InvoiceReferencedDocument>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
    <ram:IncludedSupplyChainTradeLineItem>
      <ram:AssociatedDocumentLineDocument>
        <ram:LineID>1</ram:LineID>
      </ram:AssociatedDocumentLineDocument>
      <ram:SpecifiedTradeProduct><ram:Name>Test Item</ram:Name></ram:SpecifiedTradeProduct>
      <ram:SpecifiedLineTradeAgreement>
        <ram:NetPriceProductTradePrice><ram:ChargeAmount>10.00</ram:ChargeAmount></ram:NetPriceProductTradePrice>
      </ram:SpecifiedLineTradeAgreement>
      <ram:SpecifiedLineTradeDelivery>
        <ram:BilledQuantity unitCode="C62">10</ram:BilledQuantity>
      </ram:SpecifiedLineTradeDelivery>
      <ram:SpecifiedLineTradeSettlement>
        <ram:BillingSpecifiedPeriod>
          <ram:StartDateTime><udt:DateTimeString format="102">20240201</udt:DateTimeString></ram:StartDateTime>
          <ram:EndDateTime><udt:DateTimeString format="102">20240210</udt:DateTimeString></ram:EndDateTime>
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
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("unexpected error parsing valid CII dates: %v", err)
	}

	// Verify invoice date
	if inv.InvoiceDate.Format("20060102") != "20240315" {
		t.Errorf("InvoiceDate: got %s, want 20240315", inv.InvoiceDate.Format("20060102"))
	}

	// Verify occurrence date time (BT-72)
	if inv.OccurrenceDateTime.Format("20060102") != "20240310" {
		t.Errorf("OccurrenceDateTime: got %s, want 20240310", inv.OccurrenceDateTime.Format("20060102"))
	}

	// Verify billing period (BT-73, BT-74)
	if inv.BillingSpecifiedPeriodStart.Format("20060102") != "20240201" {
		t.Errorf("BillingSpecifiedPeriodStart: got %s, want 20240201", inv.BillingSpecifiedPeriodStart.Format("20060102"))
	}
	if inv.BillingSpecifiedPeriodEnd.Format("20060102") != "20240229" {
		t.Errorf("BillingSpecifiedPeriodEnd: got %s, want 20240229", inv.BillingSpecifiedPeriodEnd.Format("20060102"))
	}

	// Verify referenced document date
	if len(inv.InvoiceReferencedDocument) != 1 {
		t.Fatalf("expected 1 referenced document, got %d", len(inv.InvoiceReferencedDocument))
	}
	if inv.InvoiceReferencedDocument[0].Date.Format("20060102") != "20240101" {
		t.Errorf("InvoiceReferencedDocument[0].Date: got %s, want 20240101", inv.InvoiceReferencedDocument[0].Date.Format("20060102"))
	}

	// Verify line billing period (BT-134, BT-135)
	if len(inv.InvoiceLines) != 1 {
		t.Fatalf("expected 1 invoice line, got %d", len(inv.InvoiceLines))
	}
	if inv.InvoiceLines[0].BillingSpecifiedPeriodStart.Format("20060102") != "20240201" {
		t.Errorf("Line BillingSpecifiedPeriodStart: got %s, want 20240201", inv.InvoiceLines[0].BillingSpecifiedPeriodStart.Format("20060102"))
	}
	if inv.InvoiceLines[0].BillingSpecifiedPeriodEnd.Format("20060102") != "20240210" {
		t.Errorf("Line BillingSpecifiedPeriodEnd: got %s, want 20240210", inv.InvoiceLines[0].BillingSpecifiedPeriodEnd.Format("20060102"))
	}
}

// TestCIIDateParsingInvalidFormat tests that invalid date formats return errors
func TestCIIDateParsingInvalidFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		dateElement string
		wantErrMsg  string
	}{
		{
			name:        "invalid invoice date format",
			dateElement: `<ram:IssueDateTime><udt:DateTimeString format="102">2024-03-15</udt:DateTimeString></ram:IssueDateTime>`,
			wantErrMsg:  "parsing time",
		},
		{
			name:        "invalid occurrence date",
			dateElement: `<ram:ActualDeliverySupplyChainEvent><ram:OccurrenceDateTime><udt:DateTimeString format="102">INVALID</udt:DateTimeString></ram:OccurrenceDateTime></ram:ActualDeliverySupplyChainEvent>`,
			wantErrMsg:  "invalid occurrence date time",
		},
		{
			name:        "invalid billing period start",
			dateElement: `<ram:BillingSpecifiedPeriod><ram:StartDateTime><udt:DateTimeString format="102">99999999</udt:DateTimeString></ram:StartDateTime></ram:BillingSpecifiedPeriod>`,
			wantErrMsg:  "invalid billing period start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var xml string
			if strings.Contains(tt.dateElement, "IssueDateTime") {
				xml = `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
                          xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter>
      <ram:ID>urn:cen.eu:en16931:2017</ram:ID>
    </ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument>
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    ` + tt.dateElement + `
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`
			} else {
				xml = `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
                          xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter>
      <ram:ID>urn:cen.eu:en16931:2017</ram:ID>
    </ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument>
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery>` + tt.dateElement + `</ram:ApplicableHeaderTradeDelivery>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      ` + (func() string {
					if strings.Contains(tt.dateElement, "BillingSpecifiedPeriod") {
						return tt.dateElement
					}
					return ""
				}()) + `
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`
			}

			_, err := ParseReader(strings.NewReader(xml))
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("%s: error message %q does not contain %q", tt.name, err.Error(), tt.wantErrMsg)
			}
		})
	}
}

// TestCIILineDateParsingInvalid tests that invalid line billing period dates return errors
func TestCIILineDateParsingInvalid(t *testing.T) {
	t.Parallel()

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
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
    <ram:IncludedSupplyChainTradeLineItem>
      <ram:AssociatedDocumentLineDocument>
        <ram:LineID>LINE-001</ram:LineID>
      </ram:AssociatedDocumentLineDocument>
      <ram:SpecifiedTradeProduct><ram:Name>Test</ram:Name></ram:SpecifiedTradeProduct>
      <ram:SpecifiedLineTradeAgreement>
        <ram:NetPriceProductTradePrice><ram:ChargeAmount>10.00</ram:ChargeAmount></ram:NetPriceProductTradePrice>
      </ram:SpecifiedLineTradeAgreement>
      <ram:SpecifiedLineTradeDelivery>
        <ram:BilledQuantity unitCode="C62">10</ram:BilledQuantity>
      </ram:SpecifiedLineTradeDelivery>
      <ram:SpecifiedLineTradeSettlement>
        <ram:BillingSpecifiedPeriod>
          <ram:StartDateTime><udt:DateTimeString format="102">BADDATE</udt:DateTimeString></ram:StartDateTime>
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
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	_, err := ParseReader(strings.NewReader(xml))
	if err == nil {
		t.Error("expected error for invalid line billing period date, got nil")
		return
	}
	if !strings.Contains(err.Error(), "invalid line billing period start date for line LINE-001") {
		t.Errorf("error message should contain line ID, got: %v", err)
	}
}

// TestCIIAttachmentBase64Encoding tests that binary attachments are properly encoded/decoded
func TestCIIAttachmentBase64Encoding(t *testing.T) {
	t.Parallel()

	// Base64 encoded "Hello, World!"
	testData := "SGVsbG8sIFdvcmxkIQ=="

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
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
      <ram:AdditionalReferencedDocument>
        <ram:IssuerAssignedID>DOC-001</ram:IssuerAssignedID>
        <ram:TypeCode>130</ram:TypeCode>
        <ram:Name>Test Document</ram:Name>
        <ram:AttachmentBinaryObject mimeCode="application/pdf" filename="test.pdf">` + testData + `</ram:AttachmentBinaryObject>
      </ram:AdditionalReferencedDocument>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.AdditionalReferencedDocument) != 1 {
		t.Fatalf("expected 1 additional document, got %d", len(inv.AdditionalReferencedDocument))
	}

	doc := inv.AdditionalReferencedDocument[0]
	if doc.AttachmentMimeCode != "application/pdf" {
		t.Errorf("AttachmentMimeCode: got %q, want %q", doc.AttachmentMimeCode, "application/pdf")
	}
	if doc.AttachmentFilename != "test.pdf" {
		t.Errorf("AttachmentFilename: got %q, want %q", doc.AttachmentFilename, "test.pdf")
	}
	if string(doc.AttachmentBinaryObject) != "Hello, World!" {
		t.Errorf("AttachmentBinaryObject: got %q, want %q", string(doc.AttachmentBinaryObject), "Hello, World!")
	}
}

// TestCIIAttachmentInvalidBase64 tests that invalid base64 data returns an error
func TestCIIAttachmentInvalidBase64(t *testing.T) {
	t.Parallel()

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
    <ram:ID>INV-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime><udt:DateTimeString format="102">20240101</udt:DateTimeString></ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:BuyerTradeParty><ram:Name>Buyer</ram:Name></ram:BuyerTradeParty>
      <ram:SellerTradeParty><ram:Name>Seller</ram:Name></ram:SellerTradeParty>
      <ram:AdditionalReferencedDocument>
        <ram:IssuerAssignedID>DOC-001</ram:IssuerAssignedID>
        <ram:TypeCode>130</ram:TypeCode>
        <ram:AttachmentBinaryObject mimeCode="application/pdf">INVALID!!!BASE64</ram:AttachmentBinaryObject>
      </ram:AdditionalReferencedDocument>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>19.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>119.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>119.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	_, err := ParseReader(strings.NewReader(xml))
	if err == nil {
		t.Error("expected error for invalid base64 data, got nil")
		return
	}
	if !strings.Contains(err.Error(), "cannot decode attachment") {
		t.Errorf("error message should contain 'cannot decode attachment', got: %v", err)
	}
}

// Benchmark tests for parser performance across different profiles

// BenchmarkParseCII benchmarks CII parsing performance across all profiles
func BenchmarkParseCII(b *testing.B) {
	benchmarks := []struct {
		name string
		file string
	}{
		{"Minimum", "testdata/cii/minimum/zugferd-minimum-rechnung.xml"},
		{"BasicWL", "testdata/cii/basicwl/zugferd-basicwl-rechnung.xml"},
		{"Basic", "testdata/cii/basic/zugferd-basic-rechnung.xml"},
		{"EN16931", "testdata/cii/en16931/zugferd_2p3_EN16931_1.xml"},
		{"Extended", "testdata/cii/extended/zugferd-extended-rechnung.xml"},
		{"XRechnung", "testdata/cii/xrechnung/zugferd-xrechnung-einfach.xml"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data, err := os.ReadFile(bm.file)
			if err != nil {
				b.Skipf("File not found: %s", bm.file)
			}

			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_, err := ParseReader(bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Fuzz tests for parser robustness

// FuzzParseCII fuzzes the CII parser to find crashes and panics
func FuzzParseCII(f *testing.F) {
	// Seed corpus with valid CII XML files from different profiles
	seeds := []string{
		"testdata/cii/minimum/zugferd-minimum-rechnung.xml",
		"testdata/cii/basicwl/zugferd-basicwl-rechnung.xml",
		"testdata/cii/basic/zugferd-basic-rechnung.xml",
		"testdata/cii/en16931/zugferd_2p3_EN16931_1.xml",
		"testdata/cii/extended/zugferd-extended-rechnung.xml",
		"testdata/cii/xrechnung/zugferd-xrechnung-einfach.xml",
	}

	for _, seed := range seeds {
		data, err := os.ReadFile(seed)
		if err == nil {
			f.Add(data)
		}
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, even with invalid input
		_, err := ParseReader(bytes.NewReader(data))
		_ = err // Error is expected for invalid inputs
	})
}
