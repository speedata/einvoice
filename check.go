package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

func (inv *Invoice) checkBRO() {
	var sum decimal.Decimal

	// BR-CO-9 VAT identifier country prefix validation
	// VAT identifiers shall have ISO 3166-1 alpha-2 country prefix (e.g., DE, FR, GB). Greece may use 'EL' or 'GR'.
	validateVATIDPrefix := func(vatID string, fieldName string) {
		if vatID == "" {
			return // Empty VAT IDs are handled by other rules
		}
		if len(vatID) < 2 {
			inv.addViolation(rules.BRCO9, fmt.Sprintf("%s must have at least 2-character country prefix", fieldName))
			return
		}
		// Extract first 2 characters as potential country code
		prefix := vatID[:2]
		// Check if it's uppercase letters (basic validation for country code format)
		if prefix[0] < 'A' || prefix[0] > 'Z' || prefix[1] < 'A' || prefix[1] > 'Z' {
			inv.addViolation(rules.BRCO9, fmt.Sprintf("%s must start with 2-letter ISO 3166-1 alpha-2 country code (got: %s)", fieldName, prefix))
		}
		// Note: Full validation against all valid ISO codes would require maintaining a complete list
	}

	validateVATIDPrefix(inv.Seller.VATaxRegistration, "Seller VAT identifier (BT-31)")
	validateVATIDPrefix(inv.Buyer.VATaxRegistration, "Buyer VAT identifier (BT-48)")
	if inv.SellerTaxRepresentativeTradeParty != nil {
		validateVATIDPrefix(inv.SellerTaxRepresentativeTradeParty.VATaxRegistration, "Seller tax representative VAT identifier (BT-63)")
	}

	// BR-CO-3 Rechnung
	// Umsatzsteuerdatum "Value added tax point date" (BT-7) und Code für das Umsatzsteuerdatum "Value added tax point date code" (BT-8)
	// schließen sich gegenseitig aus.
	for _, tax := range inv.TradeTaxes {
		if !tax.TaxPointDate.IsZero() && tax.DueDateTypeCode != "" {
			inv.addViolation(rules.BRCO3, "TaxPointDate and DueDateTypeCode are mutually exclusive")
			break
		}
	}

	// BR-CO-4 Rechnungsposition
	// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss anhand der Umsatzsteuerkategorie des in Rechnung gestellten Postens "Invoiced item VAT
	// category code" (BT-151) kategorisiert werden.
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "" {
			inv.addViolation(rules.BRCO4, fmt.Sprintf("Invoice line %s missing VAT category code", line.LineID))
		}
	}

	// BR-CO-10 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Sum of Invoice line net amount" (BT-106) entspricht der Summe aller Inhalte der Elemente "Invoice line net amount"
	// (BT-131).
	sum = decimal.Zero
	for _, line := range inv.InvoiceLines {
		sum = sum.Add(line.Total)
	}
	if !inv.LineTotal.Equal(sum) {
		inv.addViolation(rules.BRCO10, fmt.Sprintf("Line total %s does not match sum of invoice lines %s", inv.LineTotal.String(), sum.String()))
	}

	// BR-CO-11 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Sum of allowances on document level" (BT-107) entspricht der Summe aller Inhalte
	// der Elemente "Document level allowance amount" (BT-92).
	calculatedAllowanceTotal := decimal.Zero
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator {
			calculatedAllowanceTotal = calculatedAllowanceTotal.Add(ac.ActualAmount)
		}
	}
	if !inv.AllowanceTotal.Equal(calculatedAllowanceTotal) {
		inv.addViolation(rules.BRCO11, fmt.Sprintf("Allowance total %s does not match sum of document level allowances %s", inv.AllowanceTotal.String(), calculatedAllowanceTotal.String()))
	}

	// BR-CO-12 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Sum of charges on document level" (BT-108) entspricht der Summe aller Inhalte
	// der Elemente "Document level charge amount" (BT-99).
	calculatedChargeTotal := decimal.Zero
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator {
			calculatedChargeTotal = calculatedChargeTotal.Add(ac.ActualAmount)
		}
	}
	if !inv.ChargeTotal.Equal(calculatedChargeTotal) {
		inv.addViolation(rules.BRCO12, fmt.Sprintf("Charge total %s does not match sum of document level charges %s", inv.ChargeTotal.String(), calculatedChargeTotal.String()))
	}

	// BR-CO-13 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount without VAT" (BT-109) entspricht der Summe aus "Sum of Invoice line net amount"
	// (BT-106) abzüglich "Sum of allowances on document level" (BT-107) zuzüglich "Sum of charges on document level" (BT-108).
	expectedTaxBasisTotal := inv.LineTotal.Sub(inv.AllowanceTotal).Add(inv.ChargeTotal)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		inv.addViolation(rules.BRCO13, fmt.Sprintf("Tax basis total %s does not match LineTotal - AllowanceTotal + ChargeTotal = %s", inv.TaxBasisTotal.String(), expectedTaxBasisTotal.String()))
	}

	// BR-CO-14 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total VAT amount" (BT-110) entspricht der
	// Summe aller Inhalte der Elemente "VAT category tax amount" (BT-117).
	calculatedTaxTotal := decimal.Zero
	for _, tax := range inv.TradeTaxes {
		calculatedTaxTotal = calculatedTaxTotal.Add(tax.CalculatedAmount)
	}
	if !inv.TaxTotal.Equal(calculatedTaxTotal) {
		inv.addViolation(rules.BRCO14, fmt.Sprintf("Invoice total VAT amount %s does not match sum of VAT category amounts %s", inv.TaxTotal.String(), calculatedTaxTotal.String()))
	}

	// BR-CO-15 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount with VAT" (BT-112) entspricht der Summe aus "Invoice total amount without VAT"
	// (BT-109) und "Invoice total VAT amount" (BT-110).
	expectedGrandTotal := inv.TaxBasisTotal.Add(inv.TaxTotal)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		inv.addViolation(rules.BRCO15, fmt.Sprintf("Grand total %s does not match TaxBasisTotal + TaxTotal = %s", inv.GrandTotal.String(), expectedGrandTotal.String()))
	}

	// BR-CO-16 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Amount due for payment" (BT-115) entspricht der Summe aus "Invoice total amount with VAT" (BT-112)
	// abzüglich "Paid amount" (BT-113) zuzüglich "Rounding amount" (BT-114).
	expectedDuePayableAmount := inv.GrandTotal.Sub(inv.TotalPrepaid).Add(inv.RoundingAmount)
	if !inv.DuePayableAmount.Equal(expectedDuePayableAmount) {
		inv.addViolation(rules.BRCO16, fmt.Sprintf("Due payable amount %s does not match GrandTotal - TotalPrepaid + RoundingAmount = %s", inv.DuePayableAmount.String(), expectedDuePayableAmount.String()))
	}

	// BR-CO-17 Umsatzsteueraufschlüsselung
	// Der Inhalt des Elementes "VAT category tax amount" (BT-117) entspricht dem Inhalt des Elementes "VAT category taxable amount" (BT-116),
	// multipliziert mit dem Inhalt des Elementes "VAT category rate" (BT-119) geteilt durch 100, gerundet auf zwei Dezimalstellen.
	for _, tax := range inv.TradeTaxes {
		expected := tax.BasisAmount.Mul(tax.Percent).Div(decimal.NewFromInt(100)).Round(2)
		if !tax.CalculatedAmount.Equal(expected) {
			inv.addViolation(rules.BRCO17, fmt.Sprintf("VAT category tax amount %s does not match expected %s (basis %s × rate %s ÷ 100)", tax.CalculatedAmount.String(), expected.String(), tax.BasisAmount.String(), tax.Percent.String()))
		}
	}

	// BR-CO-18 Umsatzsteueraufschlüsselung
	// Eine Rechnung (INVOICE) soll mindestens eine Gruppe "VAT BREAKDOWN" (BG-23) enthalten.
	if len(inv.TradeTaxes) < 1 {
		inv.addViolation(rules.BRCO18, "Invoice should contain at least one VAT BREAKDOWN")
	}

	// BR-CO-19 Liefer- oder Rechnungszeitraum
	// Wenn die Gruppe "INVOICING PERIOD" (BG-14) verwendet wird, müssen entweder das Element "Invoicing period start date" (BT-73) oder das
	// Element "Invoicing period end date" (BT-74) oder beide gefüllt sein.
	// Note: If at least one date is set (!IsZero()), then BR-CO-19 is automatically satisfied.
	// The rule only applies when BG-14 is present in XML, which our writer ensures by only writing when at least one date exists.

	// BR-CO-20 Rechnungszeitraum auf Positionsebene
	// Wenn die Gruppe "INVOICE LINE PERIOD" (BG-26) verwendet wird, müssen entweder das Element "Invoice line period start date" (BT-134) oder
	// das Element "Invoice line period end date" (BT-135) oder beide gefüllt sein.
	// Note: If at least one date is set (!IsZero()), then BR-CO-20 is automatically satisfied.
	// The rule only applies when BG-26 is present in XML, which our writer ensures by only writing when at least one date exists.

	// BR-CO-25 Rechnung
	// Im Falle eines positiven Zahlbetrags "Amount due for payment" (BT-115) muss entweder das Element Fälligkeitsdatum "Payment due date" (BT-9)
	// oder das Element Zahlungsbedingungen "Payment terms" (BT-20) vorhanden sein.
	if inv.DuePayableAmount.GreaterThan(decimal.Zero) {
		hasPaymentDueDate := false
		hasPaymentTerms := false

		for _, term := range inv.SpecifiedTradePaymentTerms {
			if !term.DueDate.IsZero() {
				hasPaymentDueDate = true
			}
			if term.Description != "" {
				hasPaymentTerms = true
			}
		}

		if !hasPaymentDueDate && !hasPaymentTerms {
			inv.addViolation(rules.BRCO25, "If amount due for payment is positive, either payment due date or payment terms must be present")
		}
	}

	// BR-CO-26 Verkäufer
	// In order for the buyer to automatically identify a supplier, at least one of the following shall be present:
	// - Seller identifier (BT-29)
	// - Seller legal registration identifier (BT-30)
	// - Seller VAT identifier (BT-31)
	hasSellerID := len(inv.Seller.ID) > 0 || len(inv.Seller.GlobalID) > 0
	hasLegalReg := inv.Seller.SpecifiedLegalOrganization != nil && inv.Seller.SpecifiedLegalOrganization.ID != ""
	hasVATID := inv.Seller.VATaxRegistration != ""
	if !hasSellerID && !hasLegalReg && !hasVATID {
		inv.addViolation(rules.BRCO26, "At least one seller identifier must be present: Seller ID (BT-29), Legal registration (BT-30), or VAT ID (BT-31)")
	}

	// BR-CO-27 Zahlungsanweisungen
	// Either the IBAN or a Proprietary ID (BT-84) shall be used for payment account identifier.
	// Note: BT-84 can be either IBAN or Proprietary ID; this rule ensures at least one is specified
	// when payment account information is provided.
	for _, pm := range inv.PaymentMeans {
		if pm.PayeePartyCreditorFinancialAccountIBAN == "" && pm.PayeePartyCreditorFinancialAccountProprietaryID == "" {
			// Only validate if TypeCode indicates a payment method that requires account info
			// TypeCodes 30, 58 = credit transfer, which need account identifiers (already validated by BR-61)
			if pm.TypeCode == 30 || pm.TypeCode == 58 {
				inv.addViolation(rules.BRCO27, "Payment account identifier (BT-84) must be provided as either IBAN or Proprietary ID")
			}
		}
	}

	// BR-CO-5 Document level allowance reason consistency
	// Document level allowance reason code (BT-98) and Document level allowance reason (BT-97)
	// shall indicate the same type of allowance.
	// Implementation: If one is provided, the other should also be provided for consistency.
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator { // This is an allowance
			hasReasonCode := ac.ReasonCode != 0
			hasReason := ac.Reason != ""
			if hasReasonCode && !hasReason {
				inv.addViolation(rules.BRCO5, "Document level allowance reason code (BT-98) is provided but reason text (BT-97) is missing")
			} else if !hasReasonCode && hasReason {
				inv.addViolation(rules.BRCO5, "Document level allowance reason text (BT-97) is provided but reason code (BT-98) is missing")
			}
		}
	}

	// BR-CO-6 Document level charge reason consistency
	// Document level charge reason code (BT-105) and Document level charge reason (BT-104)
	// shall indicate the same type of charge.
	// Implementation: If one is provided, the other should also be provided for consistency.
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator { // This is a charge
			hasReasonCode := ac.ReasonCode != 0
			hasReason := ac.Reason != ""
			if hasReasonCode && !hasReason {
				inv.addViolation(rules.BRCO6, "Document level charge reason code (BT-105) is provided but reason text (BT-104) is missing")
			} else if !hasReasonCode && hasReason {
				inv.addViolation(rules.BRCO6, "Document level charge reason text (BT-104) is provided but reason code (BT-105) is missing")
			}
		}
	}

	// BR-CO-7 Invoice line allowance reason consistency
	// Invoice line allowance reason code (BT-140) and Invoice line allowance reason (BT-139)
	// shall indicate the same type of allowance reason.
	// Implementation: If one is provided, the other should also be provided for consistency.
	for i, line := range inv.InvoiceLines {
		for _, ac := range line.InvoiceLineAllowances {
			hasReasonCode := ac.ReasonCode != 0
			hasReason := ac.Reason != ""
			if hasReasonCode && !hasReason {
				inv.addViolation(rules.BRCO7, fmt.Sprintf("Line %d: allowance reason code (BT-140) is provided but reason text (BT-139) is missing", i+1))
			} else if !hasReasonCode && hasReason {
				inv.addViolation(rules.BRCO7, fmt.Sprintf("Line %d: allowance reason text (BT-139) is provided but reason code (BT-140) is missing", i+1))
			}
		}
	}

	// BR-CO-8 Invoice line charge reason consistency
	// Invoice line charge reason code (BT-145) and Invoice line charge reason (BT-144)
	// shall indicate the same type of charge reason.
	// Implementation: If one is provided, the other should also be provided for consistency.
	for i, line := range inv.InvoiceLines {
		for _, ac := range line.InvoiceLineCharges {
			hasReasonCode := ac.ReasonCode != 0
			hasReason := ac.Reason != ""
			if hasReasonCode && !hasReason {
				inv.addViolation(rules.BRCO8, fmt.Sprintf("Line %d: charge reason code (BT-145) is provided but reason text (BT-144) is missing", i+1))
			} else if !hasReasonCode && hasReason {
				inv.addViolation(rules.BRCO8, fmt.Sprintf("Line %d: charge reason text (BT-144) is provided but reason code (BT-145) is missing", i+1))
			}
		}
	}

}

