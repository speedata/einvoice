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

	// BR-FXEXT-*: Factur-X EXTENDED profile rules (Factur-X 1.09 / ZUGFeRD 2.5)
	// that replace the corresponding EN 16931 base rules to support sub invoice
	// lines (chapter 7.6.2). The aggregation lines (LineStatusReasonCode "GROUP"
	// or "INFORMATION", BT-X-8) are excluded from the sums, and the sum
	// comparisons allow a tolerance of 0.01 per amount. Source:
	// Factur-X_1.09_EXTENDED.sch.
	BRFXEXTCO10 = Rule{
		Code:        "BR-FXEXT-CO-10",
		Fields:      []string{"BT-106", "BT-131", "BT-X-8"},
		Description: `Absolute value of (Sum of Invoice line net amount (BT-106) - Σ Invoice line net amounts (BT-131)) <= 0,01 * Number of line net amounts, counting only lines whose subtype (BT-X-8) is "DETAIL" or unspecified.`,
	}
	BRFXEXTS08 = Rule{
		Code:        "BR-FXEXT-S-08",
		Fields:      []string{"BT-116", "BT-131", "BT-92", "BT-99", "BT-X-8"},
		Description: `For Standard rated (S) VAT, absolute value of (VAT category taxable amount (BT-116) - (Σ line net amounts - Σ document allowances + Σ document charges)) <= 0,01 per amount, counting only lines whose subtype (BT-X-8) is "DETAIL" or unspecified.`,
	}
	BRFXEXT22 = Rule{
		Code:        "BR-FXEXT-22",
		Fields:      []string{"BT-129", "BT-X-8"},
		Description: `Each detail Invoice line (subtype BT-X-8 "DETAIL" or unspecified) shall contain the Invoiced quantity (BT-129).`,
	}
	BRFXEXT23 = Rule{
		Code:        "BR-FXEXT-23",
		Fields:      []string{"BT-130", "BT-X-8"},
		Description: `Each detail Invoice line (subtype BT-X-8 "DETAIL" or unspecified) shall contain the Invoiced quantity unit of measure code (BT-130).`,
	}
	BRFXEXT26 = Rule{
		Code:        "BR-FXEXT-26",
		Fields:      []string{"BT-146", "BT-X-8"},
		Description: `Each detail Invoice line (subtype BT-X-8 "DETAIL" or unspecified) shall contain the Item net price (BT-146).`,
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
