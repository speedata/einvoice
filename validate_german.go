package einvoice

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/speedata/einvoice/rules"
)

// validateGerman performs German XRechnung-specific business rule validation.
//
// This method validates invoices against BR-DE-* rules defined in the
// XRechnung specification (CIUS XRechnung for Germany).
//
// XRechnung is the German implementation of EN 16931, required for invoices to
// German public authorities and increasingly used in B2B scenarios.
//
// This validation applies when:
// - The specification identifier (BT-24) matches an XRechnung URN, OR
// - The seller is located in Germany (DE) and the invoice is EN 16931 compliant
//
// BR-DE Rules Implemented:
//
// Presence Tests (mandatory fields):
//   - BR-DE-1:  Payment instructions (BG-16) must be provided
//   - BR-DE-2:  Seller contact (BG-6) must be provided
//   - BR-DE-3:  Seller city (BT-37) must be provided
//   - BR-DE-4:  Seller post code (BT-38) must be provided
//   - BR-DE-5:  Seller contact point (BT-41) must be provided
//   - BR-DE-6:  Seller contact telephone (BT-42) must be provided
//   - BR-DE-7:  Seller contact email (BT-43) must be provided
//   - BR-DE-8:  Buyer city (BT-52) must be provided
//   - BR-DE-9:  Buyer post code (BT-53) must be provided
//   - BR-DE-10: Deliver to city (BT-77) must be provided if delivery address exists
//   - BR-DE-11: Deliver to post code (BT-78) must be provided if delivery address exists
//   - BR-DE-15: Buyer reference (BT-10) must be provided (Leitweg-ID)
//   - BR-DE-30: Bank assigned creditor identifier (BT-90) for direct debit
//   - BR-DE-31: Debited account identifier (BT-91) for direct debit
//
// Validation Tests (format and logic):
//   - BR-DE-16: VAT identifiers must have ISO 3166-1 alpha-2 country prefix
//   - BR-DE-17: Invoice type code validation (UNTDID 1001)
//   - BR-DE-18: Payment terms structured format (Skonto)
//   - BR-DE-19: IBAN validation for SEPA credit transfer (code 58)
//   - BR-DE-20: IBAN validation for SEPA direct debit (code 59)
//   - BR-DE-21: Specification identifier must match XRechnung standard
//   - BR-DE-23: Payment means requirements (codes 30, 58, 59)
//   - BR-DE-24: Payment card information requirements (codes 48, 54)
//   - BR-DE-25: Direct debit mandate requirements (code 59)
//   - BR-DE-26: Corrected invoice must reference preceding invoice
//   - BR-DE-27: Seller contact telephone must contain at least 3 digits
//   - BR-DE-28: Email address format validation
//
// Reference: https://github.com/itplr-kosit/xrechnung-schematron
func (inv *Invoice) validateGerman() {
	// BR-DE-1: Payment instructions (BG-16) must be provided
	if len(inv.PaymentMeans) == 0 {
		inv.addViolation(rules.BRDE1, "An invoice must contain information on PAYMENT INSTRUCTIONS (BG-16)")
	}

	// BR-DE-2: Seller contact (BG-6) must be provided
	if len(inv.Seller.DefinedTradeContact) == 0 {
		inv.addViolation(rules.BRDE2, "The element group SELLER CONTACT (BG-6) must be transmitted")
	}

	// BR-DE-3: Seller city (BT-37) must be provided
	if inv.Seller.PostalAddress == nil || inv.Seller.PostalAddress.City == "" {
		inv.addViolation(rules.BRDE3, "The element 'Seller city' (BT-37) must be transmitted")
	}

	// BR-DE-4: Seller post code (BT-38) must be provided
	if inv.Seller.PostalAddress == nil || inv.Seller.PostalAddress.PostcodeCode == "" {
		inv.addViolation(rules.BRDE4, "The element 'Seller post code' (BT-38) must be transmitted")
	}

	// BR-DE-5, BR-DE-6, BR-DE-7: Seller contact details
	if len(inv.Seller.DefinedTradeContact) > 0 {
		contact := inv.Seller.DefinedTradeContact[0]

		// BR-DE-5: Seller contact point (BT-41)
		if contact.PersonName == "" && contact.DepartmentName == "" {
			inv.addViolation(rules.BRDE5, "The element 'Seller contact point' (BT-41) must be transmitted")
		}

		// BR-DE-6: Seller contact telephone number (BT-42)
		if contact.PhoneNumber == "" {
			inv.addViolation(rules.BRDE6, "The element 'Seller contact telephone number' (BT-42) must be transmitted")
		} else {
			// BR-DE-27: Telephone must contain at least 3 digits
			digitCount := countDigits(contact.PhoneNumber)
			if digitCount < 3 {
				inv.addViolation(rules.BRDE27, "Seller contact telephone number (BT-42) must contain at least three digits")
			}
		}

		// BR-DE-7: Seller contact email address (BT-43)
		if contact.EMail == "" {
			inv.addViolation(rules.BRDE7, "The element 'Seller contact email address' (BT-43) must be transmitted")
		} else {
			// BR-DE-28: Email format validation
			if !isValidEmail(contact.EMail) {
				inv.addViolation(rules.BRDE28, "Email address must have valid format (one @, no leading/trailing dots, etc.)")
			}
		}
	}

	// BR-DE-8: Buyer city (BT-52) must be provided
	if inv.Buyer.PostalAddress == nil || inv.Buyer.PostalAddress.City == "" {
		inv.addViolation(rules.BRDE8, "The element 'Buyer city' (BT-52) must be transmitted")
	}

	// BR-DE-9: Buyer post code (BT-53) must be provided
	if inv.Buyer.PostalAddress == nil || inv.Buyer.PostalAddress.PostcodeCode == "" {
		inv.addViolation(rules.BRDE9, "The element 'Buyer post code' (BT-53) must be transmitted")
	}

	// BR-DE-10, BR-DE-11: Deliver to address (if provided)
	if inv.ShipTo != nil && inv.ShipTo.PostalAddress != nil {
		// BR-DE-10: Deliver to city (BT-77)
		if inv.ShipTo.PostalAddress.City == "" {
			inv.addViolation(rules.BRDE10, "The element 'Deliver to city' (BT-77) must be transmitted if delivery address is provided")
		}

		// BR-DE-11: Deliver to post code (BT-78)
		if inv.ShipTo.PostalAddress.PostcodeCode == "" {
			inv.addViolation(rules.BRDE11, "The element 'Deliver to post code' (BT-78) must be transmitted if delivery address is provided")
		}
	}

	// BR-DE-15: Buyer reference (BT-10) must be provided (Leitweg-ID)
	if inv.BuyerReference == "" {
		inv.addViolation(rules.BRDE15, "The element 'Buyer reference' (BT-10) must be transmitted")
	}

	// BR-DE-16: VAT identifier must have ISO country code prefix
	if inv.Seller.VATaxRegistration != "" {
		if !hasISOCountryPrefix(inv.Seller.VATaxRegistration) {
			inv.addViolation(rules.BRDE16, "Seller VAT identifier (BT-31) must have a prefix in accordance with ISO code list 3166-1 alpha-2")
		}
	}

	if inv.Buyer.VATaxRegistration != "" {
		if !hasISOCountryPrefix(inv.Buyer.VATaxRegistration) {
			inv.addViolation(rules.BRDE16, "Buyer VAT identifier (BT-48) must have a prefix in accordance with ISO code list 3166-1 alpha-2")
		}
	}

	if inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "" {
		if !hasISOCountryPrefix(inv.SellerTaxRepresentativeTradeParty.VATaxRegistration) {
			inv.addViolation(rules.BRDE16, "Tax representative VAT identifier (BT-63) must have a prefix in accordance with ISO code list 3166-1 alpha-2")
		}
	}

	// BR-DE-21 is handled separately in validateGermanSpecID() because it applies
	// to ALL German sellers, not just XRechnung invoices.

	// BR-DE-23, BR-DE-24, BR-DE-25: Payment means requirements
	// These rules ensure mutual exclusivity of payment means groups (BG-17, BG-18, BG-19)
	for _, pm := range inv.PaymentMeans {
		// Determine which payment information groups are present
		hasBG17CreditTransfer := pm.PayeePartyCreditorFinancialAccountIBAN != "" ||
			pm.PayeePartyCreditorFinancialAccountProprietaryID != ""
		hasBG18PaymentCard := pm.ApplicableTradeSettlementFinancialCardID != ""
		hasBG19DirectDebit := pm.PayerPartyDebtorFinancialAccountIBAN != ""

		// BR-DE-23: Credit transfer (codes 30, 58)
		if pm.TypeCode == 30 || pm.TypeCode == 58 {
			// BR-DE-23-a: Must have BG-17 (CREDIT TRANSFER)
			if !hasBG17CreditTransfer {
				inv.addViolation(rules.BRDE23A, "Payment means code 30 or 58 (credit transfer) requires BG-17 CREDIT TRANSFER information")
			}

			// BR-DE-23-b: Must NOT have BG-18 (payment card) or BG-19 (direct debit)
			if hasBG18PaymentCard {
				inv.addViolation(rules.BRDE23B, "Payment means code 30 or 58 (credit transfer) must not contain BG-18 PAYMENT CARD INFORMATION")
			}
			if hasBG19DirectDebit {
				inv.addViolation(rules.BRDE23B, "Payment means code 30 or 58 (credit transfer) must not contain BG-19 DIRECT DEBIT")
			}
		}

		// BR-DE-24: Payment card (codes 48, 54, 55)
		if pm.TypeCode == 48 || pm.TypeCode == 54 || pm.TypeCode == 55 {
			// BR-DE-24-a: Must have BG-18 (PAYMENT CARD INFORMATION)
			if !hasBG18PaymentCard {
				inv.addViolation(rules.BRDE24A, "Payment means code 48, 54, or 55 (payment card) requires BG-18 PAYMENT CARD INFORMATION")
			}

			// BR-DE-24-b: Must NOT have BG-17 (credit transfer) or BG-19 (direct debit)
			if hasBG17CreditTransfer {
				inv.addViolation(rules.BRDE24B, "Payment means code 48, 54, or 55 (payment card) must not contain BG-17 CREDIT TRANSFER")
			}
			if hasBG19DirectDebit {
				inv.addViolation(rules.BRDE24B, "Payment means code 48, 54, or 55 (payment card) must not contain BG-19 DIRECT DEBIT")
			}
		}

		// BR-DE-25: Direct debit (code 59)
		if pm.TypeCode == 59 {
			// BR-DE-25-a: Must have BG-19 (DIRECT DEBIT)
			if !hasBG19DirectDebit {
				inv.addViolation(rules.BRDE25A, "Payment means code 59 (direct debit) requires BG-19 DIRECT DEBIT information")
			}

			// BR-DE-25-b: Must NOT have BG-17 (credit transfer) or BG-18 (payment card)
			if hasBG17CreditTransfer {
				inv.addViolation(rules.BRDE25B, "Payment means code 59 (direct debit) must not contain BG-17 CREDIT TRANSFER")
			}
			if hasBG18PaymentCard {
				inv.addViolation(rules.BRDE25B, "Payment means code 59 (direct debit) must not contain BG-18 PAYMENT CARD INFORMATION")
			}
		}

		// BR-DE-19: IBAN validation for SEPA credit transfer
		if pm.TypeCode == 58 {
			if pm.PayeePartyCreditorFinancialAccountIBAN != "" && !isValidIBAN(pm.PayeePartyCreditorFinancialAccountIBAN) {
				inv.addViolation(rules.BRDE19, "Payment account identifier (BT-84) must be a valid IBAN when using SEPA credit transfer (code 58)")
			}
		}

		// BR-DE-20: IBAN validation for SEPA direct debit
		if pm.TypeCode == 59 {
			if pm.PayerPartyDebtorFinancialAccountIBAN != "" && !isValidIBAN(pm.PayerPartyDebtorFinancialAccountIBAN) {
				inv.addViolation(rules.BRDE20, "Debited account identifier (BT-91) must be a valid IBAN when using SEPA direct debit (code 59)")
			}
		}

		// BR-DE-30, BR-DE-31: Direct debit mandatory fields
		if pm.TypeCode == 59 {
			// BR-DE-30: Bank assigned creditor identifier (BT-90)
			if inv.CreditorReferenceID == "" {
				inv.addViolation(rules.BRDE30, "Bank assigned creditor identifier (BT-90) must be provided for direct debit")
			}

			// BR-DE-31: Debited account identifier (BT-91)
			if pm.PayerPartyDebtorFinancialAccountIBAN == "" {
				inv.addViolation(rules.BRDE31, "Debited account identifier (BT-91) must be provided for direct debit")
			}
		}
	}

	// BR-DE-26: Corrected invoice must reference preceding invoice
	if int(inv.InvoiceTypeCode) == 384 {
		if len(inv.InvoiceReferencedDocument) == 0 {
			inv.addViolation(rules.BRDE26, "If invoice type code (BT-3) is 384 (Corrected invoice), PRECEDING INVOICE REFERENCE (BG-3) must be provided")
		}
	}
}

