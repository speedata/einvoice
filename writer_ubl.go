package einvoice

import (
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/beevik/etree"
)

// writeUBL writes an invoice in UBL 2.1 format (Invoice or CreditNote).
// The document type is determined by the InvoiceTypeCode.
func writeUBL(inv *Invoice, writer io.Writer) error {
	doc := etree.NewDocument()

	// Determine if this is a CreditNote (type code 381) or Invoice
	isCreditNote := inv.InvoiceTypeCode == 381

	var root *etree.Element
	var rootNS string
	var prefix string

	if isCreditNote {
		root = doc.CreateElement("CreditNote")
		rootNS = nsUBLCreditNote
		prefix = "cn:"
	} else {
		root = doc.CreateElement("Invoice")
		rootNS = nsUBLInvoice
		prefix = "inv:"
	}

	// Set up namespaces
	root.CreateAttr("xmlns", rootNS)
	root.CreateAttr("xmlns:cac", nsUBLCAC)
	root.CreateAttr("xmlns:cbc", nsUBLCBC)

	// Write the document structure
	writeUBLHeader(inv, root, prefix)
	writeUBLParties(inv, root, prefix)
	writeUBLAllowanceCharge(inv, root, prefix)
	writeUBLTaxTotal(inv, root, prefix)
	writeUBLMonetarySummation(inv, root, prefix)
	writeUBLPaymentMeans(inv, root, prefix)
	writeUBLPaymentTerms(inv, root, prefix)
	writeUBLLines(inv, root, prefix)

	doc.Indent(2)
	if _, err := doc.WriteTo(writer); err != nil {
		return fmt.Errorf("write UBL: failed to write to the writer: %w", err)
	}

	return nil
}

// addTimeUBL formats a time in ISO 8601 format (YYYY-MM-DD) for UBL
func addTimeUBL(parent *etree.Element, elementName string, date time.Time) {
	if !date.IsZero() {
		parent.CreateElement(elementName).SetText(date.Format("2006-01-02"))
	}
}

