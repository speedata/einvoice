package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATNotSubject validates BR-O-1 through BR-O-14.
//
// These rules apply to invoices with "Not subject to VAT" category (code 'O').
// This applies to transactions that fall outside the scope of VAT entirely,
// such as certain financial services, insurance, gambling, etc. Unlike VAT
// exemption, these transactions are not part of the VAT system at all.
//
// Key requirements for Not subject to VAT:
//   - Must have at least one VAT breakdown entry with category 'O'
//   - Seller or buyer must have a tax identifier (not both required)
//   - VAT rate must be 0 (not subject to any VAT rate)
//   - VAT amount must be 0
//   - Must have exemption reason explaining why not subject to VAT
//   - Only one 'O' category allowed in VAT breakdown
func (inv *Invoice) checkVATNotSubject() {
	// BR-O-1 Not subject to VAT
	// Invoice with category O must have seller OR buyer tax ID
	hasNotSubjectToVAT := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" {
			hasNotSubjectToVAT = true
			break
		}
	}
	if !hasNotSubjectToVAT {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "O" {
				hasNotSubjectToVAT = true
				break
			}
		}
	}
	if hasNotSubjectToVAT {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		hasBuyerTaxID := inv.Buyer.VATaxRegistration != "" ||
			(inv.Buyer.SpecifiedLegalOrganization != nil && inv.Buyer.SpecifiedLegalOrganization.ID != "")

		if !hasSellerTaxID && !hasBuyerTaxID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-1", InvFields: []string{"BT-31", "BT-32", "BT-63", "BT-47", "BT-48"}, Text: "Not subject to VAT requires seller or buyer tax identifier"})
		}
	}

	// BR-O-2 Not subject to VAT
	// Lines with category O require seller tax ID
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" {
			hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
				inv.Seller.FCTaxRegistration != "" ||
				(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")

			if !hasSellerTaxID {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-2", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Not subject to VAT line requires seller tax identifier"})
			}
			break
		}
	}

	// BR-O-3 Not subject to VAT
	// Lines with O must exist in VAT breakdown
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" {
			found := false
			for _, tt := range inv.TradeTaxes {
				if tt.CategoryCode == "O" {
					found = true
					break
				}
			}
			if !found {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-3", InvFields: []string{"BG-23", "BT-118"}, Text: "Invoice line with Not subject to VAT must have corresponding VAT breakdown"})
			}
			break
		}
	}

	// BR-O-4 Not subject to VAT
	// Allowances with O must exist in VAT breakdown
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" {
			found := false
			for _, tt := range inv.TradeTaxes {
				if tt.CategoryCode == "O" {
					found = true
					break
				}
			}
			if !found {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-4", InvFields: []string{"BG-23", "BT-118"}, Text: "Allowance with Not subject to VAT must have corresponding VAT breakdown"})
			}
			break
		}
	}

	// BR-O-5 Not subject to VAT
	// Charges with O must exist in VAT breakdown
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" {
			found := false
			for _, tt := range inv.TradeTaxes {
				if tt.CategoryCode == "O" {
					found = true
					break
				}
			}
			if !found {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-5", InvFields: []string{"BG-23", "BT-118"}, Text: "Charge with Not subject to VAT must have corresponding VAT breakdown"})
			}
			break
		}
	}

	// BR-O-6 Not subject to VAT
	// VAT rate must be 0 for lines with category O
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" && !line.TaxRateApplicablePercent.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-6", InvFields: []string{"BT-152"}, Text: "Not subject to VAT rate must be 0"})
		}
	}

	// BR-O-7 Not subject to VAT
	// VAT rate must be 0 for allowances with category O
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-7", InvFields: []string{"BT-96"}, Text: "Not subject to VAT allowance rate must be 0"})
		}
	}

	// BR-O-8 Not subject to VAT
	// VAT rate must be 0 for charges with category O
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-8", InvFields: []string{"BT-103"}, Text: "Not subject to VAT charge rate must be 0"})
		}
	}

	// BR-O-9 Not subject to VAT
	// Verify taxable amount calculation for category O
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "O" {
			var lineTotal decimal.Decimal
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "O" {
					lineTotal = lineTotal.Add(line.Total)
				}
			}
			var allowanceTotal decimal.Decimal
			var chargeTotal decimal.Decimal
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "O" {
					if ac.ChargeIndicator {
						chargeTotal = chargeTotal.Add(ac.ActualAmount)
					} else {
						allowanceTotal = allowanceTotal.Add(ac.ActualAmount)
					}
				}
			}
			expectedBasis := lineTotal.Sub(allowanceTotal).Add(chargeTotal)
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-9", InvFields: []string{"BT-116"}, Text: fmt.Sprintf("Not subject to VAT taxable amount mismatch: got %s, expected %s", tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2))})
			}
		}
	}

	// BR-O-10 Not subject to VAT
	// For each different VAT rate, verify taxable amount calculation
	notSubjectRateMap := make(map[string]decimal.Decimal)
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" {
			key := line.TaxRateApplicablePercent.String()
			notSubjectRateMap[key] = notSubjectRateMap[key].Add(line.Total)
		}
	}
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.CategoryTradeTaxCategoryCode == "O" {
			key := ac.CategoryTradeTaxRateApplicablePercent.String()
			if ac.ChargeIndicator {
				notSubjectRateMap[key] = notSubjectRateMap[key].Add(ac.ActualAmount)
			} else {
				notSubjectRateMap[key] = notSubjectRateMap[key].Sub(ac.ActualAmount)
			}
		}
	}
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "O" {
			key := tt.Percent.String()
			expectedBasis := notSubjectRateMap[key]
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-10", InvFields: []string{"BT-116"}, Text: fmt.Sprintf("Not subject to VAT taxable amount for rate %s: got %s, expected %s", tt.Percent.StringFixed(2), tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2))})
			}
		}
	}

	// BR-O-11 Not subject to VAT
	// VAT amount must be 0 for category O
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "O" && !tt.CalculatedAmount.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-11", InvFields: []string{"BT-117"}, Text: "Not subject to VAT amount must be 0"})
		}
	}

	// BR-O-12 Not subject to VAT
	// For each different VAT rate, VAT amount must be 0
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "O" && !tt.CalculatedAmount.IsZero() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-12", InvFields: []string{"BT-117"}, Text: "Not subject to VAT amount for each rate must be 0"})
		}
	}

	// BR-O-13 Not subject to VAT
	// Not subject to VAT breakdown must have exemption reason
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "O" && tt.ExemptionReason == "" && tt.ExemptionReasonCode == "" {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-13", InvFields: []string{"BG-23", "BT-120", "BT-121"}, Text: "Not subject to VAT breakdown must have exemption reason"})
		}
	}

	// BR-O-14 Not subject to VAT
	// Only one Not subject to VAT category allowed (all should have rate 0)
	oCount := 0
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "O" {
			oCount++
		}
	}
	if oCount > 1 {
		inv.violations = append(inv.violations, SemanticError{Rule: "BR-O-14", InvFields: []string{"BG-23"}, Text: "Only one Not subject to VAT category allowed in VAT breakdown"})
	}
}
