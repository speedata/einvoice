package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// A SemanticError contains the rule number and a text that describes the error.
type SemanticError struct {
	Rule      string
	InvFields []string
	Text      string
}

// check checks the invoice against the rule set in EN 16931 and returns true if
// the invoice looks fine or false and a list of rule validations.
func (inv *Invoice) check() {
	// Documentation is taken from the ZUGFeRD specification and in German for now
	if inv.Violations == nil {
		inv.Violations = []SemanticError{}
	}
	inv.checkBR()
	inv.checkBRO()
	inv.checkOther()

	// BR-CO-3 Rechnung
	// Umsatzsteuerdatum "Value added tax point date“ (BT-7) und Code für das Umsatzsteuerdatum "Value added tax point date code“ (BT-8)
	// schließen sich gegenseitig aus.
	// BR-CO-4 Rechnungsposition
	// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss anhand der Umsatzsteuerkategorie des in Rechnung gestellten Postens "Invoiced item VAT
	// category code“ (BT-151) kategorisiert werden.
	// BR-CO-5 Abschläge auf Dokumentenebene
	// Der Code des Grundes für den Nachlass auf der Dokumentenebene "Document level allowance reason code“ (BT-98) und der Grund für den
	// Nachlass auf der Dokumentenebene "Document level allowance reason“ (BT-97) müssen dieselbe Nachlassart anzeigen.
	// BR-CO-6 Abschläge auf Dokumentenebene
	// Der Code des Grundes für die Abgaben auf der Dokumentenebene "Document level charge reason code“ (BT-105) und der Grund für die Abgaben
	// auf der Dokumentenebene "Document level charge reason“ (BT-104) müssen dieselbe Abgabenart anzeigen.
	// BR-CO-7 Abschläge auf Ebene der Rechnungsposition
	// Der Code für den Grund des Rechnungszeilennachlasses "Invoice line allowance reason code“ (BT-140) und der Grund für den
	// Rechnungszeilennachlass "Invoice line allowance reason“ (BT-139) müssen dieselbe Nachlassart anzeigen.
	// BR-CO-8 Zuschläge auf Ebene der Rechnungsposition
	// Der Code für den Grund der Abgabe auf Rechnungspositionsebene "Invoice line charge reason code“ (BT-145) und der Grund für die Abgabe auf
	// Rechnungspositionsebene "Invoice line charge reason“ (BT-144) müssen dieselbe Abgabenart anzeigen.
	// BR-CO-9 Umsatzsteuer-Identifikationsnummern
	// Der Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31), der Umsatzsteuer-Identifikationsnummer des
	// Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) und der Umsatzsteuer-Identifikationsnummer des Erwerbers
	// "Buyer VAT identifier“ (BT-48) muss zur Kennzeichnung des Mitgliedstaats, der sie erteilt hat, jeweils ein Präfix nach dem ISO-Code 3166 Alpha-2
	// vorangestellt werden. Griechenland wird jedoch ermächtigt, das Präfix "EL“ zu verwenden.

	// BR-CO-11 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Sum of allowances on document level“ (BT-107) entspricht der Summe aller Inhalte der Elemente "Document level
	// allowance amount“ (BT-92).
	// BR-CO-12 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Sum of charges on document level“ (BT-108) entspricht der Summe aller Inhalte der Elemente "Document level charge
	// amount“ (BT-99).
	// BR-CO-13 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount without VAT“ (BT-109) entspricht der Summe aller Inhalte der Elemente "Invoice line net amount“
	// (BT-131) abzüglich der Summe aller in der Rechnung enthaltenen Nachlässe der Dokumentenebene "Sum of allowances on document level“ (BT-
	// 107) zuzüglich der Summe aller in der Rechnung enthaltenen Abgaben der Dokumentenebene "Sum of charges on document level“ (BT-108).
	// BR-CO-14 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total VAT amount“ (BT-110) entspricht der Summe aller Inhalte der Elemente "VAT category tax amount“ (BT-
	// 117).
	// BR-CO-15 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount with VAT“ (BT-112) entspricht der Summe des Inhalts des Elementes "Invoice total amount
	// without VAT“ (BT-109) und des Elementes "Invoice total VAT amount“ (BT-110).
	// BR-CO-16 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Amount due for payment“ (BT-115) entspricht dem Inhalt des Elementes "Invoice total VAT amount“ (BT-110) abzüglich
	// dem Inhalt des Elementes "Paid amount“ (BT-113) zuzüglich dem Inhalt des Elementes "Rounding amount“ (BT-114).
	// BR-CO-17 Umsatzsteueraufschlüsselung
	// Der Inhalt des Elementes "VAT category tax amount“ (BT-117) entspricht dem Inhalt des Elementes "VAT category taxable amount“ (BT-116),
	// multipliziert mit dem Inhalt des Elementes "VAT category rate“ (BT-119) geteilt durch 100, gerundet auf zwei Dezimalstellen.
	// BR-CO-18 Umsatzsteueraufschlüsselung
	// Eine Rechnung (INVOICE) soll mindestens eine Gruppe "VAT BREAKDOWN“ (BG-23) enthalten.
	// BR-CO-19 Liefer- oder Rechnungszeitraum
	// Wenn die Gruppe "INVOICING PERIOD“ (BG-14) verwendet wird, müssen entweder das Element "Invoicing period start date“ (BT-73) oder das
	// Element "Invoicing period end date“ (BT-74) oder beide gefüllt sein.
	// BR-CO-20 Rechnungszeitraum auf Positionsebene
	// Wenn die Gruppe "INVOICE LINE PERIOD“ (BG-26) verwendet wird, müssen entweder das Element "Invoice line period start date“ (BT-134) oder
	// das Element "Invoice line period end date“ (BT-135) oder beide gefüllt sein.
	// BR-CO-21 Abschläge auf Dokumentenebene
	// Jede Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) muss entweder ein Element "Document level allowance reason“ (BT-97) oder ein
	// Element "Document level allowance reason code“ (BT-98) oder beides enthalten.
	// BR-CO-22 Zuschläge auf Dokumentenebene
	// Jede Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) muss entweder ein Element "Document level charge reason“ (BT-104) oder ein Element
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 67 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024Liste der Geschäftsregeln
	// Nr. Kontext
	// "Document level charge reason code“ (BT-105) oder beides enthalten.
	// BR-CO-23 Abschläge auf Ebene der Rechnungsposition
	// Jede Gruppe "INVOICE LINE ALLOWANCES“ (BG-27) muss entweder ein Element "Invoice line allowance reason“ (BT-139) oder ein Element
	// "Invoice line allowance reason code“ (BT-140) oder beides enthalten.
	// BR-CO-24 Zuschläge auf Ebene der Rechnungsposition
	// Jede Gruppe "INVOICE LINE CHARGES“ (BG-28) muss entweder ein Element "Invoice line charge reason“ (BT-144) oder ein Element "Invoice line
	// charge reason code“ (BT-145) oder beides enthalten.
	// BR-CO-25 Rechnung
	// Im Falle eines positiven Zahlbetrags "Amount due for payment“ (BT-115) muss entweder das Element Fälligkeitsdatum "Payment due date“ (BT-9)
	// oder das Element Zahlungsbedingungen "Payment terms“ (BT-20) vorhanden sein.
	// BR-CO-26 Verkäufer
	// Damit der Erwerber den Lieferanten automatisch identifizieren kann, soll der "Seller identifier“ (BT-29), der "Seller legal registration identifier“
	// (BT-30) oder der "Seller VAT identifier“ (BT-31) vorhanden sein.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 68 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-AE-1 Umkehrung der Steuerschuldnerschaft
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf der Rechnungsebene enthält, bei der als Code der
	// Umsatzsteuerkategorie des in Rechnung gestellten Postens ("Invoiced item VAT category code“ (BT-151), "Document level allowance VAT
	// category code“ (BT-95) oder "Document level charge VAT category code“ (BT-102)) der Wert "VAT reverse charge“ angegeben ist, muss in "VAT
	// BREAKDOWN“ (BG-23) genau einen Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "VAT Reverse charge“
	// enthalten.
	// BR-AE-2 Umkehrung der Steuerschuldnerschaft
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten "Invoiced
	// item VAT category code“ (BT-151) der Wert "Exempt from VAT“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers
	// "Seller VAT identifier“ (BT-31), die Steueridentifikationsnummer des Verkäufers "Seller tax registration identifier“ (BT-32) oder die Umsatzsteuer-
	// Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) und die Umsatzsteuer-
	// Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) oder den "Buyer tax registration identifier“a enthalten.
	// BR-AE-3 Umkehrung der Steuerschuldnerschaft
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "Exempt from VAT“ hat, müssen entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller
	// tax representative VAT identifier“ (BT-63) sowie die Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) oder
	// "Buyer tax registration identifier“ vorhanden sein.
	// BR-AE-4 Umkehrung der Steuerschuldnerschaft
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "Exempt from VAT“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax
	// representative VAT identifier“ (BT-63) und die Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) vorhanden sein.
	// BR-AE-5 Umkehrung der Steuerschuldnerschaft
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "Reverse charge“ hat, muss "Invoiced item VAT
	// rate“ (BT-152) gleich "0“ sein.
	// BR-AE-6 Umkehrung der Steuerschuldnerschaft
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "Reverse charge“
	// hat, muss "Document level allowance VAT rate“ (BT-96) gleich "0“ sein.
	// BR-AE-7 Umkehrung der Steuerschuldnerschaft
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "Reverse charge“ hat, muss
	// "Document level charge VAT rate“ (BT-103) gleich "0“ sein.
	// BR-AE-8 Umkehrung der Steuerschuldnerschaft
	// In einer "VAT BREAKDOWN“ (BG-23), in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der Wert "Reverse charge“
	// angegeben ist, muss der "VAT category taxable amount“ (BT-116) gleich der Summe der "Invoice line net amount“ (BT-131) abzüglich der
	// "Document level allowance amount“ (BT-92) zuzüglich der "Document level charge amount“ (BT-99) sein, wobei als "Invoiced item VAT category
	// code“ (BT-151), als "Document level allowance VAT category code“ (BT-95) sowie als "Document level charge VAT category code“ (BT-102) jeweils
	// der Wert "Reverse charge“ angegeben wird.
	// BR-AE-9 Umkehrung der Steuerschuldnerschaft
	// Der "VAT category tax amount“ (BT-117) muss in einer "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category
	// code“ (BT-118) mit dem Wert "Reverse charge“ gleich "0“ sein.
	// BR-AE-10 Umkehrung der Steuerschuldnerschaft
	// Ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Reverse charge“ muss
	// einen "VAT exemption reason code“ (BT-121) mit dem Wert "Reverse charge“ oder einen "VAT exemption reason text“ (BT-120) des Wertes
	// "Reverse charge“ (oder das Äquivalent in einer anderen Sprache) enthalten.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 69 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-E-1 Steuerbefreit
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf der Rechnungsebene enthält, bei der als Code der
	// Umsatzsteuerkategorie des in Rechnung gestellten Postens ("Invoiced item VAT category code“ (BT-151), "Document level allowance VAT
	// category code“ (BT-95) oder "Document level charge VAT category code“ (BT-102)) der Wert "Exempt from VAT“ angegeben ist, muss genau ein
	// "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Exempt from VAT“
	// enthalten.
	// BR-E-2 Steuerbefreit
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten "Invoiced
	// item VAT category code“ (BT-151) der Wert "Exempt from VAT“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers
	// "Seller VAT identifier“ (BT-31), die Steueridentifikationsnummer des Verkäufers "Seller tax registration identifier“ (BT-32) oder die Umsatzsteuer-
	// Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) enthalten.
	// BR-E-3 Steuerbefreit
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "Exempt from VAT“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax
	// representative VAT identifier“ (BT-63) vorhanden sein.
	// BR-E-4 Steuerbefreit
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "Exempt from VAT“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax
	// representative VAT identifier“ (BT-63) vorhanden sein.
	// BR-E-5 Steuerbefreit
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "Exempt from VAT“ hat, muss "Invoiced item VAT
	// rate“ (BT-152) gleich "0“ sein.
	// BR-E-6 Steuerbefreit
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "Exempt from VAT“
	// hat, muss "Document level allowance VAT rate“ (BT-96) gleich "0“ sein.
	// BR-E-7 Steuerbefreit
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "Exempt from VAT“ hat,
	// muss "Document level charge VAT rate“ (BT-103) gleich "0“ sein.
	// BR-E-8 Steuerbefreit
	// In einer "VAT BREAKDOWN“ (BG-23), in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der Wert "Exempt from VAT“
	// angegeben ist, muss der "VAT category taxable amount“ (BT-116) gleich der Summe der "Invoice line net amount“ (BT-131) abzüglich der
	// "Document level allowance amount“ (BT-92) zuzüglich der "Document level charge amount“ (BT-99) sein, wobei als "Invoiced item VAT category
	// code“ (BT-151), als "Document level allowance VAT category code“ (BT-95) sowie als "Document level charge VAT category code“ (BT-102) jeweils
	// der Wert "Exempt from VAT“ angegeben wird.
	// BR-E-9 Steuerbefreit
	// Der "VAT category tax amount“ (BT-117) muss in einer "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category
	// code“ (BT-118) dem Wert "Exempt from VAT“ gleich "0“ sein.
	// BR-E-10 Steuerbefreit
	// Ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Exempt from VAT“ muss
	// einen Code des Umsatzsteuerbefreiungsgrundes "VAT exemption reason code“ (BT-121) oder einen Text des Umsatzsteuerbefreiungsgrundes
	// "VAT exemption reason text“ (BT-120) enthalten.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 70 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-G-1 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf der Rechnungsebene enthält, bei der als Code der
	// Umsatzsteuerkategorie des in Rechnung gestellten Postens ("Invoiced item VAT category code“ (BT-151), "Document level allowance VAT
	// category code“ (BT-95) oder "Document level charge VAT category code“ (BT-102)) der Wert "Export outside the EU“ angegeben ist, muss in "VAT
	// BREAKDOWN“ (BG-23) genau einen Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Export outside the EU“
	// enthalten.
	// BR-G-2 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten "Invoiced
	// item VAT category code“ (BT-151) der Wert "Export outside the EU“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers
	// "Seller VAT identifier“ (BT-31) oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT
	// identifier“ (BT-63) enthalten.
	// BR-G-3 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "Export outside the EU“ hat, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31) oder
	// die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) enthalten sein.
	// BR-G-4 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "Export outside the EU“ hat, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31) oder die
	// Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) enthalten sein.
	// BR-G-5 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "Export outside the EU“ hat, muss "Invoiced item
	// VAT rate“ (BT-152) gleich "0“ sein.
	// BR-G-6 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "Export outside the
	// EU“ hat, muss "Document level allowance VAT rate“ (BT-96) gleich "0“ sein.
	// BR-G-7 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "Export outside the EU“
	// hat, muss "Document level charge VAT rate“ (BT-103) gleich "0“ sein.
	// BR-G-8 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// In einer "VAT BREAKDOWN“ (BG-23), in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der Wert "Export outside the EU“
	// angegeben ist, muss der "VAT category taxable amount“ (BT-116) gleich der Summe der "Invoice line net amount“ (BT-131) abzüglich der
	// "Document level allowance amount“ (BT-92) zuzüglich der "Document level charge amount“ (BT-99) sein, wobei als "Invoiced item VAT category
	// code“ (BT-151), als "Document level allowance VAT category code“ (BT-95) sowie als "Document level charge VAT category code“ (BT-102) jeweils
	// der Wert "Export outside the EU“ angegeben wird.
	// BR-G-9 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// Der "VAT category tax amount“ (BT-117) muss in "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“
	// (BT-118) mit dem Wert "Export outside the EU“ gleich "0“ sein.
	// BR-G-10 Steuer nicht erhoben aufgrund von Export außerhalb der EU
	// Ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "intra-community supply“
	// muss einen "VAT exemption reason code“ (BT-121) mit dem Wert "Export outside the EU“ oder einen "VAT exemption reason text“ (BT-120) des
	// Wertes "Export outside the EU“ (oder das Äquivalent in einer anderen Sprache) enthalten.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 71 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-IC-1 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf der Rechnungsebene enthält, bei der als Code der
	// Umsatzsteuerkategorie des in Rechnung gestellten Postens ("Invoiced item VAT category code“ (BT-151), "Document level allowance VAT
	// category code“ (BT-95) oder "Document level charge VAT category code“ (BT-102)) der Wert "intra-community supply“ angegeben ist, muss in
	// "VAT BREAKDOWN“ (BG-23) genau einen Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "intra-community supply“
	// enthalten.
	// BR-IC-2 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten "Invoiced
	// item VAT category code“ (BT-151) der Wert "intra-community supply“ angegeben ist, müssen die Umsatzsteuer-Identifikationsnummer des
	// Verkäufers "Seller VAT identifier“ (BT-31) oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax
	// representative VAT identifier“ (BT-63) sowie die Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) enthalten.
	// BR-IC-3 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "intra-community supply“ hat, müssen die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31)
	// oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) sowie die
	// Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) enthalten sein.
	// BR-IC-4 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "intra-community supply“ hat, müssen die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31) oder die
	// Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) sowie die
	// Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) enthalten sein.
	// BR-IC-5 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "intracommunity supply“ hat, muss "Invoiced item
	// VAT rate“ (BT-152) gleich "0“ sein.
	// BR-IC-6 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "intra-community
	// supply“ hat, muss "Document level allowance VAT rate“ (BT-96) gleich "0“ sein.
	// BR-IC-7 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "intra-community supply“
	// hat, muss "Document level charge VAT rate“ (BT-103) gleich "0“ sein.
	// BR-IC-8 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer "VAT BREAKDOWN“ (BG-23), in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der Wert "intra-community
	// supply“ angegeben ist, muss der "VAT category taxable amount“ (BT-116) gleich der Summe der "Invoice line net amount“ (BT-131) abzüglich der
	// "Document level allowance amount“ (BT-92) zuzüglich der "Document level charge amount“ (BT-99) sein, wobei als "Invoiced item VAT category
	// code“ (BT-151), als "Document level allowance VAT category code“ (BT-95) sowie als "Document level charge VAT category code“ (BT-102) jeweils
	// der Wert "intra-community supply“ angegeben wird.
	// BR-IC-9 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// Der "VAT category tax amount“ (BT-117) muss in einer "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category
	// code“ (BT-118) mit dem Wert "intra-community supply“ gleich "0“ sein.
	// BR-IC-10 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// Ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "intra-community supply“
	// muss einen "VAT exemption reason code“ (BT-121) mit dem Wert "intra-community supply“ oder einen "VAT exemption reason text“ (BT-120)
	// mit dem Wert "intracommunity supply“ (oder das Äquivalent in einer anderen Sprache) enthalten.
	// BR-IC-11 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer Rechnung, die ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert
	// "intra-community supply“ enthält, dürfen "Actual delivery date“ (BT-72) oder "INVOICING PERIOD“ (BG-14) nicht leer sein.
	// BR-IC-12 Kein Ausweis der Umsatzsteuer bei innergemeinschaftlichen
	// Lieferungen
	// In einer Rechnung, die ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert
	// "intra-community supply“ enthält, darf "Deliver to country code“ (BT-80) nicht leer sein.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 72 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-IG-1 IGIC (Kanarische Inseln)
	// Eine Rechnung (INVOICE), die eine Position, eine Abgabe auf der Dokumentenebene oder einen Nachlass auf der Rechnungsebene enthält, bei der
	// als Code der Umsatzsteuerkategorie des in Rechnung gestellten Postens der Wert "IGIC“ angegeben ist, muss in der Umsatzsteueraufschlüsselung
	// "VAT BREAKDOWN“ (BG-23) mindestens einen Code der Umsatzsteuerkategorie mit dem Wert "IGIC“ enthalten.
	// BR-IG-2 IGIC (Kanarische Inseln)
	// Eine Rechnung (INVOICE), die eine Position enthält, bei der als Code der Umsatzsteuerkategorie des in Rechnung gestellten Postens der Wert
	// "IGIC“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31), den Bezeichner der
	// Steuerangaben des Verkäufers oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT
	// identifier“ (BT-63) enthalten.
	// BR-IG-3 IGIC (Kanarische Inseln)
	// Eine Rechnung (INVOICE), die eine Abgabe auf der Dokumentenebene enthält, bei dem als Code der Umsatzsteuerkategorie des in Rechnung
	// gestellten Postens der Wert "IGIC“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31),
	// den Bezeichner der Steuerangaben des Verkäufers oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax
	// representative VAT identifier“ (BT-63) enthalten.
	// BR-IG-4 IGIC (Kanarische Inseln)
	// Eine Rechnung (INVOICE), die einen Nachlass auf der Dokumentenebene enthält, bei dem als Code der Umsatzsteuerkategorie des in Rechnung
	// gestellten Postens der Wert "IGIC“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31),
	// den Bezeichner der Steuerangaben des Verkäufers oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax
	// representative VAT identifier“ (BT-63) enthalten.
	// BR-IG-5 IGIC (Kanarische Inseln)
	// In einer Rechnungsposition, in der als Code der Umsatzsteuerkategorie für den Rechnungsposten der Wert "IGIC“ angegeben ist, muss der
	// Umsatzsteuersatz für den in Rechnung gestellten Posten "0“ oder größer "0“ sein.
	// BR-IG-6 IGIC (Kanarische Inseln)
	// In einer Abgabe auf der Dokumentenebene, in dem als Code der Umsatzsteuerkategorie für den Rechnungsposten der Wert "IGIC“ angegeben ist,
	// muss der Umsatzsteuersatz für den in Rechnung gestellten Posten "0“ oder größer "0“ sein.
	// BR-IG-7 IGIC (Kanarische Inseln)
	// In einem Nachlass auf der Dokumentenebene, in dem als Code der Umsatzsteuerkategorie für den Rechnungsposten der Wert "IGIC“ angegeben
	// ist, muss der Umsatzsteuersatz für den in Rechnung gestellten Posten "0“ oder größer "0“ sein.
	// BR-IG-8 IGIC (Kanarische Inseln)
	// Für jeden anderen Wert des kategoriespezifischen Umsatzsteuersatzes, bei dem als Code der Umsatzsteuerkategorie der Wert "IGIC“ angegeben
	// ist, muss der nach der Umsatzsteuerkategorie zu versteuernde Betrag in einer Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) gleich
	// der Summe der Rechnungspositions-Nettobeträge zuzüglich der Summe der Beträge aller Abschläge auf der Dokumentenebene abzüglich der
	// Summe der Beträge aller Zuschläge auf der Dokumentenebene sein; wobei als Code der Umsatzsteuerkategorie der Wert "IGIC“ angegeben wird
	// und der Umsatzsteuersatz gleich dem kategoriespezifischen Umsatzsteuersatz ist.
	// BR-IG-9 IGIC (Kanarische Inseln)
	// Der in der Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) angegebene Betrag der nach Kategorie zu entrichtenden Umsatzsteuer, bei
	// dem als Code der Umsatzsteuerkategorie der Wert "IGIC“ angegeben ist, muss gleich dem mit dem kategoriespezifischen Umsatzsteuersatz
	// multiplizierten nach der Umsatzsteuerkategorie zu versteuernden Betrag sein.
	// BR-IG-10 IGIC (Kanarische Inseln)
	// Eine Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie der Wert "IGIC“ darf keinen Code für
	// den Umsatzsteuerbefreiungsgrund oder Text für den Umsatzsteuerbefreiungsgrund
	// haben.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 73 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-IP-1 IPSI (Ceuta/Melilla)
	// Eine Rechnung (INVOICE), die eine Position, eine Abgabe auf der Dokumentenebene oder einen Nachlass auf der Dokumentenebene enthält, bei
	// der als Code der Umsatzsteuerkategorie des in Rechnung gestellten Postens der Wert "IPSI“ angegeben ist, muss in der
	// Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) mindestens einen Code der Umsatzsteuerkategorie gleich "IPSI“ enthalten.
	// BR-IP-2 IPSI (Ceuta/Melilla)
	// Eine Rechnung (INVOICE), die eine Position enthält, bei der als Code der Umsatzsteuerkategorie des in Rechnung gestellten Postens der Wert
	// "IPSI“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31), den Bezeichner der
	// Steuerangaben des Verkäufers oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT
	// identifier“ (BT-63) enthalten.
	// BR-IP-3 IPSI (Ceuta/Melilla)
	// Eine Rechnung (INVOICE), die eine Abgabe auf der Dokumentenebene enthält, bei dem als Code der Umsatzsteuerkategorie des in Rechnung
	// gestellten Postens der Wert "IPSI“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31),
	// den Bezeichner der Steuerangaben des Verkäufers oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax
	// representative VAT identifier“ (BT-63) enthalten.
	// BR-IP-4 IPSI (Ceuta/Melilla)
	// Eine Rechnung (INVOICE), die einen Nachlass auf der Dokumentenebene enthält, bei dem als Code der Umsatzsteuerkategorie des in Rechnung
	// gestellten Postens der Wert "IPSI“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31),
	// den Bezeichner der Steuerangaben des Verkäufers oder die Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax
	// representative VAT identifier“ (BT-63) enthalten.
	// BR-IP-5 IPSI (Ceuta/Melilla)
	// In einer Rechnungsposition, in der als Code der Umsatzsteuerkategorie für den Rechnungsposten der Wert "IPSI“ angegeben ist, muss der
	// Umsatzsteuersatz für den in Rechnung gestellten Posten "0“ oder größer "0“ sein.
	// BR-IP-6 IPSI (Ceuta/Melilla)
	// In einem Abgabe auf der Dokumentenebene, in dem als Code der Umsatzsteuerkategorie für den Rechnungsposten der Wert "IPSI“ angegeben ist,
	// muss der Umsatzsteuersatz für den in Rechnung gestellten Posten "0“ oder größer "0“ sein.
	// BR-IP-7 IPSI (Ceuta/Melilla)
	// In einem Nachlass auf der Dokumentenebene, in dem als Code der Umsatzsteuerkategorie für den Rechnungsposten der Wert "IPSI“ angegeben
	// ist, muss der Umsatzsteuersatz für den in Rechnung gestellten Posten "0“ oder größer "0“ sein.
	// BR-IP-8 IPSI (Ceuta/Melilla)
	// Für jeden anderen Wert des kategoriespezifischen Umsatzsteuersatzes, bei dem als Code der Umsatzsteuerkategorie der Wert "IPSI“ angegeben
	// ist, muss der nach der Umsatzsteuerkategorie zu versteuernde Betrag in einer Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) gleich
	// der Summe der Rechnungspositions-Nettobeträge zuzüglich der Summe der Beträge aller Abschläge auf der Dokumentenebene abzüglich der
	// Summe der Beträge aller Zuschläge auf der Dokumentenebene sein; wobei als Code der Umsatzsteuerkategorie der Wert "IPSI“ angegeben wird
	// und der Umsatzsteuersatz gleich dem kategoriespezifischen Umsatzsteuersatz ist.
	// BR-IP-9 IPSI (Ceuta/Melilla)
	// Der in der Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) angegebene Betrag der nach Kategorie zu entrichtenden Umsatzsteuer, bei
	// dem als Code der Umsatzsteuerkategorie der Wert "IPSI“ angegeben ist, muss gleich dem mit dem kategoriespezifischen Umsatzsteuersatz
	// multiplizierten nach der Umsatzsteuerkategorie zu versteuernden Betrag sein.
	// BR-IP-10 IPSI (Ceuta/Melilla)
	// Eine Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie mit dem Wert "IPSI“ darf keinen Code
	// für den Umsatzsteuerbefreiungsgrund oder Text für den Umsatzsteuerbefreiungsgrund haben.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 74 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-O-1 Nicht steuerbar
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf der Rechnungsebene enthält, bei der als Code der
	// Umsatzsteuerkategorie des in Rechnung gestellten Postens ("Invoiced item VAT category code“ (BT-151), "Document level allowance VAT
	// category code“ (BT-95) oder "Document level charge VAT category code“ (BT-102)) der Wert "Not subject to VAT“ angegeben ist, muss genau eine
	// "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Not subject to VAT“
	// enthalten.
	// BR-O-2 Nicht steuerbar
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten der Wert
	// "Not subject to VAT“ angegeben ist, darf keine Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31),
	// Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) und die
	// Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) enthalten.
	// BR-O-3 Nicht steuerbar
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "Not subject to VAT“ hat, dürfen die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31), die
	// Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) sowie die
	// Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) nicht enthalten sein.
	// BR-O-4 Nicht steuerbar
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "Not subject to VAT“ hat, dürfen die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT identifier“ (BT-31), die
	// Umsatzsteuer-Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) sowie die
	// Umsatzsteuer-Identifikationsnummer des Erwerbers "Buyer VAT identifier“ (BT-48) nicht enthalten sein.
	// BR-O-5 Nicht steuerbar
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "Not subject to VAT“ hat, darf "Invoiced item VAT
	// rate“ (BT-152) nicht enthalten sein.
	// BR-O-6 Nicht steuerbar
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "Not subject to VAT“
	// hat, darf "Document level allowance VAT rate“ (BT-96) nicht enthalten sein.
	// BR-O-7 Nicht steuerbar
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "Not subject to VAT“ hat,
	// darf "Document level charge VAT rate“ (BT-103) nicht enthalten sein.
	// BR-O-8 Nicht steuerbar
	// In einer "VAT BREAKDOWN“ (BG-23), in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der Wert "Not subject to VAT“
	// angegeben ist, muss der "VAT category taxable amount“ (BT-116) gleich der Summe der "Invoice line net amount“ (BT-131) abzüglich der
	// "Document level allowance amount“ (BT-92) zuzüglich der "Document level charge amount“ (BT-99) sein, wobei als "Invoiced item VAT category
	// code“ (BT-151), als "Document level allowance VAT category code“ (BT-95) sowie als "Document level charge VAT category code“ (BT-102) jeweils
	// der Wert "Not subject to VAT“ angegeben wird.
	// BR-O-9 Nicht steuerbar
	// Der "VAT category tax amount“ (BT-117) muss in "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“
	// (BT-118) mit dem Wert "Not subject to VAT“ gleich "0“ sein.
	// BR-O-10 Nicht steuerbar
	// Ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Not subject to VAT“
	// muss einen "VAT exemption reason code“ (BT-121) mit dem Wert "Not subject to VAT“ oder einen "VAT exemption reason text“ (BT-120) des
	// Wertes "Not subject to VAT“ (oder das Äquivalent in einer anderen Sprache) enthalten.
	// BR-O-11 Nicht steuerbar
	// Eine Rechnung (INVOICE), die ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) den Wert
	// "Not subject to VAT“ enthält, darf keine weiteren "VAT BREAKDOWN“ (BG-23) enthalten.
	// BR-O-12 Nicht steuerbar
	// Eine Rechnung (INVOICE), die ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) den Wert
	// "Not subject to VAT“ enthält, darf keine Positionen enthalten, deren "Invoiced item VAT category code“ (BT-151) einen anderen Wert als "Not
	// subject to VAT“ hat.
	// BR-O-13 Nicht steuerbar
	// Eine Rechnung (INVOICE), die ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) den Wert
	// "Not subject to VAT“ enthält, darf keine "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthalten, deren "Document level allowance VAT category
	// code“ (BT-95) einen anderen Wert als "Not subject to VAT“ hat.
	// BR-O-14 Nicht steuerbar
	// Eine Rechnung (INVOICE), die ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem
	// Wert "Not subject to VAT“ enthält, darf keine "DOCUMENT LEVEL CHARGES“ (BG-21) enthalten, deren "Document level charge VAT category
	// code“ (BT-102) einen anderen Wert als "Not subject to VAT“ hat.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 75 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-S-1 Umsatzsteuer mit Normalsatz
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf Rechnungsebene enthält, in der als Code der für den in
	// Rechnung gestellten Posten geltenden Umsatzsteuerkategorie "Invoiced item VAT category code“ (BT-151), als Code für das
	// Umsatzsteuermerkmal, das auf den Nachlass auf Dokumentenebene anzuwenden ist "Document level allowance VAT category code“ (BT-95) oder
	// als Code für das Umsatzsteuermerkmal dieser Elementgruppe "Document level charge VAT category code“ (BT-102) der Wert "Standard rated“
	// angegeben ist, muss in der Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) mindestens einen Code der Umsatzsteuerkategorie "VAT
	// category code“ (BT-118) gleich dem Wert "Standard rated“ enthalten.
	// BR-S-2 Umsatzsteuer mit Normalsatz
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten "Invoiced
	// item VAT category code“ (BT-151) der Wert "Standard rated“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller
	// VAT identifier“ (BT-31), die Steueridentifikationsnummer des Verkäufers "Seller tax registration identifier“ (BT-32) oder die Umsatzsteuer-
	// Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) enthalten.
	// BR-S-3 Umsatzsteuer mit Normalsatz
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "Standard rated“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax
	// representative VAT identifier“ (BT-63) vorhanden sein.
	// BR-S-4 Umsatzsteuer mit Normalsatz
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "Standard rated“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax
	// representative VAT identifier“ (BT-63) vorhanden sein.
	// BR-S-5 Umsatzsteuer mit Normalsatz
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "Standard rated“ hat, muss "Invoiced item VAT
	// rate“ (BT-152) größer als "0“ sein.
	// BR-S-6 Umsatzsteuer mit Normalsatz
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "Standard rated“
	// hat, muss "Document level allowance VAT rate“ (BT-96) größer als "0“ sein.
	// BR-S-7 Umsatzsteuer mit Normalsatz
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "Standard rated“ hat, muss
	// "Document level charge VAT rate“ (BT-103) größer als "0“ sein.
	// BR-S-8 Umsatzsteuer mit Normalsatz
	// Für jeden anderen Wert des kategoriespezifischen Umsatzsteuersatzes "VAT category rate“ (BT-119), bei dem als Code der Umsatzsteuerkategorie
	// "VAT category code“ (BT-118) der Wert "Standard rated“ angegeben ist, muss der nach der Umsatzsteuerkategorie zu versteuernde Betrag "VAT
	// category taxable amount“ (BT-116) in einer Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) gleich der Summe der Rechnungszeilen-
	// Nettobeträge "Invoice line net amount“ (BT-131) zuzüglich der Summe der Beträge aller Abgaben auf der Dokumentenebene "Document level
	// charge amount“ (BT-99) abzüglich der Summe der Beträge aller Nachlässe auf der Dokumentenebene "Document level allowance amount“ (BT-
	// 92) sein; wobei als Code der Umsatzsteuerkategorie ("Invoiced item VAT category code“ (BT-151), "Document level charge VAT category code“
	// (BT-102) und "Document level allowance VAT category code“ (BT-95)) der Wert "Standard rated“ angegeben wird und der Umsatzsteuersatz
	// ("Invoiced item VAT rate“ (BT-152), "Document level charge VAT rate“ (BT-103) und "Document level allowance VAT rate“ (BT-96)) gleich dem
	// kategoriespezifischen Umsatzsteuersatz ist. Anmerkung: D.h. dass für jeden USt-Satz eine gesonderte Gruppe "VAT BREAKDOWN“ (BG-23)
	// anzulegen ist.
	// BR-S-9 Umsatzsteuer mit Normalsatz
	// Der in der Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) angegebene Betrag der nach Kategorie zu entrichtenden Umsatzsteuer, bei
	// dem als Umsatzsteuerkategorie der Wert "Standard rated“ angegeben ist, muss gleich dem mit dem kategoriespezifischen Umsatzsteuersatz
	// multiplizierten nach der Umsatzsteuerkategorie zu versteuernden Betrag sein.
	// BR-S-10 Umsatzsteuer mit Normalsatz
	// Eine Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der
	// Wert "Standard rated“ angegeben ist, darf keinen Code des Umsatzsteuerbefreiungsgrundes "VAT exemption reason code“ (BT-121) oder Text
	// des Umsatzsteuerbefreiungsgrundes "VAT exemption reason text“ (BT-120) enthalten.
	// ZUGFeRD 2.3.2 / Factur-X 1.07.2 76 05.07.2024
	// Generated by GEFEG.FX Copyright AWV e.V., FNFE 2012-2024ZUGFeRD 2.3.2 / Factur-X 1.07.2 Spezifikation - Technischer Anhang
	// Liste der Geschäftsregeln
	// Nr. Kontext
	// BR-Z-1 Umsatzsteuer mit Nullsatz
	// Eine Rechnung (INVOICE), die eine Position, einen Nachlass oder eine Abgabe auf der Rechnungsebene enthält, bei der als Code der
	// Umsatzsteuerkategorie des in Rechnung gestellten Postens ("Invoiced item VAT category code“ (BT-151), "Document level allowance VAT
	// category code“ (BT-95) oder "Document level charge VAT category code“ (BT-102)) der Wert "Zero rated“ angegeben ist, muss in "VAT
	// BREAKDOWN“ (BG-23) genau einen Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) gleich dem Wert "Zero rated“ enthalten.
	// BR-Z-2 Umsatzsteuer mit Nullsatz
	// Eine Rechnung (INVOICE), die eine Position enthält, in der als Code der Umsatzsteuerkategorie für den in Rechnung gestellten Posten "Invoiced
	// item VAT category code“ (BT-151) der Wert "Zero rated“ angegeben ist, muss die Umsatzsteuer-Identifikationsnummer des Verkäufers "Seller VAT
	// identifier“ (BT-31), die Steueridentifikationsnummer des Verkäufers "Seller tax registration identifier“ (BT-32) oder die Umsatzsteuer-
	// Identifikationsnummer des Steuervertreters des Verkäufers "Seller tax representative VAT identifier“ (BT-63) enthalten.
	// BR-Z-3 Umsatzsteuer mit Nullsatz
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL ALLOWANCES“ (BG-20) enthält, in der "Document level allowance VAT category code“
	// (BT-95) den Wert "Zero rated“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax
	// representative VAT identifier“ (BT-63) vorhanden sein.
	// BR-Z-4 Umsatzsteuer mit Nullsatz
	// In einer Rechnung, die eine Gruppe "DOCUMENT LEVEL CHARGES“ (BG-21) enthält, in der "Document level charge VAT category code“ (BT-102)
	// den Wert "Zero rated“ hat, muss entweder "Seller VAT identifier“ (BT-31), "Seller tax registration identifier“ (BT-32) oder "Seller tax representative
	// VAT identifier“ (BT-63) vorhanden sein.
	// BR-Z-5 Umsatzsteuer mit Nullsatz
	// In einer "INVOICE LINE“ (BG-25), in der "Invoiced item VAT category code“ (BT-151) den Wert "Zero rated“ hat, muss "Invoiced item VAT rate“
	// (BT-152) gleich "0“ sein.
	// BR-Z-6 Umsatzsteuer mit Nullsatz
	// In einer "DOCUMENT LEVEL ALLOWANCES“ (BG-20), in der "Document level allowance VAT category code“ (BT-95) den Wert "Zero rated“ hat,
	// muss "Document level allowance VAT rate“ (BT-96) gleich "0“ sein.
	// BR-Z-7 Umsatzsteuer mit Nullsatz
	// In einer "DOCUMENT LEVEL CHARGES“ (BG-21), in der "Document level charge VAT category code“ (BT-102) den Wert "Zero rated“ hat, muss
	// "Document level charge VAT rate“ (BT-103) gleich "0“ sein.
	// BR-Z-8 Umsatzsteuer mit Nullsatz
	// In einer "VAT BREAKDOWN“ (BG-23), in der als Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) der Wert "Zero rated“ angegeben
	// ist, muss der "VAT category taxable amount“ (BT-116) gleich der Summe der Informationselemente "Invoice line net amount“ (BT-131) abzüglich
	// der "Document level allowance amount“ (BT-92) zuzüglich der "Document level charge amount“ (BT-99) sein, wobei als "Invoiced item VAT
	// category code“ (BT-151), als "Document level allowance VAT category code“ (BT-95) sowie als "Document level charge VAT category code“ (BT-
	// 102) jeweils der Wert "Zero rated“ angegeben wird.
	// BR-Z-9 Umsatzsteuer mit Nullsatz
	// Der "VAT category tax amount“ (BT-117) muss in einer "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category
	// code“ (BT-118) mit dem Wert "Zero rated“ gleich "0“ sein.
	// BR-Z-10 Umsatzsteuer mit Nullsatz
	// Ein "VAT BREAKDOWN“ (BG-23) mit dem Code der Umsatzsteuerkategorie "VAT category code“ (BT-118) mit dem Wert "Zero rated“ darf keinen
	// Code des Umsatzsteuerbefreiungsgrundes "VAT exemption reason code“ (BT-121) oder Text des Umsatzsteuerbefreiungsgrundes "VAT exemption
}

