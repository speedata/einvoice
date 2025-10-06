package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// check is a private method used internally by ParseReader to run validation
// without clearing existing parsing violations. For public use, call Validate() instead.
func (inv *Invoice) check() {
	// Initialize violations slice if nil (preserves existing parsing violations)
	if inv.violations == nil {
		inv.violations = []SemanticError{}
	}

	// Run all validation checks (will append to existing violations)
	inv.checkBR()
	inv.checkBRO()
	inv.checkOther()
}

func (inv *Invoice) checkOther() {
	// Check that line total = billed quantity * net price
	for _, line := range inv.InvoiceLines {
		calcTotal := line.BilledQuantity.Mul(line.NetPrice)
		lineTotal := line.Total
		if !lineTotal.Equal(calcTotal) {
			inv.violations = append(inv.violations, SemanticError{Rule: "Check", InvFields: []string{"BT-146", "BT-149", "BT-131"}, Text: fmt.Sprintf("Line total %s does not match quantity %s * net price %s", lineTotal.String(), line.BilledQuantity.String(), calcTotal.String())})
		}
	}
}

func (inv *Invoice) checkBRO() {
	var sum decimal.Decimal
	// BR-CO-3 Rechnung
	// Umsatzsteuerdatum "Value added tax point date" (BT-7) und Code für das Umsatzsteuerdatum "Value added tax point date code" (BT-8)
	// schließen sich gegenseitig aus.
	for _, tax := range inv.TradeTaxes {
		if !tax.TaxPointDate.IsZero() && tax.DueDateTypeCode != "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-3", InvFields: []string{"BT-7", "BT-8"}, Text: "TaxPointDate and DueDateTypeCode are mutually exclusive"})
			break
		}
	}

	// BR-CO-4 Rechnungsposition
	// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss anhand der Umsatzsteuerkategorie des in Rechnung gestellten Postens "Invoiced item VAT
	// category code" (BT-151) kategorisiert werden.
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-4", InvFields: []string{"BG-25", "BT-151"}, Text: fmt.Sprintf("Invoice line %s missing VAT category code", line.LineID)})
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
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-10", InvFields: []string{"BT-106", "BT-131"}, Text: fmt.Sprintf("Line total %s does not match sum of invoice lines %s", inv.LineTotal.String(), sum.String())})
	}

	// BR-CO-13 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount without VAT" (BT-109) entspricht der Summe aus "Sum of Invoice line net amount"
	// (BT-106) abzüglich "Sum of allowances on document level" (BT-107) zuzüglich "Sum of charges on document level" (BT-108).
	expectedTaxBasisTotal := inv.LineTotal.Sub(inv.AllowanceTotal).Add(inv.ChargeTotal)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-13", InvFields: []string{"BT-109", "BT-106", "BT-107", "BT-108"}, Text: fmt.Sprintf("Tax basis total %s does not match LineTotal - AllowanceTotal + ChargeTotal = %s", inv.TaxBasisTotal.String(), expectedTaxBasisTotal.String())})
	}

	// BR-CO-15 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount with VAT" (BT-112) entspricht der Summe aus "Invoice total amount without VAT"
	// (BT-109) und "Invoice total VAT amount" (BT-110).
	expectedGrandTotal := inv.TaxBasisTotal.Add(inv.TaxTotal)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-15", InvFields: []string{"BT-112", "BT-109", "BT-110"}, Text: fmt.Sprintf("Grand total %s does not match TaxBasisTotal + TaxTotal = %s", inv.GrandTotal.String(), expectedGrandTotal.String())})
	}

	// BR-CO-16 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Amount due for payment" (BT-115) entspricht der Summe aus "Invoice total amount with VAT" (BT-112)
	// abzüglich "Paid amount" (BT-113) zuzüglich "Rounding amount" (BT-114).
	expectedDuePayableAmount := inv.GrandTotal.Sub(inv.TotalPrepaid).Add(inv.RoundingAmount)
	if !inv.DuePayableAmount.Equal(expectedDuePayableAmount) {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-16", InvFields: []string{"BT-115", "BT-112", "BT-113", "BT-114"}, Text: fmt.Sprintf("Due payable amount %s does not match GrandTotal - TotalPrepaid + RoundingAmount = %s", inv.DuePayableAmount.String(), expectedDuePayableAmount.String())})
	}

	// BR-CO-17 Umsatzsteueraufschlüsselung
	// Der Inhalt des Elementes "VAT category tax amount" (BT-117) entspricht dem Inhalt des Elementes "VAT category taxable amount" (BT-116),
	// multipliziert mit dem Inhalt des Elementes "VAT category rate" (BT-119) geteilt durch 100, gerundet auf zwei Dezimalstellen.
	for _, tax := range inv.TradeTaxes {
		expected := tax.BasisAmount.Mul(tax.Percent).Div(decimal.NewFromInt(100)).Round(2)
		if !tax.CalculatedAmount.Equal(expected) {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-17", InvFields: []string{"BT-116", "BT-117", "BT-119"}, Text: fmt.Sprintf("VAT category tax amount %s does not match expected %s (basis %s × rate %s ÷ 100)", tax.CalculatedAmount.String(), expected.String(), tax.BasisAmount.String(), tax.Percent.String())})
		}
	}

	// BR-CO-18 Umsatzsteueraufschlüsselung
	// Eine Rechnung (INVOICE) soll mindestens eine Gruppe "VAT BREAKDOWN" (BG-23) enthalten.
	if len(inv.TradeTaxes) < 1 {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-18", InvFields: []string{"BG-23"}, Text: "Invoice should contain at least one VAT BREAKDOWN"})
	}

	// BR-CO-19 Liefer- oder Rechnungszeitraum
	// Wenn die Gruppe "INVOICING PERIOD" (BG-14) verwendet wird, müssen entweder das Element "Invoicing period start date" (BT-73) oder das
	// Element "Invoicing period end date" (BT-74) oder beide gefüllt sein.
	if !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero() {
		if inv.BillingSpecifiedPeriodStart.IsZero() && inv.BillingSpecifiedPeriodEnd.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-19", InvFields: []string{"BG-14", "BT-73", "BT-74"}, Text: "If invoicing period is used, either start date or end date must be filled"})
		}
	}

	// BR-CO-20 Rechnungszeitraum auf Positionsebene
	// Wenn die Gruppe "INVOICE LINE PERIOD" (BG-26) verwendet wird, müssen entweder das Element "Invoice line period start date" (BT-134) oder
	// das Element "Invoice line period end date" (BT-135) oder beide gefüllt sein.
	for _, line := range inv.InvoiceLines {
		if !line.BillingSpecifiedPeriodStart.IsZero() || !line.BillingSpecifiedPeriodEnd.IsZero() {
			if line.BillingSpecifiedPeriodStart.IsZero() && line.BillingSpecifiedPeriodEnd.IsZero() {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-20", InvFields: []string{"BG-26", "BT-134", "BT-135"}, Text: fmt.Sprintf("Invoice line %s: if line period is used, either start date or end date must be filled", line.LineID)})
			}
		}
	}

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
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-CO-25", InvFields: []string{"BT-9", "BT-20", "BT-115"}, Text: "If amount due for payment is positive, either payment due date or payment terms must be present"})
		}
	}

}

