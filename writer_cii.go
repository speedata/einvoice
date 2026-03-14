package einvoice

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/beevik/etree"
)

// addTimeCIIQDT adds a date in CII QDT format (YYYYMMDD) to the parent element.
func addTimeCIIQDT(parent *etree.Element, date time.Time) {
	udtdts := parent.CreateElement("qdt:DateTimeString")
	udtdts.CreateAttr("format", "102")
	udtdts.CreateText(date.Format("20060102"))
}

// addTimeCIIUDT adds a date in CII UDT format (YYYYMMDD) to the parent element.
func addTimeCIIUDT(parent *etree.Element, date time.Time) {
	udtdts := parent.CreateElement("udt:DateTimeString")
	udtdts.CreateAttr("format", "102")
	udtdts.CreateText(date.Format("20060102"))
}

func writeCIIramIncludedSupplyChainTradeLineItem(invoiceLine *InvoiceLine, inv *Invoice, parent *etree.Element) {
	lineItem := parent.CreateElement("ram:IncludedSupplyChainTradeLineItem")
	adld := lineItem.CreateElement("ram:AssociatedDocumentLineDocument")
	lineID := adld.CreateElement("ram:LineID")
	lineID.SetText(invoiceLine.LineID)

	if invoiceLine.Note != "" {
		adld.CreateElement("ram:IncludedNote").CreateElement("ram:Content").SetText(invoiceLine.Note)
	}

	stp := lineItem.CreateElement("ram:SpecifiedTradeProduct")

	if invoiceLine.GlobalID != "" { // BT-157
		gid := stp.CreateElement("ram:GlobalID")
		gid.CreateAttr("schemeID", invoiceLine.GlobalIDType)
		gid.SetText(invoiceLine.GlobalID)
	}
	// BT-155 is optional in EN
	if is(levelEN16931, inv) {
		if artno := invoiceLine.ArticleNumber; artno != "" {
			stp.CreateElement("ram:SellerAssignedID").SetText(artno)
		}
	}

	// BT-156 optional in EN
	if is(levelEN16931, inv) {
		if artno := invoiceLine.ArticleNumberBuyer; artno != "" {
			stp.CreateElement("ram:BuyerAssignedID").SetText(invoiceLine.ArticleNumberBuyer)
		}
	}

	stp.CreateElement("ram:Name").SetText(invoiceLine.ItemName)

	if invoiceLine.Description != "" {
		stp.CreateElement("ram:Description").SetText(invoiceLine.Description)
	}

	// Write product characteristics
	for i := range invoiceLine.Characteristics {
		chElt := stp.CreateElement("ram:ApplicableProductCharacteristic")
		chElt.CreateElement("ram:Description").SetText(invoiceLine.Characteristics[i].Description)
		chElt.CreateElement("ram:Value").SetText(invoiceLine.Characteristics[i].Value)
	}

	// Write product classifications
	for i := range invoiceLine.ProductClassification {
		clElt := stp.CreateElement("ram:DesignatedProductClassification")
		if invoiceLine.ProductClassification[i].ClassCode != "" {
			ccElt := clElt.CreateElement("ram:ClassCode")
			if invoiceLine.ProductClassification[i].ListID != "" {
				ccElt.CreateAttr("listID", invoiceLine.ProductClassification[i].ListID)
			}
			if invoiceLine.ProductClassification[i].ListVersionID != "" {
				ccElt.CreateAttr("listVersionID", invoiceLine.ProductClassification[i].ListVersionID)
			}
			ccElt.SetText(invoiceLine.ProductClassification[i].ClassCode)
		}
	}

	// BT-159: Item country of origin
	if invoiceLine.OriginTradeCountry != "" {
		otc := stp.CreateElement("ram:OriginTradeCountry")
		otc.CreateElement("ram:ID").SetText(invoiceLine.OriginTradeCountry)
	}

	slta := lineItem.CreateElement("ram:SpecifiedLineTradeAgreement")

	// BT-132: Referenced purchase order line reference
	if invoiceLine.BuyerOrderReferencedDocument != "" {
		bord := slta.CreateElement("ram:BuyerOrderReferencedDocument")
		bord.CreateElement("ram:LineID").SetText(invoiceLine.BuyerOrderReferencedDocument)
	}

	// BT-148: Gross price (no decimal restriction per EN 16931)
	if !invoiceLine.GrossPrice.IsZero() {
		gpptp := slta.CreateElement("ram:GrossPriceProductTradePrice")
		gpptp.CreateElement("ram:ChargeAmount").SetText(invoiceLine.GrossPrice.String())

		for i := range invoiceLine.AppliedTradeAllowanceCharge {
			acElt := gpptp.CreateElement("ram:AppliedTradeAllowanceCharge")
			// BG-27, BG-28
			acElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText(strconv.FormatBool(invoiceLine.AppliedTradeAllowanceCharge[i].ChargeIndicator))

			if cp := invoiceLine.AppliedTradeAllowanceCharge[i].CalculationPercent; !cp.IsZero() {
				acElt.CreateElement("ram:CalculationPercent").SetText(formatPercent(cp))
			}

			if ba := invoiceLine.AppliedTradeAllowanceCharge[i].BasisAmount; !ba.IsZero() {
				acElt.CreateElement("ram:BasisAmount").SetText(ba.StringFixed(2))
			}

			// BT-147: Item price discount (no decimal restriction per EN 16931)
			acElt.CreateElement("ram:ActualAmount").SetText(invoiceLine.AppliedTradeAllowanceCharge[i].ActualAmount.String())
			if r := invoiceLine.AppliedTradeAllowanceCharge[i].Reason; r != "" {
				acElt.CreateElement("ram:Reason").SetText(r)
			}
		}
	}

	// BT-146: Net price (no decimal restriction per EN 16931)
	npptp := slta.CreateElement("ram:NetPriceProductTradePrice")
	npptp.CreateElement("ram:ChargeAmount").SetText(invoiceLine.NetPrice.String())

	// BT-149: Item price base quantity with unit code (goes with NetPrice)
	if !invoiceLine.BasisQuantity.IsZero() {
		bq := npptp.CreateElement("ram:BasisQuantity")
		if invoiceLine.BasisQuantityUnit != "" {
			bq.CreateAttr("unitCode", invoiceLine.BasisQuantityUnit)
		}
		// BT-149: Item price base quantity (no decimal restriction per EN 16931)
		bq.SetText(invoiceLine.BasisQuantity.String())
	}
	bq := lineItem.CreateElement("ram:SpecifiedLineTradeDelivery").CreateElement("ram:BilledQuantity")
	bq.CreateAttr("unitCode", invoiceLine.BilledQuantityUnit)
	bq.SetText(invoiceLine.BilledQuantity.StringFixed(4))

	slts := lineItem.CreateElement("ram:SpecifiedLineTradeSettlement")

	// ApplicableTradeTax must come first per CII XSD sequence
	att := slts.CreateElement("ram:ApplicableTradeTax")
	// BT-151 must be VAT
	if invoiceLine.TaxTypeCode == "" {
		invoiceLine.TaxTypeCode = "VAT"
	}

	att.CreateElement("ram:TypeCode").SetText(invoiceLine.TaxTypeCode)
	att.CreateElement("ram:CategoryCode").SetText(invoiceLine.TaxCategoryCode)
	att.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(invoiceLine.TaxRateApplicablePercent))

	// BG-26: Invoice line period (BT-134, BT-135)
	if !invoiceLine.BillingSpecifiedPeriodStart.IsZero() || !invoiceLine.BillingSpecifiedPeriodEnd.IsZero() {
		bsp := slts.CreateElement("ram:BillingSpecifiedPeriod")
		if !invoiceLine.BillingSpecifiedPeriodStart.IsZero() {
			addTimeCIIUDT(bsp.CreateElement("ram:StartDateTime"), invoiceLine.BillingSpecifiedPeriodStart)
		}
		if !invoiceLine.BillingSpecifiedPeriodEnd.IsZero() {
			addTimeCIIUDT(bsp.CreateElement("ram:EndDateTime"), invoiceLine.BillingSpecifiedPeriodEnd)
		}
	}

	// BG-27: Invoice line allowances
	for i := range invoiceLine.InvoiceLineAllowances {
		acElt := slts.CreateElement("ram:SpecifiedTradeAllowanceCharge")
		acElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText("false")

		if !invoiceLine.InvoiceLineAllowances[i].CalculationPercent.IsZero() {
			acElt.CreateElement("ram:CalculationPercent").SetText(formatPercent(invoiceLine.InvoiceLineAllowances[i].CalculationPercent))
		}

		if !invoiceLine.InvoiceLineAllowances[i].BasisAmount.IsZero() {
			acElt.CreateElement("ram:BasisAmount").SetText(invoiceLine.InvoiceLineAllowances[i].BasisAmount.StringFixed(2))
		}

		acElt.CreateElement("ram:ActualAmount").SetText(invoiceLine.InvoiceLineAllowances[i].ActualAmount.StringFixed(2))

		if invoiceLine.InvoiceLineAllowances[i].ReasonCode != "" {
			acElt.CreateElement("ram:ReasonCode").SetText(invoiceLine.InvoiceLineAllowances[i].ReasonCode)
		}

		if invoiceLine.InvoiceLineAllowances[i].Reason != "" {
			acElt.CreateElement("ram:Reason").SetText(invoiceLine.InvoiceLineAllowances[i].Reason)
		}

		// Category trade tax for allowance
		if invoiceLine.InvoiceLineAllowances[i].CategoryTradeTaxCategoryCode != "" {
			ctt := acElt.CreateElement("ram:CategoryTradeTax")
			if invoiceLine.InvoiceLineAllowances[i].CategoryTradeTaxType != "" {
				ctt.CreateElement("ram:TypeCode").SetText(invoiceLine.InvoiceLineAllowances[i].CategoryTradeTaxType)
			}
			ctt.CreateElement("ram:CategoryCode").SetText(invoiceLine.InvoiceLineAllowances[i].CategoryTradeTaxCategoryCode)
			ctt.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(invoiceLine.InvoiceLineAllowances[i].CategoryTradeTaxRateApplicablePercent))
		}
	}

	// BG-28: Invoice line charges
	for i := range invoiceLine.InvoiceLineCharges {
		acElt := slts.CreateElement("ram:SpecifiedTradeAllowanceCharge")
		acElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText("true")

		if !invoiceLine.InvoiceLineCharges[i].CalculationPercent.IsZero() {
			acElt.CreateElement("ram:CalculationPercent").SetText(formatPercent(invoiceLine.InvoiceLineCharges[i].CalculationPercent))
		}

		if !invoiceLine.InvoiceLineCharges[i].BasisAmount.IsZero() {
			acElt.CreateElement("ram:BasisAmount").SetText(invoiceLine.InvoiceLineCharges[i].BasisAmount.StringFixed(2))
		}

		acElt.CreateElement("ram:ActualAmount").SetText(invoiceLine.InvoiceLineCharges[i].ActualAmount.StringFixed(2))

		if invoiceLine.InvoiceLineCharges[i].ReasonCode != "" {
			acElt.CreateElement("ram:ReasonCode").SetText(invoiceLine.InvoiceLineCharges[i].ReasonCode)
		}

		if invoiceLine.InvoiceLineCharges[i].Reason != "" {
			acElt.CreateElement("ram:Reason").SetText(invoiceLine.InvoiceLineCharges[i].Reason)
		}

		// Category trade tax for charge
		if invoiceLine.InvoiceLineCharges[i].CategoryTradeTaxCategoryCode != "" {
			ctt := acElt.CreateElement("ram:CategoryTradeTax")
			if invoiceLine.InvoiceLineCharges[i].CategoryTradeTaxType != "" {
				ctt.CreateElement("ram:TypeCode").SetText(invoiceLine.InvoiceLineCharges[i].CategoryTradeTaxType)
			}
			ctt.CreateElement("ram:CategoryCode").SetText(invoiceLine.InvoiceLineCharges[i].CategoryTradeTaxCategoryCode)
			ctt.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(invoiceLine.InvoiceLineCharges[i].CategoryTradeTaxRateApplicablePercent))
		}
	}

	slts.CreateElement("ram:SpecifiedTradeSettlementLineMonetarySummation").CreateElement("ram:LineTotalAmount").SetText(invoiceLine.Total.StringFixed(2))

	// BT-128: Referenced document at line level
	if invoiceLine.AdditionalReferencedDocumentID != "" {
		addDoc := slts.CreateElement("ram:AdditionalReferencedDocument")
		addDoc.CreateElement("ram:IssuerAssignedID").SetText(invoiceLine.AdditionalReferencedDocumentID)
	}
}

