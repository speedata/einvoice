package einvoice

import (
	"fmt"
	"time"

	"github.com/speedata/cxpath"
)

// UBL 2.1 namespace URNs for Invoice and CreditNote documents
const (
	nsUBLInvoice      = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
	nsUBLCreditNote   = "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2"
	nsUBLCAC          = "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
	nsUBLCBC          = "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
)

// setupUBLNamespaces registers UBL 2.1 namespaces for XPath queries.
func setupUBLNamespaces(ctx *cxpath.Context) {
	ctx.SetNamespace("inv", nsUBLInvoice)
	ctx.SetNamespace("cn", nsUBLCreditNote)
	ctx.SetNamespace("cac", nsUBLCAC)
	ctx.SetNamespace("cbc", nsUBLCBC)
}

// parseTimeUBL parses ISO 8601 date format (YYYY-MM-DD) used in UBL documents.
func parseTimeUBL(ctx *cxpath.Context, path string) (time.Time, error) {
	timestring := ctx.Eval(path).String()
	if timestring == "" {
		return time.Time{}, nil
	}

	parsedDate, err := time.Parse("2006-01-02", timestring)
	if err != nil {
		return parsedDate, fmt.Errorf("invalid date %q at %s: %w", timestring, path, err)
	}

	return parsedDate, nil
}

// parseUBL parses a UBL 2.1 Invoice or CreditNote document into an Invoice struct.
// Both document types are mapped to the same Invoice struct, differentiated by InvoiceTypeCode.
func parseUBL(ctx *cxpath.Context) (*Invoice, error) {
	inv := &Invoice{SchemaType: UBL}

	root := ctx.Root()

	// Determine document type (Invoice vs CreditNote)
	localName := root.Eval("local-name()").String()

	// Set namespace prefix based on document type
	prefix := "inv:"
	if localName == "CreditNote" {
		prefix = "cn:"
	}

	// Parse all components
	if err := parseUBLHeader(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL header: %w", err)
	}

	if err := parseUBLParties(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL parties: %w", err)
	}

	if err := parseUBLAllowanceCharge(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL allowances/charges: %w", err)
	}

	if err := parseUBLTaxTotal(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL tax total: %w", err)
	}

	if err := parseUBLMonetarySummation(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL monetary summation: %w", err)
	}

	if err := parseUBLPaymentMeans(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL payment means: %w", err)
	}

	if err := parseUBLPaymentTerms(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL payment terms: %w", err)
	}

	if err := parseUBLLines(root, inv, prefix); err != nil {
		return nil, fmt.Errorf("parse UBL lines: %w", err)
	}

	return inv, nil
}

