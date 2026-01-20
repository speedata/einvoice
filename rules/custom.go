package rules

// This file contains custom rules not present in the EN 16931 schematron
// but used in the validation logic. These are manually maintained.

var (
	// BR-USER-01..04: Custom rules (not part of EN 16931 schematron) validating
	// that allowance and charge amounts are non-negative.
	BRUSER01 = Rule{
		Code:        "BR-USER-01",
		Fields:      []string{"BT-92"},
		Description: `Document level allowance amount (BT-92) must not be negative.`,
	}
	BRUSER02 = Rule{
		Code:        "BR-USER-02",
		Fields:      []string{"BT-93"},
		Description: `Document level allowance base amount (BT-93) must not be negative.`,
	}
	BRUSER03 = Rule{
		Code:        "BR-USER-03",
		Fields:      []string{"BT-99"},
		Description: `Document level charge amount (BT-99) must not be negative.`,
	}
	BRUSER04 = Rule{
		Code:        "BR-USER-04",
		Fields:      []string{"BT-100"},
		Description: `Document level charge base amount (BT-100) must not be negative.`,
	}
	BRUSER05 = Rule{
		Code:        "BR-USER-05",
		Fields:      []string{"BT-131", "BT-129", "BT-146"},
		Description: `Invoice line net amount must match calculated amount (qty × price ÷ base qty ± allowances/charges).`,
	}

	// UNEXPECTED_TAX_CURRENCY: Validates that TaxTotalAmount elements use only
	// the invoice currency (BT-5) and optionally the accounting currency (BT-6).
	// EN 16931 only defines BT-110 and BT-111, no additional currencies are allowed.
	UNEXPECTED_TAX_CURRENCY = Rule{
		Code:        "UNEXPECTED-TAX-CURRENCY",
		Fields:      []string{"BT-110", "BT-111"},
		Description: `TaxTotalAmount with unexpected currency (expected invoice currency BT-5 or accounting currency BT-6).`,
	}
)