// writeUBLHeader writes the document header elements (BT-1 to BT-24, BG-1, BG-3, BG-14, BG-24)
func writeUBLHeader(inv *Invoice, root *etree.Element, prefix string) {
	// BT-24: Specification identifier
	if inv.GuidelineSpecifiedDocumentContextParameter != "" {
		root.CreateElement("cbc:CustomizationID").SetText(inv.GuidelineSpecifiedDocumentContextParameter)
	}

	// BT-23: Business process type
	if inv.BPSpecifiedDocumentContextParameter != "" {
		root.CreateElement("cbc:ProfileID").SetText(inv.BPSpecifiedDocumentContextParameter)
	}

	// BT-1: Invoice number
	root.CreateElement("cbc:ID").SetText(inv.InvoiceNumber)

	// BT-2: Invoice issue date
	addTimeUBL(root, "cbc:IssueDate", inv.InvoiceDate)

	// BT-3: Invoice type code
	if inv.InvoiceTypeCode != 0 {
		root.CreateElement("cbc:InvoiceTypeCode").SetText(inv.InvoiceTypeCode.String())
	}

	// BG-1: Process notes
	for _, note := range inv.Notes {
		root.CreateElement("cbc:Note").SetText(note.Text)
	}

	// BT-5: Invoice currency code
	root.CreateElement("cbc:DocumentCurrencyCode").SetText(inv.InvoiceCurrencyCode)

	// BT-6: Tax currency code (optional)
	if inv.TaxCurrencyCode != "" {
		root.CreateElement("cbc:TaxCurrencyCode").SetText(inv.TaxCurrencyCode)
	}

	// BT-10: Buyer reference (optional)
	if inv.BuyerReference != "" {
		root.CreateElement("cbc:BuyerReference").SetText(inv.BuyerReference)
	}

	// BT-19: Accounting cost
	if inv.ReceivableSpecifiedTradeAccountingAccount != "" {
		root.CreateElement("cbc:AccountingCost").SetText(inv.ReceivableSpecifiedTradeAccountingAccount)
	}

	// BG-14: Invoice period (document level)
	if !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero() {
		period := root.CreateElement("cac:InvoicePeriod")
		addTimeUBL(period, "cbc:StartDate", inv.BillingSpecifiedPeriodStart)
		addTimeUBL(period, "cbc:EndDate", inv.BillingSpecifiedPeriodEnd)
	}

	// BT-13: Purchase order reference
	if inv.BuyerOrderReferencedDocument != "" {
		orderRef := root.CreateElement("cac:OrderReference")
		orderRef.CreateElement("cbc:ID").SetText(inv.BuyerOrderReferencedDocument)

		// BT-14: Sales order reference
		if inv.SellerOrderReferencedDocument != "" {
			orderRef.CreateElement("cbc:SalesOrderID").SetText(inv.SellerOrderReferencedDocument)
		}
	}

	// BG-3: Preceding invoice references
	for _, refDoc := range inv.InvoiceReferencedDocument {
		billingRef := root.CreateElement("cac:BillingReference")
		invDocRef := billingRef.CreateElement("cac:InvoiceDocumentReference")
		invDocRef.CreateElement("cbc:ID").SetText(refDoc.ID)
		addTimeUBL(invDocRef, "cbc:IssueDate", refDoc.Date)
	}

	// BT-16: Despatch advice reference
	if inv.DespatchAdviceReferencedDocument != "" {
		despatchRef := root.CreateElement("cac:DespatchDocumentReference")
		despatchRef.CreateElement("cbc:ID").SetText(inv.DespatchAdviceReferencedDocument)
	}

	// BT-15: Receiving advice reference
	if inv.ReceivingAdviceReferencedDocument != "" {
		receiptRef := root.CreateElement("cac:ReceiptDocumentReference")
		receiptRef.CreateElement("cbc:ID").SetText(inv.ReceivingAdviceReferencedDocument)
	}

	// BT-12: Contract reference
	if inv.ContractReferencedDocument != "" {
		contractRef := root.CreateElement("cac:ContractDocumentReference")
		contractRef.CreateElement("cbc:ID").SetText(inv.ContractReferencedDocument)
	}

	// BT-11: Project reference
	if inv.SpecifiedProcuringProjectID != "" {
		projectRef := root.CreateElement("cac:ProjectReference")
		projectRef.CreateElement("cbc:ID").SetText(inv.SpecifiedProcuringProjectID)
	}

	// BG-24: Additional supporting documents
	for _, doc := range inv.AdditionalReferencedDocument {
		addDocRef := root.CreateElement("cac:AdditionalDocumentReference")
		addDocRef.CreateElement("cbc:ID").SetText(doc.IssuerAssignedID)

		if doc.TypeCode != "" {
			addDocRef.CreateElement("cbc:DocumentTypeCode").SetText(doc.TypeCode)
		}

		if doc.Name != "" {
			addDocRef.CreateElement("cbc:DocumentDescription").SetText(doc.Name)
		}

		if doc.URIID != "" {
			attachment := addDocRef.CreateElement("cac:Attachment")
			externalRef := attachment.CreateElement("cac:ExternalReference")
			externalRef.CreateElement("cbc:URI").SetText(doc.URIID)
		}

		// BT-125: Handle embedded binary objects (only write if data exists - PEPPOL-EN16931-R008)
		if len(doc.AttachmentBinaryObject) > 0 {
			attachment := addDocRef.CreateElement("cac:Attachment")
			embeddedDoc := attachment.CreateElement("cbc:EmbeddedDocumentBinaryObject")

			if doc.AttachmentMimeCode != "" {
				embeddedDoc.CreateAttr("mimeCode", doc.AttachmentMimeCode)
			}
			if doc.AttachmentFilename != "" {
				embeddedDoc.CreateAttr("filename", doc.AttachmentFilename)
			}
			embeddedDoc.SetText(base64.StdEncoding.EncodeToString(doc.AttachmentBinaryObject))
		}
	}

	// BT-72: Actual delivery date (in Delivery element, handled in writeUBLParties)
}

