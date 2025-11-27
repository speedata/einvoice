package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validatePEPPOL validates the invoice against PEPPOL BIS Billing 3.0 rules.
//
// PEPPOL (Pan-European Public Procurement On-Line) BIS Billing 3.0 extends
// EN 16931 with additional validation rules required for the PEPPOL network.
//
// This method checks business logic rules that can be validated on the
// Invoice struct. Some PEPPOL rules that require XML structure validation
// (e.g., PEPPOL-EN16931-R008 for empty elements) are checked during parsing.
//
// Common PEPPOL rules implemented:
//   - PEPPOL-EN16931-R001: Business process must be provided (BT-23)
//   - PEPPOL-EN16931-R002: No more than one note on document level
//   - PEPPOL-EN16931-R003: Buyer reference or purchase order reference required (BT-10/BT-13)
//   - PEPPOL-EN16931-R010: Buyer electronic address required (BT-49)
//   - PEPPOL-EN16931-R020: Seller electronic address required (BT-34)
//   - PEPPOL-EN16931-R120: Invoice line net amount calculation validation
//   - PEPPOL-EN16931-R121: Base quantity must be positive above zero
//   - PEPPOL-EN16931-R130: Unit code of price base quantity must match invoiced quantity
//
// Note: Full PEPPOL validation also requires checking the XML structure and
// additional business rules. This is a basic implementation covering the most
// common PEPPOL requirements. Country-specific rules (DK, IT, NL, NO, SE) and
// advanced validations are not yet implemented.
//
// TODO: Implement additional PEPPOL rules:
//   - PEPPOL-EN16931-R005: VAT accounting currency code validation
//   - PEPPOL-EN16931-R006: Only one invoiced object on document level
//   - PEPPOL-EN16931-R110: Start date of line period within invoice period
//   - PEPPOL-EN16931-R111: End date of line period within invoice period
//   - Country-specific rules (DK-R-*, IT-R-*, NL-R-*, NO-R-*, SE-R-*)
//   - Code list validations (PEPPOL-EN16931-CL*)
//   - Format validations (PEPPOL-EN16931-F*)
//   - Profile-specific rules (PEPPOL-EN16931-P*)
//   - Common identifier format rules (PEPPOL-COMMON-R*)
func (inv *Invoice) validatePEPPOL() {
	// PEPPOL-EN16931-R001: Business process MUST be provided (BT-23)
	if inv.BPSpecifiedDocumentContextParameter == "" {
		inv.addViolation(rules.PEPPOLEN16931R1, "Business process MUST be provided")
	}

	// PEPPOL-EN16931-R007: Business process format validation
	if inv.BPSpecifiedDocumentContextParameter != "" {
		if err := ValidateBusinessProcessID(inv.BPSpecifiedDocumentContextParameter); err != nil {
			inv.addViolation(rules.PEPPOLEN16931R7, err.Error())
		}
	}

	// PEPPOL-EN16931-R002: No more than one note is allowed on document level
	if len(inv.Notes) > 1 {
		inv.addViolation(rules.PEPPOLEN16931R2, "No more than one note is allowed on document level")
	}

	// PEPPOL-EN16931-R003: A buyer reference or purchase order reference MUST be provided
	// BT-10 (BuyerReference) or BT-13 (BuyerOrderReferencedDocument)
	if inv.BuyerReference == "" && inv.BuyerOrderReferencedDocument == "" {
		inv.addViolation(rules.PEPPOLEN16931R3, "A buyer reference or purchase order reference MUST be provided")
	}

	// PEPPOL-EN16931-R010: Buyer electronic address MUST be provided (BT-49)
	if inv.Buyer.URIUniversalCommunication == "" {
		inv.addViolation(rules.PEPPOLEN16931R10, "Buyer electronic address MUST be provided")
	}

	// PEPPOL-EN16931-R020: Seller electronic address MUST be provided (BT-34)
	if inv.Seller.URIUniversalCommunication == "" {
		inv.addViolation(rules.PEPPOLEN16931R20, "Seller electronic address MUST be provided")
	}

	// Validate invoice line calculations (R120, R121, R130)
	inv.validatePEPPOLLineCalculations()
}

