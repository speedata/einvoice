package einvoice

import (
	"fmt"
	"regexp"
)

// PEPPOL BIS Billing 3.0 Business Process Identifiers (BT-23)
//
// These constants represent valid ProfileID values for PEPPOL BIS Billing 3.0.
// The business process identifier specifies the business process context in which
// the transaction appears, enabling the buyer to process the invoice appropriately.
//
// Reference: PEPPOL-EN16931-R007 requires format: urn:fdc:peppol.eu:2017:poacc:billing:NN:1.0
// Documentation: https://docs.peppol.eu/poacc/billing/3.0/
const (
	// BPPEPPOLBilling01 is the standard PEPPOL BIS Billing 3.0 process (most common).
	// Use this for standard billing invoices and credit notes.
	BPPEPPOLBilling01 = "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0"
)

// PEPPOL BIS Billing 3.0 Specification Identifiers (BT-24)
//
// These constants represent valid CustomizationID values for PEPPOL BIS Billing 3.0.
// The specification identifier identifies the technical specification containing the
// total set of rules regarding semantic content, cardinalities, and business rules.
//
// Reference: PEPPOL-EN16931-R004 requires the exact value below for PEPPOL BIS Billing 3.0
// Documentation: https://docs.peppol.eu/poacc/billing/3.0/
const (
	// SpecPEPPOLBilling30 is the PEPPOL BIS Billing 3.0 specification identifier.
	// This is the required value per PEPPOL-EN16931-R004.
	SpecPEPPOLBilling30 = "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0"
)

