package einvoice

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestUBLDateParsingValid tests that valid UBL dates (ISO 8601: YYYY-MM-DD) parse successfully
func TestUBLDateParsingValid(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:CustomizationID>urn:cen.eu:en16931:2017</cbc:CustomizationID>
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>2024-03-15</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:InvoicePeriod>
    <cbc:StartDate>2024-02-01</cbc:StartDate>
    <cbc:EndDate>2024-02-29</cbc:EndDate>
  </cac:InvoicePeriod>
  <cac:BillingReference>
    <cac:InvoiceDocumentReference>
      <cbc:ID>PREV-001</cbc:ID>
      <cbc:IssueDate>2024-01-15</cbc:IssueDate>
    </cac:InvoiceDocumentReference>
  </cac:BillingReference>
  <cac:Delivery>
    <cbc:ActualDeliveryDate>2024-03-10</cbc:ActualDeliveryDate>
  </cac:Delivery>
  <cac:AccountingSupplierParty>
    <cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party>
  </cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty>
    <cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party>
  </cac:AccountingCustomerParty>
  <cac:TaxTotal>
    <cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount>
  </cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
  <cac:InvoiceLine>
    <cbc:ID>1</cbc:ID>
    <cbc:InvoicedQuantity unitCode="C62">10</cbc:InvoicedQuantity>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cac:InvoicePeriod>
      <cbc:StartDate>2024-02-01</cbc:StartDate>
      <cbc:EndDate>2024-02-10</cbc:EndDate>
    </cac:InvoicePeriod>
    <cac:Item>
      <cbc:Name>Test Item</cbc:Name>
      <cac:ClassifiedTaxCategory>
        <cbc:ID>S</cbc:ID>
        <cbc:Percent>19.00</cbc:Percent>
        <cac:TaxScheme><cbc:ID>VAT</cbc:ID></cac:TaxScheme>
      </cac:ClassifiedTaxCategory>
    </cac:Item>
    <cac:Price>
      <cbc:PriceAmount currencyID="EUR">10.00</cbc:PriceAmount>
    </cac:Price>
  </cac:InvoiceLine>
</Invoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("unexpected error parsing valid UBL dates: %v", err)
	}

	// Verify invoice date
	if inv.InvoiceDate.Format("2006-01-02") != "2024-03-15" {
		t.Errorf("InvoiceDate: got %s, want 2024-03-15", inv.InvoiceDate.Format("2006-01-02"))
	}

	// Verify occurrence date (BT-72)
	if inv.OccurrenceDateTime.Format("2006-01-02") != "2024-03-10" {
		t.Errorf("OccurrenceDateTime: got %s, want 2024-03-10", inv.OccurrenceDateTime.Format("2006-01-02"))
	}

	// Verify billing period
	if inv.BillingSpecifiedPeriodStart.Format("2006-01-02") != "2024-02-01" {
		t.Errorf("BillingSpecifiedPeriodStart: got %s, want 2024-02-01", inv.BillingSpecifiedPeriodStart.Format("2006-01-02"))
	}
	if inv.BillingSpecifiedPeriodEnd.Format("2006-01-02") != "2024-02-29" {
		t.Errorf("BillingSpecifiedPeriodEnd: got %s, want 2024-02-29", inv.BillingSpecifiedPeriodEnd.Format("2006-01-02"))
	}

	// Verify referenced document date
	if len(inv.InvoiceReferencedDocument) != 1 {
		t.Fatalf("expected 1 referenced document, got %d", len(inv.InvoiceReferencedDocument))
	}
	if inv.InvoiceReferencedDocument[0].Date.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("InvoiceReferencedDocument[0].Date: got %s, want 2024-01-15", inv.InvoiceReferencedDocument[0].Date.Format("2006-01-02"))
	}

	// Verify line billing period
	if len(inv.InvoiceLines) != 1 {
		t.Fatalf("expected 1 invoice line, got %d", len(inv.InvoiceLines))
	}
	if inv.InvoiceLines[0].BillingSpecifiedPeriodStart.Format("2006-01-02") != "2024-02-01" {
		t.Errorf("Line BillingSpecifiedPeriodStart: got %s, want 2024-02-01", inv.InvoiceLines[0].BillingSpecifiedPeriodStart.Format("2006-01-02"))
	}
	if inv.InvoiceLines[0].BillingSpecifiedPeriodEnd.Format("2006-01-02") != "2024-02-10" {
		t.Errorf("Line BillingSpecifiedPeriodEnd: got %s, want 2024-02-10", inv.InvoiceLines[0].BillingSpecifiedPeriodEnd.Format("2006-01-02"))
	}
}

