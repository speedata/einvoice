package einvoice

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"
)

var percentageRE = regexp.MustCompile(`^(.*?)\.?0+$`)
var ErrWrite = errors.New("creating the XML failed")

// is returns true if the profile in the invoice is at least cp.
func is(cp CodeProfileType, inv *Invoice) bool {
	return inv.Profile >= cp
}

// formatPercent removes trailing zeros and the decimal point, if possible.
func formatPercent(d decimal.Decimal) string {
	str := d.StringFixed(4)

	return percentageRE.ReplaceAllString(str, "$1")
}

func addTimeQDT(parent *etree.Element, date time.Time) {
	udtdts := parent.CreateElement("qdt:DateTimeString")
	udtdts.CreateAttr("format", "102")
	udtdts.CreateText(date.Format("20060102"))
}

func addTimeUDT(parent *etree.Element, date time.Time) {
	udtdts := parent.CreateElement("udt:DateTimeString")
	udtdts.CreateAttr("format", "102")
	udtdts.CreateText(date.Format("20060102"))
}

func writeCIIramIncludedSupplyChainTradeLineItem(invoiceLine InvoiceLine, inv *Invoice, parent *etree.Element) {
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
	if is(CProfileEN16931, inv) {
		if artno := invoiceLine.ArticleNumber; artno != "" {
			stp.CreateElement("ram:SellerAssignedID").SetText(artno)
		}
	}

	// BT-156 optional in EN
	if is(CProfileEN16931, inv) {
		if artno := invoiceLine.ArticleNumberBuyer; artno != "" {
			stp.CreateElement("ram:BuyerAssignedID").SetText(invoiceLine.ArticleNumberBuyer)
		}
	}

	stp.CreateElement("ram:Name").SetText(invoiceLine.ItemName)

	if invoiceLine.Description != "" {
		stp.CreateElement("ram:Description").SetText(invoiceLine.Description)
	}

	slta := lineItem.CreateElement("ram:SpecifiedLineTradeAgreement")

	// BT-148
	if !invoiceLine.GrossPrice.IsZero() {
		gpptp := slta.CreateElement("ram:GrossPriceProductTradePrice")
		gpptp.CreateElement("ram:ChargeAmount").SetText(invoiceLine.GrossPrice.StringFixed(2))

		for _, allowanceCharge := range invoiceLine.AppliedTradeAllowanceCharge {
			acElt := gpptp.CreateElement("ram:AppliedTradeAllowanceCharge")
			// BG-27, BG-28
			acElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText(fmt.Sprintf("%t", allowanceCharge.ChargeIndicator))

			if cp := allowanceCharge.CalculationPercent; !cp.IsZero() {
				acElt.CreateElement("ram:CalculationPercent").SetText(formatPercent(cp))
			}

			if ba := allowanceCharge.BasisAmount; !ba.IsZero() {
				acElt.CreateElement("ram:BasisAmount").SetText(ba.StringFixed(2))
			}

			acElt.CreateElement("ram:ActualAmount").SetText(allowanceCharge.ActualAmount.StringFixed(2))
			if r := allowanceCharge.Reason; r != "" {
				acElt.CreateElement("ram:Reason").SetText(r)
			}
		}
	}
	slta.CreateElement("ram:NetPriceProductTradePrice").CreateElement("ram:ChargeAmount").SetText(invoiceLine.NetPrice.StringFixed(2))
	bq := lineItem.CreateElement("ram:SpecifiedLineTradeDelivery").CreateElement("ram:BilledQuantity")
	bq.CreateAttr("unitCode", invoiceLine.BilledQuantityUnit)
	bq.SetText(invoiceLine.BilledQuantity.StringFixed(4))

	slts := lineItem.CreateElement("ram:SpecifiedLineTradeSettlement")
	// BG-27, BG-28
	att := slts.CreateElement("ram:ApplicableTradeTax")
	// BT-151 must be VAT
	if invoiceLine.TaxTypeCode == "" {
		invoiceLine.TaxTypeCode = "VAT"
	}

	att.CreateElement("ram:TypeCode").SetText(invoiceLine.TaxTypeCode)
	att.CreateElement("ram:CategoryCode").SetText(invoiceLine.TaxCategoryCode)
	att.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(invoiceLine.TaxRateApplicablePercent))
	slts.CreateElement("ram:SpecifiedTradeSettlementLineMonetarySummation").CreateElement("ram:LineTotalAmount").SetText(invoiceLine.Total.StringFixed(2))
}