// PEPPOL Electronic Address Scheme (EAS) Codes
//
// These constants represent commonly used Electronic Address Scheme identifiers
// for BT-34 (Seller electronic address) and BT-49 (Buyer electronic address).
//
// EAS codes identify the scheme that an electronic address identifier is based on.
// The complete list is maintained by the Digital European Programme and updated twice yearly.
//
// Usage: Set these as the scheme for Party.URIUniversalCommunicationScheme
//
// Reference: https://docs.peppol.eu/poacc/billing/3.0/codelist/eas/
// Full list: https://ec.europa.eu/digital-building-blocks (EAS Code List v13+)
const (
	// EAS0002 is SIRENE - French business registry system identifier.
	EAS0002 = "0002"

	// EAS0007 is Organisationsnummer - Swedish legal entity identifier.
	EAS0007 = "0007"

	// EAS0009 is SIRET-CODE - French establishment identifier.
	EAS0009 = "0009"

	// EAS0037 is LY-tunnus - Finnish organization identifier.
	EAS0037 = "0037"

	// EAS0060 is Data Universal Numbering System (D-U-N-S) Number.
	// Widely used international business identifier (9 digits).
	EAS0060 = "0060"

	// EAS0088 is EAN Location Code / Global Location Number (GLN).
	// Used for identifying parties and locations in supply chains (13 digits).
	EAS0088 = "0088"

	// EAS0096 is Danish CVR (Central Business Register) number.
	EAS0096 = "0096"

	// EAS0130 is Global Legal Entity Identifier (LEI).
	// ISO 17442 standard identifier for legal entities (20 alphanumeric characters).
	EAS0130 = "0130"

	// EAS0135 is Swiss Business Identification Number (UIDB).
	// Also known as Enterprise Identification Number (IDI/IDE/IDI).
	EAS0135 = "0135"

	// EAS0183 is the older code for Swiss Business Identification Number (UIDB).
	// Deprecated in favor of EAS0135, but still in use for compatibility.
	EAS0183 = "0183"

	// EAS0184 is Dutch Chamber of Commerce number (KvK).
	EAS0184 = "0184"

	// EAS0188 is Belgian Crossroad Bank of Enterprises (CBE/KBO) number.
	EAS0188 = "0188"

	// EAS0190 is Norwegian Organization Number (Organisasjonsnummer).
	EAS0190 = "0190"

	// EAS0192 is German Leitweg-ID for routing to German public entities.
	EAS0192 = "0192"

	// EAS0195 is Singapore Nationwide E-Invoice Framework (InvoiceNow) identifier.
	EAS0195 = "0195"

	// EAS0196 is Icelandic national identifier (Kennitala).
	EAS0196 = "0196"

	// EAS0198 is Irish VAT registration number.
	EAS0198 = "0198"

	// EAS0204 is Portuguese VAT registration number (NIPC).
	EAS0204 = "0204"

	// EAS0208 is Belgian VAT registration number.
	EAS0208 = "0208"

	// EAS0209 is Spanish VAT registration number (NIF).
	EAS0209 = "0209"

	// EAS0210 is Italian VAT registration number (Partita IVA).
	EAS0210 = "0210"

	// EAS0211 is Dutch VAT registration number.
	EAS0211 = "0211"

	// EAS0212 is Norwegian VAT registration number (MVA).
	EAS0212 = "0212"

	// EAS0213 is Swiss VAT registration number (MWST/TVA/IVA).
	EAS0213 = "0213"

	// EAS9901 is Danish NEMHANDELSSYSTEM identifier (NemHandel).
	EAS9901 = "9901"

	// EAS9906 is Italian Codice Fiscale (tax code for individuals and entities).
	EAS9906 = "9906"

	// EAS9907 is French CHORUS Pro identifier.
	EAS9907 = "9907"

	// EAS9910 is Hungarian VAT registration number.
	EAS9910 = "9910"

	// EAS9913 is Austrian Ergänzungsregister für sonstige Betroffene (ERsB) number.
	EAS9913 = "9913"

	// EAS9914 is Austrian Firmenregister (FN) - company register number.
	EAS9914 = "9914"

	// EAS9915 is Austrian Umsatzsteueridentifikationsnummer (UID) - VAT number.
	EAS9915 = "9915"

	// EAS9918 is Latvian VAT registration number.
	EAS9918 = "9918"

	// EAS9919 is Maltese VAT registration number.
	EAS9919 = "9919"

	// EAS9920 is Slovenian VAT registration number.
	EAS9920 = "9920"

	// EAS9922 is Croatian VAT registration number (OIB).
	EAS9922 = "9922"

	// EAS9923 is Luxembourgish VAT registration number.
	EAS9923 = "9923"

	// EAS9925 is United Kingdom VAT registration number.
	EAS9925 = "9925"

	// EAS9926 is Bosnia and Herzegovina VAT registration number.
	EAS9926 = "9926"

	// EAS9927 is EU VAT registration number (generic EU VAT).
	EAS9927 = "9927"

	// EAS9928 is Belgian Crossroad Bank of Social Security number (CBSS/KSZ).
	EAS9928 = "9928"

	// EAS9929 is French SIRENE - INSEE identifier (alternative to 0002).
	EAS9929 = "9929"

	// EAS9930 is German Umsatzsteuernummer (VAT number).
	EAS9930 = "9930"

	// EAS9931 is Estonian VAT registration number.
	EAS9931 = "9931"

	// EAS9933 is Finnish Organization Identifier (Y-tunnus).
	EAS9933 = "9933"

	// EAS9934 is Swedish Organization Number (Organisationsnummer).
	EAS9934 = "9934"

	// EAS9935 is Austrian Government Agency identifier (GovAgency).
	EAS9935 = "9935"

	// EAS9936 is Norwegian Public Sector identifier.
	EAS9936 = "9936"

	// EAS9937 is Dutch Overheid identifier (Government).
	EAS9937 = "9937"

	// EAS9938 is Welsh public sector identifier.
	EAS9938 = "9938"

	// EAS9939 is Cypriot VAT registration number.
	EAS9939 = "9939"

	// EAS9940 is Danish CVR-nummer (Central Business Register).
	EAS9940 = "9940"

	// EAS9941 is Italian Sistema di Interscambio (SDI) identifier.
	// Used for the Italian electronic invoicing system.
	EAS9941 = "9941"

	// EAS9942 is German Leitweg-ID (routing identifier for public entities).
	EAS9942 = "9942"

	// EAS9943 is Estonian e-Delivery identifier.
	EAS9943 = "9943"

	// EAS9944 is Netherlands PEPPOL identifier.
	EAS9944 = "9944"

	// EAS9945 is Polish REGON (National Business Registry Number).
	EAS9945 = "9945"

	// EAS9946 is Italian SFE (Sistema Fatturazione Elettronica) identifier.
	EAS9946 = "9946"

	// EAS9947 is Romanian VAT registration number.
	EAS9947 = "9947"

	// EAS9948 is German Steuernummer (tax number).
	EAS9948 = "9948"

	// EAS9949 is Greek VAT registration number.
	EAS9949 = "9949"

	// EAS9950 is Spanish Código de Identificación Fiscal (CIF).
	EAS9950 = "9950"

	// EAS9951 is Portuguese NIPC (Número de Identificação de Pessoa Coletiva).
	EAS9951 = "9951"

	// EAS9952 is Hungarian Adószám (Tax Number).
	EAS9952 = "9952"

	// EAS9953 is Bulgarian VAT registration number.
	EAS9953 = "9953"

	// EAS9955 is Czech VAT registration number (DIČ).
	EAS9955 = "9955"

	// EAS9956 is Latvian Pievienotās vērtības nodokļa (PVN) reģistrācijas numurs.
	EAS9956 = "9956"

	// EAS9957 is French SIRET number (alternative to 0009).
	EAS9957 = "9957"

	// EAS9958 is German Vergabenummer (procurement identifier).
	EAS9958 = "9958"
)