// writeUBLParties writes all party elements (BG-4, BG-7, BG-10, BG-11, BG-13)
func writeUBLParties(inv *Invoice, root *etree.Element, prefix string) {
	// BG-4: Seller (AccountingSupplierParty)
	supplierParty := root.CreateElement("cac:AccountingSupplierParty")
	writeUBLParty(supplierParty.CreateElement("cac:Party"), inv.Seller, true)

	// BG-7: Buyer (AccountingCustomerParty)
	customerParty := root.CreateElement("cac:AccountingCustomerParty")
	writeUBLParty(customerParty.CreateElement("cac:Party"), inv.Buyer, false)

	// BG-10: Payee (optional)
	if inv.PayeeTradeParty != nil {
		payeeParty := root.CreateElement("cac:PayeeParty")
		writeUBLParty(payeeParty, *inv.PayeeTradeParty, false)
	}

	// BG-11: Seller tax representative (optional)
	if inv.SellerTaxRepresentativeTradeParty != nil {
		taxRepParty := root.CreateElement("cac:TaxRepresentativeParty")
		writeUBLParty(taxRepParty, *inv.SellerTaxRepresentativeTradeParty, false)
	}

	// BG-13: Delivery information
	if inv.ShipTo != nil || !inv.OccurrenceDateTime.IsZero() {
		delivery := root.CreateElement("cac:Delivery")

		// BT-72: Actual delivery date
		addTimeUBL(delivery, "cbc:ActualDeliveryDate", inv.OccurrenceDateTime)

		// Delivery location/party
		if inv.ShipTo != nil {
			deliveryParty := delivery.CreateElement("cac:DeliveryParty")
			writeUBLParty(deliveryParty, *inv.ShipTo, false)
		}
	}
}

