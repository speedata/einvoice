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

func writeCIIramIncludedSupplyChainTradeLineItem(ii InvoiceLine, parent *etree.Element) {

}

func writeCIIParty(inv *Invoice, party Party, parent *etree.Element, partyType CodePartyType) {
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

}

func writeCIIramApplicableHeaderTradeDelivery(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeDelivery")
	if is(CProfileBasic, inv) {
		fmt.Println("~~> is more than minimal")
	}
	_ = elt
}

func writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:SpecifiedTradeSettlementHeaderMonetarySummation")
	elt.CreateElement("ram:TaxBasisTotalAmount").SetText(inv.TaxBasisTotal.StringFixed(2))
	tta := elt.CreateElement("ram:TaxTotalAmount")
	tta.CreateAttr("currencyID", inv.TaxTotalCurrency)
	tta.SetText(inv.TaxTotal.StringFixed(2))
	elt.CreateElement("ram:GrandTotalAmount").CreateText(inv.GrandTotal.StringFixed(2))
	elt.CreateElement("ram:DuePayableAmount").CreateText(inv.DuePayableAmount.StringFixed(2))
}

func writeCIIramApplicableHeaderTradeSettlement(inv *Invoice, parent *etree.Element) {
	elt := parent.CreateElement("ram:ApplicableHeaderTradeSettlement")
	elt.CreateElement("ram:InvoiceCurrencyCode").SetText(inv.InvoiceCurrencyCode)
	writeCIIramSpecifiedTradeSettlementHeaderMonetarySummation(inv, elt)
}

func writeCIIrsmSupplyChainTradeTransaction(inv *Invoice, parent *etree.Element) {
	rsctt := parent.CreateElement("rsm:SupplyChainTradeTransaction")
	// for _, ii := range inv.InvoiceItems {
	// 	writeCIIramIncludedSupplyChainTradeLineItem(ii, rsctt)
	// }
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
		if note.SubjectCode != "" {
			in.CreateElement("ram:SubjectCode").SetText(note.SubjectCode)
		}
		rc := in.CreateElement("ram:Content")
		rc.SetText(note.Text)
	}
}

func writeCII(inv *Invoice, w io.Writer) error {
	var err error
	d := etree.NewDocument()
	root := d.CreateElement("rsm:CrossIndustryInvoice")
	root.CreateAttr("xmlns:rsm", "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100")
	root.CreateAttr("xmlns:qdt", "urn:un:unece:uncefact:data:standard:QualifiedDataType:100")
	root.CreateAttr("xmlns:udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")
	root.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	root.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	root.CreateAttr("xmlns:ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
	writeCIIrsmExchangedDocumentContext(inv, root)
	writeCIIrsmExchangedDocument(inv, root)
	writeCIIrsmSupplyChainTradeTransaction(inv, root)
	d.Indent(4)
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