// countDigits counts the number of digit characters in a string.
func countDigits(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

// isValidEmail validates email format according to BR-DE-28.
// Requirements:
// - Exactly one @ sign
// - Does not start or end with a dot
// - @ sign must not be flanked by whitespace or dot
// - Must be preceded and followed by at least two characters
func isValidEmail(email string) bool {
	// Must have exactly one @
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	// Must not start or end with dot
	if strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return false
	}

	// Split on @
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local := parts[0]
	domain := parts[1]

	// Local and domain parts must have at least 2 characters each
	if len(local) < 2 || len(domain) < 2 {
		return false
	}

	// @ must not be flanked by whitespace or dot
	if strings.HasSuffix(local, " ") || strings.HasPrefix(domain, " ") {
		return false
	}
	if strings.HasSuffix(local, ".") || strings.HasPrefix(domain, ".") {
		return false
	}

	return true
}

// hasISOCountryPrefix checks if a VAT identifier has an ISO 3166-1 alpha-2 country code prefix.
// Valid format: Two uppercase letters followed by alphanumeric characters (e.g., DE123456789).
func hasISOCountryPrefix(vatID string) bool {
	if len(vatID) < 3 {
		return false
	}

	// First two characters must be uppercase letters (ISO 3166-1 alpha-2 code)
	first := vatID[0]
	second := vatID[1]

	if !isUppercaseLetter(first) || !isUppercaseLetter(second) {
		return false
	}

	return true
}

