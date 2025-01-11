package einvoice

import (
	"encoding/base64"
	"io"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/cxpath"
)

func parseTime(ctx *cxpath.Context, path string) (time.Time, error) {
	timestring := ctx.Eval(path).String()
	// format := ctx.Eval(path + "/@format").String()
	if timestring == "" {
		return time.Time{}, nil
	}
	return time.Parse("20060102", timestring)
}

func parseParty(tradeParty *cxpath.Context) (Party, error) {
	adr := Party{}
	for id := range tradeParty.Each("ram:ID") {
		adr.ID = append(adr.ID, id.String())
	}
	for gid := range tradeParty.Each("ram:GlobalID") {
		scheme := GlobalID{
			Scheme: gid.Eval("@schemeID").String(),
			ID:     gid.String(),
		}
		adr.GlobalID = append(adr.GlobalID, scheme)
	}

	adr.Name = tradeParty.Eval("ram:Name").String()
	if tradeParty.Eval("count(ram:SpecifiedLegalOrganization) > 0").Bool() {
		slo := SpecifiedLegalOrganization{}
		slo.ID = tradeParty.Eval("ram:SpecifiedLegalOrganization/ram:ID").String()
		slo.Scheme = tradeParty.Eval("ram:SpecifiedLegalOrganization/ram:ID/@schemeID").String()
		slo.TradingBusinessName = tradeParty.Eval("ram:SpecifiedLegalOrganization/ram:TradingBusinessName").String()
		adr.SpecifiedLegalOrganization = &slo
	}

	for dtc := range tradeParty.Each("ram:DefinedTradeContact") {
		contact := DefinedTradeContact{}
		contact.EMail = dtc.Eval("ram:EmailURIUniversalCommunication/ram:URIID").String()
		contact.PhoneNumber = dtc.Eval("ram:TelephoneUniversalCommunication/ram:CompleteNumber").String()
		contact.PersonName = dtc.Eval("ram:PersonName").String()
		adr.DefinedTradeContact = append(adr.DefinedTradeContact, contact)
	}

	if tradeParty.Eval("count(ram:PostalTradeAddress)").Int() > 0 {
		pa := &PostalAddress{
			PostcodeCode: tradeParty.Eval("ram:PostalTradeAddress/ram:PostcodeCode").String(),
			Line1:        tradeParty.Eval("ram:PostalTradeAddress/ram:LineOne").String(),
			Line2:        tradeParty.Eval("ram:PostalTradeAddress/ram:LineTwo").String(),
			Line3:        tradeParty.Eval("ram:PostalTradeAddress/ram:LineThree").String(),
			City:         tradeParty.Eval("ram:PostalTradeAddress/ram:CityName").String(),
			CountryID:    tradeParty.Eval("ram:PostalTradeAddress/ram:CountryID").String(),
		}
		adr.PostalAddress = pa
	}

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
	inv.InvoiceCurrencyCode = applicableHeaderTradeSettlement.Eval("ram:InvoiceCurrencyCode").String()
	// BT-90
	inv.CreditorReferenceID = applicableHeaderTradeSettlement.Eval("ram:CreditorReferenceID").String()
	// BG-10
	if applicableHeaderTradeSettlement.Eval("count(ram:PayeeTradeParty)").Int() > 0 {
		ptp, err := parseParty(applicableHeaderTradeSettlement.Eval("ram:PayeeTradeParty"))
		if err != nil {
			return err
		}
		inv.PayeeTradeParty = &ptp
	}

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
	inv.BillingSpecifiedPeriodStart, _ = parseTime(applicableHeaderTradeSettlement, "ram:BillingSpecifiedPeriod/ram:StartDateTime/udt:DateTimeString")
	inv.BillingSpecifiedPeriodEnd, _ = parseTime(applicableHeaderTradeSettlement, "ram:BillingSpecifiedPeriod/ram:EndDateTime/udt:DateTimeString")

	// ram:SpecifiedTradePaymentTerms
	for paymentTerm := range applicableHeaderTradeSettlement.Each("ram:SpecifiedTradePaymentTerms") {
		spt := SpecifiedTradePaymentTerms{}
		spt.Description = paymentTerm.Eval("ram:Description").String()
		spt.DueDate, err = parseTime(paymentTerm, "ram:DueDateDateTime/udt:DateTimeString")
		if err != nil {
			return err
		}
		spt.Description = paymentTerm.Eval("ram:Description").String()
		spt.DirectDebitMandateID = paymentTerm.Eval("ram:DirectDebitMandateID").String()
		inv.SpecifiedTradePaymentTerms = append(inv.SpecifiedTradePaymentTerms, spt)
	}

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

	// BG-3
	for refdoc := range applicableHeaderTradeSettlement.Each("ram:InvoiceReferencedDocument") {
		rd := ReferencedDocument{}
		rd.Date, err = parseTime(refdoc, "ram:FormattedIssueDateTime/qdt:DateTimeString")
		if err != nil {
			return err
		}
		rd.ID = refdoc.Eval("ram:IssuerAssignedID").String()
		inv.InvoiceReferencedDocument = append(inv.InvoiceReferencedDocument, rd)
	}
	return nil
}

