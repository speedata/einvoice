package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATZero validates BR-Z-1 through BR-Z-10.
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
func (inv *Invoice) validateVATZero() {
	// BR-Z-1 Umsatzsteuer mit Nullsatz (Zero rated)
	// If invoice has line/allowance/charge with "Z", must have at least one "Z" in VAT breakdown
	hasZeroRated := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "Z" {
			hasZeroRated = true
			break
		}
	}
	if !hasZeroRated {
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "Z" {
				hasZeroRated = true
				break
			}
		}
	}
	if hasZeroRated {
		hasZInBreakdown := false
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "Z" {
				hasZInBreakdown = true
				break
			}
		}
		if !hasZInBreakdown {
			inv.addViolation(rules.BRZ1, "Invoice with Zero rated items must have Zero rated VAT breakdown")
		}
	}

	// BR-Z-2 Umsatzsteuer mit Nullsatz
	// If invoice line has "Z", must have seller VAT ID or tax registration or representative VAT ID
	hasZLine := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "Z" {
			hasZLine = true
			break
		}
	}
	if hasZLine {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRZ2, "Invoice with Zero rated line must have seller VAT identifier or tax registration")
		}
	}

	// BR-Z-3 Umsatzsteuer mit Nullsatz
	// If document level allowance has "Z", must have seller tax ID
	hasZAllowance := false
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "Z" {
			hasZAllowance = true
			break
		}
	}
	if hasZAllowance {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRZ3, "Invoice with Zero rated allowance must have seller VAT identifier or tax registration")
		}
	}

	// BR-Z-4 Umsatzsteuer mit Nullsatz
	// If document level charge has "Z", must have seller tax ID
	hasZCharge := false
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "Z" {
			hasZCharge = true
			break
		}
	}
	if hasZCharge {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRZ4, "Invoice with Zero rated charge must have seller VAT identifier or tax registration")
		}
	}

	// BR-Z-5 Umsatzsteuer mit Nullsatz
	// In invoice line with "Z", VAT rate must be 0
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "Z" && !inv.InvoiceLines[i].TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRZ5, "Zero rated invoice line must have VAT rate of 0")
		}
	}

	// BR-Z-6 Umsatzsteuer mit Nullsatz
	// In document level allowance with "Z", VAT rate must be 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "Z" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRZ6, "Zero rated allowance must have VAT rate of 0")
		}
	}

	// BR-Z-7 Umsatzsteuer mit Nullsatz
	// In document level charge with "Z", VAT rate must be 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "Z" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRZ7, "Zero rated charge must have VAT rate of 0")
		}
	}

	// BR-Z-8 Umsatzsteuer mit Nullsatz
	// Taxable amount must match calculated sum for Zero rated category
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "Z" {
				calculatedBasis := decimal.Zero
				for j := range inv.InvoiceLines {
					if inv.InvoiceLines[j].TaxCategoryCode == "Z" {
						calculatedBasis = calculatedBasis.Add(inv.InvoiceLines[j].Total)
					}
				}
				for j := range inv.SpecifiedTradeAllowanceCharge {
					if inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxCategoryCode == "Z" {
						if inv.SpecifiedTradeAllowanceCharge[j].ChargeIndicator {
							calculatedBasis = calculatedBasis.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						} else {
							calculatedBasis = calculatedBasis.Sub(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						}
					}
				}
				calculatedBasis = roundHalfUp(calculatedBasis, 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(calculatedBasis) {
					inv.addViolation(rules.BRZ8, fmt.Sprintf("Zero rated taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), inv.TradeTaxes[i].BasisAmount.String()))
				}
			}
		}
	}

	// BR-Z-9 Umsatzsteuer mit Nullsatz
	// VAT amount must be 0 for Zero rated
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "Z" && !inv.TradeTaxes[i].CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRZ9, "Zero rated VAT amount must be 0")
		}
	}

	// BR-Z-10 Umsatzsteuer mit Nullsatz
	// Zero rated breakdown must not have exemption reason code or text
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "Z" && (inv.TradeTaxes[i].ExemptionReason != "" || inv.TradeTaxes[i].ExemptionReasonCode != "") {
			inv.addViolation(rules.BRZ10, "Zero rated VAT breakdown must not have exemption reason")
		}
	}
}
