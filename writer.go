package einvoice

import (
	"fmt"
	"io"
	"time"

	"github.com/beevik/etree"
)

// is returns true if the profile in the invoice is at least cp.
func is(cp CodeProfileType, inv *Invoice) bool {
	return inv.Profile >= cp
}

func addTime(parent *etree.Element, date time.Time) {
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
	gid := stp.CreateElement("ram:GlobalID")
	gid.CreateAttr("schemeID", il.GlobalIDType)
	gid.SetText(il.GlobalID)
	stp.CreateElement("ram:SellerAssignedID").SetText(il.ArticleNumber)
	// BT-156 mandatory in EN
	if is(CProfileEN16931, inv) {
		stp.CreateElement("ram:BuyerAssignedID").SetText(il.ArticleNumberBuyer)
	}
	stp.CreateElement("ram:Name").SetText(il.ItemName)
	if il.Description != "" {
		stp.CreateElement("ram:Description").SetText(il.Description)
	}
	slta := li.CreateElement("ram:SpecifiedLineTradeAgreement")
	gpptp := slta.CreateElement("ram:GrossPriceProductTradePrice")
	gpptp.CreateElement("ram:ChargeAmount").SetText(il.GrossPrice.StringFixed(4))
	for _, ac := range il.AppliedTradeAllowanceCharge {
		acElt := gpptp.CreateElement("ram:AppliedTradeAllowanceCharge")
		acElt.CreateElement("ram:ActualAmount").SetText(ac.ActualAmount.StringFixed(4))
		acElt.CreateElement("ram:Reason").SetText(ac.Reason)
	}
	slta.CreateElement("ram:NetPriceProductTradePrice").CreateElement("ram:ChargeAmount").SetText(il.NetPrice.StringFixed(4))
	bq := li.CreateElement("ram:SpecifiedLineTradeDelivery").CreateElement("ram:BilledQuantity")
	bq.CreateAttr("unitCode", il.BilledQuantityUnit)
	bq.SetText(il.BilledQuantity.StringFixed(4))

	slts := li.CreateElement("ram:SpecifiedLineTradeSettlement")
	// BG-27, BG-28
	att := slts.CreateElement("ram:ApplicableTradeTax")
	att.CreateElement("ram:TypeCode").SetText(il.TaxTypeCode)
	att.CreateElement("ram:CategoryCode").SetText(il.TaxCategoryCode)
	att.CreateElement("ram:RateApplicablePercent").SetText(il.TaxRateApplicablePercent.StringFixed(2))
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
	parent.CreateElement("ram:Name").SetText(party.Name)
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
}

func writeCIIramApplicableHeaderTradeDelivery(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeDelivery")
	if is(CProfileBasic, inv) {
		odt := elt.CreateElement("ram:ActualDeliverySupplyChainEvent").CreateElement("ram:OccurrenceDateTime")
		addTime(odt, inv.OccurrenceDateTime)
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
		att.CreateElement("ram:BasisAmount").SetText(tt.BasisAmount.StringFixed(2))
		att.CreateElement("ram:CategoryCode").SetText(tt.CategoryCode)
		att.CreateElement("ram:RateApplicablePercent").SetText(tt.Percent.StringFixed(2))
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
		ctt.CreateElement("ram:RateApplicablePercent").SetText(stac.CategoryTradeTaxRateApplicablePercent.StringFixed(2))
	}
	// ram:SpecifiedTradePaymentTerms
	// BT-20
	for _, pt := range inv.SpecifiedTradePaymentTerms {
		spt := elt.CreateElement("ram:SpecifiedTradePaymentTerms")
		spt.CreateElement("ram:Description").SetText(pt.Description)

	}
	writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv, elt)
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
	addTime(ed.CreateElement("ram:IssueDateTime"), inv.InvoiceDate)
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