func parseCIIApplicableHeaderTradeDelivery(applicableHeaderTradeDelivery *cxpath.Context, inv *Invoice) error {
	inv.DespatchAdviceReferencedDocument = applicableHeaderTradeDelivery.Eval("ram:DespatchAdviceReferencedDocument").String()
	// BT-72
	inv.OccurrenceDateTime, _ = parseTime(applicableHeaderTradeDelivery, "ram:ActualDeliverySupplyChainEvent/ram:OccurrenceDateTime/udt:DateTimeString")
	if applicableHeaderTradeDelivery.Eval("count(ram:ShipToTradeParty)").Int() > 0 {
		st, _ := parseParty(applicableHeaderTradeDelivery.Eval("ram:ShipToTradeParty"))
		inv.ShipTo = &st
	}

	return nil
}
func parseCIIApplicableHeaderTradeAgreement(applicableHeaderTradeAgreement *cxpath.Context, inv *Invoice) error {
	inv.BuyerReference = applicableHeaderTradeAgreement.Eval("ram:BuyerReference").String()
	// BT-13
	inv.BuyerOrderReferencedDocument = applicableHeaderTradeAgreement.Eval("ram:BuyerOrderReferencedDocument/ram:IssuerAssignedID").String() // BT-13
	// BT-12
	inv.ContractReferencedDocument = applicableHeaderTradeAgreement.Eval("ram:ContractReferencedDocument/ram:IssuerAssignedID").String() // BT-13
	inv.Buyer, _ = parseParty(applicableHeaderTradeAgreement.Eval("ram:BuyerTradeParty"))
	inv.Seller, _ = parseParty(applicableHeaderTradeAgreement.Eval("ram:SellerTradeParty"))

	if applicableHeaderTradeAgreement.Eval("count(ram:SellerTaxRepresentativeTradeParty)").Int() > 0 {
		trp, err := parseParty(applicableHeaderTradeAgreement.Eval("ram:SellerTaxRepresentativeTradeParty"))
		if err != nil {
			return err
		}
		inv.SellerTaxRepresentativeTradeParty = &trp
	}

	for additionalDocument := range applicableHeaderTradeAgreement.Each("ram:AdditionalReferencedDocument") {
		d := Document{}
		d.IssuerAssignedID = additionalDocument.Eval("ram:IssuerAssignedID").String()
		encoded := additionalDocument.Eval("ram:AttachmentBinaryObject").String()
		if encoded != "" {
			data, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return err
			}
			d.AttachmentBinaryObject = data
		}
		d.AttachmentFilename = additionalDocument.Eval("ram:AttachmentBinaryObject/@filename").String()
		d.AttachmentMimeCode = additionalDocument.Eval("ram:AttachmentBinaryObject/@mimeCode").String()
		d.Name = additionalDocument.Eval("ram:Name").String()
		d.TypeCode = additionalDocument.Eval("ram:TypeCode").String()
		d.ReferenceTypeCode = additionalDocument.Eval("ram:ReferenceTypeCode").String()
		inv.AdditionalReferencedDocument = append(inv.AdditionalReferencedDocument, d)
	}
	return nil
}
func parseSepecifiedLineTradeAgreement(specifiedLineTradeAgreement *cxpath.Context, p *InvoiceLine) error {
	p.NetPrice = getDecimal(specifiedLineTradeAgreement, "ram:NetPriceProductTradePrice/ram:ChargeAmount")
	p.GrossPrice = getDecimal(specifiedLineTradeAgreement, "ram:GrossPriceProductTradePrice/ram:ChargeAmount")
	// ZUGFeRD extended has unbound BT-147
	for allowanceCharge := range specifiedLineTradeAgreement.Each("ram:GrossPriceProductTradePrice/ram:AppliedTradeAllowanceCharge") {
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
		if ac.ChargeIndicator {
			p.AppliedTradeAllowanceCharge = append(p.AppliedTradeAllowanceCharge, ac)
		} else {
			p.AppliedTradeAllowanceCharge = append(p.AppliedTradeAllowanceCharge, ac)
		}
	}
	return nil
}

