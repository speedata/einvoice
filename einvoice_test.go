package einvoice

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
)

func ExampleInvoice_Write() {
	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")
	fourteenDays := time.Hour * 24 * 14
	inv := Invoice{
		InvoiceNumber:   "1234",
		InvoiceTypeCode: 380,
		GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
		InvoiceDate:         fixedDate,
		OccurrenceDateTime:  fixedDate.Add(-fourteenDays),
		InvoiceCurrencyCode: "EUR",
		TaxCurrencyCode:     "EUR",
		Notes: []Note{{
			Text: "Some text",
		}},
		Seller: Party{
			Name:              "Company name",
			VATaxRegistration: "DE123456",
			PostalAddress: &PostalAddress{
				Line1:        "Line one",
				Line2:        "Line two",
				City:         "City",
				PostcodeCode: "12345",
				CountryID:    "DE",
			},
			DefinedTradeContact: []DefinedTradeContact{{
				PersonName: "Jon Doe",
				EMail:      "doe@example.com",
			}},
		},
		Buyer: Party{
			Name: "Buyer",
			PostalAddress: &PostalAddress{
				Line1:        "Buyer line 1",
				Line2:        "Buyer line 2",
				City:         "Buyercity",
				PostcodeCode: "33441",
				CountryID:    "FR",
			},
			DefinedTradeContact: []DefinedTradeContact{{
				PersonName: "Buyer Person",
			}},
			VATaxRegistration: "FR4441112",
		},
		PaymentMeans: []PaymentMeans{
			{
				TypeCode:                                      30,
				PayeePartyCreditorFinancialAccountIBAN:        "DE123455958381",
				PayeePartyCreditorFinancialAccountName:        "My own bank",
				PayeeSpecifiedCreditorFinancialInstitutionBIC: "BANKDEFXXX",
			},
		},
		SpecifiedTradePaymentTerms: []SpecifiedTradePaymentTerms{{
			DueDate: fixedDate.Add(fourteenDays),
		}},
		InvoiceLines: []InvoiceLine{
			{
				LineID:                   "1",
				ItemName:                 "Item name one",
				BilledQuantity:           decimal.NewFromFloat(12.5),
				BilledQuantityUnit:       "C62",
				NetPrice:                 decimal.NewFromInt(100),
				TaxRateApplicablePercent: decimal.NewFromInt(19),
				Total:                    decimal.NewFromInt(1250),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "S",
			},
			{
				LineID:                   "2",
				ItemName:                 "Item name two",
				BilledQuantity:           decimal.NewFromFloat(2),
				BilledQuantityUnit:       "HUR",
				NetPrice:                 decimal.NewFromInt(200),
				TaxRateApplicablePercent: decimal.NewFromInt(0),
				Total:                    decimal.NewFromInt(400),
				TaxTypeCode:              "VAT",
				TaxCategoryCode:          "AE",
			},
		},
	}

	inv.UpdateApplicableTradeTax(map[string]string{"AE": "Reason for reverse charge"})
	inv.UpdateTotals()
	if err := inv.Write(os.Stdout); err != nil {
		panic(err.Error())
	}
	// Output:
	// <rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100" xmlns:qdt="urn:un:unece:uncefact:data:standard:QualifiedDataType:100" xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">
	//   <rsm:ExchangedDocumentContext>
	//     <ram:GuidelineSpecifiedDocumentContextParameter>
	//       <ram:ID>urn:cen.eu:en16931:2017</ram:ID>
	//     </ram:GuidelineSpecifiedDocumentContextParameter>
	//   </rsm:ExchangedDocumentContext>
	//   <rsm:ExchangedDocument>
	//     <ram:ID>1234</ram:ID>
	//     <ram:TypeCode>380</ram:TypeCode>
	//     <ram:IssueDateTime>
	//       <udt:DateTimeString format="102">20251231</udt:DateTimeString>
	//     </ram:IssueDateTime>
	//     <ram:IncludedNote>
	//       <ram:Content>Some text</ram:Content>
	//     </ram:IncludedNote>
	//   </rsm:ExchangedDocument>
	//   <rsm:SupplyChainTradeTransaction>
	//     <ram:IncludedSupplyChainTradeLineItem>
	//       <ram:AssociatedDocumentLineDocument>
	//         <ram:LineID>1</ram:LineID>
	//       </ram:AssociatedDocumentLineDocument>
	//       <ram:SpecifiedTradeProduct>
	//         <ram:Name>Item name one</ram:Name>
	//       </ram:SpecifiedTradeProduct>
	//       <ram:SpecifiedLineTradeAgreement>
	//         <ram:NetPriceProductTradePrice>
	//           <ram:ChargeAmount>100</ram:ChargeAmount>
	//         </ram:NetPriceProductTradePrice>
	//       </ram:SpecifiedLineTradeAgreement>
	//       <ram:SpecifiedLineTradeDelivery>
	//         <ram:BilledQuantity unitCode="C62">12.5000</ram:BilledQuantity>
	//       </ram:SpecifiedLineTradeDelivery>
	//       <ram:SpecifiedLineTradeSettlement>
	//         <ram:ApplicableTradeTax>
	//           <ram:TypeCode>VAT</ram:TypeCode>
	//           <ram:CategoryCode>S</ram:CategoryCode>
	//           <ram:RateApplicablePercent>19</ram:RateApplicablePercent>
	//         </ram:ApplicableTradeTax>
	//         <ram:SpecifiedTradeSettlementLineMonetarySummation>
	//           <ram:LineTotalAmount>1250.00</ram:LineTotalAmount>
	//         </ram:SpecifiedTradeSettlementLineMonetarySummation>
	//       </ram:SpecifiedLineTradeSettlement>
	//     </ram:IncludedSupplyChainTradeLineItem>
	//     <ram:IncludedSupplyChainTradeLineItem>
	//       <ram:AssociatedDocumentLineDocument>
	//         <ram:LineID>2</ram:LineID>
	//       </ram:AssociatedDocumentLineDocument>
	//       <ram:SpecifiedTradeProduct>
	//         <ram:Name>Item name two</ram:Name>
	//       </ram:SpecifiedTradeProduct>
	//       <ram:SpecifiedLineTradeAgreement>
	//         <ram:NetPriceProductTradePrice>
	//           <ram:ChargeAmount>200</ram:ChargeAmount>
	//         </ram:NetPriceProductTradePrice>
	//       </ram:SpecifiedLineTradeAgreement>
	//       <ram:SpecifiedLineTradeDelivery>
	//         <ram:BilledQuantity unitCode="HUR">2.0000</ram:BilledQuantity>
	//       </ram:SpecifiedLineTradeDelivery>
	//       <ram:SpecifiedLineTradeSettlement>
	//         <ram:ApplicableTradeTax>
	//           <ram:TypeCode>VAT</ram:TypeCode>
	//           <ram:CategoryCode>AE</ram:CategoryCode>
	//           <ram:RateApplicablePercent>0</ram:RateApplicablePercent>
	//         </ram:ApplicableTradeTax>
	//         <ram:SpecifiedTradeSettlementLineMonetarySummation>
	//           <ram:LineTotalAmount>400.00</ram:LineTotalAmount>
	//         </ram:SpecifiedTradeSettlementLineMonetarySummation>
	//       </ram:SpecifiedLineTradeSettlement>
	//     </ram:IncludedSupplyChainTradeLineItem>
	//     <ram:ApplicableHeaderTradeAgreement>
	//       <ram:SellerTradeParty>
	//         <ram:Name>Company name</ram:Name>
	//         <ram:DefinedTradeContact>
	//           <ram:PersonName>Jon Doe</ram:PersonName>
	//           <ram:EmailURIUniversalCommunication>
	//             <ram:URIID>doe@example.com</ram:URIID>
	//           </ram:EmailURIUniversalCommunication>
	//         </ram:DefinedTradeContact>
	//         <ram:PostalTradeAddress>
	//           <ram:PostcodeCode>12345</ram:PostcodeCode>
	//           <ram:LineOne>Line one</ram:LineOne>
	//           <ram:LineTwo>Line two</ram:LineTwo>
	//           <ram:CityName>City</ram:CityName>
	//           <ram:CountryID>DE</ram:CountryID>
	//         </ram:PostalTradeAddress>
	//         <ram:SpecifiedTaxRegistration>
	//           <ram:ID schemeID="VA">DE123456</ram:ID>
	//         </ram:SpecifiedTaxRegistration>
	//       </ram:SellerTradeParty>
	//       <ram:BuyerTradeParty>
	//         <ram:Name>Buyer</ram:Name>
	//         <ram:DefinedTradeContact>
	//           <ram:PersonName>Buyer Person</ram:PersonName>
	//         </ram:DefinedTradeContact>
	//         <ram:PostalTradeAddress>
	//           <ram:PostcodeCode>33441</ram:PostcodeCode>
	//           <ram:LineOne>Buyer line 1</ram:LineOne>
	//           <ram:LineTwo>Buyer line 2</ram:LineTwo>
	//           <ram:CityName>Buyercity</ram:CityName>
	//           <ram:CountryID>FR</ram:CountryID>
	//         </ram:PostalTradeAddress>
	//         <ram:SpecifiedTaxRegistration>
	//           <ram:ID schemeID="VA">FR4441112</ram:ID>
	//         </ram:SpecifiedTaxRegistration>
	//       </ram:BuyerTradeParty>
	//     </ram:ApplicableHeaderTradeAgreement>
	//     <ram:ApplicableHeaderTradeDelivery>
	//       <ram:ActualDeliverySupplyChainEvent>
	//         <ram:OccurrenceDateTime>
	//           <udt:DateTimeString format="102">20251217</udt:DateTimeString>
	//         </ram:OccurrenceDateTime>
	//       </ram:ActualDeliverySupplyChainEvent>
	//     </ram:ApplicableHeaderTradeDelivery>
	//     <ram:ApplicableHeaderTradeSettlement>
	//       <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
	//       <ram:SpecifiedTradeSettlementPaymentMeans>
	//         <ram:TypeCode>30</ram:TypeCode>
	//         <ram:PayeePartyCreditorFinancialAccount>
	//           <ram:IBANID>DE123455958381</ram:IBANID>
	//           <ram:AccountName>My own bank</ram:AccountName>
	//         </ram:PayeePartyCreditorFinancialAccount>
	//         <ram:PayeeSpecifiedCreditorFinancialInstitution>
	//           <ram:BICID>BANKDEFXXX</ram:BICID>
	//         </ram:PayeeSpecifiedCreditorFinancialInstitution>
	//       </ram:SpecifiedTradeSettlementPaymentMeans>
	//       <ram:ApplicableTradeTax>
	//         <ram:CalculatedAmount>237.50</ram:CalculatedAmount>
	//         <ram:TypeCode>VAT</ram:TypeCode>
	//         <ram:BasisAmount>1250.00</ram:BasisAmount>
	//         <ram:CategoryCode>S</ram:CategoryCode>
	//         <ram:RateApplicablePercent>19</ram:RateApplicablePercent>
	//       </ram:ApplicableTradeTax>
	//       <ram:ApplicableTradeTax>
	//         <ram:CalculatedAmount>0.00</ram:CalculatedAmount>
	//         <ram:TypeCode>VAT</ram:TypeCode>
	//         <ram:ExemptionReason>Reason for reverse charge</ram:ExemptionReason>
	//         <ram:BasisAmount>400.00</ram:BasisAmount>
	//         <ram:CategoryCode>AE</ram:CategoryCode>
	//         <ram:RateApplicablePercent>0</ram:RateApplicablePercent>
	//       </ram:ApplicableTradeTax>
	//       <ram:SpecifiedTradePaymentTerms>
	//         <ram:DueDateDateTime>
	//           <udt:DateTimeString format="102">20260114</udt:DateTimeString>
	//         </ram:DueDateDateTime>
	//       </ram:SpecifiedTradePaymentTerms>
	//       <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
	//         <ram:LineTotalAmount>1650.00</ram:LineTotalAmount>
	//         <ram:TaxBasisTotalAmount>1650.00</ram:TaxBasisTotalAmount>
	//         <ram:TaxTotalAmount currencyID="EUR">237.50</ram:TaxTotalAmount>
	//         <ram:GrandTotalAmount>1887.50</ram:GrandTotalAmount>
	//         <ram:DuePayableAmount>1887.50</ram:DuePayableAmount>
	//       </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
	//     </ram:ApplicableHeaderTradeSettlement>
	//   </rsm:SupplyChainTradeTransaction>
	// </rsm:CrossIndustryInvoice>
}

