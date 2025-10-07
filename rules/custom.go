package rules

// This file contains custom rules not present in the EN 16931 schematron
// but used in the validation logic. These are manually maintained.

var (
	// Check validates that invoice line net amount equals quantity × net price
	Check = Rule{
		Code:        "Check",
		Fields:      []string{"BT-131", "BT-129", "BT-146", "BT-149"},
		Description: `Invoice line net amount (BT-131) = invoiced quantity (BT-129) × item net price (BT-146) / item price base quantity (BT-149)`,
	}

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
)
