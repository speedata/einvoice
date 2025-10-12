package einvoice

import (
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	t.Parallel()

	inv, err := ParseXMLFile("testdata/cii/en16931/zugferd_2p0_EN16931_1_Teilrechnung.xml")
	if err != nil {
		t.Error(err)
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
