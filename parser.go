package einvoice

import (
	"io"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/cxpath"
)

func parseTime(timestring string) (time.Time, error) {
	return time.Parse("20060102", timestring)
}

func parseKontakt(tradeParty *cxpath.Context) (Adresse, error) {
	adr := Adresse{}
	adr.Firmenname = tradeParty.Eval("ram:Name").String()
	adr.EMail = tradeParty.Eval("ram:DefinedTradeContact/ram:EmailURIUniversalCommunication/ram:URIID").String()
	adr.Kontakt = tradeParty.Eval("ram:DefinedTradeContact/ram:PersonName").String()
	adr.PLZ = tradeParty.Eval("ram:PostalTradeAddress/ram:PostcodeCode").String()
	adr.Straße1 = tradeParty.Eval("ram:PostalTradeAddress/ram:LineOne").String()
	adr.Straße2 = tradeParty.Eval("ram:PostalTradeAddress/ram:LineTwo").String()
	adr.Ort = tradeParty.Eval("ram:PostalTradeAddress/ram:CityName").String()
	adr.Ländercode = tradeParty.Eval("ram:PostalTradeAddress/ram:CountryID").String()
	adr.Steuernummer = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='FC']").String()
	adr.UmsatzsteuerID = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='VA']").String()
	return adr, nil
}

func getDecimal(ctx *cxpath.Context, eval string) decimal.Decimal {
	a := ctx.Eval(eval).String()
	str, _ := decimal.NewFromString(a)
	return str
}

