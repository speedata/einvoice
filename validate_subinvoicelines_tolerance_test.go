package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
)

// hasViolationCode reports whether Validate() recorded a violation with the
// given rule code on inv.
func hasViolationCode(inv *Invoice, code string) bool {
	for _, v := range inv.violations {
		if v.Rule.Code == code {
			return true
		}
	}
	return false
}

// TestSubInvoiceLines_LineTotalTolerance locks in the EXTENDED line-total
// tolerance (BR-FXEXT-CO-10), which replaces the strict BR-CO-10 when the
// invoice uses the EXTENDED profile. The official fixtures all sum exactly, so
// this builds invoices with a deliberate per-line rounding spread to exercise
// the tolerance branch and its detailCount arithmetic (Factur-X 1.09).
//
// Two detail lines of 10.00 sum to 20.00, so the tolerance is 0.01 * 2 = 0.02.
func TestSubInvoiceLines_LineTotalTolerance(t *testing.T) {
	// twoLines returns two detail lines summing to exactly 20.00.
	twoLines := func() []InvoiceLine {
		return []InvoiceLine{
			{LineID: "1", TaxCategoryCode: "S", Total: decimal.RequireFromString("10.00")},
			{LineID: "2", TaxCategoryCode: "S", Total: decimal.RequireFromString("10.00")},
		}
	}

	t.Run("extended within tolerance passes", func(t *testing.T) {
		t.Parallel()
		// LineTotal 20.01 deviates by 0.01, inside the 0.02 tolerance.
		inv := &Invoice{
			GuidelineSpecifiedDocumentContextParameter: SpecFacturXExtended,
			InvoiceLines: twoLines(),
			LineTotal:    decimal.RequireFromString("20.01"),
		}
		_ = inv.Validate()
		if hasViolationCode(inv, "BR-FXEXT-CO-10") {
			t.Error("BR-FXEXT-CO-10 reported for a 0.01 deviation within the 0.02 tolerance")
		}
		if hasViolationCode(inv, "BR-CO-10") {
			t.Error("strict BR-CO-10 reported for an EXTENDED invoice")
		}
	})

	t.Run("extended beyond tolerance fails", func(t *testing.T) {
		t.Parallel()
		// LineTotal 20.05 deviates by 0.05, outside the 0.02 tolerance.
		inv := &Invoice{
			GuidelineSpecifiedDocumentContextParameter: SpecFacturXExtended,
			InvoiceLines: twoLines(),
			LineTotal:    decimal.RequireFromString("20.05"),
		}
		_ = inv.Validate()
		if !hasViolationCode(inv, "BR-FXEXT-CO-10") {
			t.Error("expected BR-FXEXT-CO-10 for a 0.05 deviation beyond the 0.02 tolerance")
		}
	})

	t.Run("non-extended applies strict BR-CO-10", func(t *testing.T) {
		t.Parallel()
		// The same 0.01 deviation must fail the strict rule outside EXTENDED.
		inv := &Invoice{
			GuidelineSpecifiedDocumentContextParameter: SpecEN16931,
			InvoiceLines: twoLines(),
			LineTotal:    decimal.RequireFromString("20.01"),
		}
		_ = inv.Validate()
		if !hasViolationCode(inv, "BR-CO-10") {
			t.Error("expected strict BR-CO-10 for a 0.01 deviation in a non-EXTENDED invoice")
		}
		if hasViolationCode(inv, "BR-FXEXT-CO-10") {
			t.Error("BR-FXEXT-CO-10 must not apply outside the EXTENDED profile")
		}
	})
}

// TestSubInvoiceLines_NonStandardCategoryNotDoubleCounted is a regression test
// for a non-standard VAT category whose breakdown was double counted before the
// detail-line filter was applied beyond the Standard rated validator. An
// INFORMATION aggregation line may carry a VAT category and a line amount; its
// amount must not be added to the category taxable amount on top of the detail
// line. Reproduces the reviewer's reverse-charge case: a DETAIL line of 100 plus
// an INFORMATION line of 100 with a header basis of 100 previously raised
// "BR-AE-8 ... expected 200, got 100".
func TestSubInvoiceLines_NonStandardCategoryNotDoubleCounted(t *testing.T) {
	t.Parallel()
	inv := &Invoice{
		GuidelineSpecifiedDocumentContextParameter: SpecFacturXExtended,
		InvoiceLines: []InvoiceLine{
			{LineID: "1", TaxCategoryCode: "AE", Total: decimal.RequireFromString("100.00")},
			{LineID: "2", LineStatusReasonCode: "INFORMATION", TaxCategoryCode: "AE", Total: decimal.RequireFromString("100.00")},
		},
		TradeTaxes: []TradeTax{
			{CategoryCode: "AE", BasisAmount: decimal.RequireFromString("100.00")},
		},
	}
	_ = inv.Validate()
	if hasViolationCode(inv, "BR-AE-8") || hasViolationCode(inv, "BR-FXEXT-AE-08") {
		t.Error("reverse charge basis double counted the INFORMATION aggregation line")
	}
}

// TestSubInvoiceLines_UnknownSubtype verifies that an invoice line subtype
// (BT-X-8) outside the small codelist (DETAIL / GROUP / INFORMATION) is flagged
// rather than being silently treated as an aggregation line and dropped from the
// totals.
func TestSubInvoiceLines_UnknownSubtype(t *testing.T) {
	t.Parallel()
	inv := &Invoice{
		InvoiceLines: []InvoiceLine{
			{LineID: "1", TaxCategoryCode: "S", LineStatusReasonCode: "TYPO", Total: decimal.RequireFromString("10.00")},
		},
	}
	_ = inv.Validate()
	if !hasViolationCode(inv, "BR-USER-06") {
		t.Error("expected BR-USER-06 for an unknown line subtype")
	}

	// A known subtype must not raise BR-USER-06.
	ok := &Invoice{
		InvoiceLines: []InvoiceLine{
			{LineID: "1", TaxCategoryCode: "S", LineStatusReasonCode: "GROUP", Total: decimal.RequireFromString("10.00")},
		},
	}
	_ = ok.Validate()
	if hasViolationCode(ok, "BR-USER-06") {
		t.Error("BR-USER-06 raised for a valid GROUP subtype")
	}
}