// parseUBLHeader parses the document header elements (BT-1 to BT-24, BG-1, BG-3, BG-14, BG-24).
func parseUBLHeader(root *cxpath.Context, inv *Invoice, prefix string) error {
	// BT-24: CustomizationID (Specification identifier)
	inv.GuidelineSpecifiedDocumentContextParameter = root.Eval(prefix + "cbc:CustomizationID").String()

	// BT-23: ProfileID (Business process type)
	inv.BPSpecifiedDocumentContextParameter = root.Eval(prefix + "cbc:ProfileID").String()

	// BT-1: Invoice number
	inv.InvoiceNumber = root.Eval(prefix + "cbc:ID").String()

	// BT-3: Invoice type code
	inv.InvoiceTypeCode = CodeDocument(root.Eval(prefix + "cbc:InvoiceTypeCode").Int())
	if inv.InvoiceTypeCode == 0 {
		// Try CreditNoteTypeCode for credit notes
		inv.InvoiceTypeCode = CodeDocument(root.Eval(prefix + "cbc:CreditNoteTypeCode").Int())
	}

	// BT-2: Invoice date
	var err error
	inv.InvoiceDate, err = parseTimeUBL(root, prefix+"cbc:IssueDate")
	if err != nil {
		return err
	}

	// BT-72: Actual delivery date (optional, in cac:Delivery)
	inv.OccurrenceDateTime, _ = parseTimeUBL(root, prefix+"cac:Delivery/cbc:ActualDeliveryDate")

	// BT-5: Invoice currency
	inv.InvoiceCurrencyCode = root.Eval(prefix + "cbc:DocumentCurrencyCode").String()

	// BT-6: Tax currency (optional)
	inv.TaxCurrencyCode = root.Eval(prefix + "cbc:TaxCurrencyCode").String()

	// BT-10: Buyer reference (optional)
	inv.BuyerReference = root.Eval(prefix + "cbc:BuyerReference").String()

	// BT-19: Accounting cost (Buyer accounting reference)
	inv.ReceivableSpecifiedTradeAccountingAccount = root.Eval(prefix + "cbc:AccountingCost").String()

	// BT-13: Purchase order reference
	inv.BuyerOrderReferencedDocument = root.Eval(prefix + "cac:OrderReference/cbc:ID").String()

	// BT-14: Sales order reference
	inv.SellerOrderReferencedDocument = root.Eval(prefix + "cac:OrderReference/cbc:SalesOrderID").String()

	// BT-12: Contract document reference
	inv.ContractReferencedDocument = root.Eval(prefix + "cac:ContractDocumentReference/cbc:ID").String()

	// BT-11: Project reference
	inv.SpecifiedProcuringProjectID = root.Eval(prefix + "cac:ProjectReference/cbc:ID").String()

	// BT-16: Despatch advice reference
	inv.DespatchAdviceReferencedDocument = root.Eval(prefix + "cac:DespatchDocumentReference/cbc:ID").String()

	// BT-15: Receiving advice reference
	inv.ReceivingAdviceReferencedDocument = root.Eval(prefix + "cac:ReceiptDocumentReference/cbc:ID").String()

	// BG-1: Process notes
	for note := range root.Each(prefix + "cbc:Note") {
		inv.Notes = append(inv.Notes, Note{
			Text: note.String(),
			// UBL doesn't typically have subject codes in Note elements
		})
	}

	// BG-3: Preceding invoice references
	for ref := range root.Each(prefix + "cac:BillingReference/cac:InvoiceDocumentReference") {
		refDoc := ReferencedDocument{
			ID: ref.Eval("cbc:ID").String(),
		}

		refDoc.Date, _ = parseTimeUBL(ref, "cbc:IssueDate")

		inv.InvoiceReferencedDocument = append(inv.InvoiceReferencedDocument, refDoc)
	}

	// BG-14: Invoice period (document level)
	if root.Eval(fmt.Sprintf("count(%scac:InvoicePeriod)", prefix)).Int() > 0 {
		inv.BillingSpecifiedPeriodStart, _ = parseTimeUBL(root, prefix+"cac:InvoicePeriod/cbc:StartDate")
		inv.BillingSpecifiedPeriodEnd, _ = parseTimeUBL(root, prefix+"cac:InvoicePeriod/cbc:EndDate")
	}

	// BG-24: Additional supporting documents
	for doc := range root.Each(prefix + "cac:AdditionalDocumentReference") {
		addDoc := Document{
			IssuerAssignedID: doc.Eval("cbc:ID").String(),
			TypeCode:         doc.Eval("cbc:DocumentTypeCode").String(),
			Name:             doc.Eval("cbc:DocumentDescription").String(),
			URIID:            doc.Eval("cac:Attachment/cac:ExternalReference/cbc:URI").String(),
		}

		// Handle embedded binary object
		binaryData := doc.Eval("cac:Attachment/cbc:EmbeddedDocumentBinaryObject").String()
		if binaryData != "" {
			// Binary data would need base64 decoding, similar to CII parser
			addDoc.AttachmentMimeCode = doc.Eval("cac:Attachment/cbc:EmbeddedDocumentBinaryObject/@mimeCode").String()
			addDoc.AttachmentFilename = doc.Eval("cac:Attachment/cbc:EmbeddedDocumentBinaryObject/@filename").String()
			// TODO: Decode base64 if needed
		}

		inv.AdditionalReferencedDocument = append(inv.AdditionalReferencedDocument, addDoc)
	}

	return nil
}