// writeUBLParty writes a single party (reusable for Seller, Buyer, Payee, etc.)
func writeUBLParty(parent *etree.Element, party Party, isSeller bool) {
	// Electronic address (BT-34, BT-49)
	if party.URIUniversalCommunication != "" {
		endpoint := parent.CreateElement("cbc:EndpointID")
		if party.URIUniversalCommunicationScheme != "" {
			endpoint.CreateAttr("schemeID", party.URIUniversalCommunicationScheme)
		}
		endpoint.SetText(party.URIUniversalCommunication)
	}

	// Party identification
	for _, id := range party.ID {
		partyID := parent.CreateElement("cac:PartyIdentification")
		partyID.CreateElement("cbc:ID").SetText(id)
	}

	for _, gid := range party.GlobalID {
		partyID := parent.CreateElement("cac:PartyIdentification")
		idElt := partyID.CreateElement("cbc:ID")
		idElt.CreateAttr("schemeID", gid.Scheme)
		idElt.SetText(gid.ID)
	}

	// Party name
	if party.Name != "" {
		partyName := parent.CreateElement("cac:PartyName")
		partyName.CreateElement("cbc:Name").SetText(party.Name)
	}

	// Postal address
	if party.PostalAddress != nil {
		addr := parent.CreateElement("cac:PostalAddress")

		if party.PostalAddress.Line1 != "" {
			addr.CreateElement("cbc:StreetName").SetText(party.PostalAddress.Line1)
		}
		if party.PostalAddress.Line2 != "" {
			addr.CreateElement("cbc:AdditionalStreetName").SetText(party.PostalAddress.Line2)
		}
		if party.PostalAddress.Line3 != "" {
			addrLine := addr.CreateElement("cac:AddressLine")
			addrLine.CreateElement("cbc:Line").SetText(party.PostalAddress.Line3)
		}
		if party.PostalAddress.City != "" {
			addr.CreateElement("cbc:CityName").SetText(party.PostalAddress.City)
		}
		if party.PostalAddress.PostcodeCode != "" {
			addr.CreateElement("cbc:PostalZone").SetText(party.PostalAddress.PostcodeCode)
		}
		if party.PostalAddress.CountrySubDivisionName != "" {
			addr.CreateElement("cbc:CountrySubentity").SetText(party.PostalAddress.CountrySubDivisionName)
		}
		if party.PostalAddress.CountryID != "" {
			country := addr.CreateElement("cac:Country")
			country.CreateElement("cbc:IdentificationCode").SetText(party.PostalAddress.CountryID)
		}
	}

	// Legal organization
	if party.SpecifiedLegalOrganization != nil {
		legalEntity := parent.CreateElement("cac:PartyLegalEntity")
		legalEntity.CreateElement("cbc:RegistrationName").SetText(party.Name)

		if party.SpecifiedLegalOrganization.ID != "" {
			companyID := legalEntity.CreateElement("cbc:CompanyID")
			if party.SpecifiedLegalOrganization.Scheme != "" {
				companyID.CreateAttr("schemeID", party.SpecifiedLegalOrganization.Scheme)
			}
			companyID.SetText(party.SpecifiedLegalOrganization.ID)
		}
	}

	// Tax registration
	if party.VATaxRegistration != "" {
		taxScheme := parent.CreateElement("cac:PartyTaxScheme")
		taxScheme.CreateElement("cbc:CompanyID").SetText(party.VATaxRegistration)
		taxScheme.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText("VAT")
	}

	if party.FCTaxRegistration != "" {
		taxScheme := parent.CreateElement("cac:PartyTaxScheme")
		taxScheme.CreateElement("cbc:CompanyID").SetText(party.FCTaxRegistration)
		taxScheme.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText("FC")
	}

	// Contact information
	for _, contact := range party.DefinedTradeContact {
		contactElt := parent.CreateElement("cac:Contact")

		if contact.PersonName != "" {
			contactElt.CreateElement("cbc:Name").SetText(contact.PersonName)
		}
		if contact.PhoneNumber != "" {
			contactElt.CreateElement("cbc:Telephone").SetText(contact.PhoneNumber)
		}
		if contact.EMail != "" {
			contactElt.CreateElement("cbc:ElectronicMail").SetText(contact.EMail)
		}
	}
}

// writeUBLAllowanceCharge writes document-level allowances and charges (BG-20, BG-21)
func writeUBLAllowanceCharge(inv *Invoice, root *etree.Element, prefix string) {
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		acElt := root.CreateElement("cac:AllowanceCharge")
		acElt.CreateElement("cbc:ChargeIndicator").SetText(fmt.Sprintf("%t", ac.ChargeIndicator))

		if ac.ReasonCode != 0 {
			acElt.CreateElement("cbc:AllowanceChargeReasonCode").SetText(fmt.Sprintf("%d", ac.ReasonCode))
		}

		if ac.Reason != "" {
			acElt.CreateElement("cbc:AllowanceChargeReason").SetText(ac.Reason)
		}

		if !ac.CalculationPercent.IsZero() {
			acElt.CreateElement("cbc:MultiplierFactorNumeric").SetText(formatPercent(ac.CalculationPercent))
		}

		acElt.CreateElement("cbc:Amount").SetText(ac.ActualAmount.StringFixed(2))

		if !ac.BasisAmount.IsZero() {
			acElt.CreateElement("cbc:BaseAmount").SetText(ac.BasisAmount.StringFixed(2))
		}

		// Tax category
		if ac.CategoryTradeTaxCategoryCode != "" {
			taxCat := acElt.CreateElement("cac:TaxCategory")
			taxCat.CreateElement("cbc:ID").SetText(ac.CategoryTradeTaxCategoryCode)
			taxCat.CreateElement("cbc:Percent").SetText(formatPercent(ac.CategoryTradeTaxRateApplicablePercent))
			taxCat.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText(ac.CategoryTradeTaxType)
		}
	}
}