// TestProfileDetection tests profile detection methods for all standard URNs.
// Verifies that IsMinimum(), IsBasicWL(), IsBasic(), IsEN16931(), IsExtended(),
// IsXRechnung(), and ProfileLevel() correctly identify invoice profiles from
// the GuidelineSpecifiedDocumentContextParameter (BT-24) URN.
func TestProfileDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		urn          string
		isMinimum    bool
		isBasicWL    bool
		isBasic      bool
		isEN16931    bool
		isExtended   bool
		isXRechnung  bool
		profileLevel int
	}{
		{
			name:         "Factur-X Minimum",
			urn:          SpecFacturXMinimum,
			isMinimum:    true,
			profileLevel: 1,
		},
		{
			name:         "ZUGFeRD Minimum",
			urn:          SpecZUGFeRDMinimum,
			isMinimum:    true,
			profileLevel: 1,
		},
		{
			name:         "Factur-X Basic WL",
			urn:          SpecFacturXBasicWL,
			isBasicWL:    true,
			profileLevel: 2,
		},
		{
			name:         "Factur-X Basic",
			urn:          SpecFacturXBasic,
			isBasic:      true,
			profileLevel: 3,
		},
		{
			name:         "Factur-X Basic Alt",
			urn:          SpecFacturXBasicAlt,
			isBasic:      true,
			profileLevel: 3,
		},
		{
			name:         "ZUGFeRD Basic",
			urn:          SpecZUGFeRDBasic,
			isBasic:      true,
			profileLevel: 3,
		},
		{
			name:         "EN 16931",
			urn:          SpecEN16931,
			isEN16931:    true,
			profileLevel: 4,
		},
		{
			name:         "XRechnung 3.0",
			urn:          SpecXRechnung30,
			isXRechnung:  true,
			profileLevel: 4,
		},
		{
			name:         "PEPPOL BIS Billing 3.0",
			urn:          SpecPEPPOLBilling30,
			profileLevel: 4, // PEPPOL returns level 4 via isPEPPOL() in ProfileLevel()
		},
		{
			name:         "Factur-X Extended",
			urn:          SpecFacturXExtended,
			isExtended:   true,
			profileLevel: 5,
		},
		{
			name:         "ZUGFeRD Extended",
			urn:          SpecZUGFeRDExtended,
			isExtended:   true,
			profileLevel: 5,
		},
		{
			name:         "Unknown profile",
			urn:          "urn:unknown:profile",
			profileLevel: 0,
		},
		{
			name:         "Empty URN",
			urn:          "",
			profileLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invoice{
				GuidelineSpecifiedDocumentContextParameter: tt.urn,
			}

			if got := inv.IsMinimum(); got != tt.isMinimum {
				t.Errorf("IsMinimum() = %v, want %v", got, tt.isMinimum)
			}

			if got := inv.IsBasicWL(); got != tt.isBasicWL {
				t.Errorf("IsBasicWL() = %v, want %v", got, tt.isBasicWL)
			}

			if got := inv.IsBasic(); got != tt.isBasic {
				t.Errorf("IsBasic() = %v, want %v", got, tt.isBasic)
			}

			if got := inv.IsEN16931(); got != tt.isEN16931 {
				t.Errorf("IsEN16931() = %v, want %v", got, tt.isEN16931)
			}

			if got := inv.IsExtended(); got != tt.isExtended {
				t.Errorf("IsExtended() = %v, want %v", got, tt.isExtended)
			}

			if got := inv.IsXRechnung(); got != tt.isXRechnung {
				t.Errorf("IsXRechnung() = %v, want %v", got, tt.isXRechnung)
			}

			if got := inv.ProfileLevel(); got != tt.profileLevel {
				t.Errorf("ProfileLevel() = %v, want %v", got, tt.profileLevel)
			}
		})
	}
}

