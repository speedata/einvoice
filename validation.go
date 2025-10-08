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

// Validate checks the invoice against applicable business rules with intelligent auto-detection.
//
// The method automatically detects which validation rules to apply based on:
// - Specification identifier (BT-24) for PEPPOL BIS Billing 3.0 detection
// - Seller country for country-specific rules (DK, IT, NL, NO, SE)
//
// All invoices are validated against EN 16931 core rules. Additional rules are applied
// automatically when the invoice metadata indicates they are required.
//
// This method clears any previous violations and performs a fresh validation.
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

	// Always run EN 16931 core validation
	inv.checkBR()
	inv.checkBRO()
	inv.checkBRDEC()

	// Auto-detect and run PEPPOL validation based on specification identifier
	if inv.isPEPPOL() {
		inv.checkPEPPOL()
	}

	// Auto-detect country-specific rules based on seller location
	// Note: Country-specific validators are placeholders for future implementation
	if inv.isDanish() {
		// inv.checkDanish() // TODO: Implement Danish validation rules
	}

	if inv.isItalian() {
		// inv.checkItalian() // TODO: Implement Italian validation rules
	}

	if inv.isDutch() {
		// inv.checkDutch() // TODO: Implement Dutch validation rules
	}

	if inv.isNorwegian() {
		// inv.checkNorwegian() // TODO: Implement Norwegian validation rules
	}

	if inv.isSwedish() {
		// inv.checkSwedish() // TODO: Implement Swedish validation rules
	}

	// Return error if violations exist
	if len(inv.violations) > 0 {
		return &ValidationError{violations: inv.violations}
	}

	return nil
}

// isPEPPOL checks if the invoice is a PEPPOL BIS Billing 3.0 invoice
// based on the specification identifier (BT-24).
//
// PEPPOL-EN16931-R004 requires the specification identifier to be:
// "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0"
func (inv *Invoice) isPEPPOL() bool {
	return inv.GuidelineSpecifiedDocumentContextParameter == "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:3.0"
}

// isDanish checks if the seller is located in Denmark (DK).
// Used for auto-detection of Danish-specific validation rules.
func (inv *Invoice) isDanish() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "DK"
}

// isItalian checks if the seller is located in Italy (IT).
// Used for auto-detection of Italian-specific validation rules.
func (inv *Invoice) isItalian() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "IT"
}

// isDutch checks if the seller is located in the Netherlands (NL).
// Used for auto-detection of Dutch-specific validation rules.
func (inv *Invoice) isDutch() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "NL"
}

// isNorwegian checks if the seller is located in Norway (NO).
// Used for auto-detection of Norwegian-specific validation rules.
func (inv *Invoice) isNorwegian() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "NO"
}

// isSwedish checks if the seller is located in Sweden (SE).
// Used for auto-detection of Swedish-specific validation rules.
func (inv *Invoice) isSwedish() bool {
	return inv.Seller.PostalAddress != nil &&
		inv.Seller.PostalAddress.CountryID == "SE"
}
