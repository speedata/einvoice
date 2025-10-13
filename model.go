package einvoice

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type (
	// CodeSchemaType represents the type of the invoice (CII or UBL).
	CodeSchemaType int
	// CodeProfileType represents the CII subtype (extended, minimum, ...)
	CodeProfileType int
	// CodeDocument contains the UNTDID 1001 document code.
	CodeDocument int
	// CodePartyType distinguishes between seller, buyer, ..
	CodePartyType int
)

// Don't change the order. extended > EN16931 > basic > basicwl > minimum.
const (
	// CProfileUnknown is the unknown profile, zero value.
	CProfileUnknown CodeProfileType = iota
	// CProfileMinimum urn:factur-x.eu:1p0:minimum.
	CProfileMinimum
	// CProfileBasicWL urn:factur-x.eu:1p0:basicwl.
	CProfileBasicWL
	// CProfileBasic urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic.
	CProfileBasic
	// CProfileEN16931 (previously Comfort) represents urn:cen.eu:en16931:2017.
	CProfileEN16931
	// CProfileExtended represents the urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended schema.
	CProfileExtended
	// CProfileXRechnung represents an XRechnung invoice.
	CProfileXRechnung
)

func (cp CodeProfileType) String() string {
	switch cp {
	case CProfileUnknown:
		return "unknown profile"
	case CProfileXRechnung:
		return "XRechnung"
	case CProfileExtended:
		return "extended"
	case CProfileEN16931:
		return "EN 16931"
	case CProfileBasic:
		return "basic"
	case CProfileBasicWL:
		return "basic without lines"
	case CProfileMinimum:
		return "minimum"
	default:
		return "unknown"
	}
}

// ToProfileName returns the identifier for this profile such as
// urn:cen.eu:en16931:2017.
func (cp CodeProfileType) ToProfileName() string {
	switch cp {
	case CProfileUnknown:
		return "Unknown"
	case CProfileXRechnung:
		return SpecXRechnung30
	case CProfileExtended:
		return SpecFacturXExtended
	case CProfileEN16931:
		return SpecEN16931
	case CProfileBasic:
		return SpecFacturXBasic
	case CProfileBasicWL:
		return SpecFacturXBasicWL
	case CProfileMinimum:
		return SpecFacturXMinimum
	default:
		return "unknown"
	}
}