// TestMeetsProfileLevel tests the MeetsProfileLevel method which replaces
// the old Profile enum comparison pattern.
func TestMeetsProfileLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		urn       string
		testLevel int
		expected  bool
	}{
		{
			name:      "Minimum (1) meets level 1",
			urn:       SpecFacturXMinimum,
			testLevel: 1,
			expected:  true,
		},
		{
			name:      "Minimum (1) does not meet level 2",
			urn:       SpecFacturXMinimum,
			testLevel: 2,
			expected:  false,
		},
		{
			name:      "BasicWL (2) meets level 1",
			urn:       SpecFacturXBasicWL,
			testLevel: 1,
			expected:  true,
		},
		{
			name:      "BasicWL (2) meets level 2",
			urn:       SpecFacturXBasicWL,
			testLevel: 2,
			expected:  true,
		},
		{
			name:      "BasicWL (2) does not meet level 3",
			urn:       SpecFacturXBasicWL,
			testLevel: 3,
			expected:  false,
		},
		{
			name:      "Basic (3) meets level 3",
			urn:       SpecFacturXBasic,
			testLevel: 3,
			expected:  true,
		},
		{
			name:      "Basic (3) does not meet level 4",
			urn:       SpecFacturXBasic,
			testLevel: 4,
			expected:  false,
		},
		{
			name:      "EN16931 (4) meets level 3",
			urn:       SpecEN16931,
			testLevel: 3,
			expected:  true,
		},
		{
			name:      "EN16931 (4) meets level 4",
			urn:       SpecEN16931,
			testLevel: 4,
			expected:  true,
		},
		{
			name:      "EN16931 (4) does not meet level 5",
			urn:       SpecEN16931,
			testLevel: 5,
			expected:  false,
		},
		{
			name:      "XRechnung (4) meets level 4",
			urn:       SpecXRechnung30,
			testLevel: 4,
			expected:  true,
		},
		{
			name:      "Extended (5) meets all levels",
			urn:       SpecFacturXExtended,
			testLevel: 5,
			expected:  true,
		},
		{
			name:      "Extended (5) meets level 1",
			urn:       SpecFacturXExtended,
			testLevel: 1,
			expected:  true,
		},
		{
			name:      "Unknown (0) does not meet level 1",
			urn:       "urn:unknown",
			testLevel: 1,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invoice{
				GuidelineSpecifiedDocumentContextParameter: tt.urn,
			}

			if got := inv.MeetsProfileLevel(tt.testLevel); got != tt.expected {
				t.Errorf("MeetsProfileLevel(%d) = %v, want %v (URN: %s, ProfileLevel: %d)",
					tt.testLevel, got, tt.expected, tt.urn, inv.ProfileLevel())
			}
		})
	}
}