// parseUBLParties parses all party elements (BG-4, BG-7, BG-10, BG-11, BG-13).
func parseUBLParties(root *cxpath.Context, inv *Invoice, prefix string) error {
	// BG-4: Seller (AccountingSupplierParty)
	inv.Seller = parseUBLParty(root, prefix+"cac:AccountingSupplierParty/cac:Party")

	// BG-7: Buyer (AccountingCustomerParty)
	inv.Buyer = parseUBLParty(root, prefix+"cac:AccountingCustomerParty/cac:Party")

	// BG-10: Payee (optional)
	if root.Eval(fmt.Sprintf("count(%scac:PayeeParty)", prefix)).Int() > 0 {
		payee := parseUBLParty(root, prefix+"cac:PayeeParty")
		inv.PayeeTradeParty = &payee
	}

	// BG-11: Seller tax representative (optional)
	if root.Eval(fmt.Sprintf("count(%scac:TaxRepresentativeParty)", prefix)).Int() > 0 {
		taxRep := parseUBLParty(root, prefix+"cac:TaxRepresentativeParty")
		inv.SellerTaxRepresentativeTradeParty = &taxRep
	}

	// BG-13: Delivery information (optional)
	if root.Eval(fmt.Sprintf("count(%scac:Delivery)", prefix)).Int() > 0 {
		// Delivery party
		if root.Eval(fmt.Sprintf("count(%scac:Delivery/cac:DeliveryParty)", prefix)).Int() > 0 {
			shipTo := parseUBLParty(root, prefix+"cac:Delivery/cac:DeliveryParty")
			inv.ShipTo = &shipTo
		} else if root.Eval(fmt.Sprintf("count(%scac:Delivery/cac:DeliveryLocation)", prefix)).Int() > 0 {
			// If no DeliveryParty, create one from DeliveryLocation address
			shipTo := Party{}
			if root.Eval(fmt.Sprintf("count(%scac:Delivery/cac:DeliveryLocation/cac:Address)", prefix)).Int() > 0 {
				postalAddr := &PostalAddress{
					Line1:                  root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:StreetName").String(),
					Line2:                  root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:AdditionalStreetName").String(),
					Line3:                  root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cac:AddressLine/cbc:Line").String(),
					City:                   root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:CityName").String(),
					PostcodeCode:           root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:PostalZone").String(),
					CountrySubDivisionName: root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cbc:CountrySubentity").String(),
					CountryID:              root.Eval(prefix + "cac:Delivery/cac:DeliveryLocation/cac:Address/cac:Country/cbc:IdentificationCode").String(),
				}
				shipTo.PostalAddress = postalAddr
			}
			inv.ShipTo = &shipTo
		}
	}

	return nil
}

