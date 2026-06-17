package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// sumDetailLineBasis computes the VAT category taxable amount (BT-116) for the
// given category code from the invoice lines and the document-level
// allowances/charges, and returns the number of contributing amounts.
//
// Only detail lines contribute: sub invoice line aggregation lines (GROUP /
// INFORMATION, BT-X-8) are excluded via isDetailLine so their subtotal or
// informational amounts are never double counted into a VAT basis (EXTENDED,
// Factur-X 1.09 chapter 7.6.2). For invoices without sub invoice lines every
// line is a detail line, so the result is identical to a plain sum.
//
// When matchRate is true only lines and allowances/charges whose VAT rate equals
// rate contribute (used by categories that validate one breakdown entry per
// distinct rate, e.g. Standard rated); categories with a single rate pass
// matchRate=false to sum the whole category.
//
// The result is rounded to two decimals (commercial rounding) ready to compare
// against the declared breakdown amount. The returned count drives the EXTENDED
// per-amount tolerance in checkVATCategoryBasis.
func (inv *Invoice) sumDetailLineBasis(category string, rate decimal.Decimal, matchRate bool) (decimal.Decimal, int) {
	basis := decimal.Zero
	count := 0
	for i := range inv.InvoiceLines {
		line := &inv.InvoiceLines[i]
		if !line.isDetailLine() || line.TaxCategoryCode != category {
			continue
		}
		if matchRate && !line.TaxRateApplicablePercent.Equal(rate) {
			continue
		}
		basis = basis.Add(line.Total)
		count++
	}
	for i := range inv.SpecifiedTradeAllowanceCharge {
		ac := &inv.SpecifiedTradeAllowanceCharge[i]
		if ac.CategoryTradeTaxCategoryCode != category {
			continue
		}
		if matchRate && !ac.CategoryTradeTaxRateApplicablePercent.Equal(rate) {
			continue
		}
		if ac.ChargeIndicator {
			basis = basis.Add(ac.ActualAmount)
		} else {
			basis = basis.Sub(ac.ActualAmount)
		}
		count++
	}
	return roundHalfUp(basis, 2), count
}

// checkVATCategoryBasis compares the declared VAT category taxable amount
// (BT-116, declared) against the calculated detail-line basis and adds a
// violation if they differ.
//
// In the EXTENDED profile the BR-FXEXT-*-08 rule (fxextRule) replaces the strict
// BR-*-8 rule (strictRule) and tolerates a deviation of 0.01 per contributing
// amount (Factur-X 1.09); outside EXTENDED the amounts must match exactly. label
// is the human-readable category name and rateLabel, when non-empty, the VAT rate
// appended to the message ("for rate ...").
func (inv *Invoice) checkVATCategoryBasis(label, rateLabel string, declared, calculated decimal.Decimal, amountCount int, strictRule, fxextRule rules.Rule) {
	rateSuffix := ""
	if rateLabel != "" {
		rateSuffix = " for rate " + rateLabel
	}
	if inv.IsExtended() {
		tolerance := decimal.New(1, -2).Mul(decimal.NewFromInt(int64(amountCount)))
		if declared.Sub(calculated).Abs().GreaterThan(tolerance) {
			inv.addViolation(fxextRule, fmt.Sprintf("%s taxable amount must equal sum of line amounts%s within tolerance %s (expected %s, got %s)", label, rateSuffix, tolerance.String(), calculated.String(), declared.String()))
		}
		return
	}
	if !declared.Equal(calculated) {
		inv.addViolation(strictRule, fmt.Sprintf("%s taxable amount must equal sum of line amounts%s (expected %s, got %s)", label, rateSuffix, calculated.String(), declared.String()))
	}
}