func (inv *Invoice) checkBR() {
	// BR-1
	// Eine Rechnung (INVOICE) muss eine Spezifikationskennung "Specification identification“ (BT-24) enthalten.
	if inv.Profile == CProfileUnknown {
		inv.addViolation(rules.BR1, "Could not determine the profile in GuidelineSpecifiedDocumentContextParameter")
	}
	// 	BR-2 Rechnung
	// Eine Rechnung (INVOICE) muss eine Rechnungsnummer "Invoice number“ (BT-1) enthalten.
	if inv.InvoiceNumber == "" {
		inv.addViolation(rules.BR2, "No invoice number found")
	}
	// BR-3 Rechnung
	// Eine Rechnung (INVOICE) muss ein Rechnungsdatum "Invoice issue date“ (BT-2) enthalten.
	if inv.InvoiceDate.IsZero() {
		inv.addViolation(rules.BR3, "Date is zero")
	}
	// BR-4 Rechnung
	// Eine Rechnung (INVOICE) muss einen Rechnungstyp-Code "Invoice type code“ (BT-3) enthalten.
	if inv.InvoiceTypeCode == 0 {
		inv.addViolation(rules.BR4, "Invoice type code is 0")
	}
	// BR-5 Rechnung
	// Eine Rechnung (INVOICE) muss einen Währungs-Code "Invoice currency code“ (BT-5) enthalten.
	if inv.InvoiceCurrencyCode == "" {
		inv.addViolation(rules.BR5, "Invoice currency code is empty")
	}
	// BR-6 Verkäufer
	// Eine Rechnung (INVOICE) muss den Verkäufernamen "Seller name“ (BT-27) enthalten.
	if inv.Seller.Name == "" {
		inv.addViolation(rules.BR6, "Seller name is empty")
	}
	// BR-7 Käufer
	// Eine Rechnung (INVOICE) muss den Erwerbernamen "Buyer name“ (BT-44) enthalten.
	if inv.Buyer.Name == "" {
		inv.addViolation(rules.BR7, "Buyer name is empty")
	}
	// BR-8 Verkäufer
	// Eine Rechnung (INVOICE) muss die postalische Anschrift des Verkäufers "SELLER POSTAL ADDRESS“ (BG-5) enthalten.
	if inv.Seller.PostalAddress == nil {
		inv.addViolation(rules.BR8, "Seller has no postal address")
	} else {
		// BR-9 Verkäufer
		// Eine postalische Anschrift des Verkäufers "SELLER POSTAL ADDRESS“ (BG-5) muss einen Verkäufer-Ländercode "Seller country code“ (BT-40) enthalten.
		if inv.Seller.PostalAddress.CountryID == "" {
			inv.addViolation(rules.BR9, "Seller country code empty")
		}
	}
	if inv.Profile > CProfileMinimum {
		// BR-10 Käufer
		// Eine Rechnung (INVOICE) muss die postalische Anschrift des Erwerbers "BUYER POSTAL ADDRESS“ (BG-8) enthalten.
		if inv.Buyer.PostalAddress == nil {
			inv.addViolation(rules.BR10, "Buyer has no postal address")
		} else {
			// BR-11 Käufer
			// Eine postalische Anschrift des Erwerbers "BUYER POSTAL ADDRESS“ (BG-8) muss einen Erwerber-Ländercode "Buyer country code“ (BT-55)
			// enthalten.
			if inv.Buyer.PostalAddress.CountryID == "" {
				inv.addViolation(rules.BR11, "Buyer country code empty")
			}
		}
	}
	// BR-12 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss die Summe der Rechnungspositionen-Nettobeträge "Sum of Invoice line net amount“ (BT-106) enthalten.
	if inv.LineTotal.IsZero() {
		inv.addViolation(rules.BR12, "Line total is zero")
	}
	// BR-13 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den Gesamtbetrag der Rechnung ohne Umsatzsteuer "Invoice total amount without VAT“ (BT-109) enthalten.
	if inv.TaxBasisTotal.IsZero() {
		inv.addViolation(rules.BR13, "TaxBasisTotal zero")
	}
	// BR-14 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den Gesamtbetrag der Rechnung mit Umsatzsteuer "Invoice total amount with VAT“ (BT-112) enthalten.
	if inv.GrandTotal.IsZero() {
		inv.addViolation(rules.BR14, "GrandTotal is zero")
	}
	// BR-15 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den ausstehenden Betrag "Amount due for payment“ (BT-115) enthalten.
	if inv.DuePayableAmount.IsZero() {
		inv.addViolation(rules.BR15, "DuePayableAmount is zero")
	}
	// BR-16 Rechnung
	// Eine Rechnung (INVOICE) muss mindestens eine Rechnungsposition "INVOICE LINE“ (BG-25) enthalten.
	if is(CProfileBasic, inv) {
		if len(inv.InvoiceLines) == 0 {
			inv.addViolation(rules.BR16, "Invoice lines must be at least 1")
		}
	}
	// BR-17 Zahlungsempfänger
	// Eine Rechnung (INVOICE) muss den Namen des Zahlungsempfängers "Payee name“ (BT-59) enthalten, wenn sich der Zahlungsempfänger "PAYEE“
	// (BG-10) vom Verkäufer "SELLER“ (BG-4) unterscheidet.
	if inv.PayeeTradeParty != nil {
		if inv.PayeeTradeParty.Name == "" {
			inv.addViolation(rules.BR17, "Payee has no name, although different from seller")
		}
	}
	// BR-18 Steuerbevollmächtigter des Verkäufers
	// Eine Rechnung (INVOICE) muss den Namen des Steuervertreters des Verkäufers "Seller tax representative name“ (BT-62) enthalten, wenn der
	// Verkäufer "SELLER“ (BG-4) einen Steuervertreter (BG-11) hat.
	if trp := inv.SellerTaxRepresentativeTradeParty; trp != nil {
		if trp.Name == "" {
			inv.addViolation(rules.BR18, "Tax representative has no name, although seller has specified one")
		}
		// BR-19 Steuerbevollmächtigter des Verkäufers
		// Eine Rechnung (INVOICE) muss die postalische Anschrift des Steuervertreters "SELLER TAX REPRESENTATIVE POSTAL ADDRESS“ (BG-12) enthalten,
		// wenn der Verkäufer "SELLER“ (BG-4) einen Steuervertreter hat.
		if trp.PostalAddress == nil {
			inv.addViolation(rules.BR19, "Tax representative has no postal address")
		} else {
			// BR-20 Steuerbevollmächtigter des Verkäufers
			// Die postalische Anschrift des Steuervertreters des Verkäufers "SELLER TAX REPRESENTATIVE POSTAL ADDRESS" (BG-12) muss einen
			// Steuervertreter-Ländercode enthalten, wenn der Verkäufer "SELLER" (BG-4) einen Steuervertreter hat.
			if trp.PostalAddress.CountryID == "" {
				inv.addViolation(rules.BR20, "Tax representative postal address missing country code")
			}
		}
	}
	for _, line := range inv.InvoiceLines {
		// BR-21 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss eine eindeutige Bezeichnung "Invoice line identifier“ (BT-126) haben.
		if line.LineID == "" {
			inv.addViolation(rules.BR21, "Line has no line ID")
		}
		// BR-22 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss die Menge der in der betreffenden Position in Rechnung gestellten Waren oder
		// Dienstleistungen als Einzelposten "Invoiced quantity“ (BT-129) enthalten.
		if line.BilledQuantity.IsZero() {
			inv.addViolation(rules.BR22, "Line has no billed quantity")
		}
		// BR-23 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss eine Einheit zur Mengenangabe "Invoiced quantity unit of measure code" (BT-130)
		// enthalten.
		if line.BilledQuantityUnit == "" {
			inv.addViolation(rules.BR23, "Line's billed quantity has no unit")
		}

		// BR-24 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss den Nettobetrag der Rechnungsposition "Invoice line net amount" (BT-131) enthalten.
		if line.Total.IsZero() {
			inv.addViolation(rules.BR24, "Line's net amount not found")
		}

		// BR-25 Artikelinformationen
		// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss den Namen des Postens "Item name" (BT-153) enthalten.
		if line.ItemName == "" {
			inv.addViolation(rules.BR25, "Line's item name missing")
		}

		// BR-26 Detailinformationen zum Preis
		// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss den Preis des Postens, ohne Umsatzsteuer, nach Abzug des für diese Rechnungsposition
		// geltenden Rabatts "Item net price" (BT-146) beinhalten.
		if line.NetPrice.IsZero() {
			inv.addViolation(rules.BR26, "Line's item net price not found")
		}

		// BR-27 Nettopreis des Artikels
		// Der Artikel-Nettobetrag "Item net price" (BT-146) darf nicht negativ sein.
		if line.NetPrice.IsNegative() {
			inv.addViolation(rules.BR27, "Net price must not be negative")
		}
		// BR-28 Detailinformationen zum Preis
		// Der Einheitspreis ohne Umsatzsteuer vor Abzug des Postenpreisrabatts einer Rechnungsposition "Item gross price" (BT-148) darf nicht negativ
		// sein.
		if line.GrossPrice.IsNegative() {
			inv.addViolation(rules.BR28, "Gross price must not be negative")
		}
	}
	// BR-29 Rechnungszeitraum
	// Wenn Start- und Enddatum des Rechnungszeitraums gegeben sind, muss das Enddatum "Invoicing period end date“ (BT-74) nach dem Startdatum
	// "Invoicing period start date“ (BT-73) liegen oder mit diesem identisch sein.
	if inv.BillingSpecifiedPeriodEnd.Before(inv.BillingSpecifiedPeriodStart) {
		inv.addViolation(rules.BR29, "Billing period end must be after start")
	}
	for _, line := range inv.InvoiceLines {
		// BR-30 Rechnungszeitraum auf Positionsebene
		// Wenn Start- und Enddatum des Rechnungspositionenzeitraums gegeben sind, muss das Enddatum "Invoice line period end date“ (BT-135) nach
		// dem Startdatum "Invoice line period start date“ (BT-134) liegen oder mit diesem identisch sein.
		if line.BillingSpecifiedPeriodEnd.Before(line.BillingSpecifiedPeriodStart) {
			inv.addViolation(rules.BR30, "Line item billing period end must be after or identical to start")
		}
	}

	// Initialize applicableTradeTaxes map for BR-45 validation
	// Use composite key of CategoryCode + Percent to properly group by tax category
	var applicableTradeTaxes = make(map[string]decimal.Decimal, len(inv.TradeTaxes))
	for _, lineitem := range inv.InvoiceLines {
		key := lineitem.TaxCategoryCode + "_" + lineitem.TaxRateApplicablePercent.String()
		applicableTradeTaxes[key] = applicableTradeTaxes[key].Add(lineitem.Total)
	}

	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		// BR-66 Specified Trade Allowance Charge
		// Each Specified Trade Allowance Charge shall contain a Charge Indicator.
		// Note: In Go, the boolean ChargeIndicator field always has a value (true or false),
		// so this rule is implicitly satisfied. This validation is kept for documentation
		// and to align with the EN 16931 specification.

		// Add to applicableTradeTaxes for BR-45 validation
		key := ac.CategoryTradeTaxCategoryCode + "_" + ac.CategoryTradeTaxRateApplicablePercent.String()
		amount := ac.ActualAmount
		if !ac.ChargeIndicator {
			amount = amount.Neg()
		}
		applicableTradeTaxes[key] = applicableTradeTaxes[key].Add(amount)

		if ac.ChargeIndicator {
			// BR-36 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Betrag "Document level charge amount" (BT-99)
			// aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.addViolation(rules.BR36, "Charge must not be zero")
			}

			// BR-37 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Umsatzsteuer-Code "Document level charge VAT
			// category code" (BT-102) aufweisen.
			if ac.CategoryTradeTaxCategoryCode == "" {
				inv.addViolation(rules.BR37, "Charge tax category code not set")
			}
			// BR-38 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Abgabegrund "Document level charge reason" (BT-104)
			// oder einen entsprechenden Code "Document level charge reason code" (BT-105) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.addViolation(rules.BR38, "Charge reason empty or code unset")
			}
			// BR-39 Zuschläge auf Dokumentenebene
			// Der Betrag einer Abgabe auf Dokumentenebene "Document level charge amount" (BT-99) darf nicht negativ sein.
			if ac.ActualAmount.LessThan(decimal.Zero) {
				inv.addViolation(rules.BR39, "Document level charge amount must not be negative")
			}
			// BR-40 Zuschläge auf Dokumentenebene
			// Der Basisbetrag einer Abgabe auf Dokumentenebene "Document level charge base amount" (BT-100) darf nicht negativ sein.
			if ac.BasisAmount.LessThan(decimal.Zero) {
				inv.addViolation(rules.BR40, "Document level charge base amount must not be negative")
			}
		} else {
			// BR-31 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Betrag "Document level allowance amount"
			// (BT-92) aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.addViolation(rules.BR31, "Allowance must not be zero")
			}
			// BR-32 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Umsatzsteuer-Code "Document level
			// allowance VAT category code" (BT-95) aufweisen.
			if ac.CategoryTradeTaxCategoryCode == "" {
				inv.addViolation(rules.BR32, "Allowance tax category code not set")
			}
			// BR-33 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Nachlassgrund "Document level allowance
			// reason" (BT-97) oder einen entsprechenden Code "Document level allowance reason code" (BT-98") aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.addViolation(rules.BR33, "Allowance reason empty or code unset")
			}
			// BR-34 Abschläge auf Dokumentenebene
			// Der Betrag eines Nachlasses auf Dokumentenebene "Document level allowance amount" (BT-92) darf nicht negativ sein.
			if ac.ActualAmount.LessThan(decimal.Zero) {
				inv.addViolation(rules.BR34, "Document level allowance amount must not be negative")
			}
			// BR-35 Abschläge auf Dokumentenebene
			// Der Basisbetrag eines Nachlasses auf Dokumentenebene "Document level allowance base amount" (BT-93) darf nicht negativ sein.
			if ac.BasisAmount.LessThan(decimal.Zero) {
				inv.addViolation(rules.BR35, "Document level allowance base amount must not be negative")
			}
		}
	}

	for _, line := range inv.InvoiceLines {
		// BR-41 Abschläge auf Ebene der Rechnungsposition
		// Jeder Nachlass auf der Ebene der Rechnungsposition "INVOICE LINE ALLOWANCES“ (BG-27) muss einen Betrag "Invoice line allowance amount“
		// (BT-136) aufweisen.
		for _, ac := range line.InvoiceLineAllowances {
			if ac.ActualAmount.IsZero() {
				inv.addViolation(rules.BR41, "Line allowance amount zero")
			}
			// BR-42 Abschläge auf Ebene der Rechnungsposition
			// Jeder Nachlass auf der Ebene der Rechnungsposition "INVOICE LINE ALLOWANCES“ (BG-27) muss einen Nachlassgrund "Invoice line allowance
			// reason“ (BT-139) oder einen entsprechenden Code "Invoice line allowance reason code“ (BT-140) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.addViolation(rules.BR42, "Line allowance must have a reason")
			}
		}
		for _, ac := range line.InvoiceLineCharges {
			// BR-43 Charge ou frais sur ligne de facture
			// Jede Abgabe auf der Ebene der Rechnungsposition "INVOICE LINE CHARGES“ (BG-28) muss einen Betrag "Invoice line charge amount“ (BT-141)
			// aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.addViolation(rules.BR43, "Line charge amount zero")
			}
			// BR-44 Charge ou frais sur ligne de facture
			// Jede Abgabe auf der Ebene der Rechnungsposition "INVOICE LINE CHARGES“ (BG-28) muss einen Abgabegrund "Invoice line charge reason“ (BT-
			// 144) oder einen entsprechenden Code "Invoice line charge reason code“ (BT-145) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.addViolation(rules.BR44, "Line charge must have a reason")
			}
		}
	}

	for _, tt := range inv.TradeTaxes {
		// BR-46 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN" (BG-23) muss den für
		// die betreffende Umsatzsteuerkategorie zu entrichtenden Gesamtbetrag
		// "VAT category tax amount" (BT-117) aufweisen.
		// Note: Zero is a valid value for exempt categories (E, AE, Z, G, O, IC, IG, IP).
		// Category-specific rules (BR-E-9, BR-AE-9, BR-Z-9, etc.) enforce when zero is required.
		// This rule only ensures the field is present, which it always is after parsing or calculation.

		// BR-48 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN" (BG-23) muss einen
		// Umsatzsteuersatz gemäß einer Kategorie "VAT category rate" (BT-119)
		// haben. Sofern die Rechnung von der Umsatzsteuer ausgenommen ist, ist
		// "0" zu übermitteln.
		// Note: Zero is a valid and required value for categories E, AE, Z, G, O, IC, IG, IP.
		// Category-specific rules (BR-S-5, BR-E-5, BR-AE-5, etc.) enforce the correct rate per category.
		// This rule only ensures the field is present, which it always is after parsing or calculation.

		// BR-45 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN" (BG-23) muss die
		// Summe aller nach dem jeweiligen Schlüssel zu versteuernden Beträge
		// "VAT category taxable amount" (BT-116) aufweisen.
		key := tt.CategoryCode + "_" + tt.Percent.String()
		if !applicableTradeTaxes[key].Equal(tt.BasisAmount) {
			inv.addViolation(rules.BR45, "Applicable trade tax basis amount not equal to the sum of line total")

		}
		// BR-47 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN" (BG-23) muss über
		// eine codierte Bezeichnung einer Umsatzsteuerkategorie "VAT category
		// code" (BT-118) definiert werden.
		if tt.CategoryCode == "" {
			inv.addViolation(rules.BR47, "CategoryCode not set for applicable trade tax")
		}
	}
	for _, pm := range inv.PaymentMeans {
		// BR-49 Zahlungsanweisungen
		// Die Zahlungsinstruktionen "PAYMENT INSTRUCTIONS“ (BG-16) müssen den Zahlungsart-Code "Payment means type code“ (BT-81) enthalten.
		if pm.TypeCode == 0 {
			inv.addViolation(rules.BR49, "Payment means type code must be set")
		}
	}
	// BR-50 Kontoinformationen
	// Die Kennung des Kontos, auf das die Zahlung erfolgen soll "Payment
	// account identifier" (BT-84), muss angegeben werden, wenn
	// Überweisungsinformationen in der Rechnung angegeben werden.
	// Note: This is a weaker version of BR-61 and is generally covered by BR-61

	// BR-51 Karteninformationen
	// Die letzten vier bis sechs Ziffern der Kreditkartennummer "Payment card
	// primary account number" (BT-87) sollen angegeben werden, wenn
	// Informationen zur Kartenzahlung übermittelt werden.
	// Note: This uses "sollen" (should) not "muss" (must), so it's a recommendation not requirement

	// BR-52 Rechnungsbegründende Unterlagen
	// Jede rechnungsbegründende Unterlage muss einen Dokumentenbezeichner
	// "Supporting document reference" (BT-122) haben.
	for _, doc := range inv.AdditionalReferencedDocument {
		if doc.IssuerAssignedID == "" {
			inv.addViolation(rules.BR52, "Supporting document must have a reference")
		}
	}

	// BR-53 Gesamtsummen auf Dokumentenebene
	// Wenn eine Währung für die Umsatzsteuerabrechnung angegeben wurde, muss
	// der Umsatzsteuergesamtbetrag in der Abrechnungswährung "Invoice total VAT
	// amount in accounting currency" (BT-111) angegeben werden.
	if inv.TaxCurrencyCode != "" && inv.TaxTotalVAT.IsZero() {
		inv.addViolation(rules.BR53, "Tax total in accounting currency must be specified when tax currency code is provided")
	}

	// BR-54 Artikelattribute
	// Jede Eigenschaft eines in Rechnung gestellten Postens "ITEM ATTRIBUTES"
	// (BG-32) muss eine Bezeichnung "Item attribute name" (BT-160) und einen
	// Wert "Item attribute value" (BT-161) haben.
	for _, line := range inv.InvoiceLines {
		for _, char := range line.Characteristics {
			if char.Description == "" || char.Value == "" {
				inv.addViolation(rules.BR54, "Item attribute must have both name and value")
			}
		}
	}

	// BR-55 Referenz auf die vorausgegangene Rechnung
	// Jede Bezugnahme auf eine vorausgegangene Rechnung "Preceding Invoice
	// reference" (BT-25) muss die Nummer der vorausgegangenen Rechnung
	// enthalten.
	for _, ref := range inv.InvoiceReferencedDocument {
		if ref.ID == "" {
			inv.addViolation(rules.BR55, "Preceding invoice reference must contain invoice number")
		}
	}

	// BR-56 Steuerbevollmächtigter des Verkäufers
	// Jeder Steuervertreter des Verkäufers "SELLER TAX REPRESENTATIVE PARTY"
	// (BG-11) muss eine Umsatzsteuer-Identifikationsnummer "Seller tax
	// representative VAT identifier" (BT-63) haben.
	if inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration == "" {
		inv.addViolation(rules.BR56, "Seller tax representative must have VAT identifier")
	}

	// BR-57 Lieferanschrift
	// Jede Lieferadresse "DELIVER TO ADDRESS" (BG-15) muss einen entsprechenden
	// Ländercode "Deliver to country code" (BT-80) enthalten.
	if inv.ShipTo != nil && inv.ShipTo.PostalAddress != nil && inv.ShipTo.PostalAddress.CountryID == "" {
		inv.addViolation(rules.BR57, "Deliver-to address must have country code")
	}

	// BR-61 Zahlungsanweisungen
	// Wenn der Zahlungsmittel-Typ SEPA, lokale Überweisung oder
	// Nicht-SEPA-Überweisung ist, muss der "Payment account identifier" (BT-84)
	// des Zahlungsempfängers angegeben werden.
	for _, pm := range inv.PaymentMeans {
		// TypeCode 30 = Credit transfer, 58 = SEPA credit transfer
		if (pm.TypeCode == 30 || pm.TypeCode == 58) && pm.PayeePartyCreditorFinancialAccountIBAN == "" && pm.PayeePartyCreditorFinancialAccountProprietaryID == "" {
			inv.addViolation(rules.BR61, "Payment account identifier required for credit transfer payment types")
		}
	}

	// BR-62 Elektronische Adresse des Verkäufers
	// Im Element "Seller electronic address" (BT-34) muss die Komponente
	// "Scheme Identifier" vorhanden sein.
	if inv.Seller.URIUniversalCommunication != "" && inv.Seller.URIUniversalCommunicationScheme == "" {
		inv.addViolation(rules.BR62, "Seller electronic address must have scheme identifier")
	}

	// BR-63 Elektronische Adresse des Käufers
	// Im Element "Buyer electronic address" (BT-49) muss die Komponente "Scheme
	// Identifier" vorhanden sein.
	if inv.Buyer.URIUniversalCommunication != "" && inv.Buyer.URIUniversalCommunicationScheme == "" {
		inv.addViolation(rules.BR63, "Buyer electronic address must have scheme identifier")
	}

	// BR-64 Kennung eines Artikels nach registriertem Schema
	// Im Element "Item standard identifier" (BT-157) muss die Komponente
	// "Scheme Identifier" vorhanden sein.
	for _, line := range inv.InvoiceLines {
		if line.GlobalID != "" && line.GlobalIDType == "" {
			inv.addViolation(rules.BR64, "Item standard identifier must have scheme identifier")
		}
	}

	// BR-65 Kennung der Artikelklassifizierung
	// Im Element "Item classification identifier" (BT-158) muss die Komponente
	// "Scheme Identifier" vorhanden sein.
	for _, line := range inv.InvoiceLines {
		for _, classification := range line.ProductClassification {
			if classification.ClassCode != "" && classification.ListID == "" {
				inv.addViolation(rules.BR65, "Item classification identifier must have scheme identifier")
			}
		}
	}

	// BR-B-1 Split payment (Italian domestic invoices)
	// An Invoice where the VAT category code is "Split payment" (B) shall be a domestic Italian invoice.
	// This means both seller and buyer must be located in Italy (IT).
	hasSplitPayment := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "B" {
			hasSplitPayment = true
			break
		}
	}
	if !hasSplitPayment {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "B" {
				hasSplitPayment = true
				break
			}
		}
	}
	if hasSplitPayment {
		// Check seller country
		sellerCountry := ""
		if inv.Seller.PostalAddress != nil {
			sellerCountry = inv.Seller.PostalAddress.CountryID
		}
		// Check buyer country
		buyerCountry := ""
		if inv.Buyer.PostalAddress != nil {
			buyerCountry = inv.Buyer.PostalAddress.CountryID
		}

		if sellerCountry != "IT" || buyerCountry != "IT" {
			inv.addViolation(rules.BRB1, "Split payment VAT category (B) requires both seller and buyer to be in Italy (IT)")
		}
	}

	// BR-B-2 Split payment and Standard rated exclusion
	// An Invoice with Split payment (B) shall not contain Standard rated (S) VAT category.
	hasStandardRated := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "S" {
			hasStandardRated = true
			break
		}
	}
	if !hasStandardRated {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "S" {
				hasStandardRated = true
				break
			}
		}
	}
	if hasSplitPayment && hasStandardRated {
		inv.addViolation(rules.BRB2, "Invoice with Split payment VAT category (B) must not contain Standard rated (S) category")
	}

	// VAT category validations - delegated to specialized methods
	inv.checkVATStandard()
	inv.checkVATReverse()
	inv.checkVATExempt()
	inv.checkVATZero()
	inv.checkVATExport()
	inv.checkVATIntracommunity()
	inv.checkVATIGIC()
	inv.checkVATIPSI()
	inv.checkVATNotSubject()
}

