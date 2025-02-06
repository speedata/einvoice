package einvoice_test

import (
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice"
)

func ExampleInvoice_Write() {
	fixedDate, _ := time.Parse("02.01.2006", "31.12.2025")
	fourteenDays := time.Hour * 24 * 14
	inv := einvoice.Invoice{
		InvoiceNumber:       "1234",
		InvoiceTypeCode:     380,
		Profile:             einvoice.CProfileEN16931,
		InvoiceDate:         fixedDate,
		OccurrenceDateTime:  fixedDate.Add(-fourteenDays),
		InvoiceCurrencyCode: "EUR",
		TaxCurrencyCode:     "EUR",
		Notes: []einvoice.Note{{
			Text: "Some text",
		}},
		Seller: einvoice.Party{
			Name:              "Company name",
			VATaxRegistration: "DE123456",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Line one",
				Line2:        "Line two",
				City:         "City",
				PostcodeCode: "12345",
				CountryID:    "DE",
			},
			DefinedTradeContact: []einvoice.DefinedTradeContact{{
				PersonName: "Jon Doe",
				EMail:      "doe@example.com",
			}},
		},
		Buyer: einvoice.Party{
			Name: "Buyer",
			PostalAddress: &einvoice.PostalAddress{
				Line1:        "Buyer line 1",
				Line2:        "Buyer line 2",
				City:         "Buyercity",
				PostcodeCode: "33441",
				CountryID:    "FR",
			},
			DefinedTradeContact: []einvoice.DefinedTradeContact{{
				PersonName: "Buyer Person",
			}},
			VATaxRegistration: "FR4441112",
		},
		PaymentMeans: []einvoice.PaymentMeans{
			{
				TypeCode:                                      30,
				PayeePartyCreditorFinancialAccountIBAN:        "DE123455958381",
				PayeePartyCreditorFinancialAccountName:        "My own bank",
				PayeeSpecifiedCreditorFinancialInstitutionBIC: "BANKDEFXXX",
			},
		},
		SpecifiedTradePaymentTerms: []einvoice.SpecifiedTradePaymentTerms{{
			DueDate: fixedDate.Add(fourteenDays),
		}},
		InvoiceLines: []einvoice.InvoiceLine{
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
				LineID:                   "1",
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
	//           <ram:ChargeAmount>100.00</ram:ChargeAmount>
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
	//         <ram:LineID>1</ram:LineID>
	//       </ram:AssociatedDocumentLineDocument>
	//       <ram:SpecifiedTradeProduct>
	//         <ram:Name>Item name two</ram:Name>
	//       </ram:SpecifiedTradeProduct>
	//       <ram:SpecifiedLineTradeAgreement>
	//         <ram:NetPriceProductTradePrice>
	//           <ram:ChargeAmount>200.00</ram:ChargeAmount>
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
	//         <ram:ChargeTotalAmount>0.00</ram:ChargeTotalAmount>
	//         <ram:AllowanceTotalAmount>0.00</ram:AllowanceTotalAmount>
	//         <ram:TaxBasisTotalAmount>1650.00</ram:TaxBasisTotalAmount>
	//         <ram:TaxTotalAmount currencyID="EUR">237.50</ram:TaxTotalAmount>
	//         <ram:GrandTotalAmount>1887.50</ram:GrandTotalAmount>
	//         <ram:TotalPrepaidAmount>0.00</ram:TotalPrepaidAmount>
	//         <ram:DuePayableAmount>1887.50</ram:DuePayableAmount>
	//       </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
	//     </ram:ApplicableHeaderTradeSettlement>
	//   </rsm:SupplyChainTradeTransaction>
	// </rsm:CrossIndustryInvoice>
}