// ValidateBusinessProcessID validates that a business process identifier (BT-23)
// conforms to the PEPPOL-EN16931-R007 format requirement.
//
// The format must be: urn:fdc:peppol.eu:2017:poacc:billing:NN:1.0
// where NN is a two-digit process number (e.g., "01" for standard billing).
//
// Returns nil if valid, error otherwise.
func ValidateBusinessProcessID(id string) error {
	if id == "" {
		return fmt.Errorf("business process ID cannot be empty")
	}

	pattern := regexp.MustCompile(`^urn:fdc:peppol\.eu:2017:poacc:billing:\d{2}:1\.0$`)
	if !pattern.MatchString(id) {
		return fmt.Errorf("invalid business process format: expected 'urn:fdc:peppol.eu:2017:poacc:billing:NN:1.0', got '%s'", id)
	}

	return nil
}

// ValidatePEPPOLSpecificationID validates that a specification identifier (BT-24)
// matches the PEPPOL BIS Billing 3.0 requirement (PEPPOL-EN16931-R004).
//
// This is ONLY required if you need full PEPPOL BIS Billing 3.0 compliance.
// You can use PEPPOL business process (BT-23) with ANY specification identifier (BT-24).
//
// For full PEPPOL BIS Billing 3.0, the specification identifier must be exactly:
// urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0
//
// Returns nil if valid, error otherwise.
func ValidatePEPPOLSpecificationID(id string) error {
	if id == "" {
		return fmt.Errorf("specification identifier cannot be empty")
	}

	if id != SpecPEPPOLBilling30 {
		return fmt.Errorf("not a PEPPOL BIS Billing 3.0 specification identifier: expected '%s', got '%s'", SpecPEPPOLBilling30, id)
	}

	return nil
}

// ValidateEASCode validates that an Electronic Address Scheme code is in the
// correct 4-digit format (e.g., "0088", "9901").
//
// Note: This only validates the format, not whether the code is registered
// in the official EAS code list. For the complete list, see:
// https://docs.peppol.eu/poacc/billing/3.0/codelist/eas/
//
// Returns nil if valid, error otherwise.
func ValidateEASCode(code string) error {
	if code == "" {
		return fmt.Errorf("EAS code cannot be empty")
	}

	pattern := regexp.MustCompile(`^\d{4}$`)
	if !pattern.MatchString(code) {
		return fmt.Errorf("invalid EAS code format: expected 4-digit code (e.g., '0088'), got '%s'", code)
	}

	return nil
}

// UsesPEPPOLBusinessProcess checks if an invoice uses a PEPPOL business process (BT-23).
//
// This only checks the business process identifier format, NOT the specification.
// You can use PEPPOL business process with any specification (Factur-X, EN16931, etc.).
//
// Returns true if BT-23 is set to a valid PEPPOL business process URN.
func (inv *Invoice) UsesPEPPOLBusinessProcess() bool {
	return ValidateBusinessProcessID(inv.BPSpecifiedDocumentContextParameter) == nil
}
