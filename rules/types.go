package rules

// Severity indicates the level of a validation issue.
// This is used to distinguish between hard requirements ("muss"/"must") and
// recommendations ("soll"/"should") in specifications like XRechnung.
type Severity int

const (
	// SeverityError indicates a hard requirement violation ("muss"/"must").
	// Validation fails when errors exist.
	SeverityError Severity = iota

	// SeverityWarning indicates a recommendation violation ("soll"/"should").
	// Validation succeeds but warnings are reported for user attention.
	SeverityWarning

	// SeverityInfo indicates an informational note.
	// Reserved for future use.
	SeverityInfo
)

// String returns the string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// Rule represents a business rule from the EN 16931 specification.
// Each rule contains the official code, related field identifiers, and description.
type Rule struct {
	Code        string   // EN 16931 rule code (e.g., "BR-01", "BR-S-08")
	Description string   // Official specification requirement text
	Fields      []string // BT-/BG- identifiers from semantic model
}

// SyntaxRule represents one schematron rule from the EN 16931 syntax-binding
// patterns (CII-SR-*/CII-DT-* for CII, UBL-SR-*/UBL-CR-*/UBL-DT-* for UBL).
// The Context is an XPath expression selecting the nodes the asserts apply to.
// Rules of one pattern follow schematron semantics: for each node, only the
// first rule whose context matches is evaluated ("first match wins").
type SyntaxRule struct {
	Context string         // XPath selecting the context nodes
	Asserts []SyntaxAssert // Checks evaluated relative to each context node
}

// SyntaxAssert is a single schematron assert of a SyntaxRule. The Test XPath
// is evaluated relative to a context node; an effective boolean value of
// false constitutes a violation (or warning, depending on Severity).
type SyntaxAssert struct {
	Code        string   // Official rule ID (e.g., "CII-DT-031")
	Test        string   // XPath expression that must evaluate to true
	Severity    Severity // SeverityError for flag="fatal", SeverityWarning for flag="warning"
	Description string   // Official assert text
}