func writeCIIParty(inv *Invoice, party Party, parent *etree.Element, partyType CodePartyType) {
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

	if slo := party.SpecifiedLegalOrganization; slo != nil {
		sloElt := parent.CreateElement("ram:SpecifiedLegalOrganization")
		id := sloElt.CreateElement("ram:ID")
		id.CreateAttr("schemeID", slo.Scheme)
		id.SetText(slo.ID)
	}

	for _, dtc := range party.DefinedTradeContact {
		dtcElt := parent.CreateElement("ram:DefinedTradeContact")
		dtcElt.CreateElement("ram:PersonName").SetText(dtc.PersonName)

		if dtc.PhoneNumber != "" {
			dtcElt.CreateElement("ram:TelephoneUniversalCommunication").CreateElement("ram:CompleteNumber").SetText(dtc.PhoneNumber)
		}

		if dtc.EMail != "" {
			email := dtcElt.CreateElement("ram:EmailURIUniversalCommunication").CreateElement("ram:URIID")
			// email.CreateAttr("schemeID", "SMTP")
			email.SetText(dtc.EMail)
		}
	}

	if ppa := party.PostalAddress; ppa != nil {
		// profile minimum has no postal address for the buyer (BG-8)
		if partyType == CSellerParty || is(CProfileBasic, inv) {
			postalAddress := parent.CreateElement("ram:PostalTradeAddress")
			postalAddress.CreateElement("ram:PostcodeCode").SetText(ppa.PostcodeCode)

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

	writeCIIParty(inv, inv.Seller, elt.CreateElement("ram:SellerTradeParty"), CSellerParty)
	writeCIIParty(inv, inv.Buyer, elt.CreateElement("ram:BuyerTradeParty"), CBuyerParty)
	// BT-13
	if inv.BuyerOrderReferencedDocument != "" {
		elt.CreateElement("ram:BuyerOrderReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.BuyerOrderReferencedDocument)
	}
	// BT-12
	if inv.ContractReferencedDocument != "" {
		elt.CreateElement("ram:ContractReferencedDocument").CreateElement("ram:IssuerAssignedID").SetText(inv.ContractReferencedDocument)
	}
	// BG-24
	for _, doc := range inv.AdditionalReferencedDocument {
		ard := elt.CreateElement("ram:AdditionalReferencedDocument")
		ard.CreateElement("ram:IssuerAssignedID").SetText(doc.IssuerAssignedID)
		ard.CreateElement("ram:TypeCode").SetText(doc.TypeCode)
		ard.CreateElement("ram:Name").SetText(doc.Name)
		abo := ard.CreateElement("ram:AttachmentBinaryObject")
		abo.CreateAttr("mimeCode", doc.AttachmentMimeCode)
		abo.CreateAttr("filename", doc.AttachmentFilename)
		// .SetText(base64.StdEncoding.EncodeToString(doc.AttachmentBinaryObject))
		if doc.TypeCode == "130" {
			ard.CreateElement("ram:ReferenceTypeCode").SetText(doc.ReferenceTypeCode)
		}
	}
}

func writeCIIramApplicableHeaderTradeDelivery(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeDelivery")

	if inv.ShipTo != nil {
		writeCIIParty(inv, *inv.ShipTo, elt.CreateElement("ram:ShipToTradeParty"), CShipToParty)
	}

	if is(CProfileBasic, inv) && !inv.OccurrenceDateTime.IsZero() {
		// BT-72
		odt := elt.CreateElement("ram:ActualDeliverySupplyChainEvent").CreateElement("ram:OccurrenceDateTime")
		addTimeUDT(odt, inv.OccurrenceDateTime)
	}
}

func writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:SpecifiedTradeSettlementHeaderMonetarySummation")
	elt.CreateElement("ram:LineTotalAmount").SetText(inv.LineTotal.StringFixed(2))

	if is(CProfileBasicWL, inv) {
		elt.CreateElement("ram:ChargeTotalAmount").SetText(inv.ChargeTotal.StringFixed(2))
		elt.CreateElement("ram:AllowanceTotalAmount").SetText(inv.AllowanceTotal.StringFixed(2))
	}

	elt.CreateElement("ram:TaxBasisTotalAmount").SetText(inv.TaxBasisTotal.StringFixed(2))
	tta := elt.CreateElement("ram:TaxTotalAmount")

	currency := inv.TaxTotalCurrency
	if currency == "" {
		currency = inv.InvoiceCurrencyCode
	}

	tta.CreateAttr("currencyID", currency)
	tta.SetText(inv.TaxTotal.StringFixed(2))
	if is(CProfileEN16931, inv) && !inv.RoundingAmount.IsZero() {
		elt.CreateElement("ram:RoundingAmount").CreateText(inv.RoundingAmount.StringFixed(2))
	}

	elt.CreateElement("ram:GrandTotalAmount").CreateText(inv.GrandTotal.StringFixed(2))

	if is(CProfileBasicWL, inv) {
		elt.CreateElement("ram:TotalPrepaidAmount").CreateText(inv.TotalPrepaid.StringFixed(2))
	}

	elt.CreateElement("ram:DuePayableAmount").CreateText(inv.DuePayableAmount.StringFixed(2))
}

func writeCIIramApplicableHeaderTradeSettlement(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeSettlement")
	// CreditorReferenceID BT-90
	// PaymentReference BT-83
	// TaxCurrencyCode BT-6
	elt.CreateElement("ram:InvoiceCurrencyCode").SetText(inv.InvoiceCurrencyCode)

	// PayeeTradeParty BG-10
	if pt := inv.PayeeTradeParty; pt != nil {
		writeCIIParty(inv, *pt, elt, CPayeeParty)
	}

	if is(CProfileBasicWL, inv) {
		for _, paymentMeans := range inv.PaymentMeans {
			pmElt := elt.CreateElement("ram:SpecifiedTradeSettlementPaymentMeans")

			//	BT-81
			pmElt.CreateElement("ram:TypeCode").SetText(fmt.Sprintf("%d", paymentMeans.TypeCode))

			if inf := paymentMeans.Information; inf != "" {
				// BT-82
				pmElt.CreateElement("ram:Information").SetText(inf)
			}
			if paymentMeans.ApplicableTradeSettlementFinancialCardID != "" {
				fCard := pmElt.CreateElement("ram:ApplicableTradeSettlementFinancialCard")
				// BT-87
				fCard.CreateElement("ram:ID").SetText(paymentMeans.ApplicableTradeSettlementFinancialCardID)
				// BT-88
				fCard.CreateElement("ram:CardholderName").SetText(paymentMeans.ApplicableTradeSettlementFinancialCardCardholderName)
			}
			if iban := paymentMeans.PayerPartyDebtorFinancialAccountIBAN; iban != "" {
				// BT-91
				pmElt.CreateElement("ram:PayerPartyDebtorFinancialAccount").CreateElement("ram:IBANID").SetText(iban)
			}
			if iban := paymentMeans.PayeePartyCreditorFinancialAccountIBAN; iban != "" {
				// BG-17
				account := pmElt.CreateElement("ram:PayeePartyCreditorFinancialAccount")
				account.CreateElement("ram:IBANID").SetText(iban)
				// BT-85
				if name := paymentMeans.PayeePartyCreditorFinancialAccountName; name != "" {
					account.CreateElement("ram:AccountName").SetText(name)
				}
				// BT-84
				if pid := paymentMeans.PayeePartyCreditorFinancialAccountProprietaryID; pid != "" {
					account.CreateElement("ram:ProprietaryID").SetText(pid)
				}
			}
			// BT-86
			if bic := paymentMeans.PayeeSpecifiedCreditorFinancialInstitutionBIC; bic != "" {
				pmElt.CreateElement("ram:PayeeSpecifiedCreditorFinancialInstitution").CreateElement("ram:BICID").SetText(bic)
			}
		}
	}

	for _, tradeTax := range inv.TradeTaxes {
		att := elt.CreateElement("ram:ApplicableTradeTax")
		att.CreateElement("ram:CalculatedAmount").SetText(tradeTax.CalculatedAmount.StringFixed(2))

		att.CreateElement("ram:TypeCode").SetText(tradeTax.Typ)

		if er := tradeTax.ExemptionReason; er != "" {
			att.CreateElement("ram:ExemptionReason").SetText(er)
		}

		att.CreateElement("ram:BasisAmount").SetText(tradeTax.BasisAmount.StringFixed(2))
		att.CreateElement("ram:CategoryCode").SetText(tradeTax.CategoryCode)
		att.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(tradeTax.Percent))
	}
	for _, stac := range inv.SpecifiedTradeAllowanceCharge {
		stacElt := elt.CreateElement("ram:SpecifiedTradeAllowanceCharge")
		stacElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText(fmt.Sprintf("%t", stac.ChargeIndicator))
		stacElt.CreateElement("ram:BasisAmount").SetText(stac.BasisAmount.StringFixed(2))
		stacElt.CreateElement("ram:ActualAmount").SetText(stac.ActualAmount.StringFixed(2))
		stacElt.CreateElement("ram:Reason").SetText(stac.Reason)
		ctt := stacElt.CreateElement("ram:CategoryTradeTax")
		ctt.CreateElement("ram:TypeCode").SetText(stac.CategoryTradeTaxType)
		ctt.CreateElement("ram:CategoryCode").SetText(stac.CategoryTradeTaxCategoryCode)
		ctt.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(stac.CategoryTradeTaxRateApplicablePercent))
	}

	// BG-14
	if !inv.BillingSpecifiedPeriodStart.IsZero() {
		bsp := elt.CreateElement("ram:BillingSpecifiedPeriod")
		dt := bsp.CreateElement("ram:StartDateTime")
		addTimeUDT(dt, inv.BillingSpecifiedPeriodStart)
		dt = bsp.CreateElement("ram:EndDateTime")
		addTimeUDT(dt, inv.BillingSpecifiedPeriodEnd)
	}
	// BT-20
	for _, paymentTerm := range inv.SpecifiedTradePaymentTerms {
		spt := elt.CreateElement("ram:SpecifiedTradePaymentTerms")
		if desc := paymentTerm.Description; desc != "" {
			spt.CreateElement("ram:Description").SetText(paymentTerm.Description)
		}
		// BT-9
		if !paymentTerm.DueDate.IsZero() {
			addTimeUDT(spt.CreateElement("ram:DueDateDateTime"), paymentTerm.DueDate)
		}
	}

	writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv, elt)
	// BG-3
	for _, v := range inv.InvoiceReferencedDocument {
		refdoc := elt.CreateElement("ram:InvoiceReferencedDocument")
		refdoc.CreateElement("ram:IssuerAssignedID").SetText(v.ID)
		addTimeQDT(refdoc.CreateElement("ram:FormattedIssueDateTime"), v.Date)
	}
}