func writeCIIParty(inv *Invoice, party *Party, parent *etree.Element, partyType CodePartyType) {
	for _, id := range party.ID {
		parent.CreateElement("ram:ID").SetText(id)
	}
	for _, gid := range party.GlobalID {
		eGID := parent.CreateElement("ram:GlobalID")
		eGID.CreateAttr("schemeID", gid.Scheme)
		eGID.SetText(gid.ID)
	}

	if n := party.Name; n != "" {
		parent.CreateElement("ram:Name").SetText(n)
	}

	// BT-33: Seller additional legal information (or buyer/payee/etc. description)
	if desc := party.Description; desc != "" {
		parent.CreateElement("ram:Description").SetText(desc)
	}

	if slo := party.SpecifiedLegalOrganization; slo != nil {
		sloElt := parent.CreateElement("ram:SpecifiedLegalOrganization")
		id := sloElt.CreateElement("ram:ID")
		id.CreateAttr("schemeID", slo.Scheme)
		id.SetText(slo.ID)
		if slo.TradingBusinessName != "" {
			sloElt.CreateElement("ram:TradingBusinessName").SetText(slo.TradingBusinessName)
		}
	}

	for i := range party.DefinedTradeContact {
		dtcElt := parent.CreateElement("ram:DefinedTradeContact")
		if party.DefinedTradeContact[i].PersonName != "" {
			dtcElt.CreateElement("ram:PersonName").SetText(party.DefinedTradeContact[i].PersonName)
		}

		if party.DefinedTradeContact[i].DepartmentName != "" {
			dtcElt.CreateElement("ram:DepartmentName").SetText(party.DefinedTradeContact[i].DepartmentName)
		}

		if party.DefinedTradeContact[i].PhoneNumber != "" {
			dtcElt.CreateElement("ram:TelephoneUniversalCommunication").CreateElement("ram:CompleteNumber").SetText(party.DefinedTradeContact[i].PhoneNumber)
		}

		if party.DefinedTradeContact[i].EMail != "" {
			email := dtcElt.CreateElement("ram:EmailURIUniversalCommunication").CreateElement("ram:URIID")
			// email.CreateAttr("schemeID", "SMTP")
			email.SetText(party.DefinedTradeContact[i].EMail)
		}
	}

	if ppa := party.PostalAddress; ppa != nil {
		// profile minimum has no postal address for the buyer (BG-8), but BasicWL and above do
		if partyType == CSellerParty || is(levelBasicWL, inv) {
			postalAddress := parent.CreateElement("ram:PostalTradeAddress")

			// BT-38, BT-53: Postcode is optional - only create if non-empty (PEPPOL-EN16931-R008)
			if ppa.PostcodeCode != "" {
				postalAddress.CreateElement("ram:PostcodeCode").SetText(ppa.PostcodeCode)
			}

			if l1 := ppa.Line1; l1 != "" {
				postalAddress.CreateElement("ram:LineOne").SetText(l1)
			}
			if l2 := ppa.Line2; l2 != "" {
				postalAddress.CreateElement("ram:LineTwo").SetText(l2)
			}
			if l3 := ppa.Line3; l3 != "" {
				postalAddress.CreateElement("ram:LineThree").SetText(l3)
			}
			if cityName := ppa.City; cityName != "" {
				postalAddress.CreateElement("ram:CityName").SetText(cityName)
			}
			if cid := ppa.CountryID; cid != "" {
				postalAddress.CreateElement("ram:CountryID").SetText(cid)
			}
			if csd := ppa.CountrySubDivisionName; csd != "" {
				postalAddress.CreateElement("ram:CountrySubDivisionName").SetText(csd)
			}
		}
	}

	// BT-34, BT-49: Electronic address (URI Universal Communication)
	// Write element if either value or scheme is present
	if party.URIUniversalCommunication != "" || party.URIUniversalCommunicationScheme != "" {
		uuc := parent.CreateElement("ram:URIUniversalCommunication")
		uriID := uuc.CreateElement("ram:URIID")
		if party.URIUniversalCommunicationScheme != "" {
			uriID.CreateAttr("schemeID", party.URIUniversalCommunicationScheme)
		}
		uriID.SetText(party.URIUniversalCommunication)
	}

	if fc := party.FCTaxRegistration; fc != "" {
		elt := parent.CreateElement("ram:SpecifiedTaxRegistration").CreateElement("ram:ID")
		elt.CreateAttr("schemeID", "FC")
		elt.SetText(fc)
	}
	if va := party.VATaxRegistration; va != "" {
		elt := parent.CreateElement("ram:SpecifiedTaxRegistration").CreateElement("ram:ID")
		elt.CreateAttr("schemeID", "VA")
		elt.SetText(va)
	}
}

