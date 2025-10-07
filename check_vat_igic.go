package einvoice

import (
	"github.com/speedata/einvoice/rules"
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATIGIC validates BR-AF-1 through BR-AF-10.
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
func (inv *Invoice) checkVATIGIC() {
	// BR-AF-1 IGIC (Kanarische Inseln / Canary Islands)
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
			inv.addViolation(rules.BRAF1, "IGIC requires seller VAT identifier")
		}
	}

	// BR-AF-2 IGIC
	// VAT rate must be 0 or greater for lines with category L (no validation needed - rate >= 0 is implicit)

	// BR-AF-3 IGIC
	// VAT rate must be 0 or greater for allowances with category L (no validation needed - rate >= 0 is implicit)

	// BR-AF-4 IGIC
	// VAT rate must be 0 or greater for charges with category L (no validation needed - rate >= 0 is implicit)

	// BR-AF-5 IGIC
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
				inv.addViolation(rules.BRAF5, fmt.Sprintf("IGIC taxable amount mismatch: got %s, expected %s", tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
			}
		}
	}

	// BR-AF-6 IGIC
	// VAT amount must equal basis * rate
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRAF6, fmt.Sprintf("IGIC VAT amount must equal basis * rate: got %s, expected %s", tt.CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2)))
			}
		}
	}

	// BR-AF-7 IGIC
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
				inv.addViolation(rules.BRAF7, fmt.Sprintf("IGIC taxable amount for rate %s: got %s, expected %s", tt.Percent.StringFixed(2), tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
			}
		}
	}

	// BR-AF-8 IGIC
	// For each different VAT rate, verify VAT amount calculation
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRAF8, fmt.Sprintf("IGIC VAT amount for rate %s must equal basis * rate: got %s, expected %s", tt.Percent.StringFixed(2), tt.CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2)))
			}
		}
	}

	// BR-AF-9 IGIC
	// IGIC breakdown must NOT have exemption reason
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "L" && (tt.ExemptionReason != "" || tt.ExemptionReasonCode != "") {
			inv.addViolation(rules.BRAF9, "IGIC VAT breakdown must not have exemption reason")
		}
	}

	// BR-AF-10 IGIC
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
			inv.addViolation(rules.BRAF10, "IGIC requires seller VAT or tax registration identifier")
		}
		if hasBuyerVATID {
			inv.addViolation(rules.BRAF10, "IGIC must not have buyer VAT identifier")
		}
	}
}