func (cp CodeSchemaType) String() string {
	switch cp {
	case CII:
		return "ZUGFeRD/Factur-X"
	case UBL:
		return "UBL"
	case SchemaTypeUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

func (cd CodeDocument) String() string {
	return fmt.Sprintf("%d", cd)
}

// CodeSchemaType is the main XML flavor. Currently only CII is supported.
const (
	SchemaTypeUnknown CodeSchemaType = iota // Zero value for programmatically created invoices
	CII                                     // ZUGFeRD/Factur-X (Cross Industry Invoice)
	UBL                                     // Universal Business Language
)

// CodePartyType represents the type of the party.
const (
	CUnknownParty CodePartyType = iota
	CSellerParty
	CBuyerParty
	CShipToParty
	CPayeeParty
)

// Note contains text and the subject code.
type Note struct {
	Text        string // BT-22
	SubjectCode string // BT-21
}

func (n Note) String() string {
	return fmt.Sprintf("Notiz %s - %q", n.SubjectCode, n.Text)
}

// GlobalID stores a ISO/EIC 6523 encoded ID.
type GlobalID struct {
	ID     string
	Scheme string
}

// A PostalAddress belongs to the seller, buyer and some other entities.
type PostalAddress struct {
	CountryID              string
	PostcodeCode           string // BT-38, BT-53, BT-67, BT-78
	Line1                  string // BT-35, BT-50, BT-64, BT-75
	Line2                  string // BT-36, BT-51, BT-65, BT-76
	Line3                  string // BT-162, BT-163, BT-164, BT-165
	City                   string // BT-37, BT-52, BT-66, BT-77
	CountrySubDivisionName string // BT-39, BT-54, BT-68, BT-79
}

// SpecifiedLegalOrganization represents a division BT-30, BT-47, BT-61.
type SpecifiedLegalOrganization struct {
	ID                  string //  BT-30, BT-47, BT-61
	Scheme              string // BT-30, BT-61
	TradingBusinessName string // BT-28, BT-45
	// PostalAddress       *PostalAddress // BG-X-59
}

// DefinedTradeContact represents a person. BG-6, BG-9.
type DefinedTradeContact struct {
	PersonName     string // BT-41, BT-56
	DepartmentName string // BT-41, BT-56
	EMail          string // BT-43, BT-58
	PhoneNumber    string // BT-44, BT-57
	// TypeCode string // BT-X-317
}

// Party represents buyer and seller.
type Party struct {
	ID                              []string                    // BT-29, BT-46, BT-60, BT-71
	GlobalID                        []GlobalID                  // BT-29, BT-64, BT-60, BT-71
	Name                            string                      // BT-27, BT-44, BT-59, BT-62, BT-70
	DefinedTradeContact             []DefinedTradeContact       // BG-9
	Description                     string                      // BT-33
	URIUniversalCommunication       string                      // BT-34, BT-49
	URIUniversalCommunicationScheme string                      // BT-34, BT-49
	PostalAddress                   *PostalAddress              // BG-5, BG-8
	SpecifiedLegalOrganization      *SpecifiedLegalOrganization // BT-30, BT-47, BT-61
	VATaxRegistration               string                      // BT-31, BT-48, BT-63
	FCTaxRegistration               string                      // BT-32
}

// Characteristic add details to a product, BG-32.
type Characteristic struct {
	Description string // BT-160
	Value       string // BT-161
}

// Classification specifies a product classification, BT-158.
type Classification struct {
	ClassCode     string
	ListID        string
	ListVersionID string
}

// InvoiceLine represents one position of items.
type InvoiceLine struct {
	LineID                                    string            // BT-126
	ArticleNumber                             string            // BT-155 seller assigned ID
	ArticleNumberBuyer                        string            // BT-156 buyer assigned ID
	ItemName                                  string            // BT-153
	AdditionalReferencedDocumentID            string            // BT-128
	AdditionalReferencedDocumentTypeCode      string            // BT-128
	AdditionalReferencedDocumentRefTypeCode   string            // BT-128
	BillingSpecifiedPeriodStart               time.Time         // BT-134
	BillingSpecifiedPeriodEnd                 time.Time         // BT-135
	BuyerOrderReferencedDocument              string            // BT-132
	Note                                      string            // BT-127
	GlobalID                                  string            // BT-157
	GlobalIDType                              string            // BT-157
	Characteristics                           []Characteristic  // BG-32
	ProductClassification                     []Classification  // BT-158, UNTDID 7143
	Description                               string            // BT-154 (optional)
	OriginTradeCountry                        string            // BT-159 (optional) alpha-2 code ISO 3166-1 such as DE, US,...
	ReceivableSpecifiedTradeAccountingAccount string            // BT-133
	GrossPrice                                decimal.Decimal   // BT-148
	BasisQuantity                             decimal.Decimal   // BT-149
	BasisQuantityUnit                         string            // BT-149
	InvoiceLineAllowances                     []AllowanceCharge // BG-27
	InvoiceLineCharges                        []AllowanceCharge // BG-28
	AppliedTradeAllowanceCharge               []AllowanceCharge // BT-147
	NetPrice                                  decimal.Decimal   // BT-146
	NetBilledQuantity                         decimal.Decimal   // BT-149
	NetBilledQuantityUnit                     string            // BT-150
	BilledQuantity                            decimal.Decimal   // BT-129
	BilledQuantityUnit                        string            // BT-130
	TaxTypeCode                               string            // BT-151 must be VAT
	TaxCategoryCode                           string            // BT-151
	TaxRateApplicablePercent                  decimal.Decimal   // BT-152
	Total                                     decimal.Decimal   // BT-131

	// Private fields for tracking XML element presence (BR-24, BR-26)
	// These are set during parsing to distinguish between missing elements and zero values
	hasLineTotalInXML bool
	hasNetPriceInXML  bool
}

// PaymentMeans represents a payment means.
type PaymentMeans struct {
	TypeCode                                             int    // BT-81
	Information                                          string // BT-82
	PayeePartyCreditorFinancialAccountIBAN               string // BT-84
	PayeePartyCreditorFinancialAccountName               string // BT-85
	PayeePartyCreditorFinancialAccountProprietaryID      string // BT-84
	PayeeSpecifiedCreditorFinancialInstitutionBIC        string // BT-86
	PayerPartyDebtorFinancialAccountIBAN                 string // BT-91
	ApplicableTradeSettlementFinancialCardID             string // BT-87
	ApplicableTradeSettlementFinancialCardCardholderName string // BT-88

	// Private fields for tracking XML element presence (BR-61)
	// Set during parsing to distinguish between missing elements and empty elements.
	// Per EN 16931 schematron, BR-61 test is "(ram:IBANID) or (ram:ProprietaryID)" which
	// checks for element PRESENCE, not value. An empty element <ram:IBANID/> satisfies
	// the test because the element exists.
	hasPayeeAccountInXML      bool // Set when PayeePartyCreditorFinancialAccount element exists
	hasPayeeIBANInXML         bool // Set when IBANID element exists (even if empty)
	hasPayeeProprietaryIDInXML bool // Set when ProprietaryID element exists (even if empty)
}

// AllowanceCharge specifies charges and deductions.
type AllowanceCharge struct {
	ChargeIndicator                       bool            // BG-20, BG-21, BG-27, BG-28
	CalculationPercent                    decimal.Decimal // BT-94, BT-101, BT-138, BT-143
	BasisAmount                           decimal.Decimal // BT-93, BT-100, BT-137, BT-142
	ActualAmount                          decimal.Decimal // BT-92, BT-99, BT-136, BT-141
	ReasonCode                            int             // BT-98, BT-105, BT-140, BT-145
	Reason                                string          // BT-97, BT-104, BT-139, BT-144
	CategoryTradeTaxType                  string          // BT-95, BT-102
	CategoryTradeTaxCategoryCode          string          // BT-95, BT-102
	CategoryTradeTaxRateApplicablePercent decimal.Decimal // BT-96, BT-103
}

// TradeTax is the VAT breakdown for each percentage.
type TradeTax struct {
	CalculatedAmount    decimal.Decimal // BT-117
	BasisAmount         decimal.Decimal // BT-116
	TypeCode            string          // BT-118-0
	CategoryCode        string          // BT-118
	Percent             decimal.Decimal // BT-119
	ExemptionReason     string          // BT-120
	ExemptionReasonCode string          // BT-121
	TaxPointDate        time.Time       // BT-7
	DueDateTypeCode     string          // BT-8
}

func (tt TradeTax) String() string {
	var builder strings.Builder

	builder.WriteString(tt.BasisAmount.StringFixed(2))
	builder.WriteString(" + ")
	builder.WriteString(formatPercent(tt.Percent))
	builder.WriteString(" = ")
	builder.WriteString(tt.CalculatedAmount.StringFixed(2))
	builder.WriteString(", category code ")
	builder.WriteString(tt.CategoryCode)

	if tt.ExemptionReason != "" {
		builder.WriteString(tt.ExemptionReason)
	}

	return builder.String()
}

// Document contains a reference to a document or the document itself.
type Document struct {
	IssuerAssignedID       string // BT-17, BT-18, BT-122
	URIID                  string // BT-18, BT-124
	TypeCode               string // BT-17: 50, BT-18: 130  BT-122: 916
	ReferenceTypeCode      string // BT-18
	Name                   string // BT-123
	AttachmentBinaryObject []byte // BT-125
	AttachmentMimeCode     string // BT-125
	AttachmentFilename     string // BT-125
}

// SpecifiedTradePaymentTerms is unbounded in extended.
type SpecifiedTradePaymentTerms struct {
	Description          string    // BT-20
	DueDate              time.Time // BT-9
	DirectDebitMandateID string    // BT-89

}

// ReferencedDocument links to a previous invoice BG-3.
type ReferencedDocument struct {
	ID   string    // BT-25
	Date time.Time // BT-26

}

// Invoice is the main element of the e-invoice.
type Invoice struct {
	GuidelineSpecifiedDocumentContextParameter string                       // BT-24 (Specification identifier URN)
	DespatchAdviceReferencedDocument          string                       // BT-16
	ReceivingAdviceReferencedDocument         string                       // BT-15
	BuyerReference                            string                       // BT-10
	BPSpecifiedDocumentContextParameter       string                       // BT-23
	PayeeTradeParty                           *Party                       // BG-10
	PaymentMeans                              []PaymentMeans               // BG-16
	BillingSpecifiedPeriodStart               time.Time                    // BT-73
	BillingSpecifiedPeriodEnd                 time.Time                    // BT-74
	InvoiceDate                               time.Time                    // BT-2
	CreditorReferenceID                       string                       // BT-90
	PaymentReference                          string                       // BT-83
	TaxCurrencyCode                           string                       // BT-6
	InvoiceCurrencyCode                       string                       // BT-5
	LineTotal                                 decimal.Decimal              // BT-106
	AllowanceTotal                            decimal.Decimal              // BT-107
	ChargeTotal                               decimal.Decimal              // BT-108
	TaxBasisTotal                             decimal.Decimal              // BT-109
	TaxTotalCurrency                          string                       // BT-110
	TaxTotal                                  decimal.Decimal              // BT-110
	TaxTotalAccountingCurrency                string                       // BT-111
	TaxTotalAccounting                        decimal.Decimal              // BT-111
	GrandTotal                                decimal.Decimal              // BT-112
	TotalPrepaid                              decimal.Decimal              // BT-113
	RoundingAmount                            decimal.Decimal              // BT-114
	DuePayableAmount                          decimal.Decimal              // BT-115
	Buyer                                     Party                        // BG-7
	SellerTaxRepresentativeTradeParty         *Party                       // BG-11
	SellerOrderReferencedDocument             string                       // BT-14
	BuyerOrderReferencedDocument              string                       // BT-13
	ContractReferencedDocument                string                       // BT-12
	AdditionalReferencedDocument              []Document                   // BG-24
	SpecifiedProcuringProjectID               string                       // BT-11
	SpecifiedProcuringProjectName             string                       // BT-11
	Seller                                    Party                        // BG-4
	OccurrenceDateTime                        time.Time                    // BT-72
	Notes                                     []Note                       // BG-1
	InvoiceLines                              []InvoiceLine                // BG-25
	InvoiceNumber                             string                       // BT-1
	InvoiceTypeCode                           CodeDocument                 // BT-3
	TradeTaxes                                []TradeTax                   // BG-23
	SpecifiedTradeAllowanceCharge             []AllowanceCharge            // BG-20, BG-21
	ShipTo                                    *Party                       // BG-13
	SpecifiedTradePaymentTerms                []SpecifiedTradePaymentTerms // BT-20
	SchemaType                                CodeSchemaType               // UBL or CII
	InvoiceReferencedDocument                 []ReferencedDocument         // BG-3
	ReceivableSpecifiedTradeAccountingAccount string                       // BT-19

	// Private fields for tracking XML element presence (BR-12 through BR-15)
	// These are set during parsing to distinguish between missing elements and zero values
	hasLineTotalInXML        bool
	hasTaxBasisTotalInXML    bool
	hasGrandTotalInXML       bool
	hasDuePayableAmountInXML bool

	// Private field for tracking unexpected TaxTotalAmount currencies during parsing
	unexpectedTaxCurrencies []string

	violations []SemanticError // Private field - use Validate() and check error instead
}

// Profile helper methods for Invoice
// These methods check the GuidelineSpecifiedDocumentContextParameter (BT-24) URN
// to determine the invoice profile level.

// IsMinimum checks if the invoice uses the Minimum profile.
// URN: urn:factur-x.eu:1p0:minimum or urn:zugferd.de:2p0:minimum
func (inv *Invoice) IsMinimum() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == SpecFacturXMinimum ||
		inv.GuidelineSpecifiedDocumentContextParameter == SpecZUGFeRDMinimum
}

// IsBasicWL checks if the invoice uses the Basic WL (Without Lines) profile.
// URN: urn:factur-x.eu:1p0:basicwl
func (inv *Invoice) IsBasicWL() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == SpecFacturXBasicWL
}

