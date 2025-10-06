package einvoice

import "fmt"

// ValidationError is returned when invoice validation fails.
// It contains all EN 16931 business rule violations found during validation.
//
// Example usage:
//
//	err := inv.Validate()
//	if err != nil {
//	    var valErr *ValidationError
//	    if errors.As(err, &valErr) {
//	        for _, v := range valErr.Violations() {
//	            fmt.Printf("Rule %s: %s\n", v.Rule, v.Text)
//	        }
//	    }
//	}
type ValidationError struct {
	violations []SemanticError
}

// Error implements the error interface.
// Returns a human-readable description of the validation failure.
func (e *ValidationError) Error() string {
	if len(e.violations) == 0 {
		return "validation failed with no violations"
	}

	if len(e.violations) == 1 {
		v := e.violations[0]
		return fmt.Sprintf("validation failed: %s - %s", v.Rule, v.Text)
	}

	return fmt.Sprintf("validation failed with %d violations (first: %s - %s)",
		len(e.violations),
		e.violations[0].Rule,
		e.violations[0].Text)
}

// Violations returns a copy of all validation violations.
// This ensures the internal violations slice cannot be modified externally.
func (e *ValidationError) Violations() []SemanticError {
	if e.violations == nil {
		return nil
	}

	// Return a copy to prevent external modification
	violations := make([]SemanticError, len(e.violations))
	copy(violations, e.violations)
	return violations
}

// Count returns the number of validation violations.
func (e *ValidationError) Count() int {
	return len(e.violations)
}

// HasRule checks if a specific business rule violation exists.
// The rule parameter should be a business rule identifier like "BR-1", "BR-S-8", etc.
func (e *ValidationError) HasRule(rule string) bool {
	for _, v := range e.violations {
		if v.Rule == rule {
			return true
		}
	}
	return false
}