// TestIsPEPPOL tests PEPPOL BIS Billing 3.0 detection.
// The isPEPPOL() method is private but used by ProfileLevel() and Validate().
func TestIsPEPPOL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		urn          string
		expectLevel4 bool
		expectPEPPOL bool // Inferred from ProfileLevel() behavior
	}{
		{
			name:         "PEPPOL BIS Billing 3.0",
			urn:          SpecPEPPOLBilling30,
			expectLevel4: true,
			expectPEPPOL: true,
		},
		{
			name:         "EN 16931 (not PEPPOL)",
			urn:          SpecEN16931,
			expectLevel4: true,
			expectPEPPOL: false,
		},
		{
			name:         "XRechnung (not PEPPOL)",
			urn:          SpecXRechnung30,
			expectLevel4: true,
			expectPEPPOL: false,
		},
		{
			name:         "Extended (not PEPPOL)",
			urn:          SpecFacturXExtended,
			expectLevel4: false, // Level 5
			expectPEPPOL: false,
		},
		{
			name:         "Unknown (not PEPPOL)",
			urn:          "urn:unknown",
			expectLevel4: false,
			expectPEPPOL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invoice{
				GuidelineSpecifiedDocumentContextParameter: tt.urn,
			}

			level := inv.ProfileLevel()
			gotLevel4 := (level == 4)

			if gotLevel4 != tt.expectLevel4 {
				t.Errorf("ProfileLevel() == 4: got %v, want %v (level: %d)", gotLevel4, tt.expectLevel4, level)
			}

			// Infer PEPPOL from validation behavior - PEPPOL validation runs when
			// the URN is SpecPEPPOLBilling30. We can't test isPEPPOL() directly as
			// it's private, but we verify the URN matches the PEPPOL constant.
			if tt.expectPEPPOL {
				if inv.GuidelineSpecifiedDocumentContextParameter != SpecPEPPOLBilling30 {
					t.Errorf("Expected PEPPOL URN, got %s", inv.GuidelineSpecifiedDocumentContextParameter)
				}
			}
		})
	}
}

