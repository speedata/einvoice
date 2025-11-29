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
	Fields      []string // BT-/BG- identifiers from semantic model
	Description string   // Official specification requirement text
}
