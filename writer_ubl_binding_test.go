package einvoice

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"
)

// Child element sequences from the UBL 2.1 schemas, restricted to the
// elements the writer produces.
var (
	ublInvoiceSequence = []string{
		"cbc:CustomizationID", "cbc:ProfileID", "cbc:ID", "cbc:IssueDate", "cbc:DueDate",
		"cbc:InvoiceTypeCode", "cbc:Note", "cbc:DocumentCurrencyCode", "cbc:TaxCurrencyCode",
		"cbc:AccountingCost", "cbc:BuyerReference", "cac:InvoicePeriod", "cac:OrderReference",
		"cac:BillingReference", "cac:DespatchDocumentReference", "cac:ReceiptDocumentReference",
		"cac:ContractDocumentReference", "cac:AdditionalDocumentReference", "cac:ProjectReference",
		"cac:AccountingSupplierParty", "cac:AccountingCustomerParty", "cac:PayeeParty",
		"cac:TaxRepresentativeParty", "cac:Delivery", "cac:PaymentMeans", "cac:PaymentTerms",
		"cac:AllowanceCharge", "cac:TaxTotal", "cac:LegalMonetaryTotal", "cac:InvoiceLine",
	}
	ublCreditNoteSequence = []string{
		"cbc:CustomizationID", "cbc:ProfileID", "cbc:ID", "cbc:IssueDate",
		"cbc:CreditNoteTypeCode", "cbc:Note", "cbc:DocumentCurrencyCode", "cbc:TaxCurrencyCode",
		"cbc:AccountingCost", "cbc:BuyerReference", "cac:InvoicePeriod", "cac:OrderReference",
		"cac:BillingReference", "cac:DespatchDocumentReference", "cac:ReceiptDocumentReference",
		"cac:ContractDocumentReference", "cac:AdditionalDocumentReference",
		"cac:AccountingSupplierParty", "cac:AccountingCustomerParty", "cac:PayeeParty",
		"cac:TaxRepresentativeParty", "cac:Delivery", "cac:PaymentMeans", "cac:PaymentTerms",
		"cac:AllowanceCharge", "cac:TaxTotal", "cac:LegalMonetaryTotal", "cac:CreditNoteLine",
	}
	ublSequences = map[string][]string{
		"cac:Party": {
			"cbc:EndpointID", "cac:PartyIdentification", "cac:PartyName", "cac:PostalAddress",
			"cac:PartyTaxScheme", "cac:PartyLegalEntity", "cac:Contact",
		},
		"cac:PostalAddress": {
			"cbc:StreetName", "cbc:AdditionalStreetName", "cbc:CityName", "cbc:PostalZone",
			"cbc:CountrySubentity", "cac:AddressLine", "cac:Country",
		},
		"cac:Contact":            {"cbc:Name", "cbc:Telephone", "cbc:ElectronicMail"},
		"cac:PartyLegalEntity":   {"cbc:RegistrationName", "cbc:CompanyID"},
		"cac:Attachment":         {"cbc:EmbeddedDocumentBinaryObject", "cac:ExternalReference"},
		"cac:TaxCategory":        {"cbc:ID", "cbc:Percent", "cbc:TaxExemptionReasonCode", "cbc:TaxExemptionReason", "cac:TaxScheme"},
		"cac:PaymentMandate":     {"cbc:ID", "cac:PayerFinancialAccount"},
		"cac:LegalMonetaryTotal": {"cbc:LineExtensionAmount", "cbc:TaxExclusiveAmount", "cbc:TaxInclusiveAmount", "cbc:AllowanceTotalAmount", "cbc:ChargeTotalAmount", "cbc:PrepaidAmount", "cbc:PayableRoundingAmount", "cbc:PayableAmount"},
		"cac:PaymentMeans": {
			"cbc:PaymentMeansCode", "cbc:PaymentDueDate", "cbc:PaymentID",
			"cac:CardAccount", "cac:PayeeFinancialAccount", "cac:PaymentMandate",
		},
		"cac:Item": {
			"cbc:Description", "cbc:Name", "cac:BuyersItemIdentification", "cac:SellersItemIdentification",
			"cac:StandardItemIdentification", "cac:OriginCountry", "cac:CommodityClassification",
			"cac:ClassifiedTaxCategory", "cac:AdditionalItemProperty",
		},
	}
)

// checkChildOrder reports an error if the children of elt are not in the
// order given by sequence or if a child is not part of sequence.
func checkChildOrder(t *testing.T, elt *etree.Element, sequence []string) {
	t.Helper()
	last := -1
	for _, child := range elt.ChildElements() {
		pos := slices.Index(sequence, child.FullTag())
		if pos < 0 {
			t.Errorf("%s: unexpected child %s", elt.GetPath(), child.FullTag())
			continue
		}
		if pos < last {
			t.Errorf("%s: %s is out of order (expected sequence %v)", elt.GetPath(), child.FullTag(), sequence)
		}
		last = pos
	}
}

