package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATIPSI validates BR-AG-1 through BR-AG-10.
//
// These rules apply to invoices with IPSI tax (category code 'M').
// IPSI (Impuesto sobre la Produccion, los Servicios y la Importacion) is
// the production, services, and import tax applicable in Ceuta and Melilla
// instead of VAT. Similar to IGIC, it operates as a regional replacement
// for VAT with its own rates and rules.
//
// Key requirements for IPSI:
//   - Must have at least one VAT breakdown entry with category 'M'
//   - Seller must have a tax identifier
//   - IPSI rate can be 0 or greater (various rates apply)
//   - IPSI amount is calculated as basis x rate
//   - Must NOT have exemption reason (not an exemption, it's a different tax)
//   - Seller must have tax ID but buyer must NOT have VAT ID
func (inv *Invoice) validateVATIPSI() {
	// BR-AG-1 IPSI (Ceuta/Melilla)
	// Invoice with category M must have seller VAT ID
	hasIPSI := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "M" {
			hasIPSI = true
			break
		}
	}
	if !hasIPSI {
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "M" {
				hasIPSI = true
				break
			}
		}
	}
	if hasIPSI {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRAG1, "IPSI requires seller VAT identifier")
		}
	}

	// BR-AG-2 IPSI
	// VAT rate must be 0 or greater for lines with category M (no validation needed - rate >= 0 is implicit)

	// BR-AG-3 IPSI
	// VAT rate must be 0 or greater for allowances with category M (no validation needed - rate >= 0 is implicit)

	// BR-AG-4 IPSI
	// VAT rate must be 0 or greater for charges with category M (no validation needed - rate >= 0 is implicit)

	// BR-AG-5 IPSI
	// Verify taxable amount calculation for category M
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "M" {
				var lineTotal decimal.Decimal
				for j := range inv.InvoiceLines {
					if inv.InvoiceLines[j].TaxCategoryCode == "M" {
						lineTotal = lineTotal.Add(inv.InvoiceLines[j].Total)
					}
				}
				var allowanceTotal decimal.Decimal
				var chargeTotal decimal.Decimal
				for j := range inv.SpecifiedTradeAllowanceCharge {
					if inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxCategoryCode == "M" {
						if inv.SpecifiedTradeAllowanceCharge[j].ChargeIndicator {
							chargeTotal = chargeTotal.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						} else {
							allowanceTotal = allowanceTotal.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						}
					}
				}
				expectedBasis := roundHalfUp(lineTotal.Sub(allowanceTotal).Add(chargeTotal), 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(expectedBasis) {
					inv.addViolation(rules.BRAG5, fmt.Sprintf("IPSI taxable amount mismatch: got %s, expected %s", inv.TradeTaxes[i].BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
				}
			}
		}
	}

	// BR-AG-6 IPSI
	// VAT amount must equal basis * rate
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "M" {
			expectedVAT := roundHalfUp(inv.TradeTaxes[i].BasisAmount.Mul(inv.TradeTaxes[i].Percent).Div(decimal.NewFromInt(100)), 2)
			if !inv.TradeTaxes[i].CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRAG6, fmt.Sprintf("IPSI VAT amount must equal basis * rate: got %s, expected %s", inv.TradeTaxes[i].CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2)))
			}
		}
	}

	// BR-AG-7 IPSI
	// For each different VAT rate, verify taxable amount calculation
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		ipsiRateMap := make(map[string]decimal.Decimal)
		for i := range inv.InvoiceLines {
			if inv.InvoiceLines[i].TaxCategoryCode == "M" {
				key := inv.InvoiceLines[i].TaxRateApplicablePercent.String()
				ipsiRateMap[key] = ipsiRateMap[key].Add(inv.InvoiceLines[i].Total)
			}
		}
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "M" {
				key := inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.String()
				if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator {
					ipsiRateMap[key] = ipsiRateMap[key].Add(inv.SpecifiedTradeAllowanceCharge[i].ActualAmount)
				} else {
					ipsiRateMap[key] = ipsiRateMap[key].Sub(inv.SpecifiedTradeAllowanceCharge[i].ActualAmount)
				}
			}
		}
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "M" {
				key := inv.TradeTaxes[i].Percent.String()
				expectedBasis := roundHalfUp(ipsiRateMap[key], 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(expectedBasis) {
					inv.addViolation(rules.BRAG7, fmt.Sprintf("IPSI taxable amount for rate %s: got %s, expected %s", inv.TradeTaxes[i].Percent.StringFixed(2), inv.TradeTaxes[i].BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
				}
			}
		}
	}

	// BR-AG-8 IPSI
	// For each different VAT rate, verify VAT amount calculation
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "M" {
			expectedVAT := roundHalfUp(inv.TradeTaxes[i].BasisAmount.Mul(inv.TradeTaxes[i].Percent).Div(decimal.NewFromInt(100)), 2)
			if !inv.TradeTaxes[i].CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRAG8, fmt.Sprintf("IPSI VAT amount for rate %s must equal basis * rate: got %s, expected %s", inv.TradeTaxes[i].Percent.StringFixed(2), inv.TradeTaxes[i].CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2)))
			}
		}
	}

	// BR-AG-9 IPSI
	// IPSI breakdown must NOT have exemption reason
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "M" && (inv.TradeTaxes[i].ExemptionReason != "" || inv.TradeTaxes[i].ExemptionReasonCode != "") {
			inv.addViolation(rules.BRAG9, "IPSI VAT breakdown must not have exemption reason")
		}
	}

	// BR-AG-10 IPSI
	// Must have seller tax ID and must NOT have buyer VAT ID
	hasIPSIInVATBreakdown := false
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "M" {
			hasIPSIInVATBreakdown = true
			break
		}
	}
	if hasIPSIInVATBreakdown {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" || inv.Seller.FCTaxRegistration != ""
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""

		if !hasSellerTaxID {
			inv.addViolation(rules.BRAG10, "IPSI requires seller VAT or tax registration identifier")
		}
		if hasBuyerVATID {
			inv.addViolation(rules.BRAG10, "IPSI must not have buyer VAT identifier")
		}
	}
}