// parseUBLParty parses a single party (reusable for Seller, Buyer, Payee, etc.).
func parseUBLParty(ctx *cxpath.Context, partyPath string) Party {
	party := Party{}

	// Electronic address (BT-34, BT-49, BT-98)
	party.URIUniversalCommunication = ctx.Eval(partyPath + "/cbc:EndpointID").String()
	party.URIUniversalCommunicationScheme = ctx.Eval(partyPath + "/cbc:EndpointID/@schemeID").String()

	// Party identification (BT-29, BT-46, BT-60, BT-71)
	for id := range ctx.Each(partyPath + "/cac:PartyIdentification") {
		idValue := id.Eval("cbc:ID").String()
		idScheme := id.Eval("cbc:ID/@schemeID").String()

		if idScheme != "" {
			party.GlobalID = append(party.GlobalID, GlobalID{
				ID:     idValue,
				Scheme: idScheme,
			})
		} else {
			party.ID = append(party.ID, idValue)
		}
	}

	// Party name (BT-27, BT-44, BT-59, BT-70)
	party.Name = ctx.Eval(partyPath + "/cac:PartyName/cbc:Name").String()
	if party.Name == "" {
		// Fallback to PartyLegalEntity/RegistrationName
		party.Name = ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:RegistrationName").String()
	}

	// Postal address (BG-5, BG-8, BG-12, BG-15)
	if ctx.Eval(fmt.Sprintf("count(%s/cac:PostalAddress)", partyPath)).Int() > 0 {
		postalAddr := &PostalAddress{
			Line1:                  ctx.Eval(partyPath + "/cac:PostalAddress/cbc:StreetName").String(),
			Line2:                  ctx.Eval(partyPath + "/cac:PostalAddress/cbc:AdditionalStreetName").String(),
			Line3:                  ctx.Eval(partyPath + "/cac:PostalAddress/cac:AddressLine/cbc:Line").String(),
			City:                   ctx.Eval(partyPath + "/cac:PostalAddress/cbc:CityName").String(),
			PostcodeCode:           ctx.Eval(partyPath + "/cac:PostalAddress/cbc:PostalZone").String(),
			CountrySubDivisionName: ctx.Eval(partyPath + "/cac:PostalAddress/cbc:CountrySubentity").String(),
			CountryID:              ctx.Eval(partyPath + "/cac:PostalAddress/cac:Country/cbc:IdentificationCode").String(),
		}
		party.PostalAddress = postalAddr
	}

	// Legal organization (BT-30, BT-47, BT-61)
	if ctx.Eval(fmt.Sprintf("count(%s/cac:PartyLegalEntity)", partyPath)).Int() > 0 {
		legalOrg := &SpecifiedLegalOrganization{
			ID:                  ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:CompanyID").String(),
			Scheme:              ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:CompanyID/@schemeID").String(),
			TradingBusinessName: ctx.Eval(partyPath + "/cac:PartyLegalEntity/cbc:RegistrationName").String(),
		}
		party.SpecifiedLegalOrganization = legalOrg
	}

	// Tax registration (BT-31, BT-32, BT-48, BT-63)
	for taxScheme := range ctx.Each(partyPath + "/cac:PartyTaxScheme") {
		taxID := taxScheme.Eval("cbc:CompanyID").String()
		scheme := taxScheme.Eval("cac:TaxScheme/cbc:ID").String()

		if scheme == "VAT" {
			party.VATaxRegistration = taxID
		} else if scheme == "FC" {
			party.FCTaxRegistration = taxID
		}
	}

	// Contact (BG-6, BG-9)
	for contact := range ctx.Each(partyPath + "/cac:Contact") {
		dtc := DefinedTradeContact{
			PersonName:  contact.Eval("cbc:Name").String(),
			PhoneNumber: contact.Eval("cbc:Telephone").String(),
			EMail:       contact.Eval("cbc:ElectronicMail").String(),
		}
		party.DefinedTradeContact = append(party.DefinedTradeContact, dtc)
	}

	return party
}

// parseUBLAllowanceCharge parses document-level allowances and charges (BG-20, BG-21).
func parseUBLAllowanceCharge(root *cxpath.Context, inv *Invoice, prefix string) error {
	for ac := range root.Each(prefix + "cac:AllowanceCharge") {
		chargeIndicator := ac.Eval("string(cbc:ChargeIndicator) = 'true'").Bool()

		basisAmount, err := getDecimal(ac, "cbc:BaseAmount")
		if err != nil {
			return err
		}

		actualAmount, err := getDecimal(ac, "cbc:Amount")
		if err != nil {
			return err
		}

		calculationPercent, err := getDecimal(ac, "cbc:MultiplierFactorNumeric")
		if err != nil {
			return err
		}

		categoryTaxRate, err := getDecimal(ac, "cac:TaxCategory/cbc:Percent")
		if err != nil {
			return err
		}

		allowanceCharge := AllowanceCharge{
			ChargeIndicator:                       chargeIndicator,
			BasisAmount:                           basisAmount,
			ActualAmount:                          actualAmount,
			CalculationPercent:                    calculationPercent,
			ReasonCode:                            ac.Eval("cbc:AllowanceChargeReasonCode").Int(),
			Reason:                                ac.Eval("cbc:AllowanceChargeReason").String(),
			CategoryTradeTaxType:                  ac.Eval("cac:TaxCategory/cac:TaxScheme/cbc:ID").String(),
			CategoryTradeTaxCategoryCode:          ac.Eval("cac:TaxCategory/cbc:ID").String(),
			CategoryTradeTaxRateApplicablePercent: categoryTaxRate,
		}

		inv.SpecifiedTradeAllowanceCharge = append(inv.SpecifiedTradeAllowanceCharge, allowanceCharge)
	}

	return nil
}

