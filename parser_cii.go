package einvoice

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/speedata/cxpath"
)

// CII (ZUGFeRD/Factur-X) namespace URN for root element
const nsCIIRootInvoice = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"

// parseCIITime parses CII format dates (YYYYMMDD) into time.Time.
func parseCIITime(ctx *cxpath.Context, path string) (time.Time, error) {
	timestring := ctx.Eval(path).String()
	if timestring == "" {
		return time.Time{}, nil
	}

	parsedDate, err := time.Parse("20060102", timestring)
	if err != nil {
		return parsedDate, fmt.Errorf("%w", err)
	}

	return parsedDate, nil
}

// parseCIIParty parses a party (buyer, seller, payee, etc.) from CII format.
// Uses CII-specific XPath with ram: namespace prefixes.
func parseCIIParty(tradeParty *cxpath.Context) Party {
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
		postalAddress := &PostalAddress{
			PostcodeCode:           tradeParty.Eval("ram:PostalTradeAddress/ram:PostcodeCode").String(),
			Line1:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineOne").String(),
			Line2:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineTwo").String(),
			Line3:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineThree").String(),
			City:                   tradeParty.Eval("ram:PostalTradeAddress/ram:CityName").String(),
			CountryID:              tradeParty.Eval("ram:PostalTradeAddress/ram:CountryID").String(),
			CountrySubDivisionName: tradeParty.Eval("ram:PostalTradeAddress/ram:CountrySubDivisionName").String(),
		}
		adr.PostalAddress = postalAddress
	}

	adr.FCTaxRegistration = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='FC']").String()
	adr.VATaxRegistration = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='VA']").String()

	return adr
}