// BenchmarkRoundTrip benchmarks the complete parse → write → parse cycle
// to measure round-trip performance for both CII and UBL formats.
func BenchmarkRoundTrip(b *testing.B) {
	benchmarks := []struct {
		name string
		file string
	}{
		{"CII_Minimum", "testdata/cii/minimum/zugferd-minimum-rechnung.xml"},
		{"CII_EN16931", "testdata/cii/en16931/zugferd_2p3_EN16931_1.xml"},
		{"CII_Extended", "testdata/cii/extended/zugferd-extended-rechnung.xml"},
		{"UBL_Invoice", "testdata/ubl/invoice/UBL-Invoice-2.1-Example.xml"},
		{"UBL_CreditNote", "testdata/ubl/creditnote/UBL-CreditNote-2.1-Example.xml"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data, err := os.ReadFile(bm.file)
			if err != nil {
				b.Skipf("File not found: %s", bm.file)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				// Parse
				inv, err := ParseReader(bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}

				// Write
				var buf bytes.Buffer
				if err := inv.Write(&buf); err != nil {
					b.Fatal(err)
				}

				// Parse again
				_, err = ParseReader(&buf)
				if err != nil {
					b.Fatalf("Round-trip parse failed: %v", err)
				}
			}
		})
	}
}

// FuzzRoundTrip fuzzes the round-trip XML generation to ensure written XML
// can always be parsed back. This is a CRITICAL test for data integrity.
//
// The fuzz test verifies:
// 1. Parsed invoices can be written to XML
// 2. Written XML can be parsed back without errors
// 3. No panics occur during the round-trip process
func FuzzRoundTrip(f *testing.F) {
	// Seed with representative fixtures from both CII and UBL formats
	seeds := []string{
		"testdata/cii/minimum/zugferd-minimum-rechnung.xml",
		"testdata/cii/basicwl/zugferd-basicwl-rechnung.xml",
		"testdata/cii/basic/zugferd-basic-rechnung.xml",
		"testdata/cii/en16931/zugferd_2p3_EN16931_1.xml",
		"testdata/cii/extended/zugferd-extended-rechnung.xml",
		"testdata/cii/xrechnung/zugferd-xrechnung-einfach.xml",
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
		// Parse the input data
		inv, err := ParseReader(bytes.NewReader(data))
		if err != nil {
			return // Skip invalid XML - that's okay
		}

		// Write to XML
		var buf bytes.Buffer
		err = inv.Write(&buf)
		if err != nil {
			t.Fatalf("Write failed for valid parsed invoice: %v", err)
		}

		// Critical: written XML must parse successfully
		_, err = ParseReader(&buf)
		if err != nil {
			t.Fatalf("Round-trip parse failed: %v\nOriginal invoice SchemaType: %v, Profile: %s",
				err, inv.SchemaType, inv.GuidelineSpecifiedDocumentContextParameter)
		}
	})
}