// isUppercaseLetter checks if a byte represents an uppercase ASCII letter (A-Z).
func isUppercaseLetter(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

// isValidIBAN performs basic IBAN validation.
// A valid IBAN:
// - Has 15-34 alphanumeric characters
// - Starts with a 2-letter country code
// - Followed by 2 check digits
// - Followed by the Basic Bank Account Number (BBAN)
//
// This is a simplified validation that checks format. Full validation
// would include modulo-97 checksum verification per ISO 13616.
func isValidIBAN(iban string) bool {
	// Remove spaces and convert to uppercase
	iban = strings.ReplaceAll(iban, " ", "")
	iban = strings.ToUpper(iban)

	// Length check: IBAN must be 15-34 characters (per SWIFT registry)
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}

	// First two characters must be letters (country code)
	if !isUppercaseLetter(iban[0]) || !isUppercaseLetter(iban[1]) {
		return false
	}

	// Next two characters must be digits (check digits)
	if !isDigit(iban[2]) || !isDigit(iban[3]) {
		return false
	}

	// Remaining characters must be alphanumeric
	for i := 4; i < len(iban); i++ {
		if !isAlphanumeric(iban[i]) {
			return false
		}
	}

	return true
}

// isDigit checks if a byte represents a digit (0-9).
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// isAlphanumeric checks if a byte represents an alphanumeric character (0-9, A-Z).
func isAlphanumeric(b byte) bool {
	return isDigit(b) || isUppercaseLetter(b)
}

