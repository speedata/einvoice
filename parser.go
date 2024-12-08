package einvoice

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/cxpath"
)

func parseTime(timestring string) (time.Time, error) {
	if timestring == "" {
		return time.Time{}, nil
	}
	return time.Parse("20060102", timestring)
}

func parseKontakt(tradeParty *cxpath.Context) (Party, error) {
	adr := Party{}
	for id := range tradeParty.Each("ram:ID") {
		adr.ID = append(adr.ID, id.String())
	}
	adr.GlobalID = tradeParty.Eval("ram:GlobalID").String()
	adr.GlobalScheme = tradeParty.Eval("ram:GlobalID/@schemeID").String()
	adr.Name = tradeParty.Eval("ram:Name").String()
	adr.EMail = tradeParty.Eval("ram:DefinedTradeContact/ram:EmailURIUniversalCommunication/ram:URIID").String()
	adr.PersonName = tradeParty.Eval("ram:DefinedTradeContact/ram:PersonName").String()
	adr.ZIP = tradeParty.Eval("ram:PostalTradeAddress/ram:PostcodeCode").String()
	adr.Line1 = tradeParty.Eval("ram:PostalTradeAddress/ram:LineOne").String()
	adr.Line2 = tradeParty.Eval("ram:PostalTradeAddress/ram:LineTwo").String()
	adr.Line3 = tradeParty.Eval("ram:PostalTradeAddress/ram:LineThree").String()
	adr.City = tradeParty.Eval("ram:PostalTradeAddress/ram:CityName").String()
	adr.CountryID = tradeParty.Eval("ram:PostalTradeAddress/ram:CountryID").String()
	adr.FCTaxRegistration = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='FC']").String()
	adr.VATaxRegistration = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='VA']").String()
	return adr, nil
}

func getDecimal(ctx *cxpath.Context, eval string) decimal.Decimal {
	a := ctx.Eval(eval).String()
	str, _ := decimal.NewFromString(a)
	return str
}