// writeUBLTaxTotal writes the tax breakdown (BG-23)
func writeUBLTaxTotal(inv *Invoice, root *etree.Element, prefix string) {
	taxTotal := root.CreateElement("cac:TaxTotal")

	// BT-110: Total tax amount in document currency
	taxAmount := taxTotal.CreateElement("cbc:TaxAmount")
	currency := inv.TaxTotalCurrency
	if currency == "" {
		currency = inv.InvoiceCurrencyCode
	}
	taxAmount.CreateAttr("currencyID", currency)
	taxAmount.SetText(inv.TaxTotal.StringFixed(2))

	// BG-23: VAT breakdown (TaxSubtotal elements)
	for _, tradeTax := range inv.TradeTaxes {
		subtotal := taxTotal.CreateElement("cac:TaxSubtotal")

		// Taxable amount
		subtotal.CreateElement("cbc:TaxableAmount").SetText(tradeTax.BasisAmount.StringFixed(2))

		// Tax amount
		subtotal.CreateElement("cbc:TaxAmount").SetText(tradeTax.CalculatedAmount.StringFixed(2))

		// Tax category
		taxCat := subtotal.CreateElement("cac:TaxCategory")
		taxCat.CreateElement("cbc:ID").SetText(tradeTax.CategoryCode)
		taxCat.CreateElement("cbc:Percent").SetText(formatPercent(tradeTax.Percent))

		if tradeTax.ExemptionReason != "" {
			taxCat.CreateElement("cbc:TaxExemptionReason").SetText(tradeTax.ExemptionReason)
		}

		if tradeTax.ExemptionReasonCode != "" {
			taxCat.CreateElement("cbc:TaxExemptionReasonCode").SetText(tradeTax.ExemptionReasonCode)
		}

		taxScheme := taxCat.CreateElement("cac:TaxScheme")
		taxScheme.CreateElement("cbc:ID").SetText(tradeTax.TypeCode)
	}

	// BT-111: Tax total in accounting currency (if different)
	if inv.TaxCurrencyCode != "" && inv.TaxCurrencyCode != inv.InvoiceCurrencyCode {
		taxTotalVAT := root.CreateElement("cac:TaxTotal")
		taxAmountVAT := taxTotalVAT.CreateElement("cbc:TaxAmount")
		taxAmountVAT.CreateAttr("currencyID", inv.TaxCurrencyCode)
		taxAmountVAT.SetText(inv.TaxTotalAccounting.StringFixed(2))
	}
}

// writeUBLMonetarySummation writes the monetary totals (BT-106 to BT-115)
func writeUBLMonetarySummation(inv *Invoice, root *etree.Element, prefix string) {
	lmt := root.CreateElement("cac:LegalMonetaryTotal")

	// BT-106: Sum of Invoice line net amount
	lmt.CreateElement("cbc:LineExtensionAmount").SetText(inv.LineTotal.StringFixed(2))

	// BT-107: Sum of allowances on document level
	if !inv.AllowanceTotal.IsZero() {
		lmt.CreateElement("cbc:AllowanceTotalAmount").SetText(inv.AllowanceTotal.StringFixed(2))
	}

	// BT-108: Sum of charges on document level
	if !inv.ChargeTotal.IsZero() {
		lmt.CreateElement("cbc:ChargeTotalAmount").SetText(inv.ChargeTotal.StringFixed(2))
	}

	// BT-109: Invoice total amount without VAT
	lmt.CreateElement("cbc:TaxExclusiveAmount").SetText(inv.TaxBasisTotal.StringFixed(2))

	// BT-112: Invoice total amount with VAT
	lmt.CreateElement("cbc:TaxInclusiveAmount").SetText(inv.GrandTotal.StringFixed(2))

	// BT-113: Paid amount
	if !inv.TotalPrepaid.IsZero() {
		lmt.CreateElement("cbc:PrepaidAmount").SetText(inv.TotalPrepaid.StringFixed(2))
	}

	// BT-114: Rounding amount
	if !inv.RoundingAmount.IsZero() {
		lmt.CreateElement("cbc:PayableRoundingAmount").SetText(inv.RoundingAmount.StringFixed(2))
	}

	// BT-115: Amount due for payment
	lmt.CreateElement("cbc:PayableAmount").SetText(inv.DuePayableAmount.StringFixed(2))
}

