package einvoice

import (
	"github.com/shopspring/decimal"
)

// UpdateApplicableTradeTax removes the existing trade tax lines in the invoice
// and re-creates new ones from the line items. er is a map that contains
// exemption reasons for each category code.
func (inv *Invoice) UpdateApplicableTradeTax(er map[string]string) {
	var applicableTradeTaxes = make([]*TradeTax, 0, len(inv.TradeTaxes))
	for _, lineitem := range inv.InvoiceLines {
		ti := TradeTax{
			CategoryCode: lineitem.TaxCategoryCode,
			Percent:      lineitem.TaxRateApplicablePercent,
			BasisAmount:  lineitem.Total,
			Typ:          "VAT",
		}
		found := false
		for _, att := range applicableTradeTaxes {
			if att.CategoryCode == ti.CategoryCode && att.Percent.Equal(ti.Percent) {
				att.BasisAmount = att.BasisAmount.Add(lineitem.Total)
				found = true
				break
			}
		}
		if !found {
			applicableTradeTaxes = append(applicableTradeTaxes, &ti)
		}
	}
	inv.TradeTaxes = []TradeTax{}
	onehundred := decimal.NewFromInt(100)
	for _, att := range applicableTradeTaxes {
		att.CalculatedAmount = att.BasisAmount.Mul(att.Percent.Div(onehundred)).Round(2)
		if att.Percent.IsZero() {
			att.ExemptionReason = er[att.CategoryCode]
		}
		inv.TradeTaxes = append(inv.TradeTaxes, *att)
	}
}

// UpdateTotals collects the sum from the applicable trade taxes and updates the
// monetary summation. Charges and allowances are currently not considered.
func (inv *Invoice) UpdateTotals() {

	// var chargeTotal decimal.Decimal
	// var allowanceTotal decimal.Decimal
	for _, v := range inv.TradeTaxes {
		inv.LineTotal = inv.LineTotal.Add(v.BasisAmount)
		inv.TaxTotal = inv.TaxTotal.Add(v.CalculatedAmount)
	}
	inv.TaxBasisTotal = inv.LineTotal
	inv.GrandTotal = inv.TaxBasisTotal.Add(inv.TaxTotal)
	inv.DuePayableAmount = inv.GrandTotal.Sub(inv.TotalPrepaid)

}