// parseUBLTaxTotal parses the tax breakdown (BG-23).
func parseUBLTaxTotal(root *cxpath.Context, inv *Invoice, prefix string) error {
	// BT-110: Total tax amount (document currency)
	inv.TaxTotalCurrency = root.Eval(prefix + "cac:TaxTotal/cbc:TaxAmount/@currencyID").String()
	if inv.TaxTotalCurrency == "" {
		inv.TaxTotalCurrency = inv.InvoiceCurrencyCode
	}

	var err error
	inv.TaxTotal, err = getDecimal(root, prefix+"cac:TaxTotal/cbc:TaxAmount")
	if err != nil {
		return err
	}

	// BT-111: Tax total in accounting currency (if different)
	if inv.TaxCurrencyCode != "" && inv.TaxCurrencyCode != inv.InvoiceCurrencyCode {
		// Find TaxTotal with accounting currency
		for taxTotal := range root.Each(prefix + "cac:TaxTotal") {
			currency := taxTotal.Eval("cbc:TaxAmount/@currencyID").String()
			if currency == inv.TaxCurrencyCode {
				inv.TaxTotalVAT, _ = getDecimal(taxTotal, "cbc:TaxAmount")
				inv.TaxTotalVATCurrency = currency
				break
			}
		}
	}

	// BG-23: VAT breakdown (TaxSubtotal elements)
	for subtotal := range root.Each(prefix + "cac:TaxTotal/cac:TaxSubtotal") {
		tradeTax := TradeTax{}

		tradeTax.BasisAmount, err = getDecimal(subtotal, "cbc:TaxableAmount")
		if err != nil {
			return err
		}

		tradeTax.CalculatedAmount, err = getDecimal(subtotal, "cbc:TaxAmount")
		if err != nil {
			return err
		}

		tradeTax.TypeCode = subtotal.Eval("cac:TaxCategory/cac:TaxScheme/cbc:ID").String()
		if tradeTax.TypeCode == "" {
			tradeTax.TypeCode = "VAT" // Default to VAT
		}

		tradeTax.CategoryCode = subtotal.Eval("cac:TaxCategory/cbc:ID").String()

		tradeTax.Percent, err = getDecimal(subtotal, "cac:TaxCategory/cbc:Percent")
		if err != nil {
			return err
		}

		tradeTax.ExemptionReason = subtotal.Eval("cac:TaxCategory/cbc:TaxExemptionReason").String()
		tradeTax.ExemptionReasonCode = subtotal.Eval("cac:TaxCategory/cbc:TaxExemptionReasonCode").String()

		inv.TradeTaxes = append(inv.TradeTaxes, tradeTax)
	}

	return nil
}

// parseUBLMonetarySummation parses the monetary totals (BT-106 to BT-115).
func parseUBLMonetarySummation(root *cxpath.Context, inv *Invoice, prefix string) error {
	legalMonetaryTotal := root.Eval(prefix + "cac:LegalMonetaryTotal")

	// Track XML element presence for BR-12 through BR-15 validation
	inv.hasLineTotalInXML = legalMonetaryTotal.Eval("count(cbc:LineExtensionAmount)").Int() > 0
	inv.hasTaxBasisTotalInXML = legalMonetaryTotal.Eval("count(cbc:TaxExclusiveAmount)").Int() > 0
	inv.hasGrandTotalInXML = legalMonetaryTotal.Eval("count(cbc:TaxInclusiveAmount)").Int() > 0
	inv.hasDuePayableAmountInXML = legalMonetaryTotal.Eval("count(cbc:PayableAmount)").Int() > 0

	var err error

	// BT-106: Sum of Invoice line net amount
	inv.LineTotal, err = getDecimal(legalMonetaryTotal, "cbc:LineExtensionAmount")
	if err != nil {
		return err
	}

	// BT-107: Sum of allowances on document level
	inv.AllowanceTotal, err = getDecimal(legalMonetaryTotal, "cbc:AllowanceTotalAmount")
	if err != nil {
		return err
	}

	// BT-108: Sum of charges on document level
	inv.ChargeTotal, err = getDecimal(legalMonetaryTotal, "cbc:ChargeTotalAmount")
	if err != nil {
		return err
	}

	// BT-109: Invoice total amount without VAT
	inv.TaxBasisTotal, err = getDecimal(legalMonetaryTotal, "cbc:TaxExclusiveAmount")
	if err != nil {
		return err
	}

	// BT-112: Invoice total amount with VAT
	inv.GrandTotal, err = getDecimal(legalMonetaryTotal, "cbc:TaxInclusiveAmount")
	if err != nil {
		return err
	}

	// BT-113: Paid amount
	inv.TotalPrepaid, err = getDecimal(legalMonetaryTotal, "cbc:PrepaidAmount")
	if err != nil {
		return err
	}

	// BT-114: Rounding amount
	inv.RoundingAmount, err = getDecimal(legalMonetaryTotal, "cbc:PayableRoundingAmount")
	if err != nil {
		return err
	}

	// BT-115: Amount due for payment
	inv.DuePayableAmount, err = getDecimal(legalMonetaryTotal, "cbc:PayableAmount")
	if err != nil {
		return err
	}

	return nil
}