// checkUBLOrder checks the element order of the document and of all
// descendants with a known sequence.
func checkUBLOrder(t *testing.T, doc *etree.Document) {
	t.Helper()
	root := doc.Root()
	if root.Tag == "CreditNote" {
		checkChildOrder(t, root, ublCreditNoteSequence)
	} else {
		checkChildOrder(t, root, ublInvoiceSequence)
	}
	for tag, sequence := range ublSequences {
		for _, elt := range root.FindElements("//" + tag) {
			checkChildOrder(t, elt, sequence)
		}
	}
}

// newUBLBindingTestInvoice returns an invoice that uses most of the elements
// whose position or mapping in UBL is not obvious.
func newUBLBindingTestInvoice(typeCode CodeDocument) *Invoice {
	d := decimal.RequireFromString
	inv := &Invoice{
		SchemaType: UBL,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		BPSpecifiedDocumentContextParameter:        "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
		InvoiceNumber:                              "INV-185",
		InvoiceTypeCode:                            typeCode,
		InvoiceDate:                                time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		InvoiceCurrencyCode:                        "EUR",
		BuyerReference:                             "04011000-12345-34",
		ReceivableSpecifiedTradeAccountingAccount:  "4711",
		SpecifiedProcuringProjectID:                "PRJ-1",
		AdditionalReferencedDocument: []Document{{
			IssuerAssignedID:       "ATT-1",
			Name:                   "Timesheet",
			URIID:                  "https://example.com/timesheet.pdf",
			AttachmentMimeCode:     "application/pdf",
			AttachmentFilename:     "timesheet.pdf",
			AttachmentBinaryObject: []byte("%PDF-1.4"),
		}},
		Seller: Party{
			Name: "Seller Legal Name GmbH",
			PostalAddress: &PostalAddress{
				Line1: "Hauptstraße 1", Line3: "Hinterhaus", City: "Berlin",
				PostcodeCode: "10115", CountrySubDivisionName: "Berlin", CountryID: "DE",
			},
			VATaxRegistration: "DE123456789",
			SpecifiedLegalOrganization: &SpecifiedLegalOrganization{
				ID: "HRB 12345", TradingBusinessName: "Seller Trading Name",
			},
			DefinedTradeContact: []DefinedTradeContact{{
				DepartmentName: "Accounting", PhoneNumber: "+49 30 1234", EMail: "billing@example.com",
			}},
			URIUniversalCommunication:       "billing@example.com",
			URIUniversalCommunicationScheme: "EM",
		},
		Buyer: Party{
			Name: "Buyer Legal Name AG",
			PostalAddress: &PostalAddress{
				Line1: "Marktplatz 2", City: "München", PostcodeCode: "80331", CountryID: "DE",
			},
			URIUniversalCommunication:       "buyer@example.com",
			URIUniversalCommunicationScheme: "EM",
		},
		PayeeTradeParty: &Party{Name: "Payee Factoring GmbH"},
		PaymentMeans: []PaymentMeans{{
			TypeCode: 59, Information: "SEPA-Lastschrift", PayerPartyDebtorFinancialAccountIBAN: "DE02500105170137075030",
		}},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{{
			Description:          "Zahlbar innerhalb von 14 Tagen",
			DueDate:              time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			DirectDebitMandateID: "MANDATE-1",
		}},
		SpecifiedTradeAllowanceCharge: []AllowanceCharge{
			{
				ChargeIndicator: false, ActualAmount: d("3.75"), Reason: "Rabatt",
				CategoryTradeTaxType: "VAT", CategoryTradeTaxCategoryCode: "S", CategoryTradeTaxRateApplicablePercent: d("19"),
			},
			{
				ChargeIndicator: true, ActualAmount: d("4.90"), Reason: "Versand",
				CategoryTradeTaxType: "VAT", CategoryTradeTaxCategoryCode: "S", CategoryTradeTaxRateApplicablePercent: d("19"),
			},
		},
		InvoiceLines: []InvoiceLine{{
			LineID: "1", ItemName: "Widget", ArticleNumber: "S-1", ArticleNumberBuyer: "B-1",
			BilledQuantity: d("3"), BilledQuantityUnit: "H87", NetPrice: d("12.49"), Total: d("37.47"),
			TaxCategoryCode: "S", TaxTypeCode: "VAT", TaxRateApplicablePercent: d("19"),
		}},
	}
	inv.UpdateApplicableTradeTax(nil)
	inv.UpdateTotals()
	return inv
}

