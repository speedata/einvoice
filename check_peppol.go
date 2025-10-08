package einvoice

import "github.com/speedata/einvoice/rules"

// checkPEPPOL validates the invoice against PEPPOL BIS Billing 3.0 rules.
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
//
// Note: Full PEPPOL validation also requires checking the XML structure and
// additional business rules. This is a basic implementation covering the most
// common PEPPOL requirements. Country-specific rules (DK, IT, NL, NO, SE) and
// advanced validations are not yet implemented.
//
// TODO: Implement additional PEPPOL rules:
//   - PEPPOL-EN16931-R005: VAT accounting currency code validation
//   - PEPPOL-EN16931-R006: Only one invoiced object on document level
//   - Country-specific rules (DK-R-*, IT-R-*, NL-R-*, NO-R-*, SE-R-*)
//   - Code list validations (PEPPOL-EN16931-CL*)
//   - Format validations (PEPPOL-EN16931-F*)
//   - Profile-specific rules (PEPPOL-EN16931-P*)
//   - Common identifier format rules (PEPPOL-COMMON-R*)
func (inv *Invoice) checkPEPPOL() {
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
}