func (inv *Invoice) checkOther() {
	// Check that line total = billed quantity * net price
	for _, line := range inv.InvoiceLines {
		calcTotal := line.BilledQuantity.Mul(line.NetPrice)
		lineTotal := line.Total
		if !lineTotal.Equal(calcTotal) {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "Check", InvFields: []string{"BT-146", "BT-149", "BT-131"}, Text: fmt.Sprintf("Line total %s does not match quantity %s * net price %s", lineTotal.String(), line.BilledQuantity.String(), calcTotal.String())})
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
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-3", InvFields: []string{"BT-7", "BT-8"}, Text: "TaxPointDate and DueDateTypeCode are mutually exclusive"})
			break
		}
	}

	// BR-CO-4 Rechnungsposition
	// Jede Rechnungsposition "INVOICE LINE" (BG-25) muss anhand der Umsatzsteuerkategorie des in Rechnung gestellten Postens "Invoiced item VAT
	// category code" (BT-151) kategorisiert werden.
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-4", InvFields: []string{"BG-25", "BT-151"}, Text: fmt.Sprintf("Invoice line %s missing VAT category code", line.LineID)})
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
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-10", InvFields: []string{"BT-106", "BT-131"}, Text: fmt.Sprintf("Line total %s does not match sum of invoice lines %s", inv.LineTotal.String(), sum.String())})
	}

	// BR-CO-13 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount without VAT" (BT-109) entspricht der Summe aus "Sum of Invoice line net amount"
	// (BT-106) abzüglich "Sum of allowances on document level" (BT-107) zuzüglich "Sum of charges on document level" (BT-108).
	expectedTaxBasisTotal := inv.LineTotal.Sub(inv.AllowanceTotal).Add(inv.ChargeTotal)
	if !inv.TaxBasisTotal.Equal(expectedTaxBasisTotal) {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-13", InvFields: []string{"BT-109", "BT-106", "BT-107", "BT-108"}, Text: fmt.Sprintf("Tax basis total %s does not match LineTotal - AllowanceTotal + ChargeTotal = %s", inv.TaxBasisTotal.String(), expectedTaxBasisTotal.String())})
	}

	// BR-CO-15 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Invoice total amount with VAT" (BT-112) entspricht der Summe aus "Invoice total amount without VAT"
	// (BT-109) und "Invoice total VAT amount" (BT-110).
	expectedGrandTotal := inv.TaxBasisTotal.Add(inv.TaxTotal)
	if !inv.GrandTotal.Equal(expectedGrandTotal) {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-15", InvFields: []string{"BT-112", "BT-109", "BT-110"}, Text: fmt.Sprintf("Grand total %s does not match TaxBasisTotal + TaxTotal = %s", inv.GrandTotal.String(), expectedGrandTotal.String())})
	}

	// BR-CO-16 Gesamtsummen auf Dokumentenebene
	// Der Inhalt des Elementes "Amount due for payment" (BT-115) entspricht der Summe aus "Invoice total amount with VAT" (BT-112)
	// abzüglich "Paid amount" (BT-113) zuzüglich "Rounding amount" (BT-114).
	expectedDuePayableAmount := inv.GrandTotal.Sub(inv.TotalPrepaid).Add(inv.RoundingAmount)
	if !inv.DuePayableAmount.Equal(expectedDuePayableAmount) {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-16", InvFields: []string{"BT-115", "BT-112", "BT-113", "BT-114"}, Text: fmt.Sprintf("Due payable amount %s does not match GrandTotal - TotalPrepaid + RoundingAmount = %s", inv.DuePayableAmount.String(), expectedDuePayableAmount.String())})
	}

	// BR-CO-17 Umsatzsteueraufschlüsselung
	// Der Inhalt des Elementes "VAT category tax amount" (BT-117) entspricht dem Inhalt des Elementes "VAT category taxable amount" (BT-116),
	// multipliziert mit dem Inhalt des Elementes "VAT category rate" (BT-119) geteilt durch 100, gerundet auf zwei Dezimalstellen.
	for _, tax := range inv.TradeTaxes {
		expected := tax.BasisAmount.Mul(tax.Percent).Div(decimal.NewFromInt(100)).Round(2)
		if !tax.CalculatedAmount.Equal(expected) {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-17", InvFields: []string{"BT-116", "BT-117", "BT-119"}, Text: fmt.Sprintf("VAT category tax amount %s does not match expected %s (basis %s × rate %s ÷ 100)", tax.CalculatedAmount.String(), expected.String(), tax.BasisAmount.String(), tax.Percent.String())})
		}
	}

	// BR-CO-18 Umsatzsteueraufschlüsselung
	// Eine Rechnung (INVOICE) soll mindestens eine Gruppe "VAT BREAKDOWN" (BG-23) enthalten.
	if len(inv.TradeTaxes) < 1 {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-18", InvFields: []string{"BG-23"}, Text: "Invoice should contain at least one VAT BREAKDOWN"})
	}

	// BR-CO-19 Liefer- oder Rechnungszeitraum
	// Wenn die Gruppe "INVOICING PERIOD" (BG-14) verwendet wird, müssen entweder das Element "Invoicing period start date" (BT-73) oder das
	// Element "Invoicing period end date" (BT-74) oder beide gefüllt sein.
	if !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero() {
		if inv.BillingSpecifiedPeriodStart.IsZero() && inv.BillingSpecifiedPeriodEnd.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-19", InvFields: []string{"BG-14", "BT-73", "BT-74"}, Text: "If invoicing period is used, either start date or end date must be filled"})
		}
	}

	// BR-CO-20 Rechnungszeitraum auf Positionsebene
	// Wenn die Gruppe "INVOICE LINE PERIOD" (BG-26) verwendet wird, müssen entweder das Element "Invoice line period start date" (BT-134) oder
	// das Element "Invoice line period end date" (BT-135) oder beide gefüllt sein.
	for _, line := range inv.InvoiceLines {
		if !line.BillingSpecifiedPeriodStart.IsZero() || !line.BillingSpecifiedPeriodEnd.IsZero() {
			if line.BillingSpecifiedPeriodStart.IsZero() && line.BillingSpecifiedPeriodEnd.IsZero() {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-20", InvFields: []string{"BG-26", "BT-134", "BT-135"}, Text: fmt.Sprintf("Invoice line %s: if line period is used, either start date or end date must be filled", line.LineID)})
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
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-CO-25", InvFields: []string{"BT-9", "BT-20", "BT-115"}, Text: "If amount due for payment is positive, either payment due date or payment terms must be present"})
		}
	}

}