func writeAndParseUBL(t *testing.T, inv *Invoice) (*etree.Document, *Invoice) {
	t.Helper()
	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(buf.Bytes()); err != nil {
		t.Fatalf("read written XML: %v", err)
	}
	parsed, err := ParseReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}
	return doc, parsed
}

func elementText(t *testing.T, doc *etree.Document, path string) string {
	t.Helper()
	elt := doc.Root().FindElement(path)
	if elt == nil {
		t.Errorf("missing element %s", path)
		return ""
	}
	return elt.Text()
}

// TestWriteUBLBinding checks the element order against the UBL 2.1 schema
// and the EN 16931 UBL syntax binding of the written document (issue #185).
func TestWriteUBLBinding(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeCode CodeDocument
	}{
		{"Invoice", 380},
		{"CreditNote", 381},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inv := newUBLBindingTestInvoice(tc.typeCode)
			doc, parsed := writeAndParseUBL(t, inv)
			checkUBLOrder(t, doc)

			// BT-27, BT-28, BT-44: legal names in PartyLegalEntity, trading name in PartyName
			if got := elementText(t, doc, "cac:AccountingSupplierParty/cac:Party/cac:PartyLegalEntity/cbc:RegistrationName"); got != "Seller Legal Name GmbH" {
				t.Errorf("seller RegistrationName = %q", got)
			}
			if got := elementText(t, doc, "cac:AccountingSupplierParty/cac:Party/cac:PartyName/cbc:Name"); got != "Seller Trading Name" {
				t.Errorf("seller PartyName = %q", got)
			}
			if got := elementText(t, doc, "cac:AccountingCustomerParty/cac:Party/cac:PartyLegalEntity/cbc:RegistrationName"); got != "Buyer Legal Name AG" {
				t.Errorf("buyer RegistrationName = %q", got)
			}
			if doc.Root().FindElement("cac:AccountingCustomerParty/cac:Party/cac:PartyName") != nil {
				t.Error("buyer without trading name must not have a PartyName")
			}
			// BT-59: payee name stays in PartyName
			if got := elementText(t, doc, "cac:PayeeParty/cac:PartyName/cbc:Name"); got != "Payee Factoring GmbH" {
				t.Errorf("payee PartyName = %q", got)
			}
			if doc.Root().FindElement("cac:PayeeParty/cac:PartyLegalEntity") != nil {
				t.Error("payee must not have a PartyLegalEntity")
			}
			// BT-41: department as contact name
			if got := elementText(t, doc, "cac:AccountingSupplierParty/cac:Party/cac:Contact/cbc:Name"); got != "Accounting" {
				t.Errorf("seller contact name = %q", got)
			}

			// BT-9 and BT-89 are not part of PaymentTerms in UBL (UBL-CR-463, UBL-CR-453)
			if doc.Root().FindElement("cac:PaymentTerms/cbc:PaymentDueDate") != nil {
				t.Error("PaymentTerms must not contain PaymentDueDate")
			}
			if doc.Root().FindElement("cac:PaymentTerms/cbc:PaymentMeansID") != nil {
				t.Error("PaymentTerms must not contain PaymentMeansID")
			}
			dueDatePath, projectPath := "cbc:DueDate", "cac:ProjectReference/cbc:ID"
			if tc.typeCode == 381 {
				dueDatePath, projectPath = "cac:PaymentMeans/cbc:PaymentDueDate", "cac:AdditionalDocumentReference[cbc:DocumentTypeCode='50']/cbc:ID"
			}
			if got := elementText(t, doc, dueDatePath); got != "2025-03-15" {
				t.Errorf("%s = %q", dueDatePath, got)
			}
			if got := elementText(t, doc, projectPath); got != "PRJ-1" {
				t.Errorf("%s = %q", projectPath, got)
			}
			if got := elementText(t, doc, "cac:PaymentMeans/cac:PaymentMandate/cbc:ID"); got != "MANDATE-1" {
				t.Errorf("mandate ID = %q", got)
			}
			// BT-82: payment means text
			if got := doc.Root().FindElement("cac:PaymentMeans/cbc:PaymentMeansCode").SelectAttrValue("name", ""); got != "SEPA-Lastschrift" {
				t.Errorf("PaymentMeansCode/@name = %q", got)
			}

			// Round trip
			if parsed.Seller.Name != inv.Seller.Name {
				t.Errorf("seller name round trip: %q", parsed.Seller.Name)
			}
			if parsed.Seller.SpecifiedLegalOrganization == nil || parsed.Seller.SpecifiedLegalOrganization.TradingBusinessName != "Seller Trading Name" {
				t.Errorf("seller trading name round trip: %+v", parsed.Seller.SpecifiedLegalOrganization)
			}
			if parsed.Buyer.Name != inv.Buyer.Name {
				t.Errorf("buyer name round trip: %q", parsed.Buyer.Name)
			}
			if parsed.PayeeTradeParty == nil || parsed.PayeeTradeParty.Name != "Payee Factoring GmbH" {
				t.Errorf("payee round trip: %+v", parsed.PayeeTradeParty)
			}
			if parsed.SpecifiedProcuringProjectID != "PRJ-1" {
				t.Errorf("project reference round trip: %q", parsed.SpecifiedProcuringProjectID)
			}
			if len(parsed.AdditionalReferencedDocument) != 1 {
				t.Errorf("additional documents round trip: got %d, want 1", len(parsed.AdditionalReferencedDocument))
			}
			if len(parsed.PaymentMeans) != 1 || parsed.PaymentMeans[0].Information != "SEPA-Lastschrift" {
				t.Errorf("payment means round trip: %+v", parsed.PaymentMeans)
			}
			if len(parsed.SpecifiedTradePaymentTerms) != 1 {
				t.Fatalf("payment terms round trip: got %d, want 1", len(parsed.SpecifiedTradePaymentTerms))
			}
			pt := parsed.SpecifiedTradePaymentTerms[0]
			if !pt.DueDate.Equal(inv.SpecifiedTradePaymentTerms[0].DueDate) || pt.DirectDebitMandateID != "MANDATE-1" || pt.Description != "Zahlbar innerhalb von 14 Tagen" {
				t.Errorf("payment terms round trip: %+v", pt)
			}

			if err := parsed.Validate(); err != nil {
				t.Errorf("Validate after round trip: %v", err)
			}
			for _, w := range parsed.Warnings() {
				t.Errorf("warning after round trip: %s %s", w.Rule.Code, w.Text)
			}
		})
	}
}

