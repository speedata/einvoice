package einvoice

import (
	"github.com/shopspring/decimal"
)

// UpdateApplicableTradeTax removes the existing trade tax lines in the invoice
// and re-creates new ones from the line items. er is a map that contains
// exemption reasons for each category code.
func (inv *Invoice) UpdateApplicableTradeTax(exemptReason map[string]string) {
	var applicableTradeTaxes = make([]*TradeTax, 0, len(inv.TradeTaxes))

	for _, lineitem := range inv.InvoiceLines {
		tradeTax := TradeTax{
			CategoryCode: lineitem.TaxCategoryCode,
			Percent:      lineitem.TaxRateApplicablePercent,
			BasisAmount:  lineitem.Total,
			Typ:          "VAT",
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

	inv.TradeTaxes = []TradeTax{}
	onehundred := decimal.NewFromInt(100)

	for _, att := range applicableTradeTaxes {
		att.CalculatedAmount = att.BasisAmount.Mul(att.Percent.Div(onehundred)).Round(2)
		if att.Percent.IsZero() {
			att.ExemptionReason = exemptReason[att.CategoryCode]
		}

		inv.TradeTaxes = append(inv.TradeTaxes, *att)
	}
}

// UpdateTotals recalculates all monetary totals according to EN 16931 business rules.
// This function implements the following business rules:
// - BR-CO-10: LineTotal (BT-106) = Sum of all invoice line net amounts (BT-131)
// - BR-CO-13: TaxBasisTotal (BT-109) = LineTotal (BT-106) - AllowanceTotal (BT-107) + ChargeTotal (BT-108)
// - BR-CO-15: GrandTotal (BT-112) = TaxBasisTotal (BT-109) + TaxTotal (BT-110)
// - BR-CO-16: DuePayableAmount (BT-115) = GrandTotal (BT-112) - TotalPrepaid (BT-113) + RoundingAmount (BT-114)
func (inv *Invoice) UpdateTotals() {
	// Reset all calculated totals to zero to ensure idempotent behavior
	inv.LineTotal = decimal.Zero
	inv.TaxTotal = decimal.Zero

	// BR-CO-10: Calculate line total from invoice lines (BT-106)
	for _, line := range inv.InvoiceLines {
		inv.LineTotal = inv.LineTotal.Add(line.Total)
	}

	// Calculate tax total from trade taxes (BT-110)
	for _, v := range inv.TradeTaxes {
		inv.TaxTotal = inv.TaxTotal.Add(v.CalculatedAmount)
	}

	// BR-CO-13: Apply document-level allowances and charges to calculate tax basis total
	inv.TaxBasisTotal = inv.LineTotal.Sub(inv.AllowanceTotal).Add(inv.ChargeTotal)

	// BR-CO-15: Calculate grand total
	inv.GrandTotal = inv.TaxBasisTotal.Add(inv.TaxTotal)

	// BR-CO-16: Calculate due payable amount including rounding
	inv.DuePayableAmount = inv.GrandTotal.Sub(inv.TotalPrepaid).Add(inv.RoundingAmount)
}