func parseCIIApplicableHeaderTradeSettlement(applicableHeaderTradeSettlement *cxpath.Context, inv *Invoice) error {
	var err error
	inv.Currency = applicableHeaderTradeSettlement.Eval("ram:InvoiceCurrencyCode").String()
	for pm := range applicableHeaderTradeSettlement.Each("ram:SpecifiedTradeSettlementPaymentMeans") {
		bd := PaymentMeans{
			TypeCode:                                             pm.Eval("ram:TypeCode").Int(),
			Information:                                          pm.Eval("ram:Information").String(),
			PayeePartyCreditorFinancialAccountIBAN:               pm.Eval("ram:PayeePartyCreditorFinancialAccount/ram:IBANID").String(),
			PayeePartyCreditorFinancialAccountName:               pm.Eval("ram:PayeePartyCreditorFinancialAccount/ram:AccountName").String(),
			PayeePartyCreditorFinancialAccountProprietaryID:      pm.Eval("ram:PayeePartyCreditorFinancialAccount/ram:ProprietaryID").String(),
			PayeeSpecifiedCreditorFinancialInstitutionBIC:        pm.Eval("ram:PayeeSpecifiedCreditorFinancialInstitution/ram:BICID").String(),
			PayerPartyDebtorFinancialAccountIBAN:                 pm.Eval("ram:PayerPartyDebtorFinancialAccount/ram:IBANID").String(),
			ApplicableTradeSettlementFinancialCardID:             pm.Eval("ram:ApplicableTradeSettlementFinancialCard/ram:ID").String(),
			ApplicableTradeSettlementFinancialCardCardholderName: pm.Eval("ram:ApplicableTradeSettlementFinancialCard/ram:CardholderName").String(),
		}
		inv.PaymentMeans = append(inv.PaymentMeans, bd)
	}

	for allowanceCharge := range applicableHeaderTradeSettlement.Each("ram:SpecifiedTradeAllowanceCharge") {
		ac := AllowanceCharge{
			ChargeIndicator:                       allowanceCharge.Eval("string(ram:ChargeIndicator/udt:Indicator) = 'true'").Bool(),
			BasisAmount:                           getDecimal(allowanceCharge, "ram:BasisAmount"),
			ActualAmount:                          getDecimal(allowanceCharge, "ram:ActualAmount"),
			CalculationPercent:                    getDecimal(allowanceCharge, "ram:CalculationPercent"),
			ReasonCode:                            allowanceCharge.Eval("ram:ReasonCode").Int(),
			Reason:                                allowanceCharge.Eval("ram:Reason").String(),
			CategoryTradeTaxType:                  allowanceCharge.Eval("ram:CategoryTradeTax/ram:TypeCode").String(),
			CategoryTradeTaxCategoryCode:          allowanceCharge.Eval("ram:CategoryTradeTax/ram:CategoryCode").String(),
			CategoryTradeTaxRateApplicablePercent: getDecimal(allowanceCharge, "ram:CategoryTradeTax/ram:RateApplicablePercent"),
		}
		inv.SpecifiedTradeAllowanceCharge = append(inv.SpecifiedTradeAllowanceCharge, ac)

	}
	inv.BillingSpecifiedPeriodStart, _ = parseTime(applicableHeaderTradeSettlement.Eval("ram:BillingSpecifiedPeriod/ram:StartDateTime/udt:DateTimeString").String())
	inv.BillingSpecifiedPeriodEnd, _ = parseTime(applicableHeaderTradeSettlement.Eval("ram:BillingSpecifiedPeriod/ram:EndDateTime/udt:DateTimeString").String())

	// ram:SpecifiedTradePaymentTerms
	dueDate := applicableHeaderTradeSettlement.Eval("ram:SpecifiedTradePaymentTerms/ram:DueDateDateTime/udt:DateTimeString").String()
	if dueDate != "" {
		inv.DueDate, err = parseTime(dueDate)
		if err != nil {
			return err
		}
	}
	inv.TradePaymentTermsDescription = applicableHeaderTradeSettlement.Eval("ram:SpecifiedTradePaymentTerms/ram:Description").String()
	inv.DirectDebitMandateID = applicableHeaderTradeSettlement.Eval("ram:SpecifiedTradePaymentTerms/ram:DirectDebitMandateID").String()

	for att := range applicableHeaderTradeSettlement.Each("ram:ApplicableTradeTax") {
		s := TradeTax{}
		s.CalculatedAmount = getDecimal(att, "ram:CalculatedAmount")
		s.BasisAmount = getDecimal(att, "ram:BasisAmount")
		s.Typ = att.Eval("ram:TypeCode").String()
		s.ExemptionReason = att.Eval("ram:ExemptionReason").String()
		s.CategoryCode = att.Eval("ram:CategoryCode").String()
		s.Percent = getDecimal(att, "ram:RateApplicablePercent")
		inv.TradeTaxes = append(inv.TradeTaxes, s)
	}
	summation := applicableHeaderTradeSettlement.Eval("ram:SpecifiedTradeSettlementHeaderMonetarySummation")
	inv.LineTotal = getDecimal(summation, "ram:LineTotalAmount")
	inv.ChargeTotal = getDecimal(summation, "ram:ChargeTotalAmount")
	inv.AllowanceTotal = getDecimal(summation, "ram:AllowanceTotalAmount")
	inv.TaxBasisTotal = getDecimal(summation, "ram:TaxBasisTotalAmount")
	inv.TaxTotalCurrency = summation.Eval("ram:TaxTotalAmount/@currencyID").String()
	inv.TaxTotal = getDecimal(summation, "ram:TaxTotalAmount")
	inv.GrandTotal = getDecimal(summation, "ram:GrandTotalAmount")
	inv.TotalPrepaid = getDecimal(summation, "ram:TotalPrepaidAmount")
	inv.DuePayableAmount = getDecimal(summation, "ram:DuePayableAmount")

	return nil
}