// TestParseUBLLegacyPartyNames checks that documents with the legal name only
// in PartyName (as written by earlier versions of this library) still parse.
func TestParseUBLLegacyPartyNames(t *testing.T) {
	inv := newUBLBindingTestInvoice(380)
	doc, _ := writeAndParseUBL(t, inv)
	party := doc.Root().FindElement("cac:AccountingCustomerParty/cac:Party")
	party.RemoveChild(party.FindElement("cac:PartyLegalEntity"))
	name := etree.NewElement("cac:PartyName")
	name.CreateElement("cbc:Name").SetText("Buyer Legal Name AG")
	party.InsertChildAt(party.FindElement("cac:PostalAddress").Index(), name)

	data, err := doc.WriteToBytes()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}
	if parsed.Buyer.Name != "Buyer Legal Name AG" {
		t.Errorf("buyer name = %q", parsed.Buyer.Name)
	}
	if parsed.Buyer.SpecifiedLegalOrganization != nil {
		t.Errorf("buyer legal organization = %+v, want nil", parsed.Buyer.SpecifiedLegalOrganization)
	}
}

// TestWriteNotSubjectToVATAllowanceCharge checks that document level
// allowances and charges with category O have no VAT rate (BR-O-6, BR-O-7).
func TestWriteNotSubjectToVATAllowanceCharge(t *testing.T) {
	for _, schema := range []CodeSchemaType{UBL, CII} {
		inv := newUBLBindingTestInvoice(380)
		inv.SchemaType = schema
		inv.GuidelineSpecifiedDocumentContextParameter = SpecEN16931
		inv.InvoiceLines[0].TaxCategoryCode = "O"
		inv.InvoiceLines[0].TaxRateApplicablePercent = decimal.Zero
		inv.Seller.VATaxRegistration = ""
		inv.AdditionalReferencedDocument = nil
		for i := range inv.SpecifiedTradeAllowanceCharge {
			inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode = "O"
			inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent = decimal.Zero
		}
		inv.UpdateApplicableTradeTax(map[string]string{"O": "Not subject to VAT"})
		inv.UpdateTotals()

		var buf bytes.Buffer
		if err := inv.Write(&buf); err != nil {
			t.Fatalf("Write: %v", err)
		}
		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(buf.Bytes()); err != nil {
			t.Fatal(err)
		}
		path := "//cac:AllowanceCharge/cac:TaxCategory/cbc:Percent"
		if schema == CII {
			path = "//ram:SpecifiedTradeAllowanceCharge/ram:CategoryTradeTax/ram:RateApplicablePercent"
		}
		if elts := doc.FindElements(path); len(elts) > 0 {
			t.Errorf("schema %v: found %d VAT rates for category O allowances/charges", schema, len(elts))
		}
		parsed, err := ParseReader(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("ParseReader: %v", err)
		}
		if err := parsed.Validate(); err != nil {
			t.Errorf("schema %v: Validate: %+v", schema, err)
		}
	}
}