// parseCII interprets the XML file as a ZUGFeRD or Factur-X cross industry
// invoice.
func parseCII(cii *cxpath.Context) (*Rechnung, error) {
	var err error
	m := &Rechnung{}

	nc := cii.Eval("rsm:ExchangedDocumentContext").Eval("ram:GuidelineSpecifiedDocumentContextParameter").Eval("ram:ID")
	switch nc.String() {
	case "urn:cen.eu:en16931:2017":
		m.Profil = CZUGFeRD
	}

	exchangedDocument := cii.Eval("rsm:ExchangedDocument")
	m.Rechnungsnummer = exchangedDocument.Eval("ram:ID/text()").String()
	m.Rechnungstyp = CodeInvoice(exchangedDocument.Eval("ram:TypeCode").Int())

	rechnungsdatum, err := parseTime(exchangedDocument.Eval("ram:IssueDateTime/udt:DateTimeString").String())
	if err != nil {
		return nil, err
	}
	m.Belegdatum = rechnungsdatum

	for note := range exchangedDocument.Each("ram:IncludedNote") {
		n := Notiz{}
		n.SubjectCode = note.Eval("ram:SubjectCode").String()
		n.Text = note.Eval("ram:Content").String()
		m.Notizen = append(m.Notizen, n)
	}
	supplyChainTradeTransaction := cii.Eval("rsm:SupplyChainTradeTransaction")
	for lineItem := range supplyChainTradeTransaction.Each("ram:IncludedSupplyChainTradeLineItem") {
		p := Position{}
		p.Position = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:LineID").Int()
		p.Artikelnummer = lineItem.Eval("ram:SpecifiedTradeProduct/ram:SellerAssignedID").String()
		p.ArtikelName = lineItem.Eval("ram:SpecifiedTradeProduct/ram:Name").String()
		p.Anzahl = getDecimal(lineItem, "ram:SpecifiedLineTradeDelivery/ram:BilledQuantity")
		p.Einheit = lineItem.Eval("ram:SpecifiedLineTradeDelivery/ram:BilledQuantity/@unitCode").String()
		p.NettoPreis = getDecimal(lineItem, "ram:SpecifiedLineTradeAgreement/ram:NetPriceProductTradePrice/ram:ChargeAmount")
		p.BruttoPreis = getDecimal(lineItem, "ram:SpecifiedLineTradeAgreement/ram:GrossPriceProductTradePrice/ram:ChargeAmount")
		p.Total = getDecimal(lineItem, "ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeSettlementLineMonetarySummation/ram:LineTotalAmount")
		taxInfo := lineItem.Eval("ram:SpecifiedLineTradeSettlement/ram:ApplicableTradeTax")
		p.SteuerTypCode = taxInfo.Eval("ram:TypeCode").String()
		p.SteuerKategorieCode = taxInfo.Eval("ram:CategoryCode").String()
		p.Steuersatz = getDecimal(taxInfo, "ram:RateApplicablePercent")
		m.Positionen = append(m.Positionen, p)
	}
	m.Käufer, _ = parseKontakt(supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty"))
	m.Verkäufer, _ = parseKontakt(supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty"))
	leistungsdatum := supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeDelivery/ram:ActualDeliverySupplyChainEvent/ram:OccurrenceDateTime/udt:DateTimeString").String()
	m.Leistungsdatum, _ = parseTime(leistungsdatum)

	headerTradeSettlement := supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeSettlement")
	m.Währung = headerTradeSettlement.Eval("ram:InvoiceCurrencyCode").String()
	specifiedTradeSettlementPaymentMeans := headerTradeSettlement.Eval("ram:SpecifiedTradeSettlementPaymentMeans")
	m.BankKontoname = specifiedTradeSettlementPaymentMeans.Eval("ram:PayeePartyCreditorFinancialAccount/ram:AccountName").String()
	m.BankIBAN = specifiedTradeSettlementPaymentMeans.Eval("ram:PayeePartyCreditorFinancialAccount/ram:IBANID").String()
	fälligkeitsdatum := headerTradeSettlement.Eval("ram:SpecifiedTradePaymentTerms/ram:DueDateDateTime/udt:DateTimeString").String()
	if fälligkeitsdatum != "" {
		m.Fälligkeitsdatum, err = parseTime(fälligkeitsdatum)
		if err != nil {
			return nil, err
		}
	}
	for att := range headerTradeSettlement.Each("ram:ApplicableTradeTax") {
		s := Steuersatz{}
		s.BerechneterWert = getDecimal(att, "ram:CalculatedAmount")
		s.BasisWert = getDecimal(att, "ram:BasisAmount")
		s.Typ = att.Eval("ram:TypeCode").String()
		s.Ausnahmegrund = att.Eval("ram:ExemptionReason").String()
		s.KategorieCode = att.Eval("ram:CategoryCode").String()
		s.Prozent = getDecimal(att, "ram:RateApplicablePercent")
		m.Steuersätze = append(m.Steuersätze, s)
	}
	summation := headerTradeSettlement.Eval("ram:SpecifiedTradeSettlementHeaderMonetarySummation")
	m.LineTotal = getDecimal(summation, "ram:LineTotalAmount")
	m.ChargeTotal = getDecimal(summation, "ram:ChargeTotalAmount")
	m.AllowanceTotal = getDecimal(summation, "ram:AllowanceTotalAmount")
	m.TaxBasisTotal = getDecimal(summation, "ram:TaxBasisTotalAmount")
	m.TaxTotal = getDecimal(summation, "ram:TaxTotalAmount")
	m.GrandTotal = getDecimal(summation, "ram:GrandTotalAmount")
	m.TotalPrepaid = getDecimal(summation, "ram:TotalPrepaidAmount")
	m.DuePayableAmount = getDecimal(summation, "ram:DuePayableAmount")

	return m, nil
}

// ParseReader reads the XML from the reader.
func ParseReader(r io.Reader) (*Rechnung, error) {
	ctx, err := cxpath.NewFromReader(r)
	if err != nil {
		return nil, err
	}

	ctx.SetNamespace("rsm", "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100")
	ctx.SetNamespace("ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
	ctx.SetNamespace("udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")

	var m *Rechnung
	cii := ctx.Root()
	rootns := cii.Eval("namespace-uri()").String()
	switch rootns {
	case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
		m, err = parseCII(cii)
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

// ParseXMLFile reads the XML file at filename.
func ParseXMLFile(filename string) (*Rechnung, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return ParseReader(r)
}