// TestUBLDateParsingInvalid tests that invalid UBL date formats return errors
func TestUBLDateParsingInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		xml        string
		wantErrMsg string
	}{
		{
			name: "invalid invoice date format",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>20240315</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount></cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`,
			wantErrMsg: "invalid date",
		},
		{
			name: "invalid occurrence date",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>2024-03-15</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:Delivery>
    <cbc:ActualDeliveryDate>INVALID</cbc:ActualDeliveryDate>
  </cac:Delivery>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount></cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`,
			wantErrMsg: "invalid occurrence date time",
		},
		{
			name: "invalid billing period start",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>2024-03-15</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:InvoicePeriod>
    <cbc:StartDate>2024-99-99</cbc:StartDate>
  </cac:InvoicePeriod>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount></cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`,
			wantErrMsg: "invalid billing period start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseReader(strings.NewReader(tt.xml))
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

// TestUBLLineDateParsingInvalid tests that invalid line billing period dates return errors
func TestUBLLineDateParsingInvalid(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>2024-03-15</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount></cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
  <cac:InvoiceLine>
    <cbc:ID>LINE-UBL-001</cbc:ID>
    <cbc:InvoicedQuantity unitCode="C62">10</cbc:InvoicedQuantity>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cac:InvoicePeriod>
      <cbc:StartDate>NOT-A-DATE</cbc:StartDate>
    </cac:InvoicePeriod>
    <cac:Item>
      <cbc:Name>Test</cbc:Name>
      <cac:ClassifiedTaxCategory>
        <cbc:ID>S</cbc:ID>
        <cbc:Percent>19.00</cbc:Percent>
        <cac:TaxScheme><cbc:ID>VAT</cbc:ID></cac:TaxScheme>
      </cac:ClassifiedTaxCategory>
    </cac:Item>
    <cac:Price><cbc:PriceAmount currencyID="EUR">10.00</cbc:PriceAmount></cac:Price>
  </cac:InvoiceLine>
</Invoice>`

	_, err := ParseReader(strings.NewReader(xml))
	if err == nil {
		t.Error("expected error for invalid UBL line billing period date, got nil")
		return
	}
	if !strings.Contains(err.Error(), "invalid line billing period start date for line LINE-UBL-001") {
		t.Errorf("error message should contain line ID, got: %v", err)
	}
}

// TestUBLAttachmentBase64Decoding tests that UBL binary attachments are properly decoded
func TestUBLAttachmentBase64Decoding(t *testing.T) {
	t.Parallel()

	// Base64 encoded "UBL Test Data"
	testData := "VUJMIFRlc3QgRGF0YQ=="

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>2024-03-15</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:AdditionalDocumentReference>
    <cbc:ID>DOC-UBL-001</cbc:ID>
    <cbc:DocumentTypeCode>916</cbc:DocumentTypeCode>
    <cbc:DocumentDescription>Test Attachment</cbc:DocumentDescription>
    <cac:Attachment>
      <cbc:EmbeddedDocumentBinaryObject mimeCode="text/plain" filename="test.txt">` + testData + `</cbc:EmbeddedDocumentBinaryObject>
    </cac:Attachment>
  </cac:AdditionalDocumentReference>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount></cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`

	inv, err := ParseReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.AdditionalReferencedDocument) != 1 {
		t.Fatalf("expected 1 additional document, got %d", len(inv.AdditionalReferencedDocument))
	}

	doc := inv.AdditionalReferencedDocument[0]
	if doc.AttachmentMimeCode != "text/plain" {
		t.Errorf("AttachmentMimeCode: got %q, want %q", doc.AttachmentMimeCode, "text/plain")
	}
	if doc.AttachmentFilename != "test.txt" {
		t.Errorf("AttachmentFilename: got %q, want %q", doc.AttachmentFilename, "test.txt")
	}
	if string(doc.AttachmentBinaryObject) != "UBL Test Data" {
		t.Errorf("AttachmentBinaryObject: got %q, want %q", string(doc.AttachmentBinaryObject), "UBL Test Data")
	}
}

// TestUBLAttachmentInvalidBase64 tests that invalid base64 data returns an error
func TestUBLAttachmentInvalidBase64(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>UBL-001</cbc:ID>
  <cbc:IssueDate>2024-03-15</cbc:IssueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:AdditionalDocumentReference>
    <cbc:ID>DOC-001</cbc:ID>
    <cbc:DocumentTypeCode>916</cbc:DocumentTypeCode>
    <cac:Attachment>
      <cbc:EmbeddedDocumentBinaryObject mimeCode="text/plain">!!!INVALID@@@BASE64</cbc:EmbeddedDocumentBinaryObject>
    </cac:Attachment>
  </cac:AdditionalDocumentReference>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Seller</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Buyer</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">19.00</cbc:TaxAmount></cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">100.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">100.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">119.00</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">119.00</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
</Invoice>`

	_, err := ParseReader(strings.NewReader(xml))
	if err == nil {
		t.Error("expected error for invalid base64 data in UBL, got nil")
		return
	}
	if !strings.Contains(err.Error(), "cannot decode attachment") {
		t.Errorf("error message should contain 'cannot decode attachment', got: %v", err)
	}
}

// Benchmark tests for parser performance across different profiles

// BenchmarkParseUBL benchmarks UBL parsing performance
func BenchmarkParseUBL(b *testing.B) {
	benchmarks := []struct {
		name string
		file string
	}{
		{"Invoice", "testdata/ubl/invoice/UBL-Invoice-2.1-Example.xml"},
		{"CreditNote", "testdata/ubl/creditnote/UBL-CreditNote-2.1-Example.xml"},
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

// FuzzParseUBL fuzzes the UBL parser to find crashes and panics
func FuzzParseUBL(f *testing.F) {
	// Seed corpus with valid UBL XML files
	seeds := []string{
		"testdata/ubl/invoice/UBL-Invoice-2.1-Example.xml",
		"testdata/ubl/creditnote/UBL-CreditNote-2.1-Example.xml",
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