// writeUBLPaymentMeans writes payment means elements (BG-16, BG-17, BG-18, BG-19)
func writeUBLPaymentMeans(inv *Invoice, root *etree.Element, prefix string) {
	for _, pm := range inv.PaymentMeans {
		pmElt := root.CreateElement("cac:PaymentMeans")

		// BT-81: Payment means type code
		pmElt.CreateElement("cbc:PaymentMeansCode").SetText(fmt.Sprintf("%d", pm.TypeCode))

		// BT-82: Payment means text
		if pm.Information != "" {
			pmElt.CreateElement("cbc:InstructionNote").SetText(pm.Information)
		}

		// BT-83: Remittance information
		if inv.PaymentReference != "" {
			pmElt.CreateElement("cbc:PaymentID").SetText(inv.PaymentReference)
		}

		// BG-18: Payment card information
		if pm.ApplicableTradeSettlementFinancialCardID != "" {
			cardAccount := pmElt.CreateElement("cac:CardAccount")
			cardAccount.CreateElement("cbc:PrimaryAccountNumberID").SetText(pm.ApplicableTradeSettlementFinancialCardID)

			if pm.ApplicableTradeSettlementFinancialCardCardholderName != "" {
				cardAccount.CreateElement("cbc:HolderName").SetText(pm.ApplicableTradeSettlementFinancialCardCardholderName)
			}
		}

		// BG-19: Direct debit
		if pm.PayerPartyDebtorFinancialAccountIBAN != "" {
			mandate := pmElt.CreateElement("cac:PaymentMandate")
			payerAccount := mandate.CreateElement("cac:PayerFinancialAccount")
			payerAccount.CreateElement("cbc:ID").SetText(pm.PayerPartyDebtorFinancialAccountIBAN)
		}

		// BG-17: Credit transfer (IBAN/BIC)
		if pm.PayeePartyCreditorFinancialAccountIBAN != "" {
			payeeAccount := pmElt.CreateElement("cac:PayeeFinancialAccount")
			payeeAccount.CreateElement("cbc:ID").SetText(pm.PayeePartyCreditorFinancialAccountIBAN)

			if pm.PayeePartyCreditorFinancialAccountName != "" {
				payeeAccount.CreateElement("cbc:Name").SetText(pm.PayeePartyCreditorFinancialAccountName)
			}

			if pm.PayeeSpecifiedCreditorFinancialInstitutionBIC != "" {
				branch := payeeAccount.CreateElement("cac:FinancialInstitutionBranch")
				branch.CreateElement("cbc:ID").SetText(pm.PayeeSpecifiedCreditorFinancialInstitutionBIC)
			}
		}
	}
}

// writeUBLPaymentTerms writes payment terms (BT-20, BT-9, BT-89)
func writeUBLPaymentTerms(inv *Invoice, root *etree.Element, prefix string) {
	for _, pt := range inv.SpecifiedTradePaymentTerms {
		ptElt := root.CreateElement("cac:PaymentTerms")

		// BT-20: Payment terms
		if pt.Description != "" {
			ptElt.CreateElement("cbc:Note").SetText(pt.Description)
		}

		// BT-9: Payment due date
		addTimeUBL(ptElt, "cbc:PaymentDueDate", pt.DueDate)

		// BT-89: Direct debit mandate identifier
		if pt.DirectDebitMandateID != "" {
			ptElt.CreateElement("cbc:PaymentMeansID").SetText(pt.DirectDebitMandateID)
		}
	}
}