// hasMaxDecimals checks if a decimal value has at most maxDecimals decimal places.
// Returns true if the value has maxDecimals or fewer decimal places.
func hasMaxDecimals(value decimal.Decimal, maxDecimals int) bool {
	// The exponent is the negative of the number of decimal places
	// e.g., 123.45 has exponent -2, 123.456 has exponent -3
	return value.Exponent() >= -int32(maxDecimals)
}

func (inv *Invoice) checkBRDEC() {
	// Helper function to validate decimal precision
	checkDecimalPrecision := func(value decimal.Decimal, fieldName string, btCode string, rule rules.Rule) {
		if !value.IsZero() && !hasMaxDecimals(value, 2) {
			inv.addViolation(rule, fmt.Sprintf("%s (%s) has more than 2 decimal places: %s", fieldName, btCode, value.String()))
		}
	}

	// BR-DEC-01: Document level allowance amount (BT-92)
	// BR-DEC-02: Document level allowance base amount (BT-93)
	// BR-DEC-05: Document level charge amount (BT-99)
	// BR-DEC-06: Document level charge base amount (BT-100)
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator {
			// Allowance
			checkDecimalPrecision(ac.ActualAmount, "Document level allowance amount", "BT-92", rules.BRDEC1)
			checkDecimalPrecision(ac.BasisAmount, "Document level allowance base amount", "BT-93", rules.BRDEC2)
		} else {
			// Charge
			checkDecimalPrecision(ac.ActualAmount, "Document level charge amount", "BT-99", rules.BRDEC5)
			checkDecimalPrecision(ac.BasisAmount, "Document level charge base amount", "BT-100", rules.BRDEC6)
		}
	}

	// BR-DEC-09: Sum of Invoice line net amount (BT-106)
	checkDecimalPrecision(inv.LineTotal, "Sum of Invoice line net amount", "BT-106", rules.BRDEC9)

	// BR-DEC-10: Sum of allowances on document level (BT-107)
	checkDecimalPrecision(inv.AllowanceTotal, "Sum of allowances on document level", "BT-107", rules.BRDEC10)

	// BR-DEC-11: Sum of charges on document level (BT-108)
	checkDecimalPrecision(inv.ChargeTotal, "Sum of charges on document level", "BT-108", rules.BRDEC11)

	// BR-DEC-12: Invoice total amount without VAT (BT-109)
	checkDecimalPrecision(inv.TaxBasisTotal, "Invoice total amount without VAT", "BT-109", rules.BRDEC12)

	// BR-DEC-13: Invoice total VAT amount (BT-110)
	checkDecimalPrecision(inv.TaxTotal, "Invoice total VAT amount", "BT-110", rules.BRDEC13)

	// BR-DEC-14: Invoice total amount with VAT (BT-112)
	checkDecimalPrecision(inv.GrandTotal, "Invoice total amount with VAT", "BT-112", rules.BRDEC14)

	// BR-DEC-15: Invoice total VAT amount in accounting currency (BT-111)
	checkDecimalPrecision(inv.TaxTotalVAT, "Invoice total VAT amount in accounting currency", "BT-111", rules.BRDEC15)

	// BR-DEC-16: Paid amount (BT-113)
	checkDecimalPrecision(inv.TotalPrepaid, "Paid amount", "BT-113", rules.BRDEC16)

	// BR-DEC-17: Rounding amount (BT-114)
	checkDecimalPrecision(inv.RoundingAmount, "Rounding amount", "BT-114", rules.BRDEC17)

	// BR-DEC-18: Amount due for payment (BT-115)
	checkDecimalPrecision(inv.DuePayableAmount, "Amount due for payment", "BT-115", rules.BRDEC18)

	// BR-DEC-19: VAT category taxable amount (BT-116)
	// BR-DEC-20: VAT category tax amount (BT-117)
	for _, tt := range inv.TradeTaxes {
		checkDecimalPrecision(tt.BasisAmount, "VAT category taxable amount", "BT-116", rules.BRDEC19)
		checkDecimalPrecision(tt.CalculatedAmount, "VAT category tax amount", "BT-117", rules.BRDEC20)
	}

	// BR-DEC-23: Invoice line net amount (BT-131)
	// BR-DEC-24: Invoice line allowance amount (BT-136)
	// BR-DEC-25: Invoice line allowance base amount (BT-137)
	// BR-DEC-27: Invoice line charge amount (BT-141)
	// BR-DEC-28: Invoice line charge base amount (BT-142)
	for i, line := range inv.InvoiceLines {
		linePrefix := fmt.Sprintf("Line %d: ", i+1)
		checkDecimalPrecision(line.Total, linePrefix+"Invoice line net amount", "BT-131", rules.BRDEC23)

		for _, allowance := range line.InvoiceLineAllowances {
			checkDecimalPrecision(allowance.ActualAmount, linePrefix+"Invoice line allowance amount", "BT-136", rules.BRDEC24)
			checkDecimalPrecision(allowance.BasisAmount, linePrefix+"Invoice line allowance base amount", "BT-137", rules.BRDEC25)
		}

		for _, charge := range line.InvoiceLineCharges {
			checkDecimalPrecision(charge.ActualAmount, linePrefix+"Invoice line charge amount", "BT-141", rules.BRDEC27)
			checkDecimalPrecision(charge.BasisAmount, linePrefix+"Invoice line charge base amount", "BT-142", rules.BRDEC28)
		}
	}
}