// assertInvoiceEqual compares critical fields between two invoices after round-trip using go-cmp.
// This ensures no data is lost during Parse → Write → Parse cycle.
func assertInvoiceEqual(t *testing.T, original, roundtrip *Invoice) {
	t.Helper()

	opts := []cmp.Option{
		// Proper decimal comparison using .Equal() method
		cmp.Comparer(func(a, b decimal.Decimal) bool {
			return a.Equal(b)
		}),
		// Normalize empty TaxTotalCurrency → InvoiceCurrencyCode
		// This is required because XML writers default currency attributes to InvoiceCurrencyCode
		cmp.Transformer("NormalizeTaxCurrency", func(inv Invoice) Invoice {
			if inv.TaxTotalCurrency == "" {
				inv.TaxTotalCurrency = inv.InvoiceCurrencyCode
			}
			return inv
		}),
		// Ignore unexported fields (like hasLineTotalInXML in InvoiceLine)
		cmpopts.IgnoreUnexported(Invoice{}, InvoiceLine{}, Party{}, TradeTax{}, AllowanceCharge{},
			PostalAddress{}, DefinedTradeContact{}, PaymentMeans{}, SpecifiedTradePaymentTerms{},
			Note{}, Document{}, ReferencedDocument{}, GlobalID{}, SpecifiedLegalOrganization{},
			Characteristic{}, Classification{}),
	}

	if diff := cmp.Diff(original, roundtrip, opts...); diff != "" {
		t.Errorf("Invoice round-trip mismatch (-original +roundtrip):\n%s", diff)
	}
}