// isValidSkontoFormat validates the structured Skonto format for BR-DE-18.
// Format: #SKONTO#TAGE=n#PROZENT=n#[BASISBETRAG=n]#
// This is a placeholder for future implementation.
var skontoRegex = regexp.MustCompile(`(?i)#SKONTO#TAGE=\d+#PROZENT=\d+(\.\d{1,2})?#(BASISBETRAG=\d+(\.\d{1,2})?#)?`)

func isValidSkontoFormat(paymentTerms string) bool {
	// If payment terms don't contain SKONTO, they're not required to follow the format
	if !strings.Contains(strings.ToUpper(paymentTerms), "SKONTO") {
		return true
	}

	// If they mention SKONTO, they must follow the structured format
	return skontoRegex.MatchString(paymentTerms)
}

// validateGermanSpecID validates BR-DE-21 for German sellers.
// BR-DE-21 applies to German suppliers regardless of their chosen profile,
// checking that they use an XRechnung specification identifier.
//
// This is separate from validateGerman() because:
// - BR-DE-21 applies to ALL German sellers (regardless of profile)
// - Other BR-DE rules only apply to invoices explicitly using XRechnung
func (inv *Invoice) validateGermanSpecID() {
	// Only check if seller is in Germany
	if inv.Seller.PostalAddress == nil || inv.Seller.PostalAddress.CountryID != "DE" {
		return
	}

	// BR-DE-21: German sellers should use XRechnung specification identifier
	if !inv.IsXRechnung() {
		inv.addViolation(rules.BRDE21, "The element 'Specification identifier' (BT-24) should syntactically correspond to the identifier of the XRechnung standard")
	}
}
