package rules

// Rule represents a business rule from the EN 16931 specification.
// Each rule contains the official code, related field identifiers, and description.
type Rule struct {
	Code        string   // EN 16931 rule code (e.g., "BR-01", "BR-S-08")
	Fields      []string // BT-/BG- identifiers from semantic model
	Description string   // Official specification requirement text
}