func writeCIIrsmSupplyChainTradeTransaction(inv *Invoice, parent *etree.Element) {
	rsctt := parent.CreateElement("rsm:SupplyChainTradeTransaction")
	for _, il := range inv.InvoiceLines {
		writeCIIramIncludedSupplyChainTradeLineItem(il, inv, rsctt)
	}

	writeCIIramApplicableHeaderTradeAgreement(inv, rsctt)
	writeCIIramApplicableHeaderTradeDelivery(inv, rsctt)
	writeCIIramApplicableHeaderTradeSettlement(inv, rsctt)
}

func writeCIIrsmExchangedDocumentContext(inv *Invoice, root *etree.Element) {
	documentContext := root.CreateElement("rsm:ExchangedDocumentContext")
	// BusinessProcessSpecifiedDocumentContextParameter BT-23 is mandatory in extended
	if inv.BPSpecifiedDocumentContextParameter != "" || is(CProfileExtended, inv) {
		documentContext.CreateElement("ram:BusinessProcessSpecifiedDocumentContextParameter").CreateElement("ram:ID").CreateText(inv.BPSpecifiedDocumentContextParameter)
	}

	guidelineContextParameter := documentContext.CreateElement("ram:GuidelineSpecifiedDocumentContextParameter")
	guidelineContextParameter.CreateElement("ram:ID").CreateText(inv.Profile.ToProfileName())
}

func writeCIIrsmExchangedDocument(inv *Invoice, root *etree.Element) {
	exchangedDoc := root.CreateElement("rsm:ExchangedDocument")
	exchangedDoc.CreateElement("ram:ID").SetText(inv.InvoiceNumber)
	exchangedDoc.CreateElement("ram:TypeCode").SetText(inv.InvoiceTypeCode.String())
	addTimeUDT(exchangedDoc.CreateElement("ram:IssueDateTime"), inv.InvoiceDate)

	for _, note := range inv.Notes {
		in := exchangedDoc.CreateElement("ram:IncludedNote")
		rc := in.CreateElement("ram:Content")
		rc.SetText(note.Text)
		if note.SubjectCode != "" {
			in.CreateElement("ram:SubjectCode").SetText(note.SubjectCode)
		}
	}
}

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

// ErrUnsupportedSchema is returned when the library does not recognize the schema.
var ErrUnsupportedSchema = errors.New("unsupported schema")

func (inv *Invoice) Write(w io.Writer) error {
	switch inv.SchemaType {
	case UBL:
		return fmt.Errorf("unknown schema UBL %w", ErrUnsupportedSchema)
	case CII:
		return writeCII(inv, w)
	default:
		return ErrUnsupportedSchema
	}
}
