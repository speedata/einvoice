package einvoice

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type (
	// CodeSchemaType represents the type of the invoice (CII or UBL)
	CodeSchemaType int
	// CodeDocument contains the UNTDID 1001 document code
	CodeDocument int
	// CodeGlobalID is the ISO 6523 type
	CodeGlobalID int
)

func (cp CodeSchemaType) String() string {
	switch cp {
	case CII:
		return "ZUGFeRD/Factur-X"
	case UBL:
		return "UBL"
	default:
		return "unknown"
	}
}

// CodeProfile ist der Subtyp der Rechnung
const (
	CII CodeSchemaType = iota
	UBL
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
	ID                []string
	GlobalID          string
	GlobalScheme      string
	Name              string
	PersonName        string
	EMail             string
	ZIP               string
	Address1          string
	Address2          string
	City              string
	CountryID         string
	VATaxRegistration string
	FCTaxRegistration string
}

// Characteristic add details to a product
type Characteristic struct {
	Description string
	Value       string
}

// Classification specifies a product classification
type Classification struct {
	ClassCode     string
	ListID        string
	ListVersionID string
}

// InvoiceItem ist eine Rechnungszeile
type InvoiceItem struct {
	Position           int    // BT-126
	ArticleNumber      string // BT-155 seller assigned ID
	ArticleNumberBuyer string // BT-156 buyer assigned ID
	ArticleName        string // BT-153
	// BuyerOrderReferencedDocument
	Note                     string // optional
	GlobalID                 string
	GlobalIDType             CodeGlobalID
	Characteristics          []Characteristic // BG-32
	ProductClassification    []Classification // UNTDID 7143
	GrossPrice               decimal.Decimal
	NetPrice                 decimal.Decimal
	BilledQuantity           decimal.Decimal
	Unit                     string
	Description              string // BT-154 (optional)
	OriginTradeCountry       string // BT-159 (optional) alpha-2 code ISO 3166-1 such as DE, US,...
	TaxTypeCode              string // muss VAT sein
	TaxCategoryCode          string
	TaxRateApplicablePercent decimal.Decimal
	Total                    decimal.Decimal
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

// TradeTax hängt an einer Rechnung und bezeichnet alle vorkommenden
// Steuersätze
type TradeTax struct {
	CalculatedAmount decimal.Decimal
	BasisAmount      decimal.Decimal
	Typ              string
	CategoryCode     string
	Percent          decimal.Decimal
	ExemptionReason  string
}

// Invoice ist das Hauptelement der e-Invoice-Datei
type Invoice struct {
	AllowanceTotal                      decimal.Decimal
	BuyerOrderReferencedDocument        string // BT-13
	DespatchAdviceReferencedDocument    string // BT-16
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
	OccurrenceDateTime                  time.Time
	LineTotal                           decimal.Decimal
	Notizen                             []Notiz
	InvoiceItems                        []InvoiceItem
	SchemaType                          CodeSchemaType
	InvoiceNumber                       string       // BT-1
	InvoiceTypeCode                     CodeDocument // BT-3
	TradeTaxes                          []TradeTax
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
