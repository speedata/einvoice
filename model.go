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
	// CodeGlobalID is the ISO 6523 type
	CodeGlobalID int
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

// InvoiceType UNTDID 1001 Document name code
const (
	CDummy              CodeInvoice = 0
	CommercialInvoice               = 380
	CCreditNote                     = 381
	CCorrectedInvoice               = 384
	CHireInvoice                    = 387
	CSSelfBilledInvoice             = 389
	CInvoiceInformation             = 751 // Profile BASIC WL und MINIMUM
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

// Party represents buyer and seller
type Party struct {
	ID             []string
	GlobalID       string
	GlobalScheme   string
	Name           string
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

// Characteristic add details to a product
type Characteristic struct {
	Description string
	Value       string
}

// Position ist eine Rechnungszeile
type Position struct {
	Position            int
	ArticleNumber       string
	ArticleName         string
	Note                string // optional
	GlobalID            string
	GlobalIDType        CodeGlobalID
	Characteristics     []Characteristic
	BruttoPreis         decimal.Decimal
	NettoPreis          decimal.Decimal
	Anzahl              decimal.Decimal
	Einheit             string
	Description         string // BT-154
	SteuerTypCode       string // muss VAT sein
	SteuerKategorieCode string
	Steuersatz          decimal.Decimal
	Total               decimal.Decimal
}

// PaymentMeans represents a payment means
type PaymentMeans struct {
	TypeCode                                             int
	Information                                          string
	PayeePartyCreditorFinancialAccountIBAN               string
	PayeePartyCreditorFinancialAccountName               string
	PayeePartyCreditorFinancialAccountProprietaryID      string
	PayeeSpecifiedCreditorFinancialInstitutionBIC        string // BT-86
	PayerPartyDebtorFinancialAccountIBAN                 string
	ApplicableTradeSettlementFinancialCardID             string
	ApplicableTradeSettlementFinancialCardCardholderName string
}

// AllowanceCharge specifies charges and deductions
type AllowanceCharge struct {
	ChargeIndicator                       bool
	CalculationPercent                    decimal.Decimal
	BasisAmount                           decimal.Decimal
	ActualAmount                          decimal.Decimal
	ReasonCode                            int
	Reason                                string
	CategoryTradeTaxType                  string
	CategoryTradeTaxCategoryCode          string
	CategoryTradeTaxRateApplicablePercent decimal.Decimal
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

// Invoice ist das Hauptelement der e-Invoice-Datei
type Invoice struct {
	AllowanceTotal                      decimal.Decimal
	BuyerOrderReferencedDocument        string // BT-13
	DespatchAdviceReferencedDocument    string // Detailinformationen zum zugehörigen Lieferavis
	BuyerReference                      string // ApplicableHeaderTradeAgreement/BuyerReference
	BPSpecifiedDocumentContextParameter string
	PaymentMeans                        []PaymentMeans
	BillingSpecifiedPeriodStart         time.Time
	BillingSpecifiedPeriodEnd           time.Time
	InvoiceDate                         time.Time // BT-2
	ChargeTotal                         decimal.Decimal
	DuePayableAmount                    decimal.Decimal
	DueDate                             time.Time
	TradePaymentTermsDescription        string // BT-20, BR-CO-25 BT-115>0?BT-9||BT-20
	DirectDebitMandateID                string // BG-19/BT-89
	GrandTotal                          decimal.Decimal
	Buyer                               Party
	Leistungsdatum                      time.Time
	LineTotal                           decimal.Decimal
	Notizen                             []Notiz
	Positionen                          []Position
	Profile                             CodeProfile
	InvoiceNumber                       string      // BT-1
	Rechnungstyp                        CodeInvoice // BT-3
	Steuersätze                         []Steuersatz
	TaxBasisTotal                       decimal.Decimal
	TaxTotalCurrency                    string
	TaxTotal                            decimal.Decimal
	TotalPrepaid                        decimal.Decimal
	SpecifiedTradeAllowanceCharge       []AllowanceCharge
	ShipTo                              *Party
	Seller                              Party
	SpecifiedTradePaymentTerms          string
	Currency                            string // BT-5
}
