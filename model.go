package einvoice

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type (
	// CodeProfile bestimmt den Subtyp der Rechnung
	CodeProfile int
	// CodeInvoice ist der Rechnungstyp
	CodeInvoice int
)

func (cp CodeProfile) String() string {
	switch cp {
	case CZUGFeRD:
		return "ZUGFeRD/Factur-X"
	default:
		return "unknown"
	}
}

// CodeProfile ist der Subtyp der Rechnung
const (
	CZUGFeRD CodeProfile = iota
)

// Rechnungstyp
const (
	CRechnung                   CodeInvoice = 380
	CGutschrift                             = 381
	CKorrektur                              = 384
	CSelbstAusgestellteRechnung             = 389
	CWertbelastung                          = 89
	CKeineRechnung                          = 751 // Profile BASIC WL und MINIMUM
)

// Notiz ist ein Freitext für bestimmte Themen, falls SubjectCode angegeben
// wird.
type Notiz struct {
	Text        string
	SubjectCode string
}

func (n Notiz) String() string {
	return fmt.Sprintf("Notiz %s - %q", n.SubjectCode, n.Text)
}

// Adresse repräsentiert den Käufer und den Verkäufer
type Adresse struct {
	Firmenname     string
	Kontakt        string
	EMail          string
	PLZ            string
	Straße1        string
	Straße2        string
	Ort            string
	Ländercode     string
	UmsatzsteuerID string
	Steuernummer   string
}

// Position ist eine Rechnungszeile
type Position struct {
	Position            int
	Artikelnummer       string
	ArtikelName         string
	BruttoPreis         decimal.Decimal
	NettoPreis          decimal.Decimal
	Anzahl              decimal.Decimal
	Einheit             string
	Freitext            string
	SteuerTypCode       string // muss VAT sein
	SteuerKategorieCode string
	Steuersatz          decimal.Decimal
	Total               decimal.Decimal
}

// Steuersatz hängt an einer Rechnung und bezeichnet alle vorkommenden
// Steuersätze
type Steuersatz struct {
	BerechneterWert decimal.Decimal
	BasisWert       decimal.Decimal
	Typ             string
	KategorieCode   string
	Prozent         decimal.Decimal
	Ausnahmegrund   string
}

// Rechnung ist das Hauptelement der e-Rechnung-Datei
type Rechnung struct {
	AllowanceTotal   decimal.Decimal
	BankBIC          string
	BankIBAN         string
	BankKontoname    string
	Belegdatum       time.Time // BT-2
	ChargeTotal      decimal.Decimal
	DuePayableAmount decimal.Decimal
	Fälligkeitsdatum time.Time
	GrandTotal       decimal.Decimal
	Käufer           Adresse
	Leistungsdatum   time.Time
	LineTotal        decimal.Decimal
	Notizen          []Notiz
	Positionen       []Position
	Profil           CodeProfile
	Rechnungsnummer  string      // BT-1
	Rechnungstyp     CodeInvoice // BT-3
	Steuersätze      []Steuersatz
	TaxBasisTotal    decimal.Decimal
	TaxTotal         decimal.Decimal
	TotalPrepaid     decimal.Decimal
	Verkäufer        Adresse
	Währung          string // BT-5
}