func parseCIIApplicableHeaderTradeDelivery(applicableHeaderTradeDelivery *cxpath.Context, inv *Invoice) error {
	inv.DespatchAdviceReferencedDocument = applicableHeaderTradeDelivery.Eval("ram:DespatchAdviceReferencedDocument").String()
	leistungsdatum := applicableHeaderTradeDelivery.Eval("ram:ActualDeliverySupplyChainEvent/ram:OccurrenceDateTime/udt:DateTimeString").String()
	inv.OccurrenceDateTime, _ = parseTime(leistungsdatum)
	if applicableHeaderTradeDelivery.Eval("count(ram:ShipToTradeParty)").Int() > 0 {
		st, _ := parseKontakt(applicableHeaderTradeDelivery.Eval("ram:ShipToTradeParty"))
		inv.ShipTo = &st
	}

	return nil
}
func parseCIIApplicableHeaderTradeAgreement(applicableHeaderTradeAgreement *cxpath.Context, inv *Invoice) error {
	inv.BuyerReference = applicableHeaderTradeAgreement.Eval("ram:BuyerReference").String()
	inv.BuyerOrderReferencedDocument = applicableHeaderTradeAgreement.Eval("ram:BuyerOrderReferencedDocument/ram:IssuerAssignedID").String()
	inv.Buyer, _ = parseKontakt(applicableHeaderTradeAgreement.Eval("ram:BuyerTradeParty"))
	inv.Seller, _ = parseKontakt(applicableHeaderTradeAgreement.Eval("ram:SellerTradeParty"))
	/*
			    	<ram:SellerTaxRepresentativeTradeParty></ram:SellerTaxRepresentativeTradeParty>
		            <ram:SellerOrderReferencedDocument></ram:SellerOrderReferencedDocument>
		            <ram:BuyerOrderReferencedDocument></ram:BuyerOrderReferencedDocument>
		            <ram:ContractReferencedDocument></ram:ContractReferencedDocument>
		            <ram:AdditionalReferencedDocument></ram:AdditionalReferencedDocument>
		            <ram:SpecifiedProcuringProject>
		                <ram:ID></ram:ID>
		                <ram:Name></ram:Name>
		            </ram:SpecifiedProcuringProject>
	*/
	return nil
}
func parseSepecifiedLineTradeAgreement(specifiedLineTradeAgreement *cxpath.Context, p *InvoiceItem) error {
	p.NetPrice = getDecimal(specifiedLineTradeAgreement, "ram:NetPriceProductTradePrice/ram:ChargeAmount")
	p.GrossPrice = getDecimal(specifiedLineTradeAgreement, "ram:GrossPriceProductTradePrice/ram:ChargeAmount")

	return nil
}

func parseSepecifiedTradeProduct(specifiedTradeProduct *cxpath.Context, p *InvoiceItem) error {
	p.GlobalID = specifiedTradeProduct.Eval("ram:GlobalID").String()
	p.GlobalIDType = CodeGlobalID(specifiedTradeProduct.Eval("ram:GlobalID/@schemeID").Int())
	p.ArticleNumber = specifiedTradeProduct.Eval("ram:SellerAssignedID").String()
	p.ArticleNumberBuyer = specifiedTradeProduct.Eval("ram:BuyerAssignedID").String()
	p.ArticleName = specifiedTradeProduct.Eval("ram:Name").String()
	p.Description = specifiedTradeProduct.Eval("ram:Description").String()
	for itm := range specifiedTradeProduct.Each("ram:ApplicableProductCharacteristic") {
		ch := Characteristic{
			Description: itm.Eval("ram:Description").String(),
			Value:       itm.Eval("ram:Value").String(),
		}
		p.Characteristics = append(p.Characteristics, ch)
	}
	for itm := range specifiedTradeProduct.Each("ram:DesignatedProductClassification") {
		ch := Classification{
			ClassCode:     itm.Eval("ram:ClassCode").String(),
			ListID:        itm.Eval("ram:ClassCode/@listID").String(),
			ListVersionID: itm.Eval("ram:ClassCode/@listVersionID").String(),
		}
		p.ProductClassification = append(p.ProductClassification, ch)
	}
	p.OriginTradeCountry = specifiedTradeProduct.Eval("ram:OriginTradeCountry/ram:ID").String()
	return nil
}