func writeCIIramApplicableHeaderTradeAgreement(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeAgreement")
	// BT-10 optional
	if br := inv.BuyerReference; br != "" {
		elt.CreateElement("ram:BuyerReference").SetText(br)
	}

	writeCIIParty(inv, &inv.Seller, elt.CreateElement("ram:SellerTradeParty"), CSellerParty)
	writeCIIParty(inv, &inv.Buyer, elt.CreateElement("ram:BuyerTradeParty"), CBuyerParty)

	// BG-11: Seller tax representative party
	if inv.SellerTaxRepresentativeTradeParty != nil {
		writeCIIParty(inv, inv.SellerTaxRepresentativeTradeParty, elt.CreateElement("ram:SellerTaxRepresentativeTradeParty"), CSellerParty)
	}

	// BT-14: Seller order reference
	if inv.SellerOrderReferencedDocument != "" {
		elt.CreateElement("ram:SellerOrderReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.SellerOrderReferencedDocument)
	}

	// BT-13
	if inv.BuyerOrderReferencedDocument != "" {
		elt.CreateElement("ram:BuyerOrderReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.BuyerOrderReferencedDocument)
	}
	// BT-12
	if inv.ContractReferencedDocument != "" {
		elt.CreateElement("ram:ContractReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.ContractReferencedDocument)
	}
	// BG-24
	for i := range inv.AdditionalReferencedDocument {
		ard := elt.CreateElement("ram:AdditionalReferencedDocument")
		ard.CreateElement("ram:IssuerAssignedID").SetText(inv.AdditionalReferencedDocument[i].IssuerAssignedID)
		ard.CreateElement("ram:TypeCode").SetText(inv.AdditionalReferencedDocument[i].TypeCode)
		if inv.AdditionalReferencedDocument[i].Name != "" {
			ard.CreateElement("ram:Name").SetText(inv.AdditionalReferencedDocument[i].Name)
		}
		// BT-125: Only write AttachmentBinaryObject if attachment data exists (PEPPOL-EN16931-R008)
		if len(inv.AdditionalReferencedDocument[i].AttachmentBinaryObject) > 0 {
			abo := ard.CreateElement("ram:AttachmentBinaryObject")
			if inv.AdditionalReferencedDocument[i].AttachmentMimeCode != "" {
				abo.CreateAttr("mimeCode", inv.AdditionalReferencedDocument[i].AttachmentMimeCode)
			}
			if inv.AdditionalReferencedDocument[i].AttachmentFilename != "" {
				abo.CreateAttr("filename", inv.AdditionalReferencedDocument[i].AttachmentFilename)
			}
			abo.SetText(base64.StdEncoding.EncodeToString(inv.AdditionalReferencedDocument[i].AttachmentBinaryObject))
		}
		if inv.AdditionalReferencedDocument[i].TypeCode == "130" {
			ard.CreateElement("ram:ReferenceTypeCode").SetText(inv.AdditionalReferencedDocument[i].ReferenceTypeCode)
		}
	}

	// BT-11: Project reference (ID and Name)
	if inv.SpecifiedProcuringProjectID != "" || inv.SpecifiedProcuringProjectName != "" {
		spp := elt.CreateElement("ram:SpecifiedProcuringProject")
		if inv.SpecifiedProcuringProjectID != "" {
			spp.CreateElement("ram:ID").SetText(inv.SpecifiedProcuringProjectID)
		}
		if inv.SpecifiedProcuringProjectName != "" {
			spp.CreateElement("ram:Name").SetText(inv.SpecifiedProcuringProjectName)
		}
	}
}

func writeCIIramApplicableHeaderTradeDelivery(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeDelivery")

	if inv.ShipTo != nil {
		writeCIIParty(inv, inv.ShipTo, elt.CreateElement("ram:ShipToTradeParty"), CShipToParty)
	}

	// BT-72: Actual delivery date (BasicWL and above)
	if is(levelBasicWL, inv) && !inv.OccurrenceDateTime.IsZero() {
		odt := elt.CreateElement("ram:ActualDeliverySupplyChainEvent").CreateElement("ram:OccurrenceDateTime")
		addTimeCIIUDT(odt, inv.OccurrenceDateTime)
	}

	// BT-16: Despatch advice reference
	if inv.DespatchAdviceReferencedDocument != "" {
		elt.CreateElement("ram:DespatchAdviceReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.DespatchAdviceReferencedDocument)
	}
	// BT-15: Receiving advice reference
	if inv.ReceivingAdviceReferencedDocument != "" {
		elt.CreateElement("ram:ReceivingAdviceReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.ReceivingAdviceReferencedDocument)
	}
}

func writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:SpecifiedTradeSettlementHeaderMonetarySummation")
	elt.CreateElement("ram:LineTotalAmount").SetText(inv.LineTotal.StringFixed(2))

	// Only write charge/allowance totals if non-zero to avoid unnecessary elements
	if is(levelBasicWL, inv) {
		if !inv.ChargeTotal.IsZero() {
			elt.CreateElement("ram:ChargeTotalAmount").SetText(inv.ChargeTotal.StringFixed(2))
		}
		if !inv.AllowanceTotal.IsZero() {
			elt.CreateElement("ram:AllowanceTotalAmount").SetText(inv.AllowanceTotal.StringFixed(2))
		}
	}

	elt.CreateElement("ram:TaxBasisTotalAmount").SetText(inv.TaxBasisTotal.StringFixed(2))

	// BT-110: Tax total in invoice currency
	tta := elt.CreateElement("ram:TaxTotalAmount")
	currency := inv.TaxTotalCurrency
	if currency == "" {
		currency = inv.InvoiceCurrencyCode
	}
	tta.CreateAttr("currencyID", currency)
	tta.SetText(inv.TaxTotal.StringFixed(2))

	// BT-111: Tax total in accounting currency (when different from invoice currency)
	if inv.TaxCurrencyCode != "" && inv.TaxCurrencyCode != inv.InvoiceCurrencyCode {
		ttaVAT := elt.CreateElement("ram:TaxTotalAmount")
		ttaVAT.CreateAttr("currencyID", inv.TaxCurrencyCode)
		ttaVAT.SetText(inv.TaxTotalAccounting.StringFixed(2))
	}
	if is(levelEN16931, inv) && !inv.RoundingAmount.IsZero() {
		elt.CreateElement("ram:RoundingAmount").CreateText(inv.RoundingAmount.StringFixed(2))
	}

	elt.CreateElement("ram:GrandTotalAmount").CreateText(inv.GrandTotal.StringFixed(2))

	// Only write prepaid amount if non-zero to avoid unnecessary elements
	if is(levelBasicWL, inv) && !inv.TotalPrepaid.IsZero() {
		elt.CreateElement("ram:TotalPrepaidAmount").CreateText(inv.TotalPrepaid.StringFixed(2))
	}

	elt.CreateElement("ram:DuePayableAmount").CreateText(inv.DuePayableAmount.StringFixed(2))
}

func writeCIIramApplicableHeaderTradeSettlement(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeSettlement")

	// BT-90: Creditor reference ID
	if inv.CreditorReferenceID != "" {
		elt.CreateElement("ram:CreditorReferenceID").SetText(inv.CreditorReferenceID)
	}

	// BT-83: Payment reference (remittance information)
	if inv.PaymentReference != "" {
		elt.CreateElement("ram:PaymentReference").SetText(inv.PaymentReference)
	}

	// BT-6: VAT accounting currency code (optional, when different from invoice currency)
	if inv.TaxCurrencyCode != "" && inv.TaxCurrencyCode != inv.InvoiceCurrencyCode {
		elt.CreateElement("ram:TaxCurrencyCode").SetText(inv.TaxCurrencyCode)
	}

	// BT-5: Invoice currency code (required)
	elt.CreateElement("ram:InvoiceCurrencyCode").SetText(inv.InvoiceCurrencyCode)

	// PayeeTradeParty BG-10
	if pt := inv.PayeeTradeParty; pt != nil {
		writeCIIParty(inv, pt, elt.CreateElement("ram:PayeeTradeParty"), CPayeeParty)
	}

	if is(levelBasicWL, inv) {
		for i := range inv.PaymentMeans {
			pmElt := elt.CreateElement("ram:SpecifiedTradeSettlementPaymentMeans")

			//	BT-81
			pmElt.CreateElement("ram:TypeCode").SetText(strconv.Itoa(inv.PaymentMeans[i].TypeCode))

			if inf := inv.PaymentMeans[i].Information; inf != "" {
				// BT-82
				pmElt.CreateElement("ram:Information").SetText(inf)
			}
			if inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardID != "" {
				fCard := pmElt.CreateElement("ram:ApplicableTradeSettlementFinancialCard")
				// BT-87
				fCard.CreateElement("ram:ID").SetText(inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardID)
				// BT-88: Cardholder name is optional - only create if non-empty (PEPPOL-EN16931-R008)
				if inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardCardholderName != "" {
					fCard.CreateElement("ram:CardholderName").SetText(inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardCardholderName)
				}
			}
			if iban := inv.PaymentMeans[i].PayerPartyDebtorFinancialAccountIBAN; iban != "" {
				// BT-91
				pmElt.CreateElement("ram:PayerPartyDebtorFinancialAccount").CreateElement("ram:IBANID").SetText(iban)
			}
			if iban := inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountIBAN; iban != "" {
				// BG-17
				account := pmElt.CreateElement("ram:PayeePartyCreditorFinancialAccount")
				account.CreateElement("ram:IBANID").SetText(iban)
				// BT-85
				if name := inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountName; name != "" {
					account.CreateElement("ram:AccountName").SetText(name)
				}
				// BT-84
				if pid := inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountProprietaryID; pid != "" {
					account.CreateElement("ram:ProprietaryID").SetText(pid)
				}
			}
			// BT-86
			if bic := inv.PaymentMeans[i].PayeeSpecifiedCreditorFinancialInstitutionBIC; bic != "" {
				pmElt.CreateElement("ram:PayeeSpecifiedCreditorFinancialInstitution").CreateElement("ram:BICID").SetText(bic)
			}
		}
	}

	for i := range inv.TradeTaxes {
		att := elt.CreateElement("ram:ApplicableTradeTax")
		att.CreateElement("ram:CalculatedAmount").SetText(inv.TradeTaxes[i].CalculatedAmount.StringFixed(2))

		att.CreateElement("ram:TypeCode").SetText(inv.TradeTaxes[i].TypeCode)

		// BT-120: ExemptionReason must come after TypeCode and before BasisAmount
		if er := inv.TradeTaxes[i].ExemptionReason; er != "" {
			att.CreateElement("ram:ExemptionReason").SetText(er)
		}

		att.CreateElement("ram:BasisAmount").SetText(inv.TradeTaxes[i].BasisAmount.StringFixed(2))
		att.CreateElement("ram:CategoryCode").SetText(inv.TradeTaxes[i].CategoryCode)

		// BT-121: ExemptionReasonCode must come after CategoryCode
		if erc := inv.TradeTaxes[i].ExemptionReasonCode; erc != "" {
			att.CreateElement("ram:ExemptionReasonCode").SetText(erc)
		}

		att.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(inv.TradeTaxes[i].Percent))
	}
	for i := range inv.SpecifiedTradeAllowanceCharge {
		stacElt := elt.CreateElement("ram:SpecifiedTradeAllowanceCharge")
		stacElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText(strconv.FormatBool(inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator))
		// BT-93, BT-100: BasisAmount is optional - only create if non-zero (PEPPOL-EN16931-R008)
		if !inv.SpecifiedTradeAllowanceCharge[i].BasisAmount.IsZero() {
			stacElt.CreateElement("ram:BasisAmount").SetText(inv.SpecifiedTradeAllowanceCharge[i].BasisAmount.StringFixed(2))
		}
		stacElt.CreateElement("ram:ActualAmount").SetText(inv.SpecifiedTradeAllowanceCharge[i].ActualAmount.StringFixed(2))
		// BT-98, BT-105: ReasonCode is optional - only create if non-empty (PEPPOL-EN16931-R008)
		if inv.SpecifiedTradeAllowanceCharge[i].ReasonCode != "" {
			stacElt.CreateElement("ram:ReasonCode").SetText(inv.SpecifiedTradeAllowanceCharge[i].ReasonCode)
		}
		// BT-97, BT-104: Reason is optional - only create if non-empty (PEPPOL-EN16931-R008)
		if inv.SpecifiedTradeAllowanceCharge[i].Reason != "" {
			stacElt.CreateElement("ram:Reason").SetText(inv.SpecifiedTradeAllowanceCharge[i].Reason)
		}
		// BT-94, BT-101: CalculationPercent is optional - only create if non-zero (PEPPOL-EN16931-R008)
		if !inv.SpecifiedTradeAllowanceCharge[i].CalculationPercent.IsZero() {
			stacElt.CreateElement("ram:CalculationPercent").SetText(formatPercent(inv.SpecifiedTradeAllowanceCharge[i].CalculationPercent))
		}
		ctt := stacElt.CreateElement("ram:CategoryTradeTax")
		ctt.CreateElement("ram:TypeCode").SetText(inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxType)
		ctt.CreateElement("ram:CategoryCode").SetText(inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode)
		ctt.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent))
	}

	// BG-14: Billing period
	// BR-CO-19: Either start date OR end date OR both must be filled
	if !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero() {
		bsp := elt.CreateElement("ram:BillingSpecifiedPeriod")
		// Only write non-zero dates to avoid invalid XML dates like "00010101"
		if !inv.BillingSpecifiedPeriodStart.IsZero() {
			dt := bsp.CreateElement("ram:StartDateTime")
			addTimeCIIUDT(dt, inv.BillingSpecifiedPeriodStart)
		}
		if !inv.BillingSpecifiedPeriodEnd.IsZero() {
			dt := bsp.CreateElement("ram:EndDateTime")
			addTimeCIIUDT(dt, inv.BillingSpecifiedPeriodEnd)
		}
	}
	// BT-20
	for i := range inv.SpecifiedTradePaymentTerms {
		spt := elt.CreateElement("ram:SpecifiedTradePaymentTerms")
		if desc := inv.SpecifiedTradePaymentTerms[i].Description; desc != "" {
			spt.CreateElement("ram:Description").SetText(inv.SpecifiedTradePaymentTerms[i].Description)
		}
		// BT-9
		if !inv.SpecifiedTradePaymentTerms[i].DueDate.IsZero() {
			addTimeCIIUDT(spt.CreateElement("ram:DueDateDateTime"), inv.SpecifiedTradePaymentTerms[i].DueDate)
		}
		// BT-89: Direct debit mandate reference identifier
		if inv.SpecifiedTradePaymentTerms[i].DirectDebitMandateID != "" {
			spt.CreateElement("ram:DirectDebitMandateID").SetText(inv.SpecifiedTradePaymentTerms[i].DirectDebitMandateID)
		}
	}

	writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv, elt)

	// BT-19: Buyer accounting reference
	if inv.ReceivableSpecifiedTradeAccountingAccount != "" {
		rstaac := elt.CreateElement("ram:ReceivableSpecifiedTradeAccountingAccount")
		rstaac.CreateElement("ram:ID").SetText(inv.ReceivableSpecifiedTradeAccountingAccount)
	}

	// BG-3
	for i := range inv.InvoiceReferencedDocument {
		refdoc := elt.CreateElement("ram:InvoiceReferencedDocument")
		refdoc.CreateElement("ram:IssuerAssignedID").SetText(inv.InvoiceReferencedDocument[i].ID)
		addTimeCIIQDT(refdoc.CreateElement("ram:FormattedIssueDateTime"), inv.InvoiceReferencedDocument[i].Date)
	}
}