// parseUBLPaymentMeans parses payment means elements (BG-16, BG-17, BG-18, BG-19).
func parseUBLPaymentMeans(root *cxpath.Context, inv *Invoice, prefix string) error {
	for pm := range root.Each(prefix + "cac:PaymentMeans") {
		paymentMeans := PaymentMeans{
			TypeCode:    pm.Eval("cbc:PaymentMeansCode").Int(),
			Information: pm.Eval("cbc:InstructionNote").String(),
		}

		// BT-83: Remittance information
		inv.PaymentReference = pm.Eval("cbc:PaymentID").String()

		// BG-17: Credit transfer (IBAN/BIC)
		if pm.Eval("count(cac:PayeeFinancialAccount)").Int() > 0 {
			paymentMeans.PayeePartyCreditorFinancialAccountIBAN = pm.Eval("cac:PayeeFinancialAccount/cbc:ID").String()
			paymentMeans.PayeePartyCreditorFinancialAccountName = pm.Eval("cac:PayeeFinancialAccount/cbc:Name").String()
			paymentMeans.PayeePartyCreditorFinancialAccountProprietaryID = pm.Eval("cac:PayeeFinancialAccount/cac:ID").String()
			paymentMeans.PayeeSpecifiedCreditorFinancialInstitutionBIC = pm.Eval("cac:PayeeFinancialAccount/cac:FinancialInstitutionBranch/cbc:ID").String()
		}

		// BG-18: Payment card information
		if pm.Eval("count(cac:CardAccount)").Int() > 0 {
			paymentMeans.ApplicableTradeSettlementFinancialCardID = pm.Eval("cac:CardAccount/cbc:PrimaryAccountNumberID").String()
			paymentMeans.ApplicableTradeSettlementFinancialCardCardholderName = pm.Eval("cac:CardAccount/cbc:HolderName").String()
		}

		// BG-19: Direct debit
		if pm.Eval("count(cac:PaymentMandate)").Int() > 0 {
			paymentMeans.PayerPartyDebtorFinancialAccountIBAN = pm.Eval("cac:PaymentMandate/cac:PayerFinancialAccount/cbc:ID").String()
		}

		inv.PaymentMeans = append(inv.PaymentMeans, paymentMeans)
	}

	return nil
}

// parseUBLPaymentTerms parses payment terms (BT-20, BT-9, BT-89).
func parseUBLPaymentTerms(root *cxpath.Context, inv *Invoice, prefix string) error {
	for pt := range root.Each(prefix + "cac:PaymentTerms") {
		paymentTerm := SpecifiedTradePaymentTerms{
			Description: pt.Eval("cbc:Note").String(),
		}

		// BT-9: Payment due date
		var err error
		paymentTerm.DueDate, err = parseTimeUBL(pt, "cbc:PaymentDueDate")
		if err != nil {
			return err
		}

		// BT-89: Direct debit mandate identifier
		paymentTerm.DirectDebitMandateID = pt.Eval("cbc:PaymentMeansID").String()

		inv.SpecifiedTradePaymentTerms = append(inv.SpecifiedTradePaymentTerms, paymentTerm)
	}

	return nil
}

