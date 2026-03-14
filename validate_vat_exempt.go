package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATExempt validates BR-E-1 through BR-E-10.
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
func (inv *Invoice) validateVATExempt() {
	// BR-E-1 Steuerbefreit (Exempt from VAT)
	// If invoice has line/allowance/charge with "E", must have at least one "E" in VAT breakdown
	hasExempt := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "E" {
			hasExempt = true
			break
		}
	}
	if !hasExempt {
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "E" {
				hasExempt = true
				break
			}
		}
	}
	if hasExempt {
		hasEInBreakdown := false
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "E" {
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
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "E" {
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
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "E" {
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
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "E" {
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
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "E" && !inv.InvoiceLines[i].TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRE5, "Exempt from VAT invoice line must have VAT rate of 0")
		}
	}

	// BR-E-6 Steuerbefreit
	// In document level allowance with "E", VAT rate must be 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "E" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRE6, "Exempt from VAT allowance must have VAT rate of 0")
		}
	}

	// BR-E-7 Steuerbefreit
	// In document level charge with "E", VAT rate must be 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "E" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRE7, "Exempt from VAT charge must have VAT rate of 0")
		}
	}

	// BR-E-8 Steuerbefreit
	// Taxable amount must match calculated sum for Exempt from VAT category
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "E" {
				calculatedBasis := decimal.Zero
				for j := range inv.InvoiceLines {
					if inv.InvoiceLines[j].TaxCategoryCode == "E" {
						calculatedBasis = calculatedBasis.Add(inv.InvoiceLines[j].Total)
					}
				}
				for j := range inv.SpecifiedTradeAllowanceCharge {
					if inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxCategoryCode == "E" {
						if inv.SpecifiedTradeAllowanceCharge[j].ChargeIndicator {
							calculatedBasis = calculatedBasis.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						} else {
							calculatedBasis = calculatedBasis.Sub(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						}
					}
				}
				calculatedBasis = roundHalfUp(calculatedBasis, 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(calculatedBasis) {
					inv.addViolation(rules.BRE8, fmt.Sprintf("Exempt from VAT taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), inv.TradeTaxes[i].BasisAmount.String()))
				}
			}
		}
	}

	// BR-E-9 Steuerbefreit
	// VAT amount must be 0 for Exempt from VAT
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "E" && !inv.TradeTaxes[i].CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRE9, "Exempt from VAT amount must be 0")
		}
	}

	// BR-E-10 Steuerbefreit
	// Exempt from VAT breakdown must have exemption reason code or text
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "E" && inv.TradeTaxes[i].ExemptionReason == "" && inv.TradeTaxes[i].ExemptionReasonCode == "" {
			inv.addViolation(rules.BRE10, "Exempt from VAT breakdown must have exemption reason")
		}
	}
}
