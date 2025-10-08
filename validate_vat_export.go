package einvoice

import (
	"github.com/speedata/einvoice/rules"
	"fmt"

	"github.com/shopspring/decimal"
)

// validateVATExport validates BR-G-1 through BR-G-10.
//
// These rules apply to invoices with Export outside EU VAT (category code 'G').
// This category is for goods or services exported outside the European Union,
// which are typically VAT-free for the seller but may be subject to import taxes
// in the destination country.
//
// Key requirements for Export outside EU:
//   - Must have at least one VAT breakdown entry with category 'G'
//   - Seller must have a VAT identifier (not tax registration alone)
//   - VAT rate must be 0 (exports are not taxed)
//   - VAT amount must be 0
//   - Must have exemption reason explaining the export
func (inv *Invoice) validateVATExport() {
	// BR-G-1 Export außerhalb der EU (Export outside the EU)
	// If invoice has line/allowance/charge with "G", must have at least one "G" in VAT breakdown
	hasExportOutsideEU := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "G" {
			hasExportOutsideEU = true
			break
		}
	}
	if !hasExportOutsideEU {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "G" {
				hasExportOutsideEU = true
				break
			}
		}
	}
	if hasExportOutsideEU {
		hasGInBreakdown := false
		for _, tt := range inv.TradeTaxes {
			if tt.CategoryCode == "G" {
				hasGInBreakdown = true
				break
			}
		}
		if !hasGInBreakdown {
			inv.addViolation(rules.BRG1, "Invoice with Export outside EU items must have Export outside EU VAT breakdown")
		}
	}

	// BR-G-2 Export außerhalb der EU
	// If invoice line has "G", must have seller VAT ID or representative VAT ID
	hasGLine := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "G" {
			hasGLine = true
			break
		}
	}
	if hasGLine {
		hasSellerVATID := inv.Seller.VATaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerVATID {
			inv.addViolation(rules.BRG2, "Invoice with Export outside EU line must have seller VAT identifier")
		}
	}

	// BR-G-3 Export außerhalb der EU
	// If document level allowance has "G", must have seller VAT ID
	hasGAllowance := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "G" {
			hasGAllowance = true
			break
		}
	}
	if hasGAllowance {
		hasSellerVATID := inv.Seller.VATaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerVATID {
			inv.addViolation(rules.BRG3, "Invoice with Export outside EU allowance must have seller VAT identifier")
		}
	}

	// BR-G-4 Export außerhalb der EU
	// If document level charge has "G", must have seller VAT ID
	hasGCharge := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "G" {
			hasGCharge = true
			break
		}
	}
	if hasGCharge {
		hasSellerVATID := inv.Seller.VATaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerVATID {
			inv.addViolation(rules.BRG4, "Invoice with Export outside EU charge must have seller VAT identifier")
		}
	}

	// BR-G-5 Export außerhalb der EU
	// In invoice line with "G", VAT rate must be 0
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "G" && !line.TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRG5, "Export outside EU invoice line must have VAT rate of 0")
		}
	}

	// BR-G-6 Export außerhalb der EU
	// In document level allowance with "G", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "G" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRG6, "Export outside EU allowance must have VAT rate of 0")
		}
	}

	// BR-G-7 Export außerhalb der EU
	// In document level charge with "G", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "G" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRG7, "Export outside EU charge must have VAT rate of 0")
		}
	}

	// BR-G-8 Export außerhalb der EU
	// Taxable amount must match calculated sum for Export outside EU category
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "G" {
			calculatedBasis := decimal.Zero
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "G" {
					calculatedBasis = calculatedBasis.Add(line.Total)
				}
			}
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "G" {
					if ac.ChargeIndicator {
						calculatedBasis = calculatedBasis.Add(ac.ActualAmount)
					} else {
						calculatedBasis = calculatedBasis.Sub(ac.ActualAmount)
					}
				}
			}
			calculatedBasis = calculatedBasis.Round(2)
			if !tt.BasisAmount.Equal(calculatedBasis) {
				inv.addViolation(rules.BRG8, fmt.Sprintf("Export outside EU taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), tt.BasisAmount.String()))
			}
		}
	}

	// BR-G-9 Export außerhalb der EU
	// VAT amount must be 0 for Export outside EU
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "G" && !tt.CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRG9, "Export outside EU VAT amount must be 0")
		}
	}

	// BR-G-10 Export außerhalb der EU
	// Export outside EU breakdown must have exemption reason code or text
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "G" && tt.ExemptionReason == "" && tt.ExemptionReasonCode == "" {
			inv.addViolation(rules.BRG10, "Export outside EU VAT breakdown must have exemption reason")
		}
	}
}