// parseUBLLines parses all invoice line items (BG-25).
func parseUBLLines(root *cxpath.Context, inv *Invoice, prefix string) error {
	for lineItem := range root.Each(prefix + "cac:InvoiceLine") {
		invoiceLine := InvoiceLine{}

		// BT-126: Invoice line identifier
		invoiceLine.LineID = lineItem.Eval("cbc:ID").String()

		// BT-127: Invoice line note
		invoiceLine.Note = lineItem.Eval("cbc:Note").String()

		// BG-26: Invoice line period
		if lineItem.Eval("count(cac:InvoicePeriod)").Int() > 0 {
			invoiceLine.BillingSpecifiedPeriodStart, _ = parseTimeUBL(lineItem, "cac:InvoicePeriod/cbc:StartDate")
			invoiceLine.BillingSpecifiedPeriodEnd, _ = parseTimeUBL(lineItem, "cac:InvoicePeriod/cbc:EndDate")
		}

		// BT-128: Invoice line object identifier
		invoiceLine.AdditionalReferencedDocumentID = lineItem.Eval("cac:DocumentReference/cbc:ID").String()
		invoiceLine.AdditionalReferencedDocumentTypeCode = lineItem.Eval("cac:DocumentReference/cbc:DocumentTypeCode").String()

		// BT-132: Referenced purchase order line
		invoiceLine.BuyerOrderReferencedDocument = lineItem.Eval("cac:OrderLineReference/cbc:LineID").String()

		// BT-133: Invoice line Buyer accounting reference
		invoiceLine.ReceivableSpecifiedTradeAccountingAccount = lineItem.Eval("cac:AccountingCost").String()

		// BT-129: Invoiced quantity
		var err error
		invoiceLine.BilledQuantity, err = getDecimal(lineItem, "cbc:InvoicedQuantity")
		if err != nil {
			return err
		}

		// BT-130: Invoiced quantity unit of measure
		invoiceLine.BilledQuantityUnit = lineItem.Eval("cbc:InvoicedQuantity/@unitCode").String()

		// BT-131: Invoice line net amount
		// Track XML element presence for BR-24 validation
		invoiceLine.hasLineTotalInXML = lineItem.Eval("count(cbc:LineExtensionAmount)").Int() > 0
		invoiceLine.Total, err = getDecimal(lineItem, "cbc:LineExtensionAmount")
		if err != nil {
			return err
		}

		// Parse item information
		if err := parseUBLLineItem(lineItem, &invoiceLine); err != nil {
			return err
		}

		// Parse price information
		if err := parseUBLLinePrice(lineItem, &invoiceLine); err != nil {
			return err
		}

		// BG-27: Line level allowances
		// BG-28: Line level charges
		for ac := range lineItem.Each("cac:AllowanceCharge") {
			chargeIndicator := ac.Eval("string(cbc:ChargeIndicator) = 'true'").Bool()

			basisAmount, err := getDecimal(ac, "cbc:BaseAmount")
			if err != nil {
				return err
			}

			actualAmount, err := getDecimal(ac, "cbc:Amount")
			if err != nil {
				return err
			}

			calculationPercent, err := getDecimal(ac, "cbc:MultiplierFactorNumeric")
			if err != nil {
				return err
			}

			alc := AllowanceCharge{
				ChargeIndicator:    chargeIndicator,
				BasisAmount:        basisAmount,
				ActualAmount:       actualAmount,
				CalculationPercent: calculationPercent,
				ReasonCode:         ac.Eval("cbc:AllowanceChargeReasonCode").Int(),
				Reason:             ac.Eval("cbc:AllowanceChargeReason").String(),
			}

			if chargeIndicator {
				invoiceLine.InvoiceLineCharges = append(invoiceLine.InvoiceLineCharges, alc)
			} else {
				invoiceLine.InvoiceLineAllowances = append(invoiceLine.InvoiceLineAllowances, alc)
			}
		}

		// Parse line tax information
		taxInfo := lineItem.Eval("cac:Item/cac:ClassifiedTaxCategory")
		invoiceLine.TaxTypeCode = taxInfo.Eval("cac:TaxScheme/cbc:ID").String()
		if invoiceLine.TaxTypeCode == "" {
			invoiceLine.TaxTypeCode = "VAT" // Default to VAT
		}
		invoiceLine.TaxCategoryCode = taxInfo.Eval("cbc:ID").String()
		invoiceLine.TaxRateApplicablePercent, err = getDecimal(taxInfo, "cbc:Percent")
		if err != nil {
			return err
		}

		inv.InvoiceLines = append(inv.InvoiceLines, invoiceLine)
	}

	return nil
}