// parseCII interprets the XML file as a ZUGFeRD or Factur-X cross industry invoice.
// It sets up CII-specific namespaces and parses the document structure.
func parseCII(ctx *cxpath.Context) (*Invoice, error) {
	// Setup CII namespaces
	ctx.SetNamespace("rsm", nsCIIRootInvoice)
	ctx.SetNamespace("ram", "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100")
	ctx.SetNamespace("udt", "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100")
	ctx.SetNamespace("qdt", "urn:un:unece:uncefact:data:standard:QualifiedDataType:100")

	// Get root element after namespace setup
	root := ctx.Root()

	inv := &Invoice{SchemaType: CII}

	var err error
	if err = parseCIIExchangedDocumentContext(root.Eval("rsm:ExchangedDocumentContext"), inv); err != nil {
		return nil, err
	}

	if err = parseCIIExchangedDocument(root.Eval("rsm:ExchangedDocument"), inv); err != nil {
		return nil, err
	}

	if err = parseCIISupplyChainTradeTransaction(root.Eval("rsm:SupplyChainTradeTransaction"), inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func parseCIIExchangedDocumentContext(ctx *cxpath.Context, inv *Invoice) error {
	// Store the raw URN value (BT-24 - Specification identifier)
	nc := ctx.Eval("ram:GuidelineSpecifiedDocumentContextParameter").Eval("ram:ID")
	inv.GuidelineSpecifiedDocumentContextParameter = nc.String()

	// Store the business process identifier (BT-23)
	inv.BPSpecifiedDocumentContextParameter = ctx.Eval("ram:BusinessProcessSpecifiedDocumentContextParameter/ram:ID").String()

	return nil
}

func parseCIIExchangedDocument(exchangedDocument *cxpath.Context, inv *Invoice) error {
	inv.InvoiceNumber = exchangedDocument.Eval("ram:ID/text()").String()
	inv.InvoiceTypeCode = CodeDocument(exchangedDocument.Eval("ram:TypeCode").Int())

	invoiceDate, err := parseCIITime(exchangedDocument, "ram:IssueDateTime/udt:DateTimeString")
	if err != nil {
		return err
	}

	inv.InvoiceDate = invoiceDate

	for note := range exchangedDocument.Each("ram:IncludedNote") {
		n := Note{}
		n.SubjectCode = note.Eval("ram:SubjectCode").String()
		n.Text = note.Eval("ram:Content").String()
		inv.Notes = append(inv.Notes, n)
	}

	return nil
}

func parseCIISupplyChainTradeTransaction(supplyChainTradeTransaction *cxpath.Context, inv *Invoice) error {
	var err error
	// BG-25
	for lineItem := range supplyChainTradeTransaction.Each("ram:IncludedSupplyChainTradeLineItem") {
		invoiceLine := InvoiceLine{}
		invoiceLine.LineID = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:LineID").String()
		invoiceLine.Note = lineItem.Eval("ram:AssociatedDocumentLineDocument/ram:IncludedNote/ram:Content").String()

		parseSpecifiedTradeProduct(lineItem.Eval("ram:SpecifiedTradeProduct"), &invoiceLine)
		specifiedLineTradeAgreement := lineItem.Eval("ram:SpecifiedLineTradeAgreement")
		if err = parseSpecifiedLineTradeAgreement(specifiedLineTradeAgreement, &invoiceLine); err != nil {
			return err
		}

		invoiceLine.BilledQuantity, err = getDecimal(lineItem, "ram:SpecifiedLineTradeDelivery/ram:BilledQuantity")
		if err != nil {
			return err
		}
		invoiceLine.BilledQuantityUnit = lineItem.Eval("ram:SpecifiedLineTradeDelivery/ram:BilledQuantity/@unitCode").String()
		// BR-24: Track XML element presence to validate later
		invoiceLine.hasLineTotalInXML = lineItem.Eval("count(ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeSettlementLineMonetarySummation/ram:LineTotalAmount)").Int() > 0
		invoiceLine.Total, err = getDecimal(lineItem, "ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeSettlementLineMonetarySummation/ram:LineTotalAmount")
		if err != nil {
			return err
		}

		for allowanceCharge := range lineItem.Each("ram:SpecifiedLineTradeSettlement/ram:SpecifiedTradeAllowanceCharge") {
			basisAmount, err := getDecimal(allowanceCharge, "ram:BasisAmount")
			if err != nil {
				return err
			}
			actualAmount, err := getDecimal(allowanceCharge, "ram:ActualAmount")
			if err != nil {
				return err
			}
			calculationPercent, err := getDecimal(allowanceCharge, "ram:CalculationPercent")
			if err != nil {
				return err
			}
			categoryTaxRate, err := getDecimal(allowanceCharge, "ram:CategoryTradeTax/ram:RateApplicablePercent")
			if err != nil {
				return err
			}

			alc := AllowanceCharge{
				ChargeIndicator:                       allowanceCharge.Eval("string(ram:ChargeIndicator/udt:Indicator) = 'true'").Bool(),
				BasisAmount:                           basisAmount,
				ActualAmount:                          actualAmount,
				CalculationPercent:                    calculationPercent,
				ReasonCode:                            allowanceCharge.Eval("ram:ReasonCode").Int(),
				Reason:                                allowanceCharge.Eval("ram:Reason").String(),
				CategoryTradeTaxType:                  allowanceCharge.Eval("ram:CategoryTradeTax/ram:TypeCode").String(),
				CategoryTradeTaxCategoryCode:          allowanceCharge.Eval("ram:CategoryTradeTax/ram:CategoryCode").String(),
				CategoryTradeTaxRateApplicablePercent: categoryTaxRate,
			}
			// Im Fall eines Abschlags (BG-27) ist der Wert des ChargeIndicators auf "false" zu setzen.
			// Im Fall eines Zuschlags (BG-28) ist der Wert des ChargeIndicators auf "true" zu setzen.
			if alc.ChargeIndicator {
				invoiceLine.InvoiceLineCharges = append(invoiceLine.InvoiceLineCharges, alc)
			} else {
				invoiceLine.InvoiceLineAllowances = append(invoiceLine.InvoiceLineAllowances, alc)
			}
		}

		taxInfo := lineItem.Eval("ram:SpecifiedLineTradeSettlement/ram:ApplicableTradeTax")
		// BG-27, BG-28
		invoiceLine.TaxTypeCode = taxInfo.Eval("ram:TypeCode").String()
		invoiceLine.TaxCategoryCode = taxInfo.Eval("ram:CategoryCode").String()
		invoiceLine.TaxRateApplicablePercent, err = getDecimal(taxInfo, "ram:RateApplicablePercent")
		if err != nil {
			return err
		}
		invoiceLine.BillingSpecifiedPeriodStart, err = parseCIITime(lineItem, "ram:SpecifiedLineTradeSettlement/ram:BillingSpecifiedPeriod/ram:StartDateTime/udt:DateTimeString")
		if err != nil {
			return fmt.Errorf("invalid line billing period start date for line %s: %w", invoiceLine.LineID, err)
		}
		invoiceLine.BillingSpecifiedPeriodEnd, err = parseCIITime(lineItem, "ram:SpecifiedLineTradeSettlement/ram:BillingSpecifiedPeriod/ram:EndDateTime/udt:DateTimeString")
		if err != nil {
			return fmt.Errorf("invalid line billing period end date for line %s: %w", invoiceLine.LineID, err)
		}

		inv.InvoiceLines = append(inv.InvoiceLines, invoiceLine)
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

func parseCIIApplicableHeaderTradeAgreement(applicableHeaderTradeAgreement *cxpath.Context, inv *Invoice) error {
	inv.BuyerReference = applicableHeaderTradeAgreement.Eval("ram:BuyerReference").String()
	// BT-13
	inv.BuyerOrderReferencedDocument = applicableHeaderTradeAgreement.Eval("ram:BuyerOrderReferencedDocument/ram:IssuerAssignedID").String() // BT-13
	// BT-12
	inv.ContractReferencedDocument = applicableHeaderTradeAgreement.Eval("ram:ContractReferencedDocument/ram:IssuerAssignedID").String() // BT-13
	inv.Buyer = parseCIIParty(applicableHeaderTradeAgreement.Eval("ram:BuyerTradeParty"))
	inv.Seller = parseCIIParty(applicableHeaderTradeAgreement.Eval("ram:SellerTradeParty"))

	if applicableHeaderTradeAgreement.Eval("count(ram:SellerTaxRepresentativeTradeParty)").Int() > 0 {
		trp := parseCIIParty(applicableHeaderTradeAgreement.Eval("ram:SellerTaxRepresentativeTradeParty"))
		inv.SellerTaxRepresentativeTradeParty = &trp
	}

	for additionalDocument := range applicableHeaderTradeAgreement.Each("ram:AdditionalReferencedDocument") {
		doc := Document{}
		doc.IssuerAssignedID = additionalDocument.Eval("ram:IssuerAssignedID").String()
		encoded := additionalDocument.Eval("ram:AttachmentBinaryObject").String()

		if encoded != "" {
			data, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return fmt.Errorf("cannot decode attachment %w", err)
			}

			doc.AttachmentBinaryObject = data
		}

		doc.AttachmentFilename = additionalDocument.Eval("ram:AttachmentBinaryObject/@filename").String()
		doc.AttachmentMimeCode = additionalDocument.Eval("ram:AttachmentBinaryObject/@mimeCode").String()
		doc.Name = additionalDocument.Eval("ram:Name").String()
		doc.TypeCode = additionalDocument.Eval("ram:TypeCode").String()
		doc.ReferenceTypeCode = additionalDocument.Eval("ram:ReferenceTypeCode").String()
		inv.AdditionalReferencedDocument = append(inv.AdditionalReferencedDocument, doc)
	}

	return nil
}

func parseCIIApplicableHeaderTradeDelivery(applicableHeaderTradeDelivery *cxpath.Context, inv *Invoice) error {
	inv.DespatchAdviceReferencedDocument = applicableHeaderTradeDelivery.Eval("ram:DespatchAdviceReferencedDocument").String()
	// BT-72
	var err error
	inv.OccurrenceDateTime, err = parseCIITime(applicableHeaderTradeDelivery, "ram:ActualDeliverySupplyChainEvent/ram:OccurrenceDateTime/udt:DateTimeString")
	if err != nil {
		return fmt.Errorf("invalid occurrence date time: %w", err)
	}

	if applicableHeaderTradeDelivery.Eval("count(ram:ShipToTradeParty)").Int() > 0 {
		st := parseCIIParty(applicableHeaderTradeDelivery.Eval("ram:ShipToTradeParty"))
		inv.ShipTo = &st
	}
	return nil
}

func parseCIIApplicableHeaderTradeSettlement(applicableHeaderTradeSettlement *cxpath.Context, inv *Invoice) error {
	var err error

	inv.InvoiceCurrencyCode = applicableHeaderTradeSettlement.Eval("ram:InvoiceCurrencyCode").String()
	// BT-6: Tax currency code (accounting currency, if different from invoice currency)
	inv.TaxCurrencyCode = applicableHeaderTradeSettlement.Eval("ram:TaxCurrencyCode").String()
	// BT-90
	inv.CreditorReferenceID = applicableHeaderTradeSettlement.Eval("ram:CreditorReferenceID").String()
	// BG-10
	if applicableHeaderTradeSettlement.Eval("count(ram:PayeeTradeParty)").Int() > 0 {
		ptp := parseCIIParty(applicableHeaderTradeSettlement.Eval("ram:PayeeTradeParty"))
		inv.PayeeTradeParty = &ptp
	}

	for paymentMeans := range applicableHeaderTradeSettlement.Each("ram:SpecifiedTradeSettlementPaymentMeans") {
		// BG-16
		thisPaymentMeans := PaymentMeans{
			TypeCode:                                             paymentMeans.Eval("ram:TypeCode").Int(),
			Information:                                          paymentMeans.Eval("ram:Information").String(),
			PayeePartyCreditorFinancialAccountIBAN:               paymentMeans.Eval("ram:PayeePartyCreditorFinancialAccount/ram:IBANID").String(),
			PayeePartyCreditorFinancialAccountName:               paymentMeans.Eval("ram:PayeePartyCreditorFinancialAccount/ram:AccountName").String(),
			PayeePartyCreditorFinancialAccountProprietaryID:      paymentMeans.Eval("ram:PayeePartyCreditorFinancialAccount/ram:ProprietaryID").String(),
			PayeeSpecifiedCreditorFinancialInstitutionBIC:        paymentMeans.Eval("ram:PayeeSpecifiedCreditorFinancialInstitution/ram:BICID").String(),
			PayerPartyDebtorFinancialAccountIBAN:                 paymentMeans.Eval("ram:PayerPartyDebtorFinancialAccount/ram:IBANID").String(),
			ApplicableTradeSettlementFinancialCardID:             paymentMeans.Eval("ram:ApplicableTradeSettlementFinancialCard/ram:ID").String(),
			ApplicableTradeSettlementFinancialCardCardholderName: paymentMeans.Eval("ram:ApplicableTradeSettlementFinancialCard/ram:CardholderName").String(),
		}
		inv.PaymentMeans = append(inv.PaymentMeans, thisPaymentMeans)
	}

	for allowanceCharge := range applicableHeaderTradeSettlement.Each("ram:SpecifiedTradeAllowanceCharge") {
		basisAmount, err := getDecimal(allowanceCharge, "ram:BasisAmount")
		if err != nil {
			return err
		}
		actualAmount, err := getDecimal(allowanceCharge, "ram:ActualAmount")
		if err != nil {
			return err
		}
		calculationPercent, err := getDecimal(allowanceCharge, "ram:CalculationPercent")
		if err != nil {
			return err
		}
		categoryTaxRate, err := getDecimal(allowanceCharge, "ram:CategoryTradeTax/ram:RateApplicablePercent")
		if err != nil {
			return err
		}

		allowanceCharge := AllowanceCharge{
			ChargeIndicator:                       allowanceCharge.Eval("string(ram:ChargeIndicator/udt:Indicator) = 'true'").Bool(),
			BasisAmount:                           basisAmount,
			ActualAmount:                          actualAmount,
			CalculationPercent:                    calculationPercent,
			ReasonCode:                            allowanceCharge.Eval("ram:ReasonCode").Int(),
			Reason:                                allowanceCharge.Eval("ram:Reason").String(),
			CategoryTradeTaxType:                  allowanceCharge.Eval("ram:CategoryTradeTax/ram:TypeCode").String(),
			CategoryTradeTaxCategoryCode:          allowanceCharge.Eval("ram:CategoryTradeTax/ram:CategoryCode").String(),
			CategoryTradeTaxRateApplicablePercent: categoryTaxRate,
		}
		inv.SpecifiedTradeAllowanceCharge = append(inv.SpecifiedTradeAllowanceCharge, allowanceCharge)
	}
	inv.BillingSpecifiedPeriodStart, err = parseCIITime(applicableHeaderTradeSettlement, "ram:BillingSpecifiedPeriod/ram:StartDateTime/udt:DateTimeString")
	if err != nil {
		return fmt.Errorf("invalid billing period start date: %w", err)
	}
	inv.BillingSpecifiedPeriodEnd, err = parseCIITime(applicableHeaderTradeSettlement, "ram:BillingSpecifiedPeriod/ram:EndDateTime/udt:DateTimeString")
	if err != nil {
		return fmt.Errorf("invalid billing period end date: %w", err)
	}

	// ram:SpecifiedTradePaymentTerms
	for paymentTerm := range applicableHeaderTradeSettlement.Each("ram:SpecifiedTradePaymentTerms") {
		spt := SpecifiedTradePaymentTerms{}
		spt.Description = paymentTerm.Eval("ram:Description").String()
		spt.DueDate, err = parseCIITime(paymentTerm, "ram:DueDateDateTime/udt:DateTimeString")

		if err != nil {
			return err
		}

		spt.DirectDebitMandateID = paymentTerm.Eval("ram:DirectDebitMandateID").String()
		inv.SpecifiedTradePaymentTerms = append(inv.SpecifiedTradePaymentTerms, spt)
	}

	for att := range applicableHeaderTradeSettlement.Each("ram:ApplicableTradeTax") {
		tradeTax := TradeTax{}
		tradeTax.CalculatedAmount, err = getDecimal(att, "ram:CalculatedAmount")
		if err != nil {
			return err
		}
		tradeTax.BasisAmount, err = getDecimal(att, "ram:BasisAmount")
		if err != nil {
			return err
		}
		tradeTax.TypeCode = att.Eval("ram:TypeCode").String()
		tradeTax.ExemptionReason = att.Eval("ram:ExemptionReason").String()
		tradeTax.CategoryCode = att.Eval("ram:CategoryCode").String()
		tradeTax.Percent, err = getDecimal(att, "ram:RateApplicablePercent") // BT-119
		if err != nil {
			return err
		}
		inv.TradeTaxes = append(inv.TradeTaxes, tradeTax)
	}

	summation := applicableHeaderTradeSettlement.Eval("ram:SpecifiedTradeSettlementHeaderMonetarySummation")

	// BR-12 through BR-15: Track XML element presence to validate later
	// This allows validation to distinguish between missing elements and zero values
	inv.hasLineTotalInXML = summation.Eval("count(ram:LineTotalAmount)").Int() > 0
	inv.hasTaxBasisTotalInXML = summation.Eval("count(ram:TaxBasisTotalAmount)").Int() > 0
	inv.hasGrandTotalInXML = summation.Eval("count(ram:GrandTotalAmount)").Int() > 0
	inv.hasDuePayableAmountInXML = summation.Eval("count(ram:DuePayableAmount)").Int() > 0

	inv.LineTotal, err = getDecimal(summation, "ram:LineTotalAmount")
	if err != nil {
		return err
	}
	inv.ChargeTotal, err = getDecimal(summation, "ram:ChargeTotalAmount")
	if err != nil {
		return err
	}
	inv.AllowanceTotal, err = getDecimal(summation, "ram:AllowanceTotalAmount")
	if err != nil {
		return err
	}
	inv.TaxBasisTotal, err = getDecimal(summation, "ram:TaxBasisTotalAmount")
	if err != nil {
		return err
	}

	// BT-110 and BT-111: Parse TaxTotalAmount by matching currencyID (not position)
	// EN 16931 specifies which currency each total must be in, regardless of XML order
	for taxTotal := range summation.Each("ram:TaxTotalAmount") {
		currency := taxTotal.Eval("@currencyID").String()
		amount, err := getDecimal(taxTotal, ".")
		if err != nil {
			return fmt.Errorf("invalid TaxTotalAmount with currency %s: %w", currency, err)
		}

		// BT-110: Tax total in invoice currency (must match BT-5)
		if currency == inv.InvoiceCurrencyCode {
			inv.TaxTotalCurrency = currency
			inv.TaxTotal = amount
		} else if inv.TaxCurrencyCode != "" && currency == inv.TaxCurrencyCode {
			// BT-111: Tax total in accounting currency (must match BT-6)
			inv.TaxTotalAccountingCurrency = currency
			inv.TaxTotalAccounting = amount
		}
		// Ignore TaxTotalAmount with unexpected currency (XML might have extras)
	}

	inv.GrandTotal, err = getDecimal(summation, "ram:GrandTotalAmount")
	if err != nil {
		return err
	}
	inv.TotalPrepaid, err = getDecimal(summation, "ram:TotalPrepaidAmount")
	if err != nil {
		return err
	}
	inv.DuePayableAmount, err = getDecimal(summation, "ram:DuePayableAmount")
	if err != nil {
		return err
	}

	// BG-3
	for refdoc := range applicableHeaderTradeSettlement.Each("ram:InvoiceReferencedDocument") {
		refDoc := ReferencedDocument{}

		refDoc.Date, err = parseCIITime(refdoc, "ram:FormattedIssueDateTime/qdt:DateTimeString")
		if err != nil {
			return err
		}

		refDoc.ID = refdoc.Eval("ram:IssuerAssignedID").String()
		inv.InvoiceReferencedDocument = append(inv.InvoiceReferencedDocument, refDoc)
	}

	return nil
}

func parseSpecifiedLineTradeAgreement(specifiedLineTradeAgreement *cxpath.Context, invoiceLine *InvoiceLine) error {
	var err error
	// BR-26: Track XML element presence to validate later
	invoiceLine.hasNetPriceInXML = specifiedLineTradeAgreement.Eval("count(ram:NetPriceProductTradePrice/ram:ChargeAmount)").Int() > 0
	invoiceLine.NetPrice, err = getDecimal(specifiedLineTradeAgreement, "ram:NetPriceProductTradePrice/ram:ChargeAmount")
	if err != nil {
		return err
	}
	invoiceLine.GrossPrice, err = getDecimal(specifiedLineTradeAgreement, "ram:GrossPriceProductTradePrice/ram:ChargeAmount")
	if err != nil {
		return err
	}
	// ZUGFeRD extended has unbound BT-147
	for allowanceCharge := range specifiedLineTradeAgreement.Each("ram:GrossPriceProductTradePrice/ram:AppliedTradeAllowanceCharge") {
		basisAmount, err := getDecimal(allowanceCharge, "ram:BasisAmount")
		if err != nil {
			return err
		}
		actualAmount, err := getDecimal(allowanceCharge, "ram:ActualAmount")
		if err != nil {
			return err
		}
		calculationPercent, err := getDecimal(allowanceCharge, "ram:CalculationPercent")
		if err != nil {
			return err
		}
		categoryTaxRate, err := getDecimal(allowanceCharge, "ram:CategoryTradeTax/ram:RateApplicablePercent")
		if err != nil {
			return err
		}

		allowanceCharge := AllowanceCharge{
			ChargeIndicator:                       allowanceCharge.Eval("string(ram:ChargeIndicator/udt:Indicator) = 'true'").Bool(),
			BasisAmount:                           basisAmount,
			ActualAmount:                          actualAmount,
			CalculationPercent:                    calculationPercent,
			ReasonCode:                            allowanceCharge.Eval("ram:ReasonCode").Int(),
			Reason:                                allowanceCharge.Eval("ram:Reason").String(),
			CategoryTradeTaxType:                  allowanceCharge.Eval("ram:CategoryTradeTax/ram:TypeCode").String(),
			CategoryTradeTaxCategoryCode:          allowanceCharge.Eval("ram:CategoryTradeTax/ram:CategoryCode").String(),
			CategoryTradeTaxRateApplicablePercent: categoryTaxRate,
		}
		invoiceLine.AppliedTradeAllowanceCharge = append(invoiceLine.AppliedTradeAllowanceCharge, allowanceCharge)
	}
	return nil
}

func parseSpecifiedTradeProduct(specifiedTradeProduct *cxpath.Context, invoiceLine *InvoiceLine) {
	invoiceLine.GlobalID = specifiedTradeProduct.Eval("ram:GlobalID").String()
	invoiceLine.GlobalIDType = specifiedTradeProduct.Eval("ram:GlobalID/@schemeID").String()
	invoiceLine.ArticleNumber = specifiedTradeProduct.Eval("ram:SellerAssignedID").String()
	invoiceLine.ArticleNumberBuyer = specifiedTradeProduct.Eval("ram:BuyerAssignedID").String()
	invoiceLine.ItemName = specifiedTradeProduct.Eval("ram:Name").String()
	invoiceLine.Description = specifiedTradeProduct.Eval("ram:Description").String()

	for itm := range specifiedTradeProduct.Each("ram:ApplicableProductCharacteristic") {
		ch := Characteristic{
			Description: itm.Eval("ram:Description").String(),
			Value:       itm.Eval("ram:Value").String(),
		}
		invoiceLine.Characteristics = append(invoiceLine.Characteristics, ch)
	}
	for itm := range specifiedTradeProduct.Each("ram:DesignatedProductClassification") {
		ch := Classification{
			ClassCode:     itm.Eval("ram:ClassCode").String(),
			ListID:        itm.Eval("ram:ClassCode/@listID").String(),
			ListVersionID: itm.Eval("ram:ClassCode/@listVersionID").String(),
		}
		invoiceLine.ProductClassification = append(invoiceLine.ProductClassification, ch)
	}

	invoiceLine.OriginTradeCountry = specifiedTradeProduct.Eval("ram:OriginTradeCountry/ram:ID").String()
}