// writeUBLLines writes all invoice line items (BG-25)
func writeUBLLines(inv *Invoice, root *etree.Element, prefix string) {
	for _, line := range inv.InvoiceLines {
		lineElt := root.CreateElement("cac:InvoiceLine")

		// BT-126: Invoice line identifier
		lineElt.CreateElement("cbc:ID").SetText(line.LineID)

		// BT-127: Invoice line note
		if line.Note != "" {
			lineElt.CreateElement("cbc:Note").SetText(line.Note)
		}

		// BT-129: Invoiced quantity
		qty := lineElt.CreateElement("cbc:InvoicedQuantity")
		qty.CreateAttr("unitCode", line.BilledQuantityUnit)
		qty.SetText(line.BilledQuantity.StringFixed(4))

		// BT-131: Invoice line net amount
		lineElt.CreateElement("cbc:LineExtensionAmount").SetText(line.Total.StringFixed(2))

		// BT-19: Buyer accounting reference
		if line.ReceivableSpecifiedTradeAccountingAccount != "" {
			lineElt.CreateElement("cac:AccountingCost").SetText(line.ReceivableSpecifiedTradeAccountingAccount)
		}

		// BG-26: Invoice line period
		if !line.BillingSpecifiedPeriodStart.IsZero() || !line.BillingSpecifiedPeriodEnd.IsZero() {
			period := lineElt.CreateElement("cac:InvoicePeriod")
			addTimeUBL(period, "cbc:StartDate", line.BillingSpecifiedPeriodStart)
			addTimeUBL(period, "cbc:EndDate", line.BillingSpecifiedPeriodEnd)
		}

		// BT-132: Referenced purchase order line
		if line.BuyerOrderReferencedDocument != "" {
			orderLineRef := lineElt.CreateElement("cac:OrderLineReference")
			orderLineRef.CreateElement("cbc:LineID").SetText(line.BuyerOrderReferencedDocument)
		}

		// BT-128: Invoice line object identifier
		if line.AdditionalReferencedDocumentID != "" {
			docRef := lineElt.CreateElement("cac:DocumentReference")
			docRef.CreateElement("cbc:ID").SetText(line.AdditionalReferencedDocumentID)

			if line.AdditionalReferencedDocumentTypeCode != "" {
				docRef.CreateElement("cbc:DocumentTypeCode").SetText(line.AdditionalReferencedDocumentTypeCode)
			}
		}

		// BG-27: Line level allowances
		// BG-28: Line level charges
		for _, ac := range line.InvoiceLineAllowances {
			writeUBLLineAllowanceCharge(lineElt, ac, false)
		}
		for _, ac := range line.InvoiceLineCharges {
			writeUBLLineAllowanceCharge(lineElt, ac, true)
		}

		// Item information
		writeUBLLineItem(lineElt, line)

		// Price information
		writeUBLLinePrice(lineElt, line)
	}
}

// writeUBLLineAllowanceCharge writes a line-level allowance or charge
func writeUBLLineAllowanceCharge(parent *etree.Element, ac AllowanceCharge, isCharge bool) {
	acElt := parent.CreateElement("cac:AllowanceCharge")
	acElt.CreateElement("cbc:ChargeIndicator").SetText(fmt.Sprintf("%t", isCharge))

	if ac.ReasonCode != 0 {
		acElt.CreateElement("cbc:AllowanceChargeReasonCode").SetText(fmt.Sprintf("%d", ac.ReasonCode))
	}

	if ac.Reason != "" {
		acElt.CreateElement("cbc:AllowanceChargeReason").SetText(ac.Reason)
	}

	if !ac.CalculationPercent.IsZero() {
		acElt.CreateElement("cbc:MultiplierFactorNumeric").SetText(formatPercent(ac.CalculationPercent))
	}

	acElt.CreateElement("cbc:Amount").SetText(ac.ActualAmount.StringFixed(2))

	if !ac.BasisAmount.IsZero() {
		acElt.CreateElement("cbc:BaseAmount").SetText(ac.BasisAmount.StringFixed(2))
	}
}