func (inv *Invoice) checkBR() {
	// BR-1
	// Eine Rechnung (INVOICE) muss eine Spezifikationskennung "Specification identification“ (BT-24) enthalten.
	if inv.Profile == CProfileUnknown {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-1", InvFields: []string{"BT-24"}, Text: "Could not determine the profile in GuidelineSpecifiedDocumentContextParameter"})
	}
	// 	BR-2 Rechnung
	// Eine Rechnung (INVOICE) muss eine Rechnungsnummer "Invoice number“ (BT-1) enthalten.
	if inv.InvoiceNumber == "" {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-2", InvFields: []string{"BT-1"}, Text: "No invoice number found"})
	}
	// BR-3 Rechnung
	// Eine Rechnung (INVOICE) muss ein Rechnungsdatum "Invoice issue date“ (BT-2) enthalten.
	if inv.InvoiceDate.IsZero() {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-3", InvFields: []string{"BT-2"}, Text: "Date is zero"})
	}
	// BR-4 Rechnung
	// Eine Rechnung (INVOICE) muss einen Rechnungstyp-Code "Invoice type code“ (BT-3) enthalten.
	if inv.InvoiceTypeCode == 0 {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-4", InvFields: []string{"BT-3"}, Text: "Invoice type code is 0"})
	}
	// BR-5 Rechnung
	// Eine Rechnung (INVOICE) muss einen Währungs-Code "Invoice currency code“ (BT-5) enthalten.
	if inv.InvoiceCurrencyCode == "" {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-5", InvFields: []string{"BT-5"}, Text: "Invoice currency code is empty"})
	}
	// BR-6 Verkäufer
	// Eine Rechnung (INVOICE) muss den Verkäufernamen "Seller name“ (BT-27) enthalten.
	if inv.Seller.Name == "" {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-6", InvFields: []string{"BT-27"}, Text: "Seller name is empty"})
	}
	// BR-7 Käufer
	// Eine Rechnung (INVOICE) muss den Erwerbernamen "Buyer name“ (BT-44) enthalten.
	if inv.Buyer.Name == "" {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-7", InvFields: []string{"BT-44"}, Text: "Buyer name is empty"})
	}
	// BR-8 Verkäufer
	// Eine Rechnung (INVOICE) muss die postalische Anschrift des Verkäufers "SELLER POSTAL ADDRESS“ (BG-5) enthalten.
	if inv.Seller.PostalAddress == nil {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-8", InvFields: []string{"BG-5"}, Text: "Seller has no postal address"})
	} else {
		// BR-9 Verkäufer
		// Eine postalische Anschrift des Verkäufers "SELLER POSTAL ADDRESS“ (BG-5) muss einen Verkäufer-Ländercode "Seller country code“ (BT-40) enthalten.
		if inv.Seller.PostalAddress.CountryID == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-9", InvFields: []string{"BT-40"}, Text: "Seller country code empty"})
		}
	}
	if inv.Profile > CProfileMinimum {
		// BR-10 Käufer
		// Eine Rechnung (INVOICE) muss die postalische Anschrift des Erwerbers "BUYER POSTAL ADDRESS“ (BG-8) enthalten.
		if inv.Buyer.PostalAddress == nil {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-10", InvFields: []string{"BG-8"}, Text: "Buyer has no postal address"})
		} else {
			// BR-11 Käufer
			// Eine postalische Anschrift des Erwerbers "BUYER POSTAL ADDRESS“ (BG-8) muss einen Erwerber-Ländercode "Buyer country code“ (BT-55)
			// enthalten.
			if inv.Buyer.PostalAddress.CountryID == "" {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-11", InvFields: []string{"BT-55"}, Text: "Buyer country code empty"})
			}
		}
	}
	// BR-12 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss die Summe der Rechnungspositionen-Nettobeträge "Sum of Invoice line net amount“ (BT-106) enthalten.
	if inv.LineTotal.IsZero() {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-12", InvFields: []string{"BT-106"}, Text: "Line total is zero"})
	}
	// BR-13 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den Gesamtbetrag der Rechnung ohne Umsatzsteuer "Invoice total amount without VAT“ (BT-109) enthalten.
	if inv.TaxBasisTotal.IsZero() {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-13", InvFields: []string{"BT-109"}, Text: "TaxBasisTotal zero"})
	}
	// BR-14 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den Gesamtbetrag der Rechnung mit Umsatzsteuer "Invoice total amount with VAT“ (BT-112) enthalten.
	if inv.GrandTotal.IsZero() {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-14", InvFields: []string{"BT-112"}, Text: "GrandTotal is zero"})
	}
	// BR-15 Gesamtsummen auf Dokumentenebene
	// Eine Rechnung (INVOICE) muss den ausstehenden Betrag "Amount due for payment“ (BT-115) enthalten.
	if inv.DuePayableAmount.IsZero() {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-15", InvFields: []string{"BT-115"}, Text: "DuePayableAmount is zero"})
	}
	// BR-16 Rechnung
	// Eine Rechnung (INVOICE) muss mindestens eine Rechnungsposition "INVOICE LINE“ (BG-25) enthalten.
	if is(CProfileBasic, inv) {
		if len(inv.InvoiceLines) == 0 {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-16", InvFields: []string{"BG-25"}, Text: "Invoice lines must be at least 1"})
		}
	}
	// BR-17 Zahlungsempfänger
	// Eine Rechnung (INVOICE) muss den Namen des Zahlungsempfängers "Payee name“ (BT-59) enthalten, wenn sich der Zahlungsempfänger "PAYEE“
	// (BG-10) vom Verkäufer "SELLER“ (BG-4) unterscheidet.
	if inv.PayeeTradeParty != nil {
		if inv.PayeeTradeParty.Name == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-17", InvFields: []string{"BT-59", "BG-10", "BG-4"}, Text: "Payee has no name, although different from seller"})
		}
	}
	// BR-18 Steuerbevollmächtigter des Verkäufers
	// Eine Rechnung (INVOICE) muss den Namen des Steuervertreters des Verkäufers "Seller tax representative name“ (BT-62) enthalten, wenn der
	// Verkäufer "SELLER“ (BG-4) einen Steuervertreter (BG-11) hat.
	if trp := inv.SellerTaxRepresentativeTradeParty; trp != nil {
		if trp.Name == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-18", InvFields: []string{"BT-62", "BG-4", "BG-11"}, Text: "Tax representative has no name, although seller has specified one"})
		}
		// BR-19 Steuerbevollmächtigter des Verkäufers
		// Eine Rechnung (INVOICE) muss die postalische Anschrift des Steuervertreters "SELLER TAX REPRESENTATIVE POSTAL ADDRESS“ (BG-12) enthalten,
		// wenn der Verkäufer "SELLER“ (BG-4) einen Steuervertreter hat.
		if trp.PostalAddress == nil {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-19", InvFields: []string{"BG-4", "BG-12"}, Text: "Tax representative has no postal address"})
		} else {
			// BR-20 Steuerbevollmächtigter des Verkäufers
			// Die postalische Anschrift des Steuervertreters des Verkäufers "SELLER TAX REPRESENTATIVE POSTAL ADDRESS“ (BG-12) muss einen
			// Steuervertreter-Ländercode enthalten, wenn der Verkäufer "SELLER“ (BG-4) einen Steuervertreter hat.
			if trp.PostalAddress.CountryID == "" {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-20", InvFields: []string{"BG-4", "BG-12"}, Text: "Tax representative has no postal address"})
			}
		}
	}
	for _, line := range inv.InvoiceLines {
		// BR-21 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss eine eindeutige Bezeichnung "Invoice line identifier“ (BT-126) haben.
		if line.LineID == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-21", InvFields: []string{"BG-25", "BT-126"}, Text: "Line has no line ID"})
		}
		// BR-22 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss die Menge der in der betreffenden Position in Rechnung gestellten Waren oder
		// Dienstleistungen als Einzelposten "Invoiced quantity“ (BT-129) enthalten.
		if line.BilledQuantity.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-22", InvFields: []string{"BG-25", "BT-129"}, Text: "Line has no billed quantity"})
		}
		// BR-23 Rechnungsposition
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss eine Einheit zur Mengenangabe "Invoiced quantity unit of measure code“ (BT-130)
		// enthalten.
		if line.BilledQuantityUnit == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-23", InvFields: []string{"BG-25", "BT-130"}, Text: "Line's billed quantity has no unit"})
		}

		// BR-24 in parser.go

		// BR-25 Artikelinformationen
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss den Namen des Postens "Item name“ (BT-153) enthalten.
		if line.ItemName == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-25", InvFields: []string{"BG-25", "BT-153"}, Text: "Line's item name missing"})
		}
		// BR-26 Detailinformationen zum Preis
		// Jede Rechnungsposition "INVOICE LINE“ (BG-25) muss den Preis des Postens, ohne Umsatzsteuer, nach Abzug des für diese Rechnungsposition
		// geltenden Rabatts "Item net price“ (BT-146) beinhalten.

		// BR-27 Nettopreis des Artikels
		// Der Artikel-Nettobetrag "Item net price“ (BT-146) darf nicht negativ sein.
		if line.NetPrice.IsNegative() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-27", InvFields: []string{"BG-25", "BT-146"}, Text: "Net price must not be negative"})
		}
		// BR-28 Detailinformationen zum Preis
		// Der Einheitspreis ohne Umsatzsteuer vor Abzug des Postenpreisrabatts einer Rechnungsposition "Item gross price“ (BT-148) darf nicht negativ
		// sein.
		// TODO
	}
	// BR-29 Rechnungszeitraum
	// Wenn Start- und Enddatum des Rechnungszeitraums gegeben sind, muss das Enddatum "Invoicing period end date“ (BT-74) nach dem Startdatum
	// "Invoicing period start date“ (BT-73) liegen oder mit diesem identisch sein.
	if inv.BillingSpecifiedPeriodEnd.Before(inv.BillingSpecifiedPeriodStart) {
		inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-29", InvFields: []string{"BT-73", "BT-74"}, Text: "Billing period end must be after start"})
	}
	for _, line := range inv.InvoiceLines {
		// BR-30 Rechnungszeitraum auf Positionsebene
		// Wenn Start- und Enddatum des Rechnungspositionenzeitraums gegeben sind, muss das Enddatum "Invoice line period end date“ (BT-135) nach
		// dem Startdatum "Invoice line period start date“ (BT-134) liegen oder mit diesem identisch sein.
		if line.BillingSpecifiedPeriodEnd.Before(line.BillingSpecifiedPeriodStart) {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-30", InvFields: []string{"BG-25", "BT-135", "BT-134"}, Text: "Line item billing period end must be after or identical to start"})
		}
	}

	// Initialize applicableTradeTaxes map for BR-45 validation
	var applicableTradeTaxes = make(map[string]decimal.Decimal, len(inv.TradeTaxes))
	for _, lineitem := range inv.InvoiceLines {
		percentString := lineitem.TaxRateApplicablePercent.String()
		applicableTradeTaxes[percentString] = applicableTradeTaxes[percentString].Add(lineitem.Total)
	}

	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		// Add to applicableTradeTaxes for BR-45 validation
		percentString := ac.CategoryTradeTaxRateApplicablePercent.String()
		amount := ac.ActualAmount
		if !ac.ChargeIndicator {
			amount = amount.Neg()
		}
		applicableTradeTaxes[percentString] = applicableTradeTaxes[percentString].Add(amount)

		if ac.ChargeIndicator {
			// BR-36 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Betrag "Document level charge amount" (BT-99)
			// aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-36", InvFields: []string{"BG-21", "BT-99"}, Text: "Charge must not be zero"})
			}

			// BR-37 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Umsatzsteuer-Code "Document level charge VAT
			// category code" (BT-102) aufweisen.
			if ac.CategoryTradeTaxCategoryCode == "" {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-37", InvFields: []string{"BG-21", "BT-102"}, Text: "Charge tax category code not set"})
			}
			// BR-38 Zuschläge auf Dokumentenebene
			// Jede Abgabe auf Dokumentenebene "DOCUMENT LEVEL CHARGES" (BG-21) muss einen Abgabegrund "Document level charge reason" (BT-104)
			// oder einen entsprechenden Code "Document level charge reason code" (BT-105) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-38", InvFields: []string{"BG-21", "BT-104", "BT-105"}, Text: "Charge reason empty or code unset"})
			}
		} else {
			// BR-31 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Betrag "Document level allowance amount"
			// (BT-92) aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-31", InvFields: []string{"BG-20", "BT-92"}, Text: "Allowance must not be zero"})
			}
			// BR-32 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Umsatzsteuer-Code "Document level
			// allowance VAT category code" (BT-95) aufweisen.
			if ac.CategoryTradeTaxCategoryCode == "" {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-32", InvFields: []string{"BG-20", "BT-95"}, Text: "Allowance tax category code not set"})
			}
			// BR-33 Abschläge auf Dokumentenebene
			// Jeder Nachlass für die Rechnung als Ganzes "DOCUMENT LEVEL ALLOWANCES" (BG-20) muss einen Nachlassgrund "Document level allowance
			// reason" (BT-97) oder einen entsprechenden Code "Document level allowance reason code" (BT-98) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-33", InvFields: []string{"BG-20", "BT-95"}, Text: "Allowance reason empty or code unset"})
			}
		}
	}

	for _, line := range inv.InvoiceLines {
		// BR-41 Abschläge auf Ebene der Rechnungsposition
		// Jeder Nachlass auf der Ebene der Rechnungsposition "INVOICE LINE ALLOWANCES“ (BG-27) muss einen Betrag "Invoice line allowance amount“
		// (BT-136) aufweisen.
		for _, ac := range line.InvoiceLineAllowances {
			if ac.ActualAmount.IsZero() {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-41", InvFields: []string{"BG-27", "BT-136"}, Text: "Line allowance amount zero"})
			}
			// BR-42 Abschläge auf Ebene der Rechnungsposition
			// Jeder Nachlass auf der Ebene der Rechnungsposition "INVOICE LINE ALLOWANCES“ (BG-27) muss einen Nachlassgrund "Invoice line allowance
			// reason“ (BT-139) oder einen entsprechenden Code "Invoice line allowance reason code“ (BT-140) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-42", InvFields: []string{"BG-27", "BT-139", "BT-140"}, Text: "Line allowance must have a reason"})
			}
		}
		for _, ac := range line.InvoiceLineCharges {
			// BR-43 Charge ou frais sur ligne de facture
			// Jede Abgabe auf der Ebene der Rechnungsposition "INVOICE LINE CHARGES“ (BG-28) muss einen Betrag "Invoice line charge amount“ (BT-141)
			// aufweisen.
			if ac.ActualAmount.IsZero() {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-43", InvFields: []string{"BG-28", "BT-141"}, Text: "Line charge amount zero"})
			}
			// BR-44 Charge ou frais sur ligne de facture
			// Jede Abgabe auf der Ebene der Rechnungsposition "INVOICE LINE CHARGES“ (BG-28) muss einen Abgabegrund "Invoice line charge reason“ (BT-
			// 144) oder einen entsprechenden Code "Invoice line charge reason code“ (BT-145) aufweisen.
			if ac.Reason == "" && ac.ReasonCode == 0 {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-44", InvFields: []string{"BG-28", "BT-144", "BT-145"}, Text: "Line charge must have a reason"})
			}
		}
	}

	for _, tt := range inv.TradeTaxes {
		// BR-45 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) muss die
		// Summe aller nach dem jeweiligen Schlüssel zu versteuernden Beträge
		// "VAT category taxable amount“ (BT-116) aufweisen.
		if !applicableTradeTaxes[tt.Percent.String()].Equal(tt.BasisAmount) {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-45", InvFields: []string{"BG-23", "BT-116"}, Text: "Applicable trade tax basis amount not equal to the sum of line total"})

		}
		// BR-46 Umsatzsteueraufschlüsselung
		// in parser.go

		// BR-47 Umsatzsteueraufschlüsselung
		// Jede Umsatzsteueraufschlüsselung "VAT BREAKDOWN“ (BG-23) muss über
		// eine codierte Bezeichnung einer Umsatzsteuerkategorie "VAT category
		// code“ (BT-118) definiert werden.
		if tt.CategoryCode == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-47", InvFields: []string{"BG-23", "BT-118"}, Text: "CategoryCode not set for applicable trade tax"})
		}

		// BR-48 Umsatzsteueraufschlüsselung
		// in parser.go
	}
	for _, pm := range inv.PaymentMeans {
		// BR-49 Zahlungsanweisungen
		// Die Zahlungsinstruktionen "PAYMENT INSTRUCTIONS“ (BG-16) müssen den Zahlungsart-Code "Payment means type code“ (BT-81) enthalten.
		if pm.TypeCode == 0 {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-49", InvFields: []string{"BT-81"}, Text: "Payment means type code must be set"})
		}
	}
	// BR-50 Kontoinformationen
	// Die Kennung des Kontos, auf das die Zahlung erfolgen soll "Payment
	// account identifier“ (BT-84), muss angegeben werden, wenn
	// Überweisungsinformationen in der Rechnung angegeben werden.

	// BR-51 Karteninformationen
	// Die letzten vier bis sechs Ziffern der Kreditkartennummer "Payment card
	// primary account number“ (BT-87) sollen angegeben werden, wenn
	// Informationen zur Kartenzahlung übermittelt werden.

	// BR-52 Rechnungsbegründende Unterlagen
	// Jede rechnungsbegründende Unterlage muss einen Dokumentenbezeichner
	// "Supporting document reference“ (BT-122) haben.

	// BR-53 Gesamtsummen auf Dokumentenebene
	// Wenn eine Währung für die Umsatzsteuerabrechnung angegeben wurde, muss
	// der Umsatzsteuergesamtbetrag in der Abrechnungswährung "Invoice total VAT
	// amount in accounting currency“ (BT-111) angegeben werden.

	// BR-54 Artikelattribute
	// Jede Eigenschaft eines in Rechnung gestellten Postens "ITEM ATTRIBUTES“
	// (BG-32) muss eine Bezeichnung "Item attribute name“ (BT-160) und einen
	// Wert "Item attribute value“ (BT-161) haben.

	// BR-55 Referenz auf die vorausgegangene Rechnung
	// Jede Bezugnahme auf eine vorausgegangene Rechnung "Preceding Invoice
	// reference“ (BT-25) muss die Nummer der vorausgegangenen Rechnung
	// enthalten.

	// BR-56 Steuerbevollmächtigter des Verkäufers
	// Jeder Steuervertreter des Verkäufers "SELLER TAX REPRESENTATIVE PARTY“
	// (BG-11) muss eine Umsatzsteuer-Identifikationsnummer "Seller tax
	// representative VAT identifier“ (BT-63) haben.

	// BR-57 Lieferanschrift
	// Jede Lieferadresse "DELIVER TO ADDRESS“ (BG-15) muss einen entsprechenden
	// Ländercode "Deliver to country code“ (BT-80) enthalten.

	// BR-61 Zahlungsanweisungen
	// Wenn der Zahlungsmittel-Typ SEPA, lokale Überweisung oder
	// Nicht-SEPA-Überweisung ist, muss der "Payment account identifier“ (BT-84)
	// des Zahlungsempfängers angegeben werden.

	// BR-62 Elektronische Adresse des Verkäufers
	// Im Element "Seller electronic address“ (BT-34) muss die Komponente
	// "Scheme Identifier“ vorhanden sein.

	// BR-63 Elektronische Adresse des Käufers
	// Im Element "Buyer electronic address“ (BT-49) muss die Komponente "Scheme
	// Identifier“ vorhanden sein.

	// BR-64 Kennung eines Artikels nach registriertem Schema
	// Im Element "Item standard identifier“ (BT-157) muss die Komponente
	// "Scheme Identifier“ vorhanden sein.

	// BR-65 Kennung der Artikelklassifizierung
	// Im Element "Item classification identifier“ (BT-158) muss die Komponente
	// "Scheme Identifier“ vorhanden sein.
}
