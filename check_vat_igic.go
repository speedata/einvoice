package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATIGIC validates BR-IG-1 through BR-IG-10.
//
// These rules apply to invoices with IGIC tax (category code 'L').
// IGIC (Impuesto General Indirecto Canario) is the indirect general tax
// applicable in the Canary Islands instead of VAT. It operates similarly
// to VAT but with different rates and rules specific to the region.
//
// Key requirements for IGIC:
//   - Must have at least one VAT breakdown entry with category 'L'
//   - Seller must have a tax identifier
//   - IGIC rate can be 0 or greater (various rates apply)
//   - IGIC amount is calculated as basis × rate
//   - Must NOT have exemption reason (not an exemption, it's a different tax)
//   - Seller must have tax ID but buyer must NOT have VAT ID
//
// Business rules implemented:
//   - BR-IG-1: VAT breakdown must exist and seller must have tax ID
//   - BR-IG-2: Line VAT rate must be 0 or greater (implicit)
//   - BR-IG-3: Allowance VAT rate must be 0 or greater (implicit)
//   - BR-IG-4: Charge VAT rate must be 0 or greater (implicit)
//   - BR-IG-5: Taxable amount must match calculated sum
//   - BR-IG-6: IGIC amount must equal basis × rate
//   - BR-IG-7: Taxable amount per rate must match calculated sum
//   - BR-IG-8: IGIC amount per rate must equal basis × rate
//   - BR-IG-9: Must NOT have exemption reason
//   - BR-IG-10: Seller must have tax ID, buyer must NOT have VAT ID
func (inv *Invoice) checkVATIGIC() {
	// BR-IG-1 IGIC (Kanarische Inseln / Canary Islands)
	// Invoice with category L must have seller VAT ID
	hasIGIC := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "L" {
			hasIGIC = true
			break
		}
	}
	if !hasIGIC {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "L" {
				hasIGIC = true
				break
			}
		}
	}
	if hasIGIC {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-1", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "IGIC requires seller VAT identifier"})
		}
	}

	// BR-IG-2 IGIC
	// VAT rate must be 0 or greater for lines with category L (no validation needed - rate >= 0 is implicit)

	// BR-IG-3 IGIC
	// VAT rate must be 0 or greater for allowances with category L (no validation needed - rate >= 0 is implicit)

	// BR-IG-4 IGIC
	// VAT rate must be 0 or greater for charges with category L (no validation needed - rate >= 0 is implicit)

	// BR-IG-5 IGIC
	// Verify taxable amount calculation for category L
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			var lineTotal decimal.Decimal
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "L" {
					lineTotal = lineTotal.Add(line.Total)
				}
			}
			var allowanceTotal decimal.Decimal
			var chargeTotal decimal.Decimal
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "L" {
					if ac.ChargeIndicator {
						chargeTotal = chargeTotal.Add(ac.ActualAmount)
					} else {
						allowanceTotal = allowanceTotal.Add(ac.ActualAmount)
					}
				}
			}
			expectedBasis := lineTotal.Sub(allowanceTotal).Add(chargeTotal)
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-5", InvFields: []string{"BT-116"}, Text: fmt.Sprintf("IGIC taxable amount mismatch: got %s, expected %s", tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2))})
			}
		}
	}

	// BR-IG-6 IGIC
	// VAT amount must equal basis * rate
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-6", InvFields: []string{"BT-117"}, Text: fmt.Sprintf("IGIC VAT amount must equal basis * rate: got %s, expected %s", tt.CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2))})
			}
		}
	}

	// BR-IG-7 IGIC
	// For each different VAT rate, verify taxable amount calculation
	igicRateMap := make(map[string]decimal.Decimal)
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "L" {
			key := line.TaxRateApplicablePercent.String()
			igicRateMap[key] = igicRateMap[key].Add(line.Total)
		}
	}
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.CategoryTradeTaxCategoryCode == "L" {
			key := ac.CategoryTradeTaxRateApplicablePercent.String()
			if ac.ChargeIndicator {
				igicRateMap[key] = igicRateMap[key].Add(ac.ActualAmount)
			} else {
				igicRateMap[key] = igicRateMap[key].Sub(ac.ActualAmount)
			}
		}
	}
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			key := tt.Percent.String()
			expectedBasis := igicRateMap[key]
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-7", InvFields: []string{"BT-116"}, Text: fmt.Sprintf("IGIC taxable amount for rate %s: got %s, expected %s", tt.Percent.StringFixed(2), tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2))})
			}
		}
	}

	// BR-IG-8 IGIC
	// For each different VAT rate, verify VAT amount calculation
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-8", InvFields: []string{"BT-117"}, Text: fmt.Sprintf("IGIC VAT amount for rate %s must equal basis * rate: got %s, expected %s", tt.Percent.StringFixed(2), tt.CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2))})
			}
		}
	}

	// BR-IG-9 IGIC
	// IGIC breakdown must NOT have exemption reason
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" && (tt.ExemptionReason != "" || tt.ExemptionReasonCode != "") {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-9", InvFields: []string{"BG-23", "BT-120", "BT-121"}, Text: "IGIC VAT breakdown must not have exemption reason"})
		}
	}

	// BR-IG-10 IGIC
	// Must have seller tax ID and must NOT have buyer VAT ID
	hasIGICInVATBreakdown := false
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			hasIGICInVATBreakdown = true
			break
		}
	}
	if hasIGICInVATBreakdown {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" || inv.Seller.FCTaxRegistration != ""
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""

		if !hasSellerTaxID {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-10", InvFields: []string{"BT-31", "BT-32"}, Text: "IGIC requires seller VAT or tax registration identifier"})
		}
		if hasBuyerVATID {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-IG-10", InvFields: []string{"BT-48"}, Text: "IGIC must not have buyer VAT identifier"})
		}
	}
}