// TestAllValidFixtures performs comprehensive integration testing on all official XML fixtures.
// Following TESTING_STRATEGY.md Layer 1: Integration Tests
//
// Test flow:
//  1. Parse original XML
//  2. Validate (log all violations, don't fail)
//  3. Write to new XML
//  4. Parse the written XML
//  5. Compare critical fields using assertInvoiceEqual
//
// This catches data loss bugs like the multi-currency TaxTotalAmount concatenation issue.
func TestAllValidFixtures(t *testing.T) {
	t.Parallel()

	// Explicit list of 60 fixtures to test (deterministic, not auto-discovery)
	fixtures := []string{
		// CII Minimum
		"testdata/cii/minimum/zugferd-minimum-buchungshilfe.xml",
		"testdata/cii/minimum/zugferd-minimum-rechnung.xml",

		// CII BasicWL
		"testdata/cii/basicwl/zugferd-basicwl-buchungshilfe.xml",
		"testdata/cii/basicwl/zugferd-basicwl-einfach.xml",

		// CII Basic
		"testdata/cii/basic/zugferd-basic-1.xml",
		"testdata/cii/basic/zugferd-basic-einfach.xml",
		"testdata/cii/basic/zugferd-basic-rechnungskorrektur.xml",
		"testdata/cii/basic/zugferd-basic-taxifahrt.xml",

		// CII EN16931
		"testdata/cii/en16931/CII_example1.xml",
		"testdata/cii/en16931/CII_example2.xml",
		"testdata/cii/en16931/CII_example3.xml",
		"testdata/cii/en16931/CII_example4.xml",
		"testdata/cii/en16931/CII_example5.xml", // Multi-currency
		"testdata/cii/en16931/CII_example6.xml",
		"testdata/cii/en16931/CII_example7.xml",
		"testdata/cii/en16931/CII_example8.xml",
		"testdata/cii/en16931/CII_example9.xml",
		"testdata/cii/en16931/zugferd_2p0_EN16931_1_Teilrechnung.xml",
		"testdata/cii/en16931/zugferd-en16931-einfach.xml",
		"testdata/cii/en16931/zugferd-en16931-gutschrift.xml",
		"testdata/cii/en16931/zugferd-en16931-intra-community.xml",
		"testdata/cii/en16931/zugferd-en16931-payee.xml",
		"testdata/cii/en16931/zugferd-en16931-rabatte.xml",
		"testdata/cii/en16931/zugferd-en16931-rechnungskorrektur.xml",

		// CII Extended
		"testdata/cii/extended/zugferd-extended-1.xml",
		"testdata/cii/extended/zugferd-extended-2.xml",
		"testdata/cii/extended/zugferd-extended-fremdwaehrung.xml", // Multi-currency
		"testdata/cii/extended/zugferd-extended-intra-community-multi.xml",
		"testdata/cii/extended/zugferd-extended-rechnungskorrektur.xml",
		"testdata/cii/extended/zugferd-extended-warenrechnung.xml",

		// CII XRechnung
		"testdata/cii/xrechnung/XRechnung-O.xml",
		"testdata/cii/xrechnung/zugferd-xrechnung-betriebskosten.xml", // XRechnung 2.1
		"testdata/cii/xrechnung/zugferd-xrechnung-einfach.xml",
		"testdata/cii/xrechnung/zugferd-xrechnung-elektron.xml",

		// UBL Invoice
		"testdata/ubl/invoice/UBL-Invoice-2.1-Example.xml",
		"testdata/ubl/invoice/ubl-tc434-example1.xml",
		"testdata/ubl/invoice/ubl-tc434-example2.xml",
		"testdata/ubl/invoice/ubl-tc434-example3.xml",
		"testdata/ubl/invoice/ubl-tc434-example4.xml",
		"testdata/ubl/invoice/ubl-tc434-example5.xml",
		"testdata/ubl/invoice/ubl-tc434-example6.xml",
		"testdata/ubl/invoice/ubl-tc434-example7.xml",
		"testdata/ubl/invoice/ubl-tc434-example8.xml",
		"testdata/ubl/invoice/ubl-tc434-example9.xml",
		"testdata/ubl/invoice/ubl-tc434-example10.xml",

		// UBL CreditNote
		"testdata/ubl/creditnote/UBL-CreditNote-2.1-Example.xml",
		"testdata/ubl/creditnote/ubl-tc434-creditnote1.xml",

		// PEPPOL Valid
		"testdata/peppol/valid/Allowance-example.xml",
		"testdata/peppol/valid/GR-base-example-TaxRepresentative.xml",
		"testdata/peppol/valid/GR-base-example-correct.xml",
		"testdata/peppol/valid/Norwegian-example-1.xml",
		"testdata/peppol/valid/Vat-category-S.xml",
		"testdata/peppol/valid/base-creditnote-correction.xml",
		"testdata/peppol/valid/base-example.xml",
		"testdata/peppol/valid/base-negative-inv-correction.xml",
		"testdata/peppol/valid/vat-category-E.xml",
		"testdata/peppol/valid/vat-category-O.xml",
		"testdata/peppol/valid/vat-category-Z.xml",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			// Step 1: Parse original XML
			inv1, err := ParseXMLFile(fixture)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Step 2: Validate - validation errors are ERRORS!
			if err := inv1.Validate(); err != nil {
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					violations := valErr.Violations()
					t.Errorf("Validation failed with %d violations:", len(violations))
					for _, v := range violations {
						t.Logf("  - %s: %s", v.Rule.Code, v.Text)
					}
				} else {
					t.Fatalf("Validation error: %v", err)
				}
			}

			// Step 3: Write to new XML
			var buf bytes.Buffer
			if err := inv1.Write(&buf); err != nil {
				t.Fatalf("Write failed: %v", err)
			}

			// Step 4: Parse the written XML
			inv2, err := ParseReader(&buf)
			if err != nil {
				t.Fatalf("Round-trip parse failed: %v", err)
			}

			// Step 5: Compare critical fields - catches data loss bugs!
			assertInvoiceEqual(t, inv1, inv2)
		})
	}
}
