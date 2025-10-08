package einvoice

// Profile Specification Identifiers (BT-24 / GuidelineSpecifiedDocumentContextParameter)
//
// These constants represent the standard profile URNs for electronic invoicing in Europe.
// The specification identifier (BT-24) indicates which semantic and syntax rules apply
// to the invoice, determining the required and optional fields.
//
// Profile Hierarchy (from most basic to most comprehensive):
//   1. Minimum - Absolute minimum required data
//   2. Basic WL (Without Lines) - No line item details required
//   3. Basic - Basic invoicing with line items
//   4. EN 16931 - European standard, PEPPOL, XRechnung (all EN 16931 compliant)
//   5. Extended - Full feature set including complex scenarios
//
// References:
//   - Factur-X: https://fnfe-mpe.org/factur-x/
//   - ZUGFeRD: https://www.ferd-net.de/standards/zugferd-2.3/index.html
//   - EN 16931: https://ec.europa.eu/digital-building-blocks/sites/display/DIGITAL/Electronic+Invoicing
//   - XRechnung: https://xeinkauf.de/xrechnung/
//
// Usage:
//   invoice.GuidelineSpecifiedDocumentContextParameter = SpecFacturXBasic
//   if invoice.GuidelineSpecifiedDocumentContextParameter == SpecEN16931 { ... }

// Factur-X Specification Identifiers
//
// Factur-X is the French/European standard for electronic invoicing, developed jointly
// by French and German organizations. It's based on the Cross Industry Invoice (CII)
// format and provides multiple profile levels.
const (
	// SpecFacturXMinimum is the Factur-X Minimum profile (lowest level).
	// Use for basic invoice data without line details.
	// Version: 1.0.7
	SpecFacturXMinimum = "urn:factur-x.eu:1p0:minimum"

	// SpecFacturXBasicWL is the Factur-X Basic WL (Without Lines) profile.
	// Use when invoice totals are needed but line item details are not required.
	// Version: 1.0.7
	SpecFacturXBasicWL = "urn:factur-x.eu:1p0:basicwl"

	// SpecFacturXBasic is the Factur-X Basic profile (EN 16931 subset).
	// This is EN 16931 compliant and includes essential invoice line items.
	// Version: 1.0.7
	SpecFacturXBasic = "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic"

	// SpecFacturXBasicAlt is an alternative format for Factur-X Basic profile.
	// Uses colons instead of hash symbols. Some systems accept this variant.
	// Version: 1.0
	SpecFacturXBasicAlt = "urn:cen.eu:en16931:2017:compliant:factur-x.eu:1p0:basic"

	// SpecFacturXExtended is the Factur-X Extended profile (most comprehensive).
	// This is EN 16931 conformant (superset) and includes advanced features.
	// Use for complex invoicing scenarios with all optional data.
	// Version: 1.0.7
	SpecFacturXExtended = "urn:cen.eu:en16931:2017#conformant#urn:factur-x.eu:1p0:extended"
)

// ZUGFeRD Specification Identifiers
//
// ZUGFeRD is the German standard for electronic invoicing. Version 2.0 and later
// are aligned with Factur-X. ZUGFeRD 2.x is technically equivalent to Factur-X 1.0.
const (
	// SpecZUGFeRDMinimum is the ZUGFeRD 2.0 Minimum profile.
	// Equivalent to Factur-X Minimum. German variant.
	// Version: 2.0+
	SpecZUGFeRDMinimum = "urn:zugferd.de:2p0:minimum"

	// SpecZUGFeRDBasic is the ZUGFeRD 2.0 Basic profile (EN 16931 subset).
	// Equivalent to Factur-X Basic. German variant.
	// Version: 2.0+
	SpecZUGFeRDBasic = "urn:cen.eu:en16931:2017#compliant#urn:zugferd.de:2p0:basic"

	// SpecZUGFeRDExtended is the ZUGFeRD 2.0 Extended profile.
	// Equivalent to Factur-X Extended. German variant.
	// Version: 2.0+
	SpecZUGFeRDExtended = "urn:cen.eu:en16931:2017#conformant#urn:zugferd.de:2p0:extended"
)

// EN 16931 Specification Identifier
//
// EN 16931 is the European standard for electronic invoicing, mandated by the
// European Directive 2014/55/EU for invoices to public sector buyers.
const (
	// SpecEN16931 is the pure EN 16931:2017 specification.
	// This is the base European standard without any country or domain extensions.
	// Use for standard EN 16931 compliant invoices.
	SpecEN16931 = "urn:cen.eu:en16931:2017"
)

// XRechnung Specification Identifier
//
// XRechnung is the German implementation of EN 16931, required for invoices to
// German public sector entities. It adds German-specific business rules and
// extensions to the base EN 16931 standard.
const (
	// SpecXRechnung30 is the XRechnung 3.0 specification identifier.
	// Required for German public sector invoicing (B2G).
	// This is EN 16931 compliant with German extensions.
	// Version: 3.0
	SpecXRechnung30 = "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0"
)

// IsProfileURN checks if the given string is a recognized profile specification URN.
// This helper function can validate user input or configuration values.
//
// Returns true if the URN matches any known profile specification identifier.
func IsProfileURN(urn string) bool {
	switch urn {
	case SpecFacturXMinimum, SpecFacturXBasicWL, SpecFacturXBasic, SpecFacturXBasicAlt, SpecFacturXExtended,
		SpecZUGFeRDMinimum, SpecZUGFeRDBasic, SpecZUGFeRDExtended,
		SpecEN16931,
		SpecXRechnung30:
		return true
	default:
		return false
	}
}

// GetProfileName returns a human-readable name for a profile URN.
// This is useful for display purposes and user interfaces.
//
// Returns the profile name (e.g., "Factur-X Basic", "XRechnung 3.0") or
// "Unknown" if the URN is not recognized.
func GetProfileName(urn string) string {
	switch urn {
	case SpecFacturXMinimum:
		return "Factur-X Minimum"
	case SpecFacturXBasicWL:
		return "Factur-X Basic WL"
	case SpecFacturXBasic, SpecFacturXBasicAlt:
		return "Factur-X Basic"
	case SpecFacturXExtended:
		return "Factur-X Extended"
	case SpecZUGFeRDMinimum:
		return "ZUGFeRD Minimum"
	case SpecZUGFeRDBasic:
		return "ZUGFeRD Basic"
	case SpecZUGFeRDExtended:
		return "ZUGFeRD Extended"
	case SpecEN16931:
		return "EN 16931"
	case SpecXRechnung30:
		return "XRechnung 3.0"
	default:
		return "Unknown"
	}
}
