package rules

// This file contains custom rules not present in the EN 16931 schematron
// but used in the validation logic. These are manually maintained.

var (
	// BR34-40: These rules are not in the EN 16931 schematron but validate
	// that allowance and charge amounts are non-negative.
	BR34 = Rule{
		Code:        "BR-34",
		Fields:      []string{"BT-92"},
		Description: `Document level allowance amount (BT-92) must not be negative.`,
	}
	BR35 = Rule{
		Code:        "BR-35",
		Fields:      []string{"BT-93"},
		Description: `Document level allowance base amount (BT-93) must not be negative.`,
	}
	BR39 = Rule{
		Code:        "BR-39",
		Fields:      []string{"BT-99"},
		Description: `Document level charge amount (BT-99) must not be negative.`,
	}
	BR40 = Rule{
		Code:        "BR-40",
		Fields:      []string{"BT-100"},
		Description: `Document level charge base amount (BT-100) must not be negative.`,
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
