package einvoice

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	// Write the document structure. The call order follows the element
	// sequence of the UBL 2.1 Invoice and CreditNote schemas.
	writeUBLHeader(inv, root, prefix)
	writeUBLParties(inv, root, prefix)
	writeUBLPaymentMeans(inv, root, prefix)
	writeUBLPaymentTerms(inv, root, prefix)
	writeUBLAllowanceCharge(inv, root, prefix)
	writeUBLTaxTotal(inv, root, prefix)
	writeUBLMonetarySummation(inv, root, prefix)
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

	// BT-9: Payment due date. A CreditNote has no DueDate element, there BT-9
	// is written as PaymentMeans/PaymentDueDate (see writeUBLPaymentMeans).
	if prefix != "cn:" {
		addTimeUBL(root, "cbc:DueDate", paymentDueDate(inv))
	}

	// BT-3: Invoice type code (or CreditNoteTypeCode for credit notes)
	if inv.InvoiceTypeCode != 0 {
		typeCodeElementName := "cbc:InvoiceTypeCode"
		if prefix == "cn:" {
			typeCodeElementName = "cbc:CreditNoteTypeCode"
		}
		root.CreateElement(typeCodeElementName).SetText(inv.InvoiceTypeCode.String())
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

	// BT-19: Accounting cost
	if inv.ReceivableSpecifiedTradeAccountingAccount != "" {
		root.CreateElement("cbc:AccountingCost").SetText(inv.ReceivableSpecifiedTradeAccountingAccount)
	}

	// BT-10: Buyer reference (optional)
	if inv.BuyerReference != "" {
		root.CreateElement("cbc:BuyerReference").SetText(inv.BuyerReference)
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
	for i := range inv.InvoiceReferencedDocument {
		billingRef := root.CreateElement("cac:BillingReference")
		invDocRef := billingRef.CreateElement("cac:InvoiceDocumentReference")
		invDocRef.CreateElement("cbc:ID").SetText(inv.InvoiceReferencedDocument[i].ID)
		if !inv.InvoiceReferencedDocument[i].Date.IsZero() {
			addTimeUBL(invDocRef, "cbc:IssueDate", inv.InvoiceReferencedDocument[i].Date)
		}
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

	// BT-17: Tender or lot reference (a CII document with type code 50).
	// The position of OriginatorDocumentReference differs between Invoice
	// and CreditNote.
	tenderRef := findReferencedDocument(inv, "50")
	if tenderRef != nil && prefix != "cn:" {
		root.CreateElement("cac:OriginatorDocumentReference").CreateElement("cbc:ID").SetText(tenderRef.IssuerAssignedID)
	}

	// BT-12: Contract reference
	if inv.ContractReferencedDocument != "" {
		contractRef := root.CreateElement("cac:ContractDocumentReference")
		contractRef.CreateElement("cbc:ID").SetText(inv.ContractReferencedDocument)
	}

	// BT-11: Project reference. A CreditNote has no ProjectReference element,
	// there BT-11 is an AdditionalDocumentReference with DocumentTypeCode 50.
	if inv.SpecifiedProcuringProjectID != "" && prefix == "cn:" {
		projectRef := root.CreateElement("cac:AdditionalDocumentReference")
		projectRef.CreateElement("cbc:ID").SetText(inv.SpecifiedProcuringProjectID)
		projectRef.CreateElement("cbc:DocumentTypeCode").SetText("50")
	}

	// BT-18: Invoiced object identifier (type code 130, at most once, no
	// description or attachment)
	if objectRef := findReferencedDocument(inv, "130"); objectRef != nil {
		addDocRef := root.CreateElement("cac:AdditionalDocumentReference")
		idElt := addDocRef.CreateElement("cbc:ID")
		if objectRef.ReferenceTypeCode != "" {
			idElt.CreateAttr("schemeID", objectRef.ReferenceTypeCode)
		}
		idElt.SetText(objectRef.IssuerAssignedID)
		addDocRef.CreateElement("cbc:DocumentTypeCode").SetText("130")
	}

	// BG-24: Additional supporting documents (without type code in UBL)
	for i := range inv.AdditionalReferencedDocument {
		doc := &inv.AdditionalReferencedDocument[i]
		if doc.TypeCode == "50" || doc.TypeCode == "130" {
			continue
		}
		addDocRef := root.CreateElement("cac:AdditionalDocumentReference")
		addDocRef.CreateElement("cbc:ID").SetText(doc.IssuerAssignedID)

		// BT-123: Supporting document description
		if doc.Name != "" {
			addDocRef.CreateElement("cbc:DocumentDescription").SetText(doc.Name)
		}

		// UBL allows a single Attachment holding the embedded object
		// followed by the external reference.
		var attachment *etree.Element
		if len(doc.AttachmentBinaryObject) > 0 || doc.URIID != "" {
			attachment = addDocRef.CreateElement("cac:Attachment")
		}

		// BT-125: Handle embedded binary objects (only write if data exists - PEPPOL-EN16931-R008)
		if len(doc.AttachmentBinaryObject) > 0 {
			embeddedDoc := attachment.CreateElement("cbc:EmbeddedDocumentBinaryObject")

			if doc.AttachmentMimeCode != "" {
				embeddedDoc.CreateAttr("mimeCode", doc.AttachmentMimeCode)
			}
			if doc.AttachmentFilename != "" {
				embeddedDoc.CreateAttr("filename", doc.AttachmentFilename)
			}
			embeddedDoc.SetText(base64.StdEncoding.EncodeToString(doc.AttachmentBinaryObject))
		}

		// BT-124: External document location
		if doc.URIID != "" {
			externalRef := attachment.CreateElement("cac:ExternalReference")
			externalRef.CreateElement("cbc:URI").SetText(doc.URIID)
		}
	}

	if tenderRef != nil && prefix == "cn:" {
		root.CreateElement("cac:OriginatorDocumentReference").CreateElement("cbc:ID").SetText(tenderRef.IssuerAssignedID)
	}

	// BT-11: Project reference (Invoice only, after the document references)
	if inv.SpecifiedProcuringProjectID != "" && prefix != "cn:" {
		projectRef := root.CreateElement("cac:ProjectReference")
		projectRef.CreateElement("cbc:ID").SetText(inv.SpecifiedProcuringProjectID)
	}

	// BT-72: Actual delivery date (in Delivery element, handled in writeUBLParties)
}

// findReferencedDocument returns the first additional referenced document
// with the given type code or nil.
func findReferencedDocument(inv *Invoice, typeCode string) *Document {
	for i := range inv.AdditionalReferencedDocument {
		if inv.AdditionalReferencedDocument[i].TypeCode == typeCode {
			return &inv.AdditionalReferencedDocument[i]
		}
	}
	return nil
}

// writeUBLParties writes all party elements (BG-4, BG-7, BG-10, BG-11, BG-13)
func writeUBLParties(inv *Invoice, root *etree.Element, prefix string) {
	// BT-90: Bank assigned creditor identifier. UBL has no element of its
	// own, it is a party identifier of the payee or, without a payee, of the
	// seller.
	sellerCreditorID, payeeCreditorID := inv.CreditorReferenceID, ""
	if inv.PayeeTradeParty != nil {
		sellerCreditorID, payeeCreditorID = "", inv.CreditorReferenceID
	}

	// BG-4: Seller (AccountingSupplierParty)
	supplierParty := root.CreateElement("cac:AccountingSupplierParty")
	writeUBLParty(supplierParty.CreateElement("cac:Party"), &inv.Seller, ublSeller, sellerCreditorID)

	// BG-7: Buyer (AccountingCustomerParty)
	customerParty := root.CreateElement("cac:AccountingCustomerParty")
	writeUBLParty(customerParty.CreateElement("cac:Party"), &inv.Buyer, ublBuyer, "")

	// BG-10: Payee (optional)
	if inv.PayeeTradeParty != nil {
		payeeParty := root.CreateElement("cac:PayeeParty")
		writeUBLParty(payeeParty, inv.PayeeTradeParty, ublPayee, payeeCreditorID)
	}

	// BG-11: Seller tax representative (optional)
	if inv.SellerTaxRepresentativeTradeParty != nil {
		taxRepParty := root.CreateElement("cac:TaxRepresentativeParty")
		writeUBLParty(taxRepParty, inv.SellerTaxRepresentativeTradeParty, ublTaxRepresentative, "")
	}

	// BG-13: Delivery information
	if inv.ShipTo != nil || !inv.OccurrenceDateTime.IsZero() {
		delivery := root.CreateElement("cac:Delivery")

		// BT-72: Actual delivery date
		addTimeUBL(delivery, "cbc:ActualDeliveryDate", inv.OccurrenceDateTime)

		if inv.ShipTo != nil {
			writeUBLDeliveryLocation(delivery, inv.ShipTo)

			// BT-70: Deliver to party name
			if inv.ShipTo.Name != "" {
				deliveryParty := delivery.CreateElement("cac:DeliveryParty")
				deliveryParty.CreateElement("cac:PartyName").CreateElement("cbc:Name").SetText(inv.ShipTo.Name)
			}
		}
	}
}

// writeUBLDeliveryLocation writes the deliver to location identifier (BT-71)
// and the deliver to address (BG-15). UBL allows a single location
// identifier, so a global identifier is preferred over a local one.
func writeUBLDeliveryLocation(delivery *etree.Element, shipTo *Party) {
	var id, scheme string
	if len(shipTo.GlobalID) > 0 {
		id, scheme = shipTo.GlobalID[0].ID, shipTo.GlobalID[0].Scheme
	} else if len(shipTo.ID) > 0 {
		id = shipTo.ID[0]
	}
	if id == "" && shipTo.PostalAddress == nil {
		return
	}

	location := delivery.CreateElement("cac:DeliveryLocation")
	if id != "" {
		idElt := location.CreateElement("cbc:ID")
		if scheme != "" {
			idElt.CreateAttr("schemeID", scheme)
		}
		idElt.SetText(id)
	}
	writeUBLAddress(location, "cac:Address", shipTo.PostalAddress)
}

// writeUBLAddress writes a postal address (BG-5, BG-8, BG-12, BG-15) as a
// child element of parent with the given name. Nothing is written for a nil
// address.
func writeUBLAddress(parent *etree.Element, name string, address *PostalAddress) {
	if address == nil {
		return
	}
	addr := parent.CreateElement(name)

	if address.Line1 != "" {
		addr.CreateElement("cbc:StreetName").SetText(address.Line1)
	}
	if address.Line2 != "" {
		addr.CreateElement("cbc:AdditionalStreetName").SetText(address.Line2)
	}
	if address.City != "" {
		addr.CreateElement("cbc:CityName").SetText(address.City)
	}
	if address.PostcodeCode != "" {
		addr.CreateElement("cbc:PostalZone").SetText(address.PostcodeCode)
	}
	if address.CountrySubDivisionName != "" {
		addr.CreateElement("cbc:CountrySubentity").SetText(address.CountrySubDivisionName)
	}
	if address.Line3 != "" {
		addrLine := addr.CreateElement("cac:AddressLine")
		addrLine.CreateElement("cbc:Line").SetText(address.Line3)
	}
	if address.CountryID != "" {
		country := addr.CreateElement("cac:Country")
		country.CreateElement("cbc:IdentificationCode").SetText(address.CountryID)
	}
}

// ublPartyRole selects which elements writeUBLParty writes. The EN 16931 UBL
// binding allows different subsets of the UBL Party elements per role.
type ublPartyRole int

const (
	ublSeller            ublPartyRole = iota // BG-4
	ublBuyer                                 // BG-7
	ublPayee                                 // BG-10
	ublTaxRepresentative                     // BG-11
)

// writeUBLParty writes a single party with the elements the role allows:
//   - Seller and buyer: their name (BT-27, BT-44) is the legal registration
//     name in PartyLegalEntity, PartyName holds the trading name (BT-28,
//     BT-45). The buyer has at most one identifier (BT-46).
//   - Payee: name (BT-59), at most one identifier (BT-60) and the legal
//     registration identifier (BT-61).
//   - Tax representative: name (BT-62), address (BG-12) and VAT
//     identifier (BT-63).
//
// A non-empty creditorID (BT-90) is written as an additional party
// identifier with scheme SEPA.
func writeUBLParty(parent *etree.Element, party *Party, role ublPartyRole, creditorID string) {
	isLegalParty := role == ublSeller || role == ublBuyer

	// Electronic address (BT-34, BT-49)
	// Write element if either value or scheme is present
	if isLegalParty && (party.URIUniversalCommunication != "" || party.URIUniversalCommunicationScheme != "") {
		endpoint := parent.CreateElement("cbc:EndpointID")
		if party.URIUniversalCommunicationScheme != "" {
			endpoint.CreateAttr("schemeID", party.URIUniversalCommunicationScheme)
		}
		endpoint.SetText(party.URIUniversalCommunication)
	}

	// Party identification (BT-29, BT-46, BT-60)
	switch role {
	case ublSeller:
		for _, id := range party.ID {
			parent.CreateElement("cac:PartyIdentification").CreateElement("cbc:ID").SetText(id)
		}
		for _, gid := range party.GlobalID {
			writeUBLPartyIdentification(parent, gid.ID, gid.Scheme)
		}
	case ublBuyer, ublPayee:
		// Only one identifier is allowed, a global one is preferred
		if len(party.GlobalID) > 0 {
			writeUBLPartyIdentification(parent, party.GlobalID[0].ID, party.GlobalID[0].Scheme)
		} else if len(party.ID) > 0 {
			parent.CreateElement("cac:PartyIdentification").CreateElement("cbc:ID").SetText(party.ID[0])
		}
	}

	if creditorID != "" {
		writeUBLPartyIdentification(parent, creditorID, "SEPA")
	}

	// Party name (BT-28, BT-45 for seller/buyer; BT-59, BT-62 otherwise)
	partyName := party.Name
	if isLegalParty {
		partyName = ""
		if party.SpecifiedLegalOrganization != nil {
			partyName = party.SpecifiedLegalOrganization.TradingBusinessName
		}
	}
	if partyName != "" {
		parent.CreateElement("cac:PartyName").CreateElement("cbc:Name").SetText(partyName)
	}

	// Postal address (BG-5, BG-8, BG-12)
	if role != ublPayee {
		writeUBLAddress(parent, "cac:PostalAddress", party.PostalAddress)
	}

	// Tax registration (BT-31, BT-32, BT-48, BT-63)
	if role != ublPayee && party.VATaxRegistration != "" {
		taxScheme := parent.CreateElement("cac:PartyTaxScheme")
		taxScheme.CreateElement("cbc:CompanyID").SetText(party.VATaxRegistration)
		taxScheme.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText("VAT")
	}

	if role == ublSeller && party.FCTaxRegistration != "" {
		taxScheme := parent.CreateElement("cac:PartyTaxScheme")
		taxScheme.CreateElement("cbc:CompanyID").SetText(party.FCTaxRegistration)
		taxScheme.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText("FC")
	}

	// Legal organization (BT-27, BT-44 and BT-30, BT-47, BT-61)
	hasLegalID := role != ublTaxRepresentative && party.SpecifiedLegalOrganization != nil && party.SpecifiedLegalOrganization.ID != ""
	if (isLegalParty && party.Name != "") || hasLegalID {
		legalEntity := parent.CreateElement("cac:PartyLegalEntity")
		if isLegalParty && party.Name != "" {
			legalEntity.CreateElement("cbc:RegistrationName").SetText(party.Name)
		}

		if hasLegalID {
			companyID := legalEntity.CreateElement("cbc:CompanyID")
			if party.SpecifiedLegalOrganization.Scheme != "" {
				companyID.CreateAttr("schemeID", party.SpecifiedLegalOrganization.Scheme)
			}
			companyID.SetText(party.SpecifiedLegalOrganization.ID)
		}
	}

	// Contact information (BG-6, BG-9). UBL allows a single contact.
	if isLegalParty && len(party.DefinedTradeContact) > 0 {
		contact := &party.DefinedTradeContact[0]
		contactElt := parent.CreateElement("cac:Contact")

		// BT-41, BT-56: Contact point. UBL has no department element, so the
		// department name is used when no person name is given.
		contactName := contact.PersonName
		if contactName == "" {
			contactName = contact.DepartmentName
		}
		if contactName != "" {
			contactElt.CreateElement("cbc:Name").SetText(contactName)
		}
		if contact.PhoneNumber != "" {
			contactElt.CreateElement("cbc:Telephone").SetText(contact.PhoneNumber)
		}
		if contact.EMail != "" {
			contactElt.CreateElement("cbc:ElectronicMail").SetText(contact.EMail)
		}
	}
}

// writeUBLPartyIdentification writes a party identifier with an optional
// scheme.
func writeUBLPartyIdentification(parent *etree.Element, id, scheme string) {
	idElt := parent.CreateElement("cac:PartyIdentification").CreateElement("cbc:ID")
	if scheme != "" {
		idElt.CreateAttr("schemeID", scheme)
	}
	idElt.SetText(id)
}

// writeUBLAllowanceCharge writes document-level allowances and charges (BG-20, BG-21)
func writeUBLAllowanceCharge(inv *Invoice, root *etree.Element, prefix string) {
	for i := range inv.SpecifiedTradeAllowanceCharge {
		acElt := root.CreateElement("cac:AllowanceCharge")
		acElt.CreateElement("cbc:ChargeIndicator").SetText(strconv.FormatBool(inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator))

		if inv.SpecifiedTradeAllowanceCharge[i].ReasonCode != "" {
			acElt.CreateElement("cbc:AllowanceChargeReasonCode").SetText(inv.SpecifiedTradeAllowanceCharge[i].ReasonCode)
		}

		if inv.SpecifiedTradeAllowanceCharge[i].Reason != "" {
			acElt.CreateElement("cbc:AllowanceChargeReason").SetText(inv.SpecifiedTradeAllowanceCharge[i].Reason)
		}

		if !inv.SpecifiedTradeAllowanceCharge[i].CalculationPercent.IsZero() {
			acElt.CreateElement("cbc:MultiplierFactorNumeric").SetText(formatPercent(inv.SpecifiedTradeAllowanceCharge[i].CalculationPercent))
		}

		// BT-92/BT-99: Allowance/Charge amount with currency
		amt := acElt.CreateElement("cbc:Amount")
		amt.CreateAttr("currencyID", inv.InvoiceCurrencyCode)
		amt.SetText(inv.SpecifiedTradeAllowanceCharge[i].ActualAmount.StringFixed(2))

		// BT-93/BT-100: Allowance/Charge base amount with currency
		if !inv.SpecifiedTradeAllowanceCharge[i].BasisAmount.IsZero() {
			baseAmt := acElt.CreateElement("cbc:BaseAmount")
			baseAmt.CreateAttr("currencyID", inv.InvoiceCurrencyCode)
			baseAmt.SetText(inv.SpecifiedTradeAllowanceCharge[i].BasisAmount.StringFixed(2))
		}

		// Tax category
		if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode != "" {
			taxCat := acElt.CreateElement("cac:TaxCategory")
			taxCat.CreateElement("cbc:ID").SetText(inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode)
			// BR-O-6, BR-O-7: no VAT rate for "Not subject to VAT"
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode != "O" {
				taxCat.CreateElement("cbc:Percent").SetText(formatPercent(inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent))
			}
			taxCat.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText(inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxType)
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
	for i := range inv.TradeTaxes {
		subtotal := taxTotal.CreateElement("cac:TaxSubtotal")

		// Taxable amount
		taxableAmt := subtotal.CreateElement("cbc:TaxableAmount")
		taxableAmt.CreateAttr("currencyID", currency)
		taxableAmt.SetText(inv.TradeTaxes[i].BasisAmount.StringFixed(2))

		// Tax amount
		taxAmt := subtotal.CreateElement("cbc:TaxAmount")
		taxAmt.CreateAttr("currencyID", currency)
		taxAmt.SetText(inv.TradeTaxes[i].CalculatedAmount.StringFixed(2))

		// Tax category
		taxCat := subtotal.CreateElement("cac:TaxCategory")
		taxCat.CreateElement("cbc:ID").SetText(inv.TradeTaxes[i].CategoryCode)
		taxCat.CreateElement("cbc:Percent").SetText(formatPercent(inv.TradeTaxes[i].Percent))

		if inv.TradeTaxes[i].ExemptionReasonCode != "" {
			taxCat.CreateElement("cbc:TaxExemptionReasonCode").SetText(inv.TradeTaxes[i].ExemptionReasonCode)
		}

		if inv.TradeTaxes[i].ExemptionReason != "" {
			taxCat.CreateElement("cbc:TaxExemptionReason").SetText(inv.TradeTaxes[i].ExemptionReason)
		}

		taxScheme := taxCat.CreateElement("cac:TaxScheme")
		taxScheme.CreateElement("cbc:ID").SetText(inv.TradeTaxes[i].TypeCode)
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
	currency := inv.InvoiceCurrencyCode

	// BT-106: Sum of Invoice line net amount
	lineExt := lmt.CreateElement("cbc:LineExtensionAmount")
	lineExt.CreateAttr("currencyID", currency)
	lineExt.SetText(inv.LineTotal.StringFixed(2))

	// BT-109: Invoice total amount without VAT
	taxExcl := lmt.CreateElement("cbc:TaxExclusiveAmount")
	taxExcl.CreateAttr("currencyID", currency)
	taxExcl.SetText(inv.TaxBasisTotal.StringFixed(2))

	// BT-112: Invoice total amount with VAT
	taxIncl := lmt.CreateElement("cbc:TaxInclusiveAmount")
	taxIncl.CreateAttr("currencyID", currency)
	taxIncl.SetText(inv.GrandTotal.StringFixed(2))

	// BT-107: Sum of allowances on document level
	if !inv.AllowanceTotal.IsZero() {
		allowTotal := lmt.CreateElement("cbc:AllowanceTotalAmount")
		allowTotal.CreateAttr("currencyID", currency)
		allowTotal.SetText(inv.AllowanceTotal.StringFixed(2))
	}

	// BT-108: Sum of charges on document level
	if !inv.ChargeTotal.IsZero() {
		chargeTotal := lmt.CreateElement("cbc:ChargeTotalAmount")
		chargeTotal.CreateAttr("currencyID", currency)
		chargeTotal.SetText(inv.ChargeTotal.StringFixed(2))
	}

	// BT-113: Paid amount
	if !inv.TotalPrepaid.IsZero() {
		prepaid := lmt.CreateElement("cbc:PrepaidAmount")
		prepaid.CreateAttr("currencyID", currency)
		prepaid.SetText(inv.TotalPrepaid.StringFixed(2))
	}

	// BT-114: Rounding amount
	if !inv.RoundingAmount.IsZero() {
		rounding := lmt.CreateElement("cbc:PayableRoundingAmount")
		rounding.CreateAttr("currencyID", currency)
		rounding.SetText(inv.RoundingAmount.StringFixed(2))
	}

	// BT-115: Amount due for payment
	payable := lmt.CreateElement("cbc:PayableAmount")
	payable.CreateAttr("currencyID", currency)
	payable.SetText(inv.DuePayableAmount.StringFixed(2))
}

// paymentDueDate returns the first payment due date (BT-9) of the payment
// terms. UBL has a single due date per document.
func paymentDueDate(inv *Invoice) time.Time {
	for i := range inv.SpecifiedTradePaymentTerms {
		if !inv.SpecifiedTradePaymentTerms[i].DueDate.IsZero() {
			return inv.SpecifiedTradePaymentTerms[i].DueDate
		}
	}
	return time.Time{}
}

// directDebitMandateID returns the first mandate reference identifier (BT-89)
// of the payment terms. In UBL it is part of the payment means.
func directDebitMandateID(inv *Invoice) string {
	for i := range inv.SpecifiedTradePaymentTerms {
		if inv.SpecifiedTradePaymentTerms[i].DirectDebitMandateID != "" {
			return inv.SpecifiedTradePaymentTerms[i].DirectDebitMandateID
		}
	}
	return ""
}

// isDirectDebit reports whether the payment means carries direct debit
// information (BG-19).
func isDirectDebit(pm *PaymentMeans) bool {
	return pm.PayerPartyDebtorFinancialAccountIBAN != "" || pm.TypeCode == 49 || pm.TypeCode == 59
}

// writeUBLPaymentMeans writes payment means elements (BG-16, BG-17, BG-18, BG-19)
func writeUBLPaymentMeans(inv *Invoice, root *etree.Element, prefix string) {
	// BT-89 belongs to a direct debit payment means. If there is none, it is
	// attached to the first payment means so that it does not get lost.
	mandateID := directDebitMandateID(inv)
	mandateIndex := -1
	if mandateID != "" {
		mandateIndex = 0
		for i := range inv.PaymentMeans {
			if isDirectDebit(&inv.PaymentMeans[i]) {
				mandateIndex = i
				break
			}
		}
	}

	for i := range inv.PaymentMeans {
		pmElt := root.CreateElement("cac:PaymentMeans")

		// BT-81: Payment means type code, BT-82: Payment means text
		code := pmElt.CreateElement("cbc:PaymentMeansCode")
		if inv.PaymentMeans[i].Information != "" {
			code.CreateAttr("name", inv.PaymentMeans[i].Information)
		}
		code.SetText(strconv.Itoa(inv.PaymentMeans[i].TypeCode))

		// BT-9: Payment due date (CreditNote only, an Invoice has cbc:DueDate)
		if prefix == "cn:" && i == 0 {
			addTimeUBL(pmElt, "cbc:PaymentDueDate", paymentDueDate(inv))
		}

		// BT-83: Remittance information
		if inv.PaymentReference != "" {
			pmElt.CreateElement("cbc:PaymentID").SetText(inv.PaymentReference)
		}

		// BG-18: Payment card information
		if inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardID != "" {
			cardAccount := pmElt.CreateElement("cac:CardAccount")
			cardAccount.CreateElement("cbc:PrimaryAccountNumberID").SetText(inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardID)

			if inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardCardholderName != "" {
				cardAccount.CreateElement("cbc:HolderName").SetText(inv.PaymentMeans[i].ApplicableTradeSettlementFinancialCardCardholderName)
			}
		}

		// BG-17: Credit transfer (IBAN/BIC)
		if inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountIBAN != "" {
			payeeAccount := pmElt.CreateElement("cac:PayeeFinancialAccount")
			payeeAccount.CreateElement("cbc:ID").SetText(inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountIBAN)

			if inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountName != "" {
				payeeAccount.CreateElement("cbc:Name").SetText(inv.PaymentMeans[i].PayeePartyCreditorFinancialAccountName)
			}

			if inv.PaymentMeans[i].PayeeSpecifiedCreditorFinancialInstitutionBIC != "" {
				branch := payeeAccount.CreateElement("cac:FinancialInstitutionBranch")
				branch.CreateElement("cbc:ID").SetText(inv.PaymentMeans[i].PayeeSpecifiedCreditorFinancialInstitutionBIC)
			}
		}

		// BG-19: Direct debit (BT-89 mandate reference, BT-91 debited account)
		if i == mandateIndex || inv.PaymentMeans[i].PayerPartyDebtorFinancialAccountIBAN != "" {
			mandate := pmElt.CreateElement("cac:PaymentMandate")
			if i == mandateIndex {
				mandate.CreateElement("cbc:ID").SetText(mandateID)
			}
			if inv.PaymentMeans[i].PayerPartyDebtorFinancialAccountIBAN != "" {
				payerAccount := mandate.CreateElement("cac:PayerFinancialAccount")
				payerAccount.CreateElement("cbc:ID").SetText(inv.PaymentMeans[i].PayerPartyDebtorFinancialAccountIBAN)
			}
		}
	}
}

// writeUBLPaymentTerms writes payment terms (BT-20). The due date (BT-9) and
// the mandate reference (BT-89) are written elsewhere in UBL.
func writeUBLPaymentTerms(inv *Invoice, root *etree.Element, prefix string) {
	// BT-20: Payment terms. UBL allows a single note, so the descriptions of
	// several payment terms (Extended profile) are joined.
	var descriptions []string
	for i := range inv.SpecifiedTradePaymentTerms {
		if inv.SpecifiedTradePaymentTerms[i].Description != "" {
			descriptions = append(descriptions, inv.SpecifiedTradePaymentTerms[i].Description)
		}
	}
	if len(descriptions) > 0 {
		ptElt := root.CreateElement("cac:PaymentTerms")
		ptElt.CreateElement("cbc:Note").SetText(strings.Join(descriptions, "\n"))
	}
}

// writeUBLLines writes all invoice line items (BG-25)
func writeUBLLines(inv *Invoice, root *etree.Element, prefix string) {
	// Determine line element and quantity element names based on document type
	lineElementName := "cac:InvoiceLine"
	quantityElementName := "cbc:InvoicedQuantity"
	if prefix == "cn:" {
		lineElementName = "cac:CreditNoteLine"
		quantityElementName = "cbc:CreditedQuantity"
	}

	for i := range inv.InvoiceLines {
		lineElt := root.CreateElement(lineElementName)

		// BT-126: Invoice line identifier
		lineElt.CreateElement("cbc:ID").SetText(inv.InvoiceLines[i].LineID)

		// BT-127: Invoice line note
		if inv.InvoiceLines[i].Note != "" {
			lineElt.CreateElement("cbc:Note").SetText(inv.InvoiceLines[i].Note)
		}

		// BT-129: Invoiced quantity (or Credited quantity for credit notes)
		qty := lineElt.CreateElement(quantityElementName)
		qty.CreateAttr("unitCode", inv.InvoiceLines[i].BilledQuantityUnit)
		qty.SetText(inv.InvoiceLines[i].BilledQuantity.StringFixed(4))

		// BT-131: Invoice line net amount
		lineExtAmt := lineElt.CreateElement("cbc:LineExtensionAmount")
		lineExtAmt.CreateAttr("currencyID", inv.InvoiceCurrencyCode)
		lineExtAmt.SetText(inv.InvoiceLines[i].Total.StringFixed(2))

		// BT-133: Invoice line Buyer accounting reference
		if inv.InvoiceLines[i].ReceivableSpecifiedTradeAccountingAccount != "" {
			lineElt.CreateElement("cbc:AccountingCost").SetText(inv.InvoiceLines[i].ReceivableSpecifiedTradeAccountingAccount)
		}

		// BG-26: Invoice line period
		if !inv.InvoiceLines[i].BillingSpecifiedPeriodStart.IsZero() || !inv.InvoiceLines[i].BillingSpecifiedPeriodEnd.IsZero() {
			period := lineElt.CreateElement("cac:InvoicePeriod")
			addTimeUBL(period, "cbc:StartDate", inv.InvoiceLines[i].BillingSpecifiedPeriodStart)
			addTimeUBL(period, "cbc:EndDate", inv.InvoiceLines[i].BillingSpecifiedPeriodEnd)
		}

		// BT-132: Referenced purchase order line
		if inv.InvoiceLines[i].BuyerOrderReferencedDocument != "" {
			orderLineRef := lineElt.CreateElement("cac:OrderLineReference")
			orderLineRef.CreateElement("cbc:LineID").SetText(inv.InvoiceLines[i].BuyerOrderReferencedDocument)
		}

		// BT-128: Invoice line object identifier
		if inv.InvoiceLines[i].AdditionalReferencedDocumentID != "" {
			docRef := lineElt.CreateElement("cac:DocumentReference")
			docRef.CreateElement("cbc:ID").SetText(inv.InvoiceLines[i].AdditionalReferencedDocumentID)

			if inv.InvoiceLines[i].AdditionalReferencedDocumentTypeCode != "" {
				docRef.CreateElement("cbc:DocumentTypeCode").SetText(inv.InvoiceLines[i].AdditionalReferencedDocumentTypeCode)
			}
		}

		// BG-27: Line level allowances (BT-136 - must be rounded to 2 decimals)
		// BG-28: Line level charges (BT-141 - must be rounded to 2 decimals)
		for j := range inv.InvoiceLines[i].InvoiceLineAllowances {
			writeUBLLineAllowanceCharge(lineElt, &inv.InvoiceLines[i].InvoiceLineAllowances[j], false, inv.InvoiceCurrencyCode)
		}
		for j := range inv.InvoiceLines[i].InvoiceLineCharges {
			writeUBLLineAllowanceCharge(lineElt, &inv.InvoiceLines[i].InvoiceLineCharges[j], true, inv.InvoiceCurrencyCode)
		}

		// Item information
		writeUBLLineItem(lineElt, &inv.InvoiceLines[i])

		// Price information
		writeUBLLinePrice(lineElt, &inv.InvoiceLines[i], inv.InvoiceCurrencyCode)
	}
}

// writeUBLLineAllowanceCharge writes a line-level allowance or charge (BG-27, BG-28)
func writeUBLLineAllowanceCharge(parent *etree.Element, ac *AllowanceCharge, isCharge bool, currency string) {
	acElt := parent.CreateElement("cac:AllowanceCharge")
	acElt.CreateElement("cbc:ChargeIndicator").SetText(strconv.FormatBool(isCharge))

	if ac.ReasonCode != "" {
		acElt.CreateElement("cbc:AllowanceChargeReasonCode").SetText(ac.ReasonCode)
	}

	if ac.Reason != "" {
		acElt.CreateElement("cbc:AllowanceChargeReason").SetText(ac.Reason)
	}

	if !ac.CalculationPercent.IsZero() {
		acElt.CreateElement("cbc:MultiplierFactorNumeric").SetText(formatPercent(ac.CalculationPercent))
	}

	// BT-136/BT-141: Amount with currency, 2 decimals (BR-DEC-24, BR-DEC-27)
	amt := acElt.CreateElement("cbc:Amount")
	amt.CreateAttr("currencyID", currency)
	amt.SetText(ac.ActualAmount.StringFixed(2))

	// BT-137/BT-142: Base amount with currency
	if !ac.BasisAmount.IsZero() {
		baseAmt := acElt.CreateElement("cbc:BaseAmount")
		baseAmt.CreateAttr("currencyID", currency)
		baseAmt.SetText(ac.BasisAmount.StringFixed(2))
	}
}

// writeUBLLineItem writes item-specific information within a line
func writeUBLLineItem(parent *etree.Element, line *InvoiceLine) {
	item := parent.CreateElement("cac:Item")

	// BT-154: Item description (before name per UBL ordering)
	if line.Description != "" {
		item.CreateElement("cbc:Description").SetText(line.Description)
	}

	// BT-153: Item name
	item.CreateElement("cbc:Name").SetText(line.ItemName)

	// BT-156: Item Buyer's identifier
	if line.ArticleNumberBuyer != "" {
		buyersID := item.CreateElement("cac:BuyersItemIdentification")
		buyersID.CreateElement("cbc:ID").SetText(line.ArticleNumberBuyer)
	}

	// BT-155: Item Seller's identifier
	if line.ArticleNumber != "" {
		sellersID := item.CreateElement("cac:SellersItemIdentification")
		sellersID.CreateElement("cbc:ID").SetText(line.ArticleNumber)
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
	for j := range line.ProductClassification {
		commodityClass := item.CreateElement("cac:CommodityClassification")
		classCode := commodityClass.CreateElement("cbc:ItemClassificationCode")
		if line.ProductClassification[j].ListID != "" {
			classCode.CreateAttr("listID", line.ProductClassification[j].ListID)
		}
		if line.ProductClassification[j].ListVersionID != "" {
			classCode.CreateAttr("listVersionID", line.ProductClassification[j].ListVersionID)
		}
		classCode.SetText(line.ProductClassification[j].ClassCode)
	}

	// Tax information (before additional properties per UBL ordering)
	taxCat := item.CreateElement("cac:ClassifiedTaxCategory")
	taxCat.CreateElement("cbc:ID").SetText(line.TaxCategoryCode)
	if line.TaxCategoryCode != "O" {
		taxCat.CreateElement("cbc:Percent").SetText(formatPercent(line.TaxRateApplicablePercent))
	}
	taxCat.CreateElement("cac:TaxScheme").CreateElement("cbc:ID").SetText(line.TaxTypeCode)

	// BG-32: Item attributes
	for j := range line.Characteristics {
		prop := item.CreateElement("cac:AdditionalItemProperty")
		prop.CreateElement("cbc:Name").SetText(line.Characteristics[j].Description)
		prop.CreateElement("cbc:Value").SetText(line.Characteristics[j].Value)
	}
}

// writeUBLLinePrice writes price information within a line
func writeUBLLinePrice(parent *etree.Element, line *InvoiceLine, currency string) {
	price := parent.CreateElement("cac:Price")

	// BT-146: Item net price (no decimal restriction per EN 16931)
	priceAmt := price.CreateElement("cbc:PriceAmount")
	priceAmt.CreateAttr("currencyID", currency)
	priceAmt.SetText(line.NetPrice.String())

	// BT-149: Item price base quantity with unit code
	if !line.BasisQuantity.IsZero() {
		baseQty := price.CreateElement("cbc:BaseQuantity")
		// Only add unitCode if BasisQuantityUnit was explicitly set
		// Don't fallback to BilledQuantityUnit to preserve round-trip accuracy
		if line.BasisQuantityUnit != "" {
			baseQty.CreateAttr("unitCode", line.BasisQuantityUnit)
		}
		// BT-149: Item price base quantity (no decimal restriction per EN 16931)
		baseQty.SetText(line.BasisQuantity.String())
	}

	// BT-147: Item price discount and BT-148: Item gross price (no decimal
	// restriction per EN 16931). UBL allows a single price allowance, so
	// several allowances and charges on the gross price (Extended profile)
	// are combined into one.
	acs := line.AppliedTradeAllowanceCharge
	if len(acs) == 0 && line.GrossPrice.IsZero() {
		return
	}
	var ac AllowanceCharge
	switch len(acs) {
	case 0:
		ac.ActualAmount = line.GrossPrice.Sub(line.NetPrice)
	case 1:
		ac = acs[0]
	default:
		for j := range acs {
			if acs[j].ChargeIndicator {
				ac.ActualAmount = ac.ActualAmount.Sub(acs[j].ActualAmount)
			} else {
				ac.ActualAmount = ac.ActualAmount.Add(acs[j].ActualAmount)
			}
			if ac.BasisAmount.IsZero() {
				ac.BasisAmount = acs[j].BasisAmount
			}
		}
		if ac.ActualAmount.IsNegative() {
			ac.ChargeIndicator = true
			ac.ActualAmount = ac.ActualAmount.Neg()
		}
	}
	if !line.GrossPrice.IsZero() {
		ac.BasisAmount = line.GrossPrice
	}

	acElt := price.CreateElement("cac:AllowanceCharge")
	acElt.CreateElement("cbc:ChargeIndicator").SetText(strconv.FormatBool(ac.ChargeIndicator))
	if ac.ReasonCode != "" {
		acElt.CreateElement("cbc:AllowanceChargeReasonCode").SetText(ac.ReasonCode)
	}
	if ac.Reason != "" {
		acElt.CreateElement("cbc:AllowanceChargeReason").SetText(ac.Reason)
	}
	if !ac.CalculationPercent.IsZero() {
		acElt.CreateElement("cbc:MultiplierFactorNumeric").SetText(formatPercent(ac.CalculationPercent))
	}
	amt := acElt.CreateElement("cbc:Amount")
	amt.CreateAttr("currencyID", currency)
	amt.SetText(ac.ActualAmount.String())
	if !ac.BasisAmount.IsZero() {
		baseAmt := acElt.CreateElement("cbc:BaseAmount")
		baseAmt.CreateAttr("currencyID", currency)
		baseAmt.SetText(ac.BasisAmount.String())
	}
}
