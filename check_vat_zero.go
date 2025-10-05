package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATZero validates BR-Z-1 through BR-Z-10.
//
// These rules apply to invoices with Zero rated VAT (category code 'Z').
// Zero rating means VAT is charged at 0% rate, different from exemption.
// Common for exports within certain trading blocs or specific goods/services.
//
// Key requirements for Zero rated VAT:
//   - Must have at least one VAT breakdown entry with category 'Z'
//   - Seller must have a VAT identifier or tax registration
//   - VAT rate must be 0 (taxable but at 0% rate)
//   - VAT amount must be 0
//   - Must NOT have exemption reason (it's rated, not exempt)
func (inv *Invoice) checkVATZero() {
	// BR-Z-1 Umsatzsteuer mit Nullsatz (Zero rated)
	// If invoice has line/allowance/charge with "Z", must have at least one "Z" in VAT breakdown
	hasZeroRated := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "Z" {
			hasZeroRated = true
			break
		}
	}
	if !hasZeroRated {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "Z" {
				hasZeroRated = true
				break
			}
		}
	}
	if hasZeroRated {
		hasZInBreakdown := false
		for _, tt := range inv.TradeTaxes {
			if tt.CategoryCode == "Z" {
				hasZInBreakdown = true
				break
			}
		}
		if !hasZInBreakdown {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-1", InvFields: []string{"BG-23", "BT-118"}, Text: "Invoice with Zero rated items must have Zero rated VAT breakdown"})
		}
	}

	// BR-Z-2 Umsatzsteuer mit Nullsatz
	// If invoice line has "Z", must have seller VAT ID or tax registration or representative VAT ID
	hasZLine := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "Z" {
			hasZLine = true
			break
		}
	}
	if hasZLine {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-2", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Invoice with Zero rated line must have seller VAT identifier or tax registration"})
		}
	}

	// BR-Z-3 Umsatzsteuer mit Nullsatz
	// If document level allowance has "Z", must have seller tax ID
	hasZAllowance := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "Z" {
			hasZAllowance = true
			break
		}
	}
	if hasZAllowance {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-3", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Invoice with Zero rated allowance must have seller VAT identifier or tax registration"})
		}
	}

	// BR-Z-4 Umsatzsteuer mit Nullsatz
	// If document level charge has "Z", must have seller tax ID
	hasZCharge := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "Z" {
			hasZCharge = true
			break
		}
	}
	if hasZCharge {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-4", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Invoice with Zero rated charge must have seller VAT identifier or tax registration"})
		}
	}

	// BR-Z-5 Umsatzsteuer mit Nullsatz
	// In invoice line with "Z", VAT rate must be 0
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "Z" && !line.TaxRateApplicablePercent.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-5", InvFields: []string{"BG-25", "BT-152"}, Text: "Zero rated invoice line must have VAT rate of 0"})
		}
	}

	// BR-Z-6 Umsatzsteuer mit Nullsatz
	// In document level allowance with "Z", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "Z" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-6", InvFields: []string{"BG-20", "BT-96"}, Text: "Zero rated allowance must have VAT rate of 0"})
		}
	}

	// BR-Z-7 Umsatzsteuer mit Nullsatz
	// In document level charge with "Z", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "Z" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-7", InvFields: []string{"BG-21", "BT-103"}, Text: "Zero rated charge must have VAT rate of 0"})
		}
	}

	// BR-Z-8 Umsatzsteuer mit Nullsatz
	// Taxable amount must match calculated sum for Zero rated category
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "Z" {
			calculatedBasis := decimal.Zero
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "Z" {
					calculatedBasis = calculatedBasis.Add(line.Total)
				}
			}
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "Z" {
					if ac.ChargeIndicator {
						calculatedBasis = calculatedBasis.Add(ac.ActualAmount)
					} else {
						calculatedBasis = calculatedBasis.Sub(ac.ActualAmount)
					}
				}
			}
			calculatedBasis = calculatedBasis.Round(2)
			if !tt.BasisAmount.Equal(calculatedBasis) {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-8", InvFields: []string{"BG-23", "BT-116"}, Text: fmt.Sprintf("Zero rated taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), tt.BasisAmount.String())})
			}
		}
	}

	// BR-Z-9 Umsatzsteuer mit Nullsatz
	// VAT amount must be 0 for Zero rated
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "Z" && !tt.CalculatedAmount.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-9", InvFields: []string{"BG-23", "BT-117"}, Text: "Zero rated VAT amount must be 0"})
		}
	}

	// BR-Z-10 Umsatzsteuer mit Nullsatz
	// Zero rated breakdown must not have exemption reason code or text
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "Z" && (tt.ExemptionReason != "" || tt.ExemptionReasonCode != "") {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-Z-10", InvFields: []string{"BG-23", "BT-120", "BT-121"}, Text: "Zero rated VAT breakdown must not have exemption reason"})
		}
	}
}