func writeCIIrsmSupplyChainTradeTransaction(inv *Invoice, parent *etree.Element) {
	rsctt := parent.CreateElement("rsm:SupplyChainTradeTransaction")
	for i := range inv.InvoiceLines {
		writeCIIramIncludedSupplyChainTradeLineItem(&inv.InvoiceLines[i], inv, rsctt)
	}

	writeCIIramApplicableHeaderTradeAgreement(inv, rsctt)
	writeCIIramApplicableHeaderTradeDelivery(inv, rsctt)
	writeCIIramApplicableHeaderTradeSettlement(inv, rsctt)
}

func writeCIIrsmExchangedDocumentContext(inv *Invoice, root *etree.Element) {
	documentContext := root.CreateElement("rsm:ExchangedDocumentContext")
	// BusinessProcessSpecifiedDocumentContextParameter BT-23 is mandatory in extended, optional otherwise
	// Only output if it has a value OR if we're in Extended profile (where it's mandatory)
	if inv.BPSpecifiedDocumentContextParameter != "" || is(levelExtended, inv) {
		documentContext.CreateElement("ram:BusinessProcessSpecifiedDocumentContextParameter").CreateElement("ram:ID").CreateText(inv.BPSpecifiedDocumentContextParameter)
	}

	// GuidelineSpecifiedDocumentContextParameter BT-24 is mandatory
	guidelineContextParameter := documentContext.CreateElement("ram:GuidelineSpecifiedDocumentContextParameter")
	guidelineContextParameter.CreateElement("ram:ID").CreateText(inv.GuidelineSpecifiedDocumentContextParameter)
}

