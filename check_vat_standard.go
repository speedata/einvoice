package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATStandard validates BR-S-1 through BR-S-10.
//
// These rules apply to invoices with Standard rated VAT (category code 'S').
// Standard rate is the normal VAT rate applied to most goods and services.
//
// Key requirements for Standard rated VAT:
//   - Must have at least one VAT breakdown entry with category 'S'
//   - Seller must have a VAT identifier or tax registration
//   - VAT rate must be greater than 0 (not zero)
//   - VAT amount is calculated as basis amount × rate
//   - Must NOT have exemption reason or code
func (inv *Invoice) checkVATStandard() {
	// BR-S-1 Umsatzsteuer mit Normalsatz
	// If invoice has line/allowance/charge with "Standard rated" (S), must have at least one "S" in VAT breakdown
	hasStandardRated := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "S" {
			hasStandardRated = true
			break
		}
	}
	if !hasStandardRated {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "S" {
				hasStandardRated = true
				break
			}
		}
	}
	if hasStandardRated {
		hasStandardInBreakdown := false
		for _, tt := range inv.TradeTaxes {
			if tt.CategoryCode == "S" {
				hasStandardInBreakdown = true
				break
			}
		}
		if !hasStandardInBreakdown {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-1", InvFields: []string{"BG-23", "BT-118"}, Text: "Invoice with Standard rated items must have Standard rated VAT breakdown"})
		}
	}

	// BR-S-2 Umsatzsteuer mit Normalsatz
	// If invoice line has "Standard rated", must have seller VAT ID, tax reg, or rep VAT ID
	hasStandardLine := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "S" {
			hasStandardLine = true
			break
		}
	}
	if hasStandardLine {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-2", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Invoice with Standard rated line must have seller VAT identifier or tax registration"})
		}
	}

	// BR-S-3 Umsatzsteuer mit Normalsatz
	// If document level allowance has "Standard rated", must have seller tax ID
	hasStandardAllowance := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "S" {
			hasStandardAllowance = true
			break
		}
	}
	if hasStandardAllowance {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-3", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Invoice with Standard rated allowance must have seller VAT identifier or tax registration"})
		}
	}

	// BR-S-4 Umsatzsteuer mit Normalsatz
	// If document level charge has "Standard rated", must have seller tax ID
	hasStandardCharge := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "S" {
			hasStandardCharge = true
			break
		}
	}
	if hasStandardCharge {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-4", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "Invoice with Standard rated charge must have seller VAT identifier or tax registration"})
		}
	}

	// BR-S-5 Umsatzsteuer mit Normalsatz
	// In invoice line with "Standard rated", VAT rate must be > 0
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "S" && !line.TaxRateApplicablePercent.IsPositive() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-5", InvFields: []string{"BG-25", "BT-152"}, Text: "Standard rated invoice line must have VAT rate greater than 0"})
		}
	}

	// BR-S-6 Umsatzsteuer mit Normalsatz
	// In document level allowance with "Standard rated", VAT rate must be > 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "S" && !ac.CategoryTradeTaxRateApplicablePercent.IsPositive() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-6", InvFields: []string{"BG-20", "BT-96"}, Text: "Standard rated allowance must have VAT rate greater than 0"})
		}
	}

	// BR-S-7 Umsatzsteuer mit Normalsatz
	// In document level charge with "Standard rated", VAT rate must be > 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "S" && !ac.CategoryTradeTaxRateApplicablePercent.IsPositive() {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-7", InvFields: []string{"BG-21", "BT-103"}, Text: "Standard rated charge must have VAT rate greater than 0"})
		}
	}

	// BR-S-8 Umsatzsteuer mit Normalsatz
	// For each distinct rate in Standard rated category, taxable amount must match calculated sum
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "S" {
			// Calculate sum: lines - allowances + charges for this rate
			calculatedBasis := decimal.Zero
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "S" && line.TaxRateApplicablePercent.Equal(tt.Percent) {
					calculatedBasis = calculatedBasis.Add(line.Total)
				}
			}
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "S" && ac.CategoryTradeTaxRateApplicablePercent.Equal(tt.Percent) {
					if ac.ChargeIndicator {
						calculatedBasis = calculatedBasis.Add(ac.ActualAmount)
					} else {
						calculatedBasis = calculatedBasis.Sub(ac.ActualAmount)
					}
				}
			}
			// Round to 2 decimals for comparison
			calculatedBasis = calculatedBasis.Round(2)
			if !tt.BasisAmount.Equal(calculatedBasis) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-8", InvFields: []string{"BG-23", "BT-116"}, Text: fmt.Sprintf("Standard rated taxable amount must equal sum of line amounts for rate %s (expected %s, got %s)", tt.Percent.String(), calculatedBasis.String(), tt.BasisAmount.String())})
			}
		}
	}

	// BR-S-9 Umsatzsteuer mit Normalsatz
	// VAT amount must equal taxable amount * rate
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "S" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-9", InvFields: []string{"BG-23", "BT-117"}, Text: fmt.Sprintf("Standard rated VAT amount must equal basis * rate (expected %s, got %s)", expectedVAT.String(), tt.CalculatedAmount.String())})
			}
		}
	}

	// BR-S-10 Umsatzsteuer mit Normalsatz
	// Standard rated breakdown must not have exemption reason or code
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "S" && (tt.ExemptionReason != "" || tt.ExemptionReasonCode != "") {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-S-10", InvFields: []string{"BG-23", "BT-120", "BT-121"}, Text: "Standard rated VAT breakdown must not have exemption reason"})
		}
	}
}
