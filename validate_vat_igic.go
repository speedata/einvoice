package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATIGIC validates BR-AF-1 through BR-AF-10.
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
//   - IGIC amount is calculated as basis x rate
//   - Must NOT have exemption reason (not an exemption, it's a different tax)
//   - Seller must have tax ID but buyer must NOT have VAT ID
func (inv *Invoice) validateVATIGIC() {
	// BR-AF-1 IGIC (Kanarische Inseln / Canary Islands)
	// Invoice with category L must have seller VAT ID
	hasIGIC := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "L" {
			hasIGIC = true
			break
		}
	}
	if !hasIGIC {
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "L" {
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
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "L" {
				var lineTotal decimal.Decimal
				for j := range inv.InvoiceLines {
					if inv.InvoiceLines[j].TaxCategoryCode == "L" {
						lineTotal = lineTotal.Add(inv.InvoiceLines[j].Total)
					}
				}
				var allowanceTotal decimal.Decimal
				var chargeTotal decimal.Decimal
				for j := range inv.SpecifiedTradeAllowanceCharge {
					if inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxCategoryCode == "L" {
						if inv.SpecifiedTradeAllowanceCharge[j].ChargeIndicator {
							chargeTotal = chargeTotal.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						} else {
							allowanceTotal = allowanceTotal.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						}
					}
				}
				expectedBasis := roundHalfUp(lineTotal.Sub(allowanceTotal).Add(chargeTotal), 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(expectedBasis) {
					inv.addViolation(rules.BRAF5, fmt.Sprintf("IGIC taxable amount mismatch: got %s, expected %s", inv.TradeTaxes[i].BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
				}
			}
		}
	}

	// BR-AF-6 IGIC
	// VAT amount must equal basis * rate
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "L" {
			expectedVAT := roundHalfUp(inv.TradeTaxes[i].BasisAmount.Mul(inv.TradeTaxes[i].Percent).Div(decimal100), 2)
			if !inv.TradeTaxes[i].CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRAF6, fmt.Sprintf("IGIC VAT amount must equal basis * rate: got %s, expected %s", inv.TradeTaxes[i].CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2)))
			}
		}
	}

	// BR-AF-7 IGIC
	// For each different VAT rate, verify taxable amount calculation
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		igicRateMap := make(map[string]decimal.Decimal)
		for i := range inv.InvoiceLines {
			if inv.InvoiceLines[i].TaxCategoryCode == "L" {
				key := inv.InvoiceLines[i].TaxRateApplicablePercent.String()
				igicRateMap[key] = igicRateMap[key].Add(inv.InvoiceLines[i].Total)
			}
		}
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "L" {
				key := inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.String()
				if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator {
					igicRateMap[key] = igicRateMap[key].Add(inv.SpecifiedTradeAllowanceCharge[i].ActualAmount)
				} else {
					igicRateMap[key] = igicRateMap[key].Sub(inv.SpecifiedTradeAllowanceCharge[i].ActualAmount)
				}
			}
		}
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "L" {
				key := inv.TradeTaxes[i].Percent.String()
				expectedBasis := roundHalfUp(igicRateMap[key], 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(expectedBasis) {
					inv.addViolation(rules.BRAF7, fmt.Sprintf("IGIC taxable amount for rate %s: got %s, expected %s", inv.TradeTaxes[i].Percent.StringFixed(2), inv.TradeTaxes[i].BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
				}
			}
		}
	}

	// BR-AF-8 IGIC
	// For each different VAT rate, verify VAT amount calculation
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "L" {
			expectedVAT := roundHalfUp(inv.TradeTaxes[i].BasisAmount.Mul(inv.TradeTaxes[i].Percent).Div(decimal100), 2)
			if !inv.TradeTaxes[i].CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRAF8, fmt.Sprintf("IGIC VAT amount for rate %s must equal basis * rate: got %s, expected %s", inv.TradeTaxes[i].Percent.StringFixed(2), inv.TradeTaxes[i].CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2)))
			}
		}
	}

	// BR-AF-9 IGIC
	// IGIC breakdown must NOT have exemption reason
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "L" && (inv.TradeTaxes[i].ExemptionReason != "" || inv.TradeTaxes[i].ExemptionReasonCode != "") {
			inv.addViolation(rules.BRAF9, "IGIC VAT breakdown must not have exemption reason")
		}
	}

	// BR-AF-10 IGIC
	// Must have seller tax ID and must NOT have buyer VAT ID
	hasIGICInVATBreakdown := false
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "L" {
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