func writeCIIrsmExchangedDocument(inv *Invoice, root *etree.Element) {
	exchangedDoc := root.CreateElement("rsm:ExchangedDocument")
	exchangedDoc.CreateElement("ram:ID").SetText(inv.InvoiceNumber)
	exchangedDoc.CreateElement("ram:TypeCode").SetText(inv.InvoiceTypeCode.String())
	addTimeCIIUDT(exchangedDoc.CreateElement("ram:IssueDateTime"), inv.InvoiceDate)

	for _, note := range inv.Notes {
		in := exchangedDoc.CreateElement("ram:IncludedNote")
		rc := in.CreateElement("ram:Content")
		rc.SetText(note.Text)
		if note.SubjectCode != "" {
			in.CreateElement("ram:SubjectCode").SetText(note.SubjectCode)
		}
	}
}

// writeCII writes an invoice in CII (Cross Industry Invoice) format used by ZUGFeRD/Factur-X.
func writeCII(inv *Invoice, writer io.Writer) error {
	var err error

	doc := etree.NewDocument()
	root := doc.CreateElement("rsm:CrossIndustryInvoice")
	root.CreateAttr("xmlns:rsm", "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100")
	root.CreateAttr("xmlns:qdt", "urn:un:unece:uncefact:data:standard:QualifiedDataType:100")
	root.CreateAttr("xmlns:ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
	root.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	root.CreateAttr("xmlns:udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")
	writeCIIrsmExchangedDocumentContext(inv, root)
	writeCIIrsmExchangedDocument(inv, root)
	writeCIIrsmSupplyChainTradeTransaction(inv, root)

	doc.Indent(2)
	if _, err = doc.WriteTo(writer); err != nil {
		return fmt.Errorf("write CII: failed to write to the writer %w", err)
	}

	return nil
}
