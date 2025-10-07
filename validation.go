package einvoice

import (
	"fmt"

	"github.com/speedata/einvoice/rules"
)

// SemanticError contains a business rule violation found during validation.
type SemanticError struct {
	Rule rules.Rule // The business rule that was violated
	Text string     // Human-readable description with actual values
}

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
//	            fmt.Printf("Rule %s: %s\n", v.Rule.Code, v.Text)
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
		return fmt.Sprintf("validation failed: %s - %s", v.Rule.Code, v.Text)
	}

	return fmt.Sprintf("validation failed with %d violations (first: %s - %s)",
		len(e.violations),
		e.violations[0].Rule.Code,
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
// Accepts a Rule constant (e.g., rules.BR1, rules.BRS8, rules.BRCO14).
//
// Example:
//
//	if valErr.HasRule(rules.BR1) {
//	    // Handle missing specification identifier
//	}
func (e *ValidationError) HasRule(rule rules.Rule) bool {
	for _, v := range e.violations {
		if v.Rule.Code == rule.Code {
			return true
		}
	}
	return false
}

// HasRuleCode checks if a specific business rule code violation exists.
// The code parameter should be a business rule identifier string like "BR-1", "BR-S-8", etc.
//
// Example:
//
//	if valErr.HasRuleCode("BR-1") {
//	    // Handle missing specification identifier
//	}
func (e *ValidationError) HasRuleCode(code string) bool {
	for _, v := range e.violations {
		if v.Rule.Code == code {
			return true
		}
	}
	return false
}

// addViolation is a helper method that appends a business rule violation to the invoice.
// It is used internally by validation methods to record rule violations in a type-safe way.
//
// Example:
//
//	inv.addViolation(rules.BRCO14, fmt.Sprintf(
//	    "Invoice total VAT amount %s does not match sum %s",
//	    inv.TaxTotal.String(), calculatedTaxTotal.String()))
func (inv *Invoice) addViolation(rule rules.Rule, text string) {
	inv.violations = append(inv.violations, SemanticError{
		Rule: rule,
		Text: text,
	})
}

// Validate checks the invoice against EN 16931 business rules.
// It clears any previous violations and performs a fresh validation.
// Returns a ValidationError if violations exist, nil if invoice is valid.
//
// This method should be called:
// - After building an invoice programmatically
// - After modifying invoice data (e.g., after UpdateTotals)
// - Before writing XML to ensure compliance
//
// Example:
//
//	err := inv.Validate()
//	if err != nil {
//	    var valErr *ValidationError
//	    if errors.As(err, &valErr) {
//	        for _, v := range valErr.Violations() {
//	            fmt.Printf("Rule %s: %s\n", v.Rule.Code, v.Text)
//	        }
//	    }
//	    return err
//	}
func (inv *Invoice) Validate() error {
	// Always clear previous violations to ensure idempotency
	inv.violations = []SemanticError{}

	// Run all validation checks
	inv.checkBR()
	inv.checkBRO()
	inv.checkOther()

	// Return error if violations exist
	if len(inv.violations) > 0 {
		return &ValidationError{violations: inv.violations}
	}

	return nil
}
