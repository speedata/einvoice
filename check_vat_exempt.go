package einvoice

import (
	"github.com/speedata/einvoice/rules"
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATExempt validates BR-E-1 through BR-E-10.
//
// These rules apply to invoices with Exempt from VAT (category code 'E').
// VAT exemption applies to specific goods/services that are excluded from VAT
// under national or EU law (e.g., financial services, education, healthcare).
//
// Key requirements for Exempt from VAT:
//   - Must have at least one VAT breakdown entry with category 'E'
//   - Seller must have a VAT identifier or tax registration
//   - VAT rate must be 0 (exempt, not taxable)
//   - VAT amount must be 0
//   - Must have exemption reason explaining why VAT is exempt
func (inv *Invoice) checkVATExempt() {
	// BR-E-1 Steuerbefreit (Exempt from VAT)
	// If invoice has line/allowance/charge with "E", must have at least one "E" in VAT breakdown
	hasExempt := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "E" {
			hasExempt = true
			break
		}
	}
	if !hasExempt {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "E" {
				hasExempt = true
				break
			}
		}
	}
	if hasExempt {
		hasEInBreakdown := false
		for _, tt := range inv.TradeTaxes {
			if tt.CategoryCode == "E" {
				hasEInBreakdown = true
				break
			}
		}
		if !hasEInBreakdown {
			inv.addViolation(rules.BRE1, "Invoice with Exempt from VAT items must have Exempt from VAT breakdown")
		}
	}

	// BR-E-2 Steuerbefreit
	// If invoice line has "E", must have seller VAT ID or tax registration or representative VAT ID
	hasELine := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "E" {
			hasELine = true
			break
		}
	}
	if hasELine {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRE2, "Invoice with Exempt from VAT line must have seller VAT identifier or tax registration")
		}
	}

	// BR-E-3 Steuerbefreit
	// If document level allowance has "E", must have seller tax ID
	hasEAllowance := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "E" {
			hasEAllowance = true
			break
		}
	}
	if hasEAllowance {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRE3, "Invoice with Exempt from VAT allowance must have seller VAT identifier or tax registration")
		}
	}

	// BR-E-4 Steuerbefreit
	// If document level charge has "E", must have seller tax ID
	hasECharge := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "E" {
			hasECharge = true
			break
		}
	}
	if hasECharge {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRE4, "Invoice with Exempt from VAT charge must have seller VAT identifier or tax registration")
		}
	}

	// BR-E-5 Steuerbefreit
	// In invoice line with "E", VAT rate must be 0
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "E" && !line.TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRE5, "Exempt from VAT invoice line must have VAT rate of 0")
		}
	}

	// BR-E-6 Steuerbefreit
	// In document level allowance with "E", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "E" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRE6, "Exempt from VAT allowance must have VAT rate of 0")
		}
	}

	// BR-E-7 Steuerbefreit
	// In document level charge with "E", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "E" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRE7, "Exempt from VAT charge must have VAT rate of 0")
		}
	}

	// BR-E-8 Steuerbefreit
	// Taxable amount must match calculated sum for Exempt from VAT category
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "E" {
			calculatedBasis := decimal.Zero
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "E" {
					calculatedBasis = calculatedBasis.Add(line.Total)
				}
			}
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "E" {
					if ac.ChargeIndicator {
						calculatedBasis = calculatedBasis.Add(ac.ActualAmount)
					} else {
						calculatedBasis = calculatedBasis.Sub(ac.ActualAmount)
					}
				}
			}
			calculatedBasis = calculatedBasis.Round(2)
			if !tt.BasisAmount.Equal(calculatedBasis) {
				inv.addViolation(rules.BRE8, fmt.Sprintf("Exempt from VAT taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), tt.BasisAmount.String()))
			}
		}
	}

	// BR-E-9 Steuerbefreit
	// VAT amount must be 0 for Exempt from VAT
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "E" && !tt.CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRE9, "Exempt from VAT amount must be 0")
		}
	}

	// BR-E-10 Steuerbefreit
	// Exempt from VAT breakdown must have exemption reason code or text
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "E" && tt.ExemptionReason == "" && tt.ExemptionReasonCode == "" {
			inv.addViolation(rules.BRE10, "Exempt from VAT breakdown must have exemption reason")
		}
	}
}
