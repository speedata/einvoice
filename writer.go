package einvoice

import (
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/beevik/etree"
	"github.com/shopspring/decimal"
)

var percentageRE = regexp.MustCompile(`^(.*?)\.?0+$`)

// is returns true if the profile in the invoice is at least cp.
func is(cp CodeProfileType, inv *Invoice) bool {
	return inv.Profile >= cp
}

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

func writeCIIramIncludedSupplyChainTradeLineItem(il InvoiceLine, inv *Invoice, parent *etree.Element) {
	li := parent.CreateElement("ram:IncludedSupplyChainTradeLineItem")
	adld := li.CreateElement("ram:AssociatedDocumentLineDocument")
	lineID := adld.CreateElement("ram:LineID")
	lineID.SetText(il.LineID)
	if il.Note != "" {
		adld.CreateElement("ram:IncludedNote").CreateElement("ram:Content").SetText(il.Note)
	}
	stp := li.CreateElement("ram:SpecifiedTradeProduct")
	if il.GlobalID != "" { // BT-157
		gid := stp.CreateElement("ram:GlobalID")
		gid.CreateAttr("schemeID", il.GlobalIDType)
		gid.SetText(il.GlobalID)
	}
	// BT-155 is optional in EN
	if is(CProfileEN16931, inv) {
		if artno := il.ArticleNumber; artno != "" {
			stp.CreateElement("ram:SellerAssignedID").SetText(artno)
		}
	}

	// BT-156 optional in EN
	if is(CProfileEN16931, inv) {
		if artno := il.ArticleNumberBuyer; artno != "" {
			stp.CreateElement("ram:BuyerAssignedID").SetText(il.ArticleNumberBuyer)
		}
	}
	stp.CreateElement("ram:Name").SetText(il.ItemName)
	if il.Description != "" {
		stp.CreateElement("ram:Description").SetText(il.Description)
	}
	slta := li.CreateElement("ram:SpecifiedLineTradeAgreement")
	// BT-148
	if !il.GrossPrice.IsZero() {
		gpptp := slta.CreateElement("ram:GrossPriceProductTradePrice")
		gpptp.CreateElement("ram:ChargeAmount").SetText(il.GrossPrice.StringFixed(12))
		for _, ac := range il.AppliedTradeAllowanceCharge {
			acElt := gpptp.CreateElement("ram:AppliedTradeAllowanceCharge")
			// BG-27, BG-28
			acElt.CreateElement("ram:ChargeIndicator").CreateElement("udt:Indicator").SetText(fmt.Sprintf("%t", ac.ChargeIndicator))
			if cp := ac.CalculationPercent; !cp.IsZero() {
				acElt.CreateElement("ram:CalculationPercent").SetText(formatPercent(cp))
			}
			if ba := ac.BasisAmount; !ba.IsZero() {
				acElt.CreateElement("ram:BasisAmount").SetText(ba.StringFixed(2))
			}
			acElt.CreateElement("ram:ActualAmount").SetText(ac.ActualAmount.StringFixed(2))
			if r := ac.Reason; r != "" {
				acElt.CreateElement("ram:Reason").SetText(r)
			}
		}
	}
	slta.CreateElement("ram:NetPriceProductTradePrice").CreateElement("ram:ChargeAmount").SetText(il.NetPrice.StringFixed(2))
	bq := li.CreateElement("ram:SpecifiedLineTradeDelivery").CreateElement("ram:BilledQuantity")
	bq.CreateAttr("unitCode", il.BilledQuantityUnit)
	bq.SetText(il.BilledQuantity.StringFixed(4))

	slts := li.CreateElement("ram:SpecifiedLineTradeSettlement")
	// BG-27, BG-28
	att := slts.CreateElement("ram:ApplicableTradeTax")
	att.CreateElement("ram:TypeCode").SetText(il.TaxTypeCode)
	att.CreateElement("ram:CategoryCode").SetText(il.TaxCategoryCode)
	att.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(il.TaxRateApplicablePercent))
	slts.CreateElement("ram:SpecifiedTradeSettlementLineMonetarySummation").CreateElement("ram:LineTotalAmount").SetText(il.Total.StringFixed(2))
}