// writeUBLLineItem writes item-specific information within a line
func writeUBLLineItem(parent *etree.Element, line InvoiceLine) {
	item := parent.CreateElement("cac:Item")

	// BT-153: Item name
	item.CreateElement("cbc:Name").SetText(line.ItemName)

	// BT-154: Item description
	if line.Description != "" {
		item.CreateElement("cbc:Description").SetText(line.Description)
	}

	// BT-155: Item Seller's identifier
	if line.ArticleNumber != "" {
		sellersID := item.CreateElement("cac:SellersItemIdentification")
		sellersID.CreateElement("cbc:ID").SetText(line.ArticleNumber)
	}

	// BT-156: Item Buyer's identifier
	if line.ArticleNumberBuyer != "" {
		buyersID := item.CreateElement("cac:BuyersItemIdentification")
		buyersID.CreateElement("cbc:ID").SetText(line.ArticleNumberBuyer)
	}

	// BT-157: Item standard identifier
	if line.GlobalID != "" {
		standardID := item.CreateElement("cac:StandardItemIdentification")
		idElt := standardID.CreateElement("cbc:ID")
		if line.GlobalIDType != "" {
			idElt.CreateAttr("schemeID", line.GlobalIDType)
		}
		idElt.SetText(line.GlobalID)
	}

	// BT-159: Item country of origin
	if line.OriginTradeCountry != "" {
		originCountry := item.CreateElement("cac:OriginCountry")
		originCountry.CreateElement("cbc:IdentificationCode").SetText(line.OriginTradeCountry)
	}

	// BT-158: Item classification identifier
	for _, class := range line.ProductClassification {
		commodityClass := item.CreateElement("cac:CommodityClassification")
		classCode := commodityClass.CreateElement("cbc:ItemClassificationCode")
		if class.ListID != "" {
			classCode.CreateAttr("listID", class.ListID)
		}
		if class.ListVersionID != "" {
			classCode.CreateAttr("listVersionID", class.ListVersionID)
		}
		classCode.SetText(class.ClassCode)
	}

	// BG-32: Item attributes
	for _, char := range line.Characteristics {
		prop := item.CreateElement("cac:AdditionalItemProperty")
		prop.CreateElement("cbc:Name").SetText(char.Description)
		prop.CreateElement("cbc:Value").SetText(char.Value)
	}

	// Tax information
	taxCat := item.CreateElement("cac:ClassifiedTaxCategory")
	taxCat.CreateElement("cbc:ID").SetText(line.TaxCategoryCode)
	taxCat.CreateElement("cbc:Percent").SetText(formatPercent(line.TaxRateApplicablePercent))
	taxCat.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText(line.TaxTypeCode)
}

// writeUBLLinePrice writes price information within a line
func writeUBLLinePrice(parent *etree.Element, line InvoiceLine) {
	price := parent.CreateElement("cac:Price")

	// BT-146: Item net price
	price.CreateElement("cbc:PriceAmount").SetText(line.NetPrice.StringFixed(2))

	// BT-149: Item price base quantity
	if !line.BasisQuantity.IsZero() {
		baseQty := price.CreateElement("cbc:BaseQuantity")
		baseQty.CreateAttr("unitCode", line.BilledQuantityUnit)
		baseQty.SetText(line.BasisQuantity.StringFixed(4))
	}

	// BT-147: Item price allowances (from GrossPrice)
	for _, ac := range line.AppliedTradeAllowanceCharge {
		writeUBLLineAllowanceCharge(price, ac, ac.ChargeIndicator)
	}
}
