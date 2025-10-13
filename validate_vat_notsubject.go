package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATNotSubject validates BR-O-01 through BR-O-14.
//
// These rules apply to invoices with "Not subject to VAT" category (code 'O').
// This applies to transactions that fall outside the scope of VAT entirely,
// such as certain financial services, insurance, gambling, etc. Unlike VAT
// exemption, these transactions are not part of the VAT system at all.
//
// Key requirements for Not subject to VAT:
//   - Must have exactly one VAT breakdown entry with category 'O'
//   - Must NOT contain Seller VAT identifier, Seller tax rep VAT identifier, or Buyer VAT identifier
//   - Must NOT contain VAT rates (they should be absent, not 0)
//   - VAT amount must be 0
//   - Must have exemption reason explaining why not subject to VAT
//   - All lines/allowances/charges must be category 'O' (cannot mix with other categories)
func (inv *Invoice) validateVATNotSubject() {
	// Check if invoice has any "Not subject to VAT" items
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

	// Only validate if invoice uses category O
	if !hasNotSubjectToVAT {
		return
	}

	// BR-O-01: Must have exactly one VAT breakdown with category O
	oBreakdownCount := 0
	var oBreakdown *TradeTax
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "O" {
			oBreakdownCount++
			oBreakdown = &inv.TradeTaxes[i]
		}
	}
	if oBreakdownCount != 1 {
		inv.addViolation(rules.BRO1, fmt.Sprintf("Invoice with 'Not subject to VAT' must have exactly one VAT breakdown with category O (found %d)", oBreakdownCount))
	}

	// BR-O-02: Invoice lines with category O must NOT contain VAT identifiers
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" {
			hasVATIdentifiers := inv.Seller.VATaxRegistration != "" ||
				inv.Buyer.VATaxRegistration != "" ||
				(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")

			if hasVATIdentifiers {
				inv.addViolation(rules.BRO2, "Invoice line with 'Not subject to VAT' shall not contain Seller VAT (BT-31), Buyer VAT (BT-48), or Seller tax rep VAT (BT-63)")
				break
			}
		}
	}

	// BR-O-03: Document level allowances with category O must NOT contain VAT identifiers
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" {
			hasVATIdentifiers := inv.Seller.VATaxRegistration != "" ||
				inv.Buyer.VATaxRegistration != "" ||
				(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")

			if hasVATIdentifiers {
				inv.addViolation(rules.BRO3, "Allowance with 'Not subject to VAT' shall not contain Seller VAT (BT-31), Buyer VAT (BT-48), or Seller tax rep VAT (BT-63)")
				break
			}
		}
	}

	// BR-O-04: Document level charges with category O must NOT contain VAT identifiers
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" {
			hasVATIdentifiers := inv.Seller.VATaxRegistration != "" ||
				inv.Buyer.VATaxRegistration != "" ||
				(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")

			if hasVATIdentifiers {
				inv.addViolation(rules.BRO4, "Charge with 'Not subject to VAT' shall not contain Seller VAT (BT-31), Buyer VAT (BT-48), or Seller tax rep VAT (BT-63)")
				break
			}
		}
	}

	// BR-O-05: Invoice lines with category O must NOT contain VAT rate
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "O" && !line.TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRO5, "Invoice line with 'Not subject to VAT' shall not contain VAT rate (BT-152)")
		}
	}

	// BR-O-06: Allowances with category O must NOT contain VAT rate
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRO6, "Allowance with 'Not subject to VAT' shall not contain VAT rate (BT-96)")
		}
	}

	// BR-O-07: Charges with category O must NOT contain VAT rate
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "O" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRO7, "Charge with 'Not subject to VAT' shall not contain VAT rate (BT-103)")
		}
	}

	// BR-O-08: VAT category taxable amount calculation
	if oBreakdown != nil {
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
		expectedBasis := roundHalfUp(lineTotal.Sub(allowanceTotal).Add(chargeTotal), 2)
		if !oBreakdown.BasisAmount.Equal(expectedBasis) {
			inv.addViolation(rules.BRO8, fmt.Sprintf("'Not subject to VAT' taxable amount mismatch: got %s, expected %s", oBreakdown.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
		}
	}

	// BR-O-09: VAT amount must be 0
	if oBreakdown != nil && !oBreakdown.CalculatedAmount.IsZero() {
		inv.addViolation(rules.BRO9, "'Not subject to VAT' tax amount must be 0")
	}

	// BR-O-10: Must have exemption reason
	if oBreakdown != nil && oBreakdown.ExemptionReason == "" && oBreakdown.ExemptionReasonCode == "" {
		inv.addViolation(rules.BRO10, "'Not subject to VAT' breakdown must have exemption reason (BT-120) or reason code (BT-121)")
	}

	// BR-O-11: Only one VAT breakdown allowed (already checked in BR-O-01)
	// This rule means if you have category O, you cannot have any other categories
	if oBreakdownCount == 1 && len(inv.TradeTaxes) > 1 {
		inv.addViolation(rules.BRO11, "Invoice with 'Not subject to VAT' breakdown shall not contain other VAT categories")
	}

	// BR-O-12: All invoice lines must be category O
	if oBreakdownCount > 0 {
		for _, line := range inv.InvoiceLines {
			if line.TaxCategoryCode != "O" {
				inv.addViolation(rules.BRO12, fmt.Sprintf("Invoice with 'Not subject to VAT' breakdown shall not contain invoice lines with other categories (found %s)", line.TaxCategoryCode))
				break
			}
		}
	}

	// BR-O-13: All allowances must be category O
	if oBreakdownCount > 0 {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode != "O" {
				inv.addViolation(rules.BRO13, fmt.Sprintf("Invoice with 'Not subject to VAT' breakdown shall not contain allowances with other categories (found %s)", ac.CategoryTradeTaxCategoryCode))
				break
			}
		}
	}

	// BR-O-14: All charges must be category O
	if oBreakdownCount > 0 {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode != "O" {
				inv.addViolation(rules.BRO14, fmt.Sprintf("Invoice with 'Not subject to VAT' breakdown shall not contain charges with other categories (found %s)", ac.CategoryTradeTaxCategoryCode))
				break
			}
		}
	}
}