// parseUBLLineItem parses item-specific information within a line.
func parseUBLLineItem(lineItem *cxpath.Context, invoiceLine *InvoiceLine) error {
	item := lineItem.Eval("cac:Item")

	// BT-153: Item name
	invoiceLine.ItemName = item.Eval("cbc:Name").String()

	// BT-154: Item description
	invoiceLine.Description = item.Eval("cbc:Description").String()

	// BT-155: Item Seller's identifier
	invoiceLine.ArticleNumber = item.Eval("cac:SellersItemIdentification/cbc:ID").String()

	// BT-156: Item Buyer's identifier
	invoiceLine.ArticleNumberBuyer = item.Eval("cac:BuyersItemIdentification/cbc:ID").String()

	// BT-157: Item standard identifier
	invoiceLine.GlobalID = item.Eval("cac:StandardItemIdentification/cbc:ID").String()
	invoiceLine.GlobalIDType = item.Eval("cac:StandardItemIdentification/cbc:ID/@schemeID").String()

	// BT-158: Item classification identifier
	for class := range item.Each("cac:CommodityClassification") {
		classification := Classification{
			ClassCode:     class.Eval("cbc:ItemClassificationCode").String(),
			ListID:        class.Eval("cbc:ItemClassificationCode/@listID").String(),
			ListVersionID: class.Eval("cbc:ItemClassificationCode/@listVersionID").String(),
		}
		invoiceLine.ProductClassification = append(invoiceLine.ProductClassification, classification)
	}

	// BT-159: Item country of origin
	invoiceLine.OriginTradeCountry = item.Eval("cac:OriginCountry/cbc:IdentificationCode").String()

	// BG-32: Item attributes
	for attr := range item.Each("cac:AdditionalItemProperty") {
		characteristic := Characteristic{
			Description: attr.Eval("cbc:Name").String(),
			Value:       attr.Eval("cbc:Value").String(),
		}
		invoiceLine.Characteristics = append(invoiceLine.Characteristics, characteristic)
	}

	return nil
}

// parseUBLLinePrice parses price information within a line.
func parseUBLLinePrice(lineItem *cxpath.Context, invoiceLine *InvoiceLine) error {
	price := lineItem.Eval("cac:Price")

	var err error

	// BT-146: Item net price
	// Track XML element presence for BR-26 validation
	invoiceLine.hasNetPriceInXML = price.Eval("count(cbc:PriceAmount)").Int() > 0
	invoiceLine.NetPrice, err = getDecimal(price, "cbc:PriceAmount")
	if err != nil {
		return err
	}

	// BT-149: Item price base quantity
	invoiceLine.BasisQuantity, err = getDecimal(price, "cbc:BaseQuantity")
	if err != nil {
		return err
	}

	// BT-148: Item gross price (price before allowances)
	// UBL doesn't have a direct gross price field, but may have allowances on price
	// For now, calculate from net price if allowances exist on price
	for ac := range price.Each("cac:AllowanceCharge") {
		chargeIndicator := ac.Eval("string(cbc:ChargeIndicator) = 'true'").Bool()

		basisAmount, err := getDecimal(ac, "cbc:BaseAmount")
		if err != nil {
			return err
		}

		actualAmount, err := getDecimal(ac, "cbc:Amount")
		if err != nil {
			return err
		}

		calculationPercent, err := getDecimal(ac, "cbc:MultiplierFactorNumeric")
		if err != nil {
			return err
		}

		allowanceCharge := AllowanceCharge{
			ChargeIndicator:    chargeIndicator,
			BasisAmount:        basisAmount,
			ActualAmount:       actualAmount,
			CalculationPercent: calculationPercent,
			ReasonCode:         ac.Eval("cbc:AllowanceChargeReasonCode").Int(),
			Reason:             ac.Eval("cbc:AllowanceChargeReason").String(),
		}

		invoiceLine.AppliedTradeAllowanceCharge = append(invoiceLine.AppliedTradeAllowanceCharge, allowanceCharge)

		// Calculate gross price if we have basis amount
		if !basisAmount.IsZero() && invoiceLine.GrossPrice.IsZero() {
			invoiceLine.GrossPrice = basisAmount
		}
	}

	return nil
}
