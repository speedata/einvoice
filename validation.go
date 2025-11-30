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
// It contains all EN 16931 business rule violations found during validation,
// as well as any warnings (recommendation violations).
//
// Example usage:
//
//	err := inv.Validate()
//	if err != nil {
//	    var valErr *ValidationError
//	    if errors.As(err, &valErr) {
//	        for _, v := range valErr.Violations() {
//	            fmt.Printf("Error: %s - %s\n", v.Rule.Code, v.Text)
//	        }
//	        for _, w := range valErr.Warnings() {
//	            fmt.Printf("Warning: %s - %s\n", w.Rule.Code, w.Text)
//	        }
//	    }
//	}
type ValidationError struct {
	violations []SemanticError
	warnings   []SemanticError
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

// Count returns the number of validation violations (errors).
func (e *ValidationError) Count() int {
	return len(e.violations)
}

// Warnings returns a copy of all validation warnings.
// Warnings are recommendation violations ("soll"/"should") that don't cause
// validation to fail but are reported for user attention.
// This ensures the internal warnings slice cannot be modified externally.
func (e *ValidationError) Warnings() []SemanticError {
	if e.warnings == nil {
		return nil
	}

	// Return a copy to prevent external modification
	warnings := make([]SemanticError, len(e.warnings))
	copy(warnings, e.warnings)
	return warnings
}

// WarningCount returns the number of validation warnings.
func (e *ValidationError) WarningCount() int {
	return len(e.warnings)
}

// HasWarnings returns true if there are any warnings.
func (e *ValidationError) HasWarnings() bool {
	return len(e.warnings) > 0
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

// addWarning is a helper method that appends a recommendation violation (warning) to the invoice.
// Warnings are for "soll"/"should" requirements that don't cause validation to fail
// but should be reported to the user for attention.
//
// Example:
//
//	inv.addWarning(rules.BRDE17, "Invoice type code should be within expected range")
//
//nolint:unused // Reserved for future warning-level validation rules
func (inv *Invoice) addWarning(rule rules.Rule, text string) {
	inv.warnings = append(inv.warnings, SemanticError{
		Rule: rule,
		Text: text,
	})
}

// Warnings returns a copy of all validation warnings found during the last Validate() call.
// Warnings are recommendation violations ("soll"/"should") that don't cause validation
// to fail but are reported for user attention.
//
// This method should be called after Validate() to check for recommendations even when
// validation passes (returns nil).
//
// Example:
//
//	err := inv.Validate()
//	if err == nil {
//	    // Validation passed, but check for recommendations
//	    for _, w := range inv.Warnings() {
//	        fmt.Printf("Recommendation: %s - %s\n", w.Rule.Code, w.Text)
//	    }
//	}
func (inv *Invoice) Warnings() []SemanticError {
	if inv.warnings == nil {
		return nil
	}

	// Return a copy to prevent external modification
	warnings := make([]SemanticError, len(inv.warnings))
	copy(warnings, inv.warnings)
	return warnings
}

// HasWarnings returns true if there are any warnings from the last Validate() call.
func (inv *Invoice) HasWarnings() bool {
	return len(inv.warnings) > 0
}

// Validate checks the invoice against applicable business rules with intelligent auto-detection.
//
// The method automatically detects which validation rules to apply based on:
// - Specification identifier (BT-24) for PEPPOL BIS Billing 3.0 detection
// - Seller country for country-specific rules (DK, IT, NL, NO, SE)
//
// All invoices are validated against EN 16931 core rules. Additional rules are applied
// automatically when the invoice metadata indicates they are required.
//
// This method clears any previous violations and warnings, then performs a fresh validation.
// Returns a ValidationError if violations (errors) exist, nil if invoice is valid.
// Warnings do NOT cause validation to fail - use Invoice.Warnings() or
// ValidationError.Warnings() to access them.
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
//	            fmt.Printf("Error: %s - %s\n", v.Rule.Code, v.Text)
//	        }
//	        for _, w := range valErr.Warnings() {
//	            fmt.Printf("Warning: %s - %s\n", w.Rule.Code, w.Text)
//	        }
//	    }
//	    return err
//	} else {
//	    // Validation passed, but check for recommendations
//	    for _, w := range inv.Warnings() {
//	        fmt.Printf("Recommendation: %s - %s\n", w.Rule.Code, w.Text)
//	    }
//	}
func (inv *Invoice) Validate() error {
	// Always clear previous violations and warnings to ensure idempotency
	inv.violations = []SemanticError{}
	inv.warnings = []SemanticError{}

	// Determine if we should validate:
	// - For parsed invoices (CII/UBL): Only validate if they claim EN 16931 compliance via BT-24
	// - For programmatically built invoices (SchemaTypeUnknown): Always validate
	shouldValidate := inv.SchemaType == SchemaTypeUnknown || inv.isEN16931Compliant()

	if shouldValidate {
		inv.validateCore()
		inv.validateCalculations()
		inv.validateDecimals()

		// Auto-detect and run PEPPOL validation based on specification identifier
		if inv.isPEPPOL() {
			inv.validatePEPPOL()
		}

		// Auto-detect country-specific rules
		// BR-DE-1 through BR-DE-31: Only for XRechnung invoices
		if inv.isGerman() {
			inv.validateGerman()
		}

		// TODO: Implement additional country-specific validation rules for:
		//   - Denmark (isDanish)
		//   - Italy (isItalian)
		//   - Netherlands (isDutch)
		//   - Norway (isNorwegian)
		//   - Sweden (isSwedish)
	}

	// Return error if violations exist (include warnings for convenience)
	if len(inv.violations) > 0 {
		return &ValidationError{
			violations: inv.violations,
			warnings:   inv.warnings,
		}
	}

	// No violations = success, even if warnings exist
	// User can call inv.Warnings() to check for recommendations
	return nil
}

// isEN16931Compliant checks if the invoice claims to be EN 16931 compliant
// based on the specification identifier (BT-24).
//
// Returns true if the CustomizationID (BT-24) contains any of the following:
// - "en16931" (EN 16931 core, PEPPOL, XRechnung, etc.)
// - "factur-x" (Factur-X/ZUGFeRD profiles)
// - "zugferd" (ZUGFeRD profiles)
//
// Pure UBL 2.1 or CII documents without BT-24 are NOT EN 16931 compliant and
// will not be validated against EN 16931 business rules.
func (inv *Invoice) isEN16931Compliant() bool {
	if inv.GuidelineSpecifiedDocumentContextParameter == "" {
		// Empty BT-24: Document does not claim EN 16931 compliance
		return false
	}

	urn := inv.GuidelineSpecifiedDocumentContextParameter
	// Check for EN 16931 compliance indicators
	return contains(urn, "en16931") || contains(urn, "factur-x") || contains(urn, "zugferd")
}

// contains checks if a string contains a substring (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

// isPEPPOL checks if the invoice is a PEPPOL BIS Billing 3.0 invoice
// based on the specification identifier (BT-24).
//
// PEPPOL-EN16931-R004 requires the specification identifier defined in
// SpecPEPPOLBilling30 constant (peppol_constants.go).
func (inv *Invoice) isPEPPOL() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == SpecPEPPOLBilling30
}

// isGerman checks if the invoice uses an XRechnung specification identifier.
// Used for auto-detection of German XRechnung-specific validation rules (BR-DE-*).
//
// Note: BR-DE rules apply specifically to XRechnung invoices, not to all German
// invoices. German sellers using Factur-X, PEPPOL, or plain EN 16931 profiles
// are not subject to BR-DE rules unless they explicitly use an XRechnung spec ID.
func (inv *Invoice) isGerman() bool {
	// Only apply German BR-DE validation when invoice explicitly uses XRechnung
	return inv.IsXRechnung()
}

// isDanish checks if the seller is located in Denmark (DK).
// Used for auto-detection of Danish-specific validation rules.
//
//nolint:unused // Reserved for future Danish validation rules
func (inv *Invoice) isDanish() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "DK"
}

// isItalian checks if the seller is located in Italy (IT).
// Used for auto-detection of Italian-specific validation rules.
//
//nolint:unused // Reserved for future Italian validation rules
func (inv *Invoice) isItalian() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "IT"
}

// isDutch checks if the seller is located in the Netherlands (NL).
// Used for auto-detection of Dutch-specific validation rules.
//
//nolint:unused // Reserved for future Dutch validation rules
func (inv *Invoice) isDutch() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "NL"
}

// isNorwegian checks if the seller is located in Norway (NO).
// Used for auto-detection of Norwegian-specific validation rules.
//
//nolint:unused // Reserved for future Norwegian validation rules
func (inv *Invoice) isNorwegian() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "NO"
}

// isSwedish checks if the seller is located in Sweden (SE).
// Used for auto-detection of Swedish-specific validation rules.
//
//nolint:unused // Reserved for future Swedish validation rules
func (inv *Invoice) isSwedish() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "SE"
}