func (inv *Invoice) checkBR() {
	// BR-1
	// Eine Rechnung (INVOICE) muss eine Spezifikationskennung "Specification identification“ (BT-24) enthalten.
	if inv.Profile == CProfileUnknown {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-1", InvFields: []string{"BT-24"}, Text: "Could not determine the profile in GuidelineSpecifiedDocumentContextParameter"})
	}
	// 	BR-2 Rechnung
	// Eine Rechnung (INVOICE) muss eine Rechnungsnummer "Invoice number“ (BT-1) enthalten.
	if inv.InvoiceNumber == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-2", InvFields: []string{"BT-1"}, Text: "No invoice number found"})
	}
	// BR-3 Rechnung
	// Eine Rechnung (INVOICE) muss ein Rechnungsdatum "Invoice issue date“ (BT-2) enthalten.
	if inv.InvoiceDate.IsZero() {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-3", InvFields: []string{"BT-2"}, Text: "Date is zero"})
	}
	// BR-4 Rechnung
	// Eine Rechnung (INVOICE) muss einen Rechnungstyp-Code "Invoice type code“ (BT-3) enthalten.
	if inv.InvoiceTypeCode == 0 {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-4", InvFields: []string{"BT-3"}, Text: "Invoice type code is 0"})
	}
	// BR-5 Rechnung
	// Eine Rechnung (INVOICE) muss einen Währungs-Code "Invoice currency code“ (BT-5) enthalten.
	if inv.InvoiceCurrencyCode == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-5", InvFields: []string{"BT-5"}, Text: "Invoice currency code is empty"})
	}
	// BR-6 Verkäufer
	// Eine Rechnung (INVOICE) muss den Verkäufernamen "Seller name“ (BT-27) enthalten.
	if inv.Seller.Name == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-6", InvFields: []string{"BT-27"}, Text: "Seller name is empty"})
	}
	// BR-7 Käufer
	// Eine Rechnung (INVOICE) muss den Erwerbernamen "Buyer name“ (BT-44) enthalten.
	if inv.Buyer.Name == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-7", InvFields: []string{"BT-44"}, Text: "Buyer name is empty"})
	}
	// BR-8 Verkäufer
	// Eine Rechnung (INVOICE) muss die postalische Anschrift des Verkäufers "SELLER POSTAL ADDRESS“ (BG-5) enthalten.
	if inv.Seller.PostalAddress == nil {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-8", InvFields: []string{"BG-5"}, Text: "Seller has no postal address"})
	} else {
		// BR-9 Verkäufer
		// Eine postalische Anschrift des Verkäufers "SELLER POSTAL ADDRESS“ (BG-5) muss einen Verkäufer-Ländercode "Seller country code“ (BT-40) enthalten.
		if inv.Seller.PostalAddress.CountryID == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-9", InvFields: []string{"BT-40"}, Text: "Seller country code empty"})
		}
	}
	if inv.Profile > CProfileMinimum {
		// BR-10 Käufer
		// Eine Rechnung (INVOICE) muss die postalische Anschrift des Erwerbers "BUYER POSTAL ADDRESS“ (BG-8) enthalten.
		if inv.Buyer.PostalAddress == nil {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-10", InvFields: []string{"BG-8"}, Text: "Buyer has no postal address"})
		} else {
			// BR-11 Käufer
			// Eine postalische Anschrift des Erwerbers "BUYER POSTAL ADDRESS“ (BG-8) muss einen Erwerber-Ländercode "Buyer country code“ (BT-55)
			// enthalten.
			if inv.Buyer.PostalAddress.CountryID == "" {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-11", InvFields: []string{"BT-55"}, Text: "Buyer country code empty"})
			}
		}
	}
	// BR-12 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss die Summe der Rechnungspositionen-Nettobeträge "Sum of Invoice line net amount“ (BT-106) enthalten.
	if inv.LineTotal.IsZero() {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-12", InvFields: []string{"BT-106"}, Text: "Line total is zero"})
	}
	// BR-13 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den Gesamtbetrag der Rechnung ohne Umsatzsteuer "Invoice total amount without VAT“ (BT-109) enthalten.
	if inv.TaxBasisTotal.IsZero() {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-13", InvFields: []string{"BT-109"}, Text: "TaxBasisTotal zero"})
	}
	// BR-14 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den Gesamtbetrag der Rechnung mit Umsatzsteuer "Invoice total amount with VAT“ (BT-112) enthalten.
	if inv.GrandTotal.IsZero() {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-14", InvFields: []string{"BT-112"}, Text: "GrandTotal is zero"})
	}
	// BR-15 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den ausstehenden Betrag "Amount due for payment“ (BT-115) enthalten.
	if inv.DuePayableAmount.IsZero() {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-15", InvFields: []string{"BT-115"}, Text: "DuePayableAmount is zero"})
	}
	// BR-16 Rechnung
	// Eine Rechnung (INVOICE) muss mindestens eine Rechnungsposition "INVOICE LINE“ (BG-25) enthalten.
	if is(CProfileBasic, inv) {
		if len(inv.InvoiceLines) == 0 {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-16", InvFields: []string{"BG-25"}, Text: "Invoice lines must be at least 1"})
		}
	}
	// BR-17 Zahlungsempfänger
	// Eine Rechnung (INVOICE) muss den Namen des Zahlungsempfängers "Payee name“ (BT-59) enthalten, wenn sich der Zahlungsempfänger "PAYEE“
	// (BG-10) vom Verkäufer "SELLER“ (BG-4) unterscheidet.
	if inv.PayeeTradeParty != nil {
		if inv.PayeeTradeParty.Name == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-17", InvFields: []string{"BT-59", "BG-10", "BG-4"}, Text: "Payee has no name, although different from seller"})
		}
	}
	// BR-18 Steuerbevollmächtigter des Verkäufers
	// Eine Rechnung (INVOICE) muss den Namen des Steuervertreters des Verkäufers "Seller tax representative name“ (BT-62) enthalten, wenn der
	// Verkäufer "SELLER“ (BG-4) einen Steuervertreter (BG-11) hat.
	if trp := inv.SellerTaxRepresentativeTradeParty; trp != nil {
		if trp.Name == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-18", InvFields: []string{"BT-62", "BG-4", "BG-11"}, Text: "Tax representative has no name, although seller has specified one"})
		}
		// BR-19 Steuerbevollmächtigter des Verkäufers
		// Eine Rechnung (INVOICE) muss die postalische Anschrift des Steuervertreters "SELLER TAX REPRESENTATIVE POSTAL ADDRESS“ (BG-12) enthalten,
		// wenn der Verkäufer "SELLER“ (BG-4) einen Steuervertreter hat.
		if trp.PostalAddress == nil {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-19", InvFields: []string{"BG-4", "BG-12"}, Text: "Tax representative has no postal address"})
		} else {
			// BR-20 Steuerbevollmächtigter des Verkäufers
			// Die postalische Anschrift des Steuervertreters des Verkäufers "SELLER TAX REPRESENTATIVE POSTAL ADDRESS“ (BG-12) muss einen
			// Steuervertreter-Ländercode enthalten, wenn der Verkäufer "SELLER“ (BG-4) einen Steuervertreter hat.
			if trp.PostalAddress.CountryID == "" {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-20", InvFields: []string{"BG-4", "BG-12"}, Text: "Tax representative has no postal address"})
			}
		}
	}
	for _, line := range inv.InvoiceLines {
		// BR-21 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss eine eindeutige Bezeichnung "Invoice line identifier“ (BT-126) haben.
		if line.LineID == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-21", InvFields: []string{"BG-25", "BT-126"}, Text: "Line has no line ID"})
		}
		// BR-22 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss die Menge der in der betreffenden Position in Rechnung gestellten Waren oder
		// Dienstleistungen als Einzelposten "Invoiced quantity“ (BT-129) enthalten.
		if line.BilledQuantity.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-22", InvFields: []string{"BG-25", "BT-129"}, Text: "Line has no billed quantity"})
		}
		// BR-23 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss eine Einheit zur Mengenangabe "Invoiced quantity unit of measure code“ (BT-130)
		// enthalten.
		if line.BilledQuantityUnit == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-23", InvFields: []string{"BG-25", "BT-130"}, Text: "Line's billed quantity has no unit"})
		}

		// BR-25 Artikelinformationen
		// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss den Namen des Postens "Item name" (BT-153) enthalten.
		if line.ItemName == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-25", InvFields: []string{"BG-25", "BT-153"}, Text: "Line's item name missing"})
		}

		// BR-26 Detailinformationen zum Preis
		// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss den Preis des Postens, ohne Umsatzsteuer, nach Abzug des für diese Rechnungsposition
		// geltenden Rabatts "Item net price" (BT-146) beinhalten.
		if line.NetPrice.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-26", InvFields: []string{"BG-25", "BT-146"}, Text: "Line's item net price not found"})
		}

		// BR-27 Nettopreis des Artikels
		// Der Artikel-Nettobetrag "Item net price" (BT-146) darf nicht negativ sein.
		if line.NetPrice.IsNegative() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-27", InvFields: []string{"BG-25", "BT-146"}, Text: "Net price must not be negative"})
		}
		// BR-28 Detailinformationen zum Preis
		// Der Einheitspreis ohne Umsatzsteuer vor Abzug des Postenpreisrabatts einer Rechnungsposition "Item gross price" (BT-148) darf nicht negativ
		// sein.
		if line.GrossPrice.IsNegative() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-28", InvFields: []string{"BG-25", "BT-148"}, Text: "Gross price must not be negative"})
		}
	}
	// BR-29 Rechnungszeitraum
	// Wenn Start- und Enddatum des Rechnungszeitraums gegeben sind, muss das Enddatum "Invoicing period end date“ (BT-74) nach dem Startdatum
	// "Invoicing period start date“ (BT-73) liegen oder mit diesem identisch sein.
	if inv.BillingSpecifiedPeriodEnd.Before(inv.BillingSpecifiedPeriodStart) {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-29", InvFields: []string{"BT-73", "BT-74"}, Text: "Billing period end must be after start"})
	}
	for _, line := range inv.InvoiceLines {
		// BR-30 Rechnungszeitraum auf Positionsebene
		// Wenn Start- und Enddatum des Rechnungspositionenzeitraums gegeben sind, muss das Enddatum "Invoice line period end date“ (BT-135) nach
		// dem Startdatum "Invoice line period start date“ (BT-134) liegen oder mit diesem identisch sein.
		if line.BillingSpecifiedPeriodEnd.Before(line.BillingSpecifiedPeriodStart) {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-30", InvFields: []string{"BG-25", "BT-135", "BT-134"}, Text: "Line item billing period end must be after or identical to start"})
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
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-36", InvFields: []string{"BG-21", "BT-99"}, Text: "Charge must not be zero"})
			}

			// BR-37 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Umsatzsteuer-Code "Document level charge VAT
			// category code" (BT-102) aufweisen.
			if ac.CategoryTradeTaxCategoryCode == "" {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-37", InvFields: []string{"BG-21", "BT-102"}, Text: "Charge tax category code not set"})
			}
			// BR-38 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Abgabegrund "Document level charge reason" (BT-104)
			// oder einen entsprechenden Code "Document level charge reason code" (BT-105) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-38", InvFields: []string{"BG-21", "BT-104", "BT-105"}, Text: "Charge reason empty or code unset"})
			}
			// BR-39 Zuschläge auf Dokumentenebene
			// Der Betrag einer Abgabe auf Dokumentenebene "Document level charge amount" (BT-99) darf nicht negativ sein.
			if ac.ActualAmount.LessThan(decimal.Zero) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-39", InvFields: []string{"BG-21", "BT-99"}, Text: "Document level charge amount must not be negative"})
			}
			// BR-40 Zuschläge auf Dokumentenebene
			// Der Basisbetrag einer Abgabe auf Dokumentenebene "Document level charge base amount" (BT-100) darf nicht negativ sein.
			if ac.BasisAmount.LessThan(decimal.Zero) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-40", InvFields: []string{"BG-21", "BT-100"}, Text: "Document level charge base amount must not be negative"})
			}
		} else {
			// BR-31 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Betrag "Document level allowance amount"
			// (BT-92) aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-31", InvFields: []string{"BG-20", "BT-92"}, Text: "Allowance must not be zero"})
			}
			// BR-32 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Umsatzsteuer-Code "Document level
			// allowance VAT category code" (BT-95) aufweisen.
			if ac.CategoryTradeTaxCategoryCode == "" {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-32", InvFields: []string{"BG-20", "BT-95"}, Text: "Allowance tax category code not set"})
			}
			// BR-33 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Nachlassgrund "Document level allowance
			// reason" (BT-97) oder einen entsprechenden Code "Document level allowance reason code" (BT-98") aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-33", InvFields: []string{"BG-20", "BT-95"}, Text: "Allowance reason empty or code unset"})
			}
			// BR-34 Abschläge auf Dokumentenebene
			// Der Betrag eines Nachlasses auf Dokumentenebene "Document level allowance amount" (BT-92) darf nicht negativ sein.
			if ac.ActualAmount.LessThan(decimal.Zero) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-34", InvFields: []string{"BG-20", "BT-92"}, Text: "Document level allowance amount must not be negative"})
			}
			// BR-35 Abschläge auf Dokumentenebene
			// Der Basisbetrag eines Nachlasses auf Dokumentenebene "Document level allowance base amount" (BT-93) darf nicht negativ sein.
			if ac.BasisAmount.LessThan(decimal.Zero) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-35", InvFields: []string{"BG-20", "BT-93"}, Text: "Document level allowance base amount must not be negative"})
			}
		}
	}

	for _, line := range inv.InvoiceLines {
		// BR-41 Abschläge auf Ebene der Rechnungsposition
		// Jeder Nachlass auf der Ebene der Rechnungsposition "INVOICE LINE ALLOWANCES“ (BG-27) muss einen Betrag "Invoice line allowance amount“
		// (BT-136) aufweisen.
		for _, ac := range line.InvoiceLineAllowances {
			if ac.ActualAmount.IsZero() {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-41", InvFields: []string{"BG-27", "BT-136"}, Text: "Line allowance amount zero"})
			}
			// BR-42 Abschläge auf Ebene der Rechnungsposition
			// Jeder Nachlass auf der Ebene der Rechnungsposition "INVOICE LINE ALLOWANCES“ (BG-27) muss einen Nachlassgrund "Invoice line allowance
			// reason“ (BT-139) oder einen entsprechenden Code "Invoice line allowance reason code“ (BT-140) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-42", InvFields: []string{"BG-27", "BT-139", "BT-140"}, Text: "Line allowance must have a reason"})
			}
		}
		for _, ac := range line.InvoiceLineCharges {
			// BR-43 Charge ou frais sur ligne de facture
			// Jede Abgabe auf der Ebene der Rechnungsposition "INVOICE LINE CHARGES“ (BG-28) muss einen Betrag "Invoice line charge amount“ (BT-141)
			// aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-43", InvFields: []string{"BG-28", "BT-141"}, Text: "Line charge amount zero"})
			}
			// BR-44 Charge ou frais sur ligne de facture
			// Jede Abgabe auf der Ebene der Rechnungsposition "INVOICE LINE CHARGES“ (BG-28) muss einen Abgabegrund "Invoice line charge reason“ (BT-
			// 144) oder einen entsprechenden Code "Invoice line charge reason code“ (BT-145) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-44", InvFields: []string{"BG-28", "BT-144", "BT-145"}, Text: "Line charge must have a reason"})
			}
		}
	}

	for _, tt := range inv.TradeTaxes {
		// BR-45 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN" (BG-23) muss die
		// Summe aller nach dem jeweiligen Schlüssel zu versteuernden Beträge
		// "VAT category taxable amount" (BT-116) aufweisen.
		key := tt.CategoryCode + "_" + tt.Percent.String()
		if !applicableTradeTaxes[key].Equal(tt.BasisAmount) {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-45", InvFields: []string{"BG-23", "BT-116"}, Text: "Applicable trade tax basis amount not equal to the sum of line total"})

		}
		// BR-47 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) muss über
		// eine codierte Bezeichnung einer Umsatzsteuerkategorie "VAT category
		// code“ (BT-118) definiert werden.
		if tt.CategoryCode == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-47", InvFields: []string{"BG-23", "BT-118"}, Text: "CategoryCode not set for applicable trade tax"})
		}
	}
	for _, pm := range inv.PaymentMeans {
		// BR-49 Zahlungsanweisungen
		// Die Zahlungsinstruktionen "PAYMENT INSTRUCTIONS“ (BG-16) müssen den Zahlungsart-Code "Payment means type code“ (BT-81) enthalten.
		if pm.TypeCode == 0 {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-49", InvFields: []string{"BT-81"}, Text: "Payment means type code must be set"})
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
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-52", InvFields: []string{"BG-24", "BT-122"}, Text: "Supporting document must have a reference"})
		}
	}

	// BR-53 Gesamtsummen auf Dokumentenebene
	// Wenn eine Währung für die Umsatzsteuerabrechnung angegeben wurde, muss
	// der Umsatzsteuergesamtbetrag in der Abrechnungswährung "Invoice total VAT
	// amount in accounting currency" (BT-111) angegeben werden.
	if inv.TaxCurrencyCode != "" && inv.TaxTotalVAT.IsZero() {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-53", InvFields: []string{"BT-6", "BT-111"}, Text: "Tax total in accounting currency must be specified when tax currency code is provided"})
	}

	// BR-54 Artikelattribute
	// Jede Eigenschaft eines in Rechnung gestellten Postens "ITEM ATTRIBUTES"
	// (BG-32) muss eine Bezeichnung "Item attribute name" (BT-160) und einen
	// Wert "Item attribute value" (BT-161) haben.
	for _, line := range inv.InvoiceLines {
		for _, char := range line.Characteristics {
			if char.Description == "" || char.Value == "" {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-54", InvFields: []string{"BG-32", "BT-160", "BT-161"}, Text: "Item attribute must have both name and value"})
			}
		}
	}

	// BR-55 Referenz auf die vorausgegangene Rechnung
	// Jede Bezugnahme auf eine vorausgegangene Rechnung "Preceding Invoice
	// reference" (BT-25) muss die Nummer der vorausgegangenen Rechnung
	// enthalten.
	for _, ref := range inv.InvoiceReferencedDocument {
		if ref.ID == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-55", InvFields: []string{"BG-3", "BT-25"}, Text: "Preceding invoice reference must contain invoice number"})
		}
	}

	// BR-56 Steuerbevollmächtigter des Verkäufers
	// Jeder Steuervertreter des Verkäufers "SELLER TAX REPRESENTATIVE PARTY"
	// (BG-11) muss eine Umsatzsteuer-Identifikationsnummer "Seller tax
	// representative VAT identifier" (BT-63) haben.
	if inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-56", InvFields: []string{"BG-11", "BT-63"}, Text: "Seller tax representative must have VAT identifier"})
	}

	// BR-57 Lieferanschrift
	// Jede Lieferadresse "DELIVER TO ADDRESS" (BG-15) muss einen entsprechenden
	// Ländercode "Deliver to country code" (BT-80) enthalten.
	if inv.ShipTo != nil && inv.ShipTo.PostalAddress != nil && inv.ShipTo.PostalAddress.CountryID == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-57", InvFields: []string{"BG-15", "BT-80"}, Text: "Deliver-to address must have country code"})
	}

	// BR-61 Zahlungsanweisungen
	// Wenn der Zahlungsmittel-Typ SEPA, lokale Überweisung oder
	// Nicht-SEPA-Überweisung ist, muss der "Payment account identifier" (BT-84)
	// des Zahlungsempfängers angegeben werden.
	for _, pm := range inv.PaymentMeans {
		// TypeCode 30 = Credit transfer, 58 = SEPA credit transfer
		if (pm.TypeCode == 30 || pm.TypeCode == 58) && pm.PayeePartyCreditorFinancialAccountIBAN == "" && pm.PayeePartyCreditorFinancialAccountProprietaryID == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-61", InvFields: []string{"BG-16", "BT-84"}, Text: "Payment account identifier required for credit transfer payment types"})
		}
	}

	// BR-62 Elektronische Adresse des Verkäufers
	// Im Element "Seller electronic address" (BT-34) muss die Komponente
	// "Scheme Identifier" vorhanden sein.
	if inv.Seller.URIUniversalCommunication != "" && inv.Seller.URIUniversalCommunicationScheme == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-62", InvFields: []string{"BT-34"}, Text: "Seller electronic address must have scheme identifier"})
	}

	// BR-63 Elektronische Adresse des Käufers
	// Im Element "Buyer electronic address" (BT-49) muss die Komponente "Scheme
	// Identifier" vorhanden sein.
	if inv.Buyer.URIUniversalCommunication != "" && inv.Buyer.URIUniversalCommunicationScheme == "" {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-63", InvFields: []string{"BT-49"}, Text: "Buyer electronic address must have scheme identifier"})
	}

	// BR-64 Kennung eines Artikels nach registriertem Schema
	// Im Element "Item standard identifier" (BT-157) muss die Komponente
	// "Scheme Identifier" vorhanden sein.
	for _, line := range inv.InvoiceLines {
		if line.GlobalID != "" && line.GlobalIDType == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-64", InvFields: []string{"BG-31", "BT-157"}, Text: "Item standard identifier must have scheme identifier"})
		}
	}

	// BR-65 Kennung der Artikelklassifizierung
	// Im Element "Item classification identifier" (BT-158) muss die Komponente
	// "Scheme Identifier" vorhanden sein.
	for _, line := range inv.InvoiceLines {
		for _, classification := range line.ProductClassification {
			if classification.ClassCode != "" && classification.ListID == "" {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-65", InvFields: []string{"BG-31", "BT-158"}, Text: "Item classification identifier must have scheme identifier"})
			}
		}
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