func writeCIIParty(inv *Invoice, party Party, parent *etree.Element, partyType CodePartyType) {
	for _, id := range party.ID {
		parent.CreateElement("ram:ID").SetText(id)
	}
	for _, gid := range party.GlobalID {
		eGid := parent.CreateElement("ram:GlobalID")
		eGid.CreateAttr("schemeID", gid.Scheme)
		eGid.SetText(gid.ID)
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
			email.CreateAttr("schemeID", "SMTP")
			email.SetText(dtc.EMail)
		}

	}
	if ppa := party.PostalAddress; ppa != nil {
		// profile minimum has no postal address for the buyer (BG-8)
		if partyType == CSellerParty || is(CProfileBasic, inv) {
			pa := parent.CreateElement("ram:PostalTradeAddress")
			pa.CreateElement("ram:PostcodeCode").SetText(ppa.PostcodeCode)
			if l1 := ppa.Line1; l1 != "" {
				pa.CreateElement("ram:LineOne").SetText(l1)
			}
			if l2 := ppa.Line2; l2 != "" {
				pa.CreateElement("ram:LineTwo").SetText(l2)
			}
			if l3 := ppa.Line3; l3 != "" {
				pa.CreateElement("ram:LineThree").SetText(l3)
			}
			if cityName := ppa.City; cityName != "" {
				pa.CreateElement("ram:CityName").SetText(cityName)
			}
			if cid := ppa.CountryID; cid != "" {
				pa.CreateElement("ram:CountryID").SetText(cid)
			}
			if csd := ppa.CountrySubDivisionName; csd != "" {
				pa.CreateElement("ram:CountrySubDivisionName").SetText(csd)
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
	tta.CreateAttr("currencyID", inv.TaxTotalCurrency)
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
	elt.CreateElement("ram:InvoiceCurrencyCode").SetText(inv.InvoiceCurrencyCode)
	for _, tt := range inv.TradeTaxes {
		att := elt.CreateElement("ram:ApplicableTradeTax")
		att.CreateElement("ram:CalculatedAmount").SetText(tt.CalculatedAmount.StringFixed(2))
		att.CreateElement("ram:TypeCode").SetText(tt.Typ)
		if er := tt.ExemptionReason; er != "" {
			att.CreateElement("ram:ExemptionReason").SetText(er)
		}
		att.CreateElement("ram:BasisAmount").SetText(tt.BasisAmount.StringFixed(2))
		att.CreateElement("ram:CategoryCode").SetText(tt.CategoryCode)
		att.CreateElement("ram:RateApplicablePercent").SetText(formatPercent(tt.Percent))
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
	for _, pt := range inv.SpecifiedTradePaymentTerms {
		spt := elt.CreateElement("ram:SpecifiedTradePaymentTerms")
		if desc := pt.Description; desc != "" {
			spt.CreateElement("ram:Description").SetText(pt.Description)
		}
		// BT-9
		if !pt.DueDate.IsZero() {
			addTimeUDT(spt.CreateElement("ram:DueDateDateTime"), pt.DueDate)
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
	ed := root.CreateElement("rsm:ExchangedDocument")
	ed.CreateElement("ram:ID").SetText(inv.InvoiceNumber)
	ed.CreateElement("ram:TypeCode").SetText(inv.InvoiceTypeCode.String())
	addTimeUDT(ed.CreateElement("ram:IssueDateTime"), inv.InvoiceDate)
	for _, note := range inv.Notes {
		in := ed.CreateElement("ram:IncludedNote")
		rc := in.CreateElement("ram:Content")
		rc.SetText(note.Text)
		if note.SubjectCode != "" {
			in.CreateElement("ram:SubjectCode").SetText(note.SubjectCode)
		}
	}
}

func writeCII(inv *Invoice, w io.Writer) error {
	var err error
	d := etree.NewDocument()
	root := d.CreateElement("rsm:CrossIndustryInvoice")
	root.CreateAttr("xmlns:rsm", "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100")
	root.CreateAttr("xmlns:qdt", "urn:un:unece:uncefact:data:standard:QualifiedDataType:100")
	root.CreateAttr("xmlns:ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
	root.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	root.CreateAttr("xmlns:udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")
	writeCIIrsmExchangedDocumentContext(inv, root)
	writeCIIrsmExchangedDocument(inv, root)
	writeCIIrsmSupplyChainTradeTransaction(inv, root)
	d.Indent(2)
	_, err = d.WriteTo(w)
	return err
}

func (inv *Invoice) Write(w io.Writer) error {
	switch inv.SchemaType {
	case UBL:
		return fmt.Errorf("UBL writing is not supported yet")
	default:
		return writeCII(inv, w)
	}
}