func parseSepecifiedTradeProduct(specifiedTradeProduct *cxpath.Context, p *InvoiceLine) error {
	p.GlobalID = specifiedTradeProduct.Eval("ram:GlobalID").String()
	p.GlobalIDType = specifiedTradeProduct.Eval("ram:GlobalID/@schemeID").String()
	p.ArticleNumber = specifiedTradeProduct.Eval("ram:SellerAssignedID").String()
	p.ArticleNumberBuyer = specifiedTradeProduct.Eval("ram:BuyerAssignedID").String()
	p.ItemName = specifiedTradeProduct.Eval("ram:Name").String()
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
	// BG-25
	for lineItem := range supplyChainTradeTransaction.Each("ram:IncludedSupplyChainTradeLineItem") {
		p := InvoiceLine{}
		p.LineID = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:LineID").String()
		p.Note = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:IncludedNote/ram:Content").String()

		err = parseSepecifiedTradeProduct(lineItem.Eval("ram:SpecifiedTradeProduct"), &p)
		err = parseSepecifiedLineTradeAgreement(lineItem.Eval("ram:SpecifiedLineTradeAgreement"), &p)

		p.BilledQuantity = getDecimal(lineItem, "ram:SpecifiedLineTradeDelivery/ram:BilledQuantity")
		p.BilledQuantityUnit = lineItem.Eval("ram:SpecifiedLineTradeDelivery/ram:BilledQuantity/@unitCode").String()
		if lineItem.Eval("count(ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeSettlementLineMonetarySummation/ram:LineTotalAmount)").Int() > 0 {
			// TODO: add marker for test BR-24
			p.Total = getDecimal(lineItem, "ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeSettlementLineMonetarySummation/ram:LineTotalAmount")
		}
		for allowanceCharge := range lineItem.Each("ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeAllowanceCharge") {
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
			// Im Fall eines Abschlags (BG-27) ist der Wert des ChargeIndicators auf "false" zu setzen.
			// Im Fall eines Zuschlags (BG-28) ist der Wert des ChargeIndicators auf "true" zu setzen.
			if ac.ChargeIndicator {
				p.InvoiceLineCharges = append(p.InvoiceLineCharges, ac)
			} else {
				p.InvoiceLineAllowances = append(p.InvoiceLineAllowances, ac)
			}
		}
		taxInfo := lineItem.Eval("ram:SpecifiedLineTradeSettlement/ram:ApplicableTradeTax")
		// BG-27, BG-28
		p.TaxTypeCode = taxInfo.Eval("ram:TypeCode").String()
		p.TaxCategoryCode = taxInfo.Eval("ram:CategoryCode").String()
		p.TaxRateApplicablePercent = getDecimal(taxInfo, "ram:RateApplicablePercent")
		p.BillingSpecifiedPeriodStart, _ = parseTime(lineItem, "ram:SpecifiedLineTradeSettlement/ram:BillingSpecifiedPeriod/ram:StartDateTime/udt:DateTimeString")
		p.BillingSpecifiedPeriodEnd, _ = parseTime(lineItem, "ram:SpecifiedLineTradeSettlement/ram:BillingSpecifiedPeriod/ram:EndDateTime/udt:DateTimeString")

		inv.InvoiceLines = append(inv.InvoiceLines, p)
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

	invoiceDate, err := parseTime(exchangedDocument, "ram:IssueDateTime/udt:DateTimeString")
	if err != nil {
		return err
	}
	rg.InvoiceDate = invoiceDate

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
	switch nc.String() {
	case "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0":
		rg.Profile = CProfileXRechnung
	case "urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended", "urn:cen.eu:en16931:2017#conformant#urn:zugferd.de:2p0:extended":
		rg.Profile = CProfileExtended
	case "urn:cen.eu:en16931:2017":
		rg.Profile = CProfileEN16931
	case "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic",
		"urn:cen.eu:en16931:2017#compliant#urn:zugferd.de:2p0:basic",
		"urn:cen.eu:en16931:2017:compliant:factur-x.eu:1p0:basic":
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
	ctx.SetNamespace("qdt", "urn:un:unece:uncefact:data:standard:QualifiedDataType:100")

	var m *Invoice
	cii := ctx.Root()
	rootns := cii.Eval("namespace-uri()").String()
	switch rootns {
	case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
		m, err = parseCII(cii)
	}
	if err != nil {
		return nil, err
	}
	m.SchemaType = CII

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