// validatePEPPOLLineCalculations validates line-level calculation rules.
//
// This implements the following PEPPOL BIS Billing 3.0 rules:
//   - PEPPOL-EN16931-R120: Invoice line net amount MUST equal
//     (Invoiced quantity × (Item net price / item price base quantity)
//     + Sum of invoice line charge amount - sum of invoice line allowance amount
//   - PEPPOL-EN16931-R121: Base quantity MUST be a positive number above zero
//   - PEPPOL-EN16931-R130: Unit code of price base quantity MUST be same as invoiced quantity
//
// These rules ensure that line-level calculations are mathematically correct,
// catching errors before they cascade to document-level totals.
func (inv *Invoice) validatePEPPOLLineCalculations() {
	for i, line := range inv.InvoiceLines {
		// Create line reference for error messages
		lineRef := line.LineID
		if lineRef == "" {
			lineRef = fmt.Sprintf("%d", i+1)
		}

		// PEPPOL-EN16931-R121: Base quantity MUST be a positive number above zero
		// Only validate if BasisQuantity was explicitly set (non-zero in parsed XML)
		// When element is missing, parser returns zero and we default to 1 for calculation
		if !line.BasisQuantity.IsZero() && !line.BasisQuantity.GreaterThan(decimal.Zero) {
			inv.addViolation(rules.PEPPOLEN16931R121,
				fmt.Sprintf("Line %s: Base quantity MUST be a positive number above zero (got %s)",
					lineRef, line.BasisQuantity))
		}

		// PEPPOL-EN16931-R130: Unit code of price base quantity MUST be same as invoiced quantity
		// Only validate if BasisQuantityUnit is specified (element present in XML)
		if line.BasisQuantityUnit != "" && line.BasisQuantityUnit != line.BilledQuantityUnit {
			inv.addViolation(rules.PEPPOLEN16931R130,
				fmt.Sprintf("Line %s: Unit code of price base quantity (%s) MUST be same as invoiced quantity (%s)",
					lineRef, line.BasisQuantityUnit, line.BilledQuantityUnit))
		}

		// PEPPOL-EN16931-R120: Invoice line net amount calculation
		// Formula: (quantity × price / baseQty) + charges - allowances
		baseQty := line.BasisQuantity
		if baseQty.IsZero() {
			// Default to 1 when not specified (per EN 16931)
			baseQty = decimal.NewFromInt(1)
		}

		// Calculate: BilledQuantity × NetPrice / BasisQuantity
		calculated := line.BilledQuantity.Mul(line.NetPrice).Div(baseQty)

		// Add line-level charges (BG-28)
		chargeTotal := decimal.Zero
		for _, charge := range line.InvoiceLineCharges {
			calculated = calculated.Add(charge.ActualAmount)
			chargeTotal = chargeTotal.Add(charge.ActualAmount)
		}

		// Subtract line-level allowances (BG-27)
		allowanceTotal := decimal.Zero
		for _, allowance := range line.InvoiceLineAllowances {
			calculated = calculated.Sub(allowance.ActualAmount)
			allowanceTotal = allowanceTotal.Add(allowance.ActualAmount)
		}

		// Round to 2 decimal places (per PEPPOL schematron)
		expected := roundHalfUp(calculated, 2)

		if !line.Total.Equal(expected) {
			inv.addViolation(rules.PEPPOLEN16931R120,
				fmt.Sprintf("Line %s: Invoice line net amount %s does not match calculated %s "+
					"(qty %s × price %s / baseQty %s + charges %s - allowances %s)",
					lineRef, line.Total, expected,
					line.BilledQuantity, line.NetPrice, baseQty,
					chargeTotal, allowanceTotal))
		}
	}
}
