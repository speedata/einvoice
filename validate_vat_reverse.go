package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATReverse validates BR-AE-1 through BR-AE-10.
//
// These rules apply to invoices with Reverse charge VAT (category code 'AE').
// Reverse charge is when the buyer, not the seller, is liable to pay the VAT.
//
// Key requirements for Reverse charge VAT:
//   - Must have at least one VAT breakdown entry with category 'AE'
//   - Both seller and buyer must have VAT identifiers
//   - VAT rate must be 0 (liability transferred to buyer)
//   - VAT amount must be 0 in the invoice
//   - Must have exemption reason explaining the reverse charge
func (inv *Invoice) validateVATReverse() {
	// BR-AE-1 Umkehrung der Steuerschuldnerschaft (Reverse charge)
	// If invoice has line/allowance/charge with "AE", must have at least one "AE" in VAT breakdown
	hasReverseCharge := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "AE" {
			hasReverseCharge = true
			break
		}
	}
	if !hasReverseCharge {
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "AE" {
				hasReverseCharge = true
				break
			}
		}
	}
	if hasReverseCharge {
		hasAEInBreakdown := false
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "AE" {
				hasAEInBreakdown = true
				break
			}
		}
		if !hasAEInBreakdown {
			inv.addViolation(rules.BRAE1, "Invoice with Reverse charge items must have Reverse charge VAT breakdown")
		}
	}

	// BR-AE-2 Umkehrung der Steuerschuldnerschaft
	// If invoice line has "AE", must have seller VAT ID/tax reg/rep VAT ID AND buyer VAT ID
	hasAELine := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "AE" {
			hasAELine = true
			break
		}
	}
	if hasAELine {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""
		if !hasSellerTaxID || !hasBuyerVATID {
			inv.addViolation(rules.BRAE2, "Invoice with Reverse charge line must have seller and buyer VAT identifiers")
		}
	}

	// BR-AE-3 Umkehrung der Steuerschuldnerschaft
	// If document level allowance has "AE", must have seller and buyer tax IDs
	hasAEAllowance := false
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "AE" {
			hasAEAllowance = true
			break
		}
	}
	if hasAEAllowance {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""
		if !hasSellerTaxID || !hasBuyerVATID {
			inv.addViolation(rules.BRAE3, "Invoice with Reverse charge allowance must have seller and buyer VAT identifiers")
		}
	}

	// BR-AE-4 Umkehrung der Steuerschuldnerschaft
	// If document level charge has "AE", must have seller and buyer tax IDs
	hasAECharge := false
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "AE" {
			hasAECharge = true
			break
		}
	}
	if hasAECharge {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""
		if !hasSellerTaxID || !hasBuyerVATID {
			inv.addViolation(rules.BRAE4, "Invoice with Reverse charge charge must have seller and buyer VAT identifiers")
		}
	}

	// BR-AE-5 Umkehrung der Steuerschuldnerschaft
	// In invoice line with "AE", VAT rate must be 0
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "AE" && !inv.InvoiceLines[i].TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRAE5, "Reverse charge invoice line must have VAT rate of 0")
		}
	}

	// BR-AE-6 Umkehrung der Steuerschuldnerschaft
	// In document level allowance with "AE", VAT rate must be 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "AE" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRAE6, "Reverse charge allowance must have VAT rate of 0")
		}
	}

	// BR-AE-7 Umkehrung der Steuerschuldnerschaft
	// In document level charge with "AE", VAT rate must be 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "AE" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRAE7, "Reverse charge charge must have VAT rate of 0")
		}
	}

	// BR-AE-8 Umkehrung der Steuerschuldnerschaft
	// Taxable amount must match calculated sum for Reverse charge category
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "AE" {
				calculatedBasis := decimal.Zero
				for j := range inv.InvoiceLines {
					if inv.InvoiceLines[j].TaxCategoryCode == "AE" {
						calculatedBasis = calculatedBasis.Add(inv.InvoiceLines[j].Total)
					}
				}
				for j := range inv.SpecifiedTradeAllowanceCharge {
					if inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxCategoryCode == "AE" {
						if inv.SpecifiedTradeAllowanceCharge[j].ChargeIndicator {
							calculatedBasis = calculatedBasis.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						} else {
							calculatedBasis = calculatedBasis.Sub(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						}
					}
				}
				calculatedBasis = roundHalfUp(calculatedBasis, 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(calculatedBasis) {
					inv.addViolation(rules.BRAE8, fmt.Sprintf("Reverse charge taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), inv.TradeTaxes[i].BasisAmount.String()))
				}
			}
		}
	}

	// BR-AE-9 Umkehrung der Steuerschuldnerschaft
	// VAT amount must be 0 for Reverse charge
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "AE" && !inv.TradeTaxes[i].CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRAE9, "Reverse charge VAT amount must be 0")
		}
	}

	// BR-AE-10 Umkehrung der Steuerschuldnerschaft
	// Reverse charge breakdown must have exemption reason code or text
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "AE" && inv.TradeTaxes[i].ExemptionReason == "" && inv.TradeTaxes[i].ExemptionReasonCode == "" {
			inv.addViolation(rules.BRAE10, "Reverse charge VAT breakdown must have exemption reason")
		}
	}
}