func parseCIISupplyChainTradeTransaction(supplyChainTradeTransaction *cxpath.Context, inv *Invoice) error {
	var err error
	for lineItem := range supplyChainTradeTransaction.Each("ram:IncludedSupplyChainTradeLineItem") {
		p := InvoiceItem{}
		p.Position = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:LineID").Int()
		p.Note = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:IncludedNote/ram:Content").String()

		err = parseSepecifiedTradeProduct(lineItem.Eval("ram:SpecifiedTradeProduct"), &p)
		err = parseSepecifiedLineTradeAgreement(lineItem.Eval("ram:SpecifiedLineTradeAgreement"), &p)

		p.BilledQuantity = getDecimal(lineItem, "ram:SpecifiedLineTradeDelivery/ram:BilledQuantity")
		p.Unit = lineItem.Eval("ram:SpecifiedLineTradeDelivery/ram:BilledQuantity/@unitCode").String()
		p.Total = getDecimal(lineItem, "ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeSettlementLineMonetarySummation/ram:LineTotalAmount")
		taxInfo := lineItem.Eval("ram:SpecifiedLineTradeSettlement/ram:ApplicableTradeTax")
		p.TaxTypeCode = taxInfo.Eval("ram:TypeCode").String()
		p.TaxCategoryCode = taxInfo.Eval("ram:CategoryCode").String()
		p.TaxRateApplicablePercent = getDecimal(taxInfo, "ram:RateApplicablePercent")
		inv.InvoiceItems = append(inv.InvoiceItems, p)
	}
	if err = parseCIIApplicableHeaderTradeAgreement(supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeAgreement"), inv); err != nil {
		return err
	}
	if err = parseCIIApplicableHeaderTradeDelivery(supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeDelivery"), inv); err != nil {
		return err
	}
	if err = parseCIIApplicableHeaderTradeSettlement(supplyChainTradeTransaction.Eval("ram:ApplicableHeaderTradeSettlement"), inv); err != nil {
		return err
	}

	return nil
}

func parseCIIExchangedDocument(exchangedDocument *cxpath.Context, rg *Invoice) error {
	rg.InvoiceNumber = exchangedDocument.Eval("ram:ID/text()").String()
	rg.InvoiceTypeCode = CodeDocument(exchangedDocument.Eval("ram:TypeCode").Int())

	rechnungsdatum, err := parseTime(exchangedDocument.Eval("ram:IssueDateTime/udt:DateTimeString").String())
	if err != nil {
		return err
	}
	rg.InvoiceDate = rechnungsdatum

	for note := range exchangedDocument.Each("ram:IncludedNote") {
		n := Note{}
		n.SubjectCode = note.Eval("ram:SubjectCode").String()
		n.Text = note.Eval("ram:Content").String()
		rg.Notes = append(rg.Notes, n)
	}

	return nil
}

func parseCIIExchangedDocumentContext(ctx *cxpath.Context, rg *Invoice) error {
	nc := ctx.Eval("ram:GuidelineSpecifiedDocumentContextParameter").Eval("ram:ID")
	fmt.Println(`~~> nc.String()`, nc.String())
	switch nc.String() {
	case "urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended", "urn:cen.eu:en16931:2017#conformant#urn:zugferd.de:2p0:extended":
		rg.Profile = CProfileExtended
	case "urn:cen.eu:en16931:2017":
		rg.Profile = CProfileEN16931
	case "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic", "urn:cen.eu:en16931:2017#compliant#urn:zugferd.de:2p0:basic":
		rg.Profile = CProfileBasic
	case "urn:factur-x.eu:1p0:basicwl":
		rg.Profile = CProfileBasicWL
	case "urn:factur-x.eu:1p0:minimum", "urn:zugferd.de:2p0:minimum":
		rg.Profile = CProfileMinimum
	}
	rg.BPSpecifiedDocumentContextParameter = ctx.Eval("ram:BusinessProcessSpecifiedDocumentContextParameter/ram:ID").String()
	return nil
}

// parseCII interprets the XML file as a ZUGFeRD or Factur-X cross industry
// invoice.
func parseCII(cii *cxpath.Context) (*Invoice, error) {
	var err error
	inv := &Invoice{}

	err = parseCIIExchangedDocumentContext(cii.Eval("rsm:ExchangedDocumentContext"), inv)
	if err != nil {
		return nil, err
	}
	if err = parseCIIExchangedDocument(cii.Eval("rsm:ExchangedDocument"), inv); err != nil {
		return nil, err
	}
	if err = parseCIISupplyChainTradeTransaction(cii.Eval("rsm:SupplyChainTradeTransaction"), inv); err != nil {
		return nil, err
	}

	return inv, nil
}

// ParseReader reads the XML from the reader.
func ParseReader(r io.Reader) (*Invoice, error) {
	ctx, err := cxpath.NewFromReader(r)
	if err != nil {
		return nil, err
	}

	ctx.SetNamespace("rsm", "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100")
	ctx.SetNamespace("ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
	ctx.SetNamespace("udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")

	var m *Invoice
	cii := ctx.Root()
	rootns := cii.Eval("namespace-uri()").String()
	switch rootns {
	case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
		m, err = parseCII(cii)
		m.SchemaType = CII
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

// ParseXMLFile reads the XML file at filename.
func ParseXMLFile(filename string) (*Invoice, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return ParseReader(r)
}
