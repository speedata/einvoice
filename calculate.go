package einvoice

import (
	"github.com/shopspring/decimal"
)

// roundHalfUp performs commercial rounding (round half up) to the specified number of decimal places.
// This is required by EN 16931 for VAT calculations, as opposed to banker's rounding.
// For example, 34700.0458 rounded to 2 places becomes 34700.05 (not 34700.04).
func roundHalfUp(d decimal.Decimal, places int32) decimal.Decimal {
	// Add 0.5 of the smallest unit and truncate
	// For 2 decimal places, we add 0.005 and truncate
	adjustment := decimal.New(5, -(places + 1))
	if d.IsNegative() {
		return d.Sub(adjustment).Truncate(places)
	}
	return d.Add(adjustment).Truncate(places)
}

// UpdateApplicableTradeTax removes the existing trade tax lines in the invoice
// and re-creates new ones from the line items and document-level allowances/charges.
// er is a map that contains exemption reasons for each category code.
// According to BR-45 and category-specific rules (BR-S-8, BR-AE-8, BR-E-8, etc.),
// the VAT category taxable amount must include:
// - Sum of invoice line net amounts for that category
// - Minus document level allowance amounts for that category
// - Plus document level charge amounts for that category
func (inv *Invoice) UpdateApplicableTradeTax(exemptReason map[string]string) {
	var applicableTradeTaxes = make([]*TradeTax, 0, len(inv.TradeTaxes))

	// Process invoice lines
	for _, lineitem := range inv.InvoiceLines {
		tradeTax := TradeTax{
			CategoryCode: lineitem.TaxCategoryCode,
			Percent:      lineitem.TaxRateApplicablePercent,
			BasisAmount:  lineitem.Total,
			TypeCode:     "VAT",
		}
		found := false

		for _, att := range applicableTradeTaxes {
			if att.CategoryCode == tradeTax.CategoryCode && att.Percent.Equal(tradeTax.Percent) {
				att.BasisAmount = att.BasisAmount.Add(lineitem.Total)
				found = true

				break
			}
		}

		if !found {
			applicableTradeTaxes = append(applicableTradeTaxes, &tradeTax)
		}
	}

	// Process document-level allowances and charges
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		found := false
		for _, att := range applicableTradeTaxes {
			if att.CategoryCode == ac.CategoryTradeTaxCategoryCode && att.Percent.Equal(ac.CategoryTradeTaxRateApplicablePercent) {
				// Charges add to the basis, allowances subtract
				if ac.ChargeIndicator {
					att.BasisAmount = att.BasisAmount.Add(ac.ActualAmount)
				} else {
					att.BasisAmount = att.BasisAmount.Sub(ac.ActualAmount)
				}
				found = true
				break
			}
		}

		// If not found, create a new tax entry for this category
		if !found {
			basisAmount := ac.ActualAmount
			if !ac.ChargeIndicator {
				basisAmount = basisAmount.Neg()
			}
			tradeTax := &TradeTax{
				CategoryCode: ac.CategoryTradeTaxCategoryCode,
				Percent:      ac.CategoryTradeTaxRateApplicablePercent,
				BasisAmount:  basisAmount,
				TypeCode:     "VAT",
			}
			applicableTradeTaxes = append(applicableTradeTaxes, tradeTax)
		}
	}

	inv.TradeTaxes = []TradeTax{}
	onehundred := decimal.NewFromInt(100)

	for _, att := range applicableTradeTaxes {
		// BR-DEC-19: VAT category taxable amount (BT-116) must have max 2 decimal places
		att.BasisAmount = roundHalfUp(att.BasisAmount, 2)
		att.CalculatedAmount = roundHalfUp(att.BasisAmount.Mul(att.Percent.Div(onehundred)), 2)
		if att.Percent.IsZero() {
			att.ExemptionReason = exemptReason[att.CategoryCode]
		}

		inv.TradeTaxes = append(inv.TradeTaxes, *att)
	}
}

// updateAllowancesAndCharges recalculates the document-level allowance and charge totals
// according to EN 16931 business rules.
// This function implements the following business rules:
// - BR-CO-11: AllowanceTotal (BT-107) = Sum of all document level allowance amounts (BT-92)
// - BR-CO-12: ChargeTotal (BT-108) = Sum of all document level charge amounts (BT-99)
func (inv *Invoice) updateAllowancesAndCharges() {
	// Reset totals to zero to ensure idempotent behavior
	inv.AllowanceTotal = decimal.Zero
	inv.ChargeTotal = decimal.Zero

	// BR-CO-11 and BR-CO-12: Calculate allowance and charge totals
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator {
			inv.ChargeTotal = inv.ChargeTotal.Add(ac.ActualAmount)
		} else {
			inv.AllowanceTotal = inv.AllowanceTotal.Add(ac.ActualAmount)
		}
	}
}

// UpdateTotals recalculates all monetary totals according to EN 16931 business rules.
// This function implements the following business rules in calculation order:
// - BR-CO-10: LineTotal (BT-106) = Sum of all invoice line net amounts (BT-131)
// - BR-CO-11: AllowanceTotal (BT-107) = Sum of all document level allowance amounts (BT-92)
// - BR-CO-12: ChargeTotal (BT-108) = Sum of all document level charge amounts (BT-99)
// - BR-CO-13: TaxBasisTotal (BT-109) = LineTotal - AllowanceTotal + ChargeTotal
// - BR-CO-15: GrandTotal (BT-112) = TaxBasisTotal + TaxTotal (BT-110)
// - BR-CO-16: DuePayableAmount (BT-115) = GrandTotal - TotalPrepaid (BT-113) + RoundingAmount (BT-114)
//
// Profile handling: Minimum (level 1) and BasicWL (level 2) profiles do not have line items by design.
// For these profiles, totals are provided directly and cannot be recalculated from line items.
// This function skips recalculation for profiles below level 3 (Basic, EN16931, Extended, etc.).
func (inv *Invoice) UpdateTotals() {
	// Skip recalculation for Minimum and BasicWL profiles - they don't have line items
	// and provide totals directly in the invoice
	if inv.ProfileLevel() < 3 {
		return
	}

	// BR-CO-10: Calculate line total from invoice lines (BT-106)
	inv.LineTotal = decimal.Zero
	for _, line := range inv.InvoiceLines {
		inv.LineTotal = inv.LineTotal.Add(line.Total)
	}

	// BR-CO-11 & BR-CO-12: Calculate allowance and charge totals from document-level items
	inv.updateAllowancesAndCharges()

	// Calculate tax total from trade taxes (BT-110)
	inv.TaxTotal = decimal.Zero
	for _, v := range inv.TradeTaxes {
		inv.TaxTotal = inv.TaxTotal.Add(v.CalculatedAmount)
	}

	// BR-CO-13: TaxBasisTotal = LineTotal - AllowanceTotal + ChargeTotal
	inv.TaxBasisTotal = inv.LineTotal.Sub(inv.AllowanceTotal).Add(inv.ChargeTotal)

	// BR-CO-15: GrandTotal = TaxBasisTotal + TaxTotal
	inv.GrandTotal = inv.TaxBasisTotal.Add(inv.TaxTotal)

	// BR-CO-16: DuePayableAmount = GrandTotal - TotalPrepaid + RoundingAmount
	inv.DuePayableAmount = inv.GrandTotal.Sub(inv.TotalPrepaid).Add(inv.RoundingAmount)
}