// IsBasic checks if the invoice uses the Basic profile.
// URN: urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic
func (inv *Invoice) IsBasic() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == SpecFacturXBasic ||
		inv.GuidelineSpecifiedDocumentContextParameter == SpecZUGFeRDBasic ||
		inv.GuidelineSpecifiedDocumentContextParameter == SpecFacturXBasicAlt
}

// IsEN16931 checks if the invoice uses the EN 16931 profile.
// URN: urn:cen.eu:en16931:2017
func (inv *Invoice) IsEN16931() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == SpecEN16931
}

// IsExtended checks if the invoice uses the Extended profile.
// URN: urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended
func (inv *Invoice) IsExtended() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == SpecFacturXExtended ||
		inv.GuidelineSpecifiedDocumentContextParameter == SpecZUGFeRDExtended
}

// IsXRechnung checks if the invoice uses the XRechnung profile.
// Supports XRechnung 2.0, 2.1, 2.2, 2.3, and 3.0.
// URN examples:
//   - urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0
//   - urn:cen.eu:en16931:2017#compliant#urn:xoev-de:kosit:standard:xrechnung_2.1
func (inv *Invoice) IsXRechnung() bool {
	urn := inv.GuidelineSpecifiedDocumentContextParameter
	return urn == SpecXRechnung20 ||
		urn == SpecXRechnung21 ||
		urn == SpecXRechnung22 ||
		urn == SpecXRechnung23 ||
		urn == SpecXRechnung30
}

// ProfileLevel returns an integer representing the profile hierarchy level.
// Higher numbers indicate more inclusive profiles.
// This is used by the writer to determine which fields to include.
//
// Levels: 0=Unknown, 1=Minimum, 2=BasicWL, 3=Basic, 4=EN16931/PEPPOL/XRechnung, 5=Extended
func (inv *Invoice) ProfileLevel() int {
	if inv.IsExtended() {
		return 5
	}
	if inv.IsXRechnung() || inv.isPEPPOL() || inv.IsEN16931() {
		return 4
	}
	if inv.IsBasic() {
		return 3
	}
	if inv.IsBasicWL() {
		return 2
	}
	if inv.IsMinimum() {
		return 1
	}
	return 0 // Unknown
}

// MeetsProfileLevel checks if the invoice meets or exceeds a minimum profile level.
// This replaces the old Profile enum comparison (inv.Profile >= targetProfile).
//
// Example: inv.MeetsProfileLevel(3) returns true if profile is Basic, EN16931, Extended, or XRechnung
func (inv *Invoice) MeetsProfileLevel(minLevel int) bool {
	return inv.ProfileLevel() >= minLevel
}
