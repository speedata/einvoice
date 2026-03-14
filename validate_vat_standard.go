package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// validateVATStandard validates BR-S-1 through BR-S-10.
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
func (inv *Invoice) validateVATStandard() {
	// BR-S-1 Umsatzsteuer mit Normalsatz
	// If invoice has line/allowance/charge with "Standard rated" (S), must have at least one "S" in VAT breakdown
	hasStandardRated := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "S" {
			hasStandardRated = true
			break
		}
	}
	if !hasStandardRated {
		for i := range inv.SpecifiedTradeAllowanceCharge {
			if inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "S" {
				hasStandardRated = true
				break
			}
		}
	}
	if hasStandardRated {
		hasStandardInBreakdown := false
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "S" {
				hasStandardInBreakdown = true
				break
			}
		}
		if !hasStandardInBreakdown {
			inv.addViolation(rules.BRS1, "Invoice with Standard rated items must have Standard rated VAT breakdown")
		}
	}

	// BR-S-2 Umsatzsteuer mit Normalsatz
	// If invoice line has "Standard rated", must have seller VAT ID, tax reg, or rep VAT ID
	hasStandardLine := false
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "S" {
			hasStandardLine = true
			break
		}
	}
	if hasStandardLine {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRS2, "Invoice with Standard rated line must have seller VAT identifier or tax registration")
		}
	}

	// BR-S-3 Umsatzsteuer mit Normalsatz
	// If document level allowance has "Standard rated", must have seller tax ID
	hasStandardAllowance := false
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "S" {
			hasStandardAllowance = true
			break
		}
	}
	if hasStandardAllowance {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRS3, "Invoice with Standard rated allowance must have seller VAT identifier or tax registration")
		}
	}

	// BR-S-4 Umsatzsteuer mit Normalsatz
	// If document level charge has "Standard rated", must have seller tax ID
	hasStandardCharge := false
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "S" {
			hasStandardCharge = true
			break
		}
	}
	if hasStandardCharge {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.addViolation(rules.BRS4, "Invoice with Standard rated charge must have seller VAT identifier or tax registration")
		}
	}

	// BR-S-5 Umsatzsteuer mit Normalsatz
	// In invoice line with "Standard rated", VAT rate must be > 0
	for i := range inv.InvoiceLines {
		if inv.InvoiceLines[i].TaxCategoryCode == "S" && !inv.InvoiceLines[i].TaxRateApplicablePercent.IsPositive() {
			inv.addViolation(rules.BRS5, "Standard rated invoice line must have VAT rate greater than 0")
		}
	}

	// BR-S-6 Umsatzsteuer mit Normalsatz
	// In document level allowance with "Standard rated", VAT rate must be > 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if !inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "S" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsPositive() {
			inv.addViolation(rules.BRS6, "Standard rated allowance must have VAT rate greater than 0")
		}
	}

	// BR-S-7 Umsatzsteuer mit Normalsatz
	// In document level charge with "Standard rated", VAT rate must be > 0
	for i := range inv.SpecifiedTradeAllowanceCharge {
		if inv.SpecifiedTradeAllowanceCharge[i].ChargeIndicator && inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode == "S" && !inv.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxRateApplicablePercent.IsPositive() {
			inv.addViolation(rules.BRS7, "Standard rated charge must have VAT rate greater than 0")
		}
	}

	// BR-S-8 Umsatzsteuer mit Normalsatz
	// For each distinct rate in Standard rated category, taxable amount must match calculated sum
	// Note: This validation only applies to profiles with line items (>= Basic, level 3).
	// BasicWL profile (level 2) provides BasisAmount directly without line items.
	if inv.ProfileLevel() >= levelBasic || (inv.ProfileLevel() == 0 && len(inv.InvoiceLines) > 0) {
		for i := range inv.TradeTaxes {
			if inv.TradeTaxes[i].CategoryCode == "S" {
				// Calculate sum: lines - allowances + charges for this rate
				calculatedBasis := decimal.Zero
				for j := range inv.InvoiceLines {
					if inv.InvoiceLines[j].TaxCategoryCode == "S" && inv.InvoiceLines[j].TaxRateApplicablePercent.Equal(inv.TradeTaxes[i].Percent) {
						calculatedBasis = calculatedBasis.Add(inv.InvoiceLines[j].Total)
					}
				}
				for j := range inv.SpecifiedTradeAllowanceCharge {
					if inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxCategoryCode == "S" && inv.SpecifiedTradeAllowanceCharge[j].CategoryTradeTaxRateApplicablePercent.Equal(inv.TradeTaxes[i].Percent) {
						if inv.SpecifiedTradeAllowanceCharge[j].ChargeIndicator {
							calculatedBasis = calculatedBasis.Add(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						} else {
							calculatedBasis = calculatedBasis.Sub(inv.SpecifiedTradeAllowanceCharge[j].ActualAmount)
						}
					}
				}
				// Round to 2 decimals for comparison using commercial rounding (round half up)
				calculatedBasis = roundHalfUp(calculatedBasis, 2)
				if !inv.TradeTaxes[i].BasisAmount.Equal(calculatedBasis) {
					inv.addViolation(rules.BRS8, fmt.Sprintf("Standard rated taxable amount must equal sum of line amounts for rate %s (expected %s, got %s)", inv.TradeTaxes[i].Percent.String(), calculatedBasis.String(), inv.TradeTaxes[i].BasisAmount.String()))
				}
			}
		}
	}

	// BR-S-9 Umsatzsteuer mit Normalsatz
	// VAT amount must equal taxable amount * rate
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "S" {
			expectedVAT := roundHalfUp(inv.TradeTaxes[i].BasisAmount.Mul(inv.TradeTaxes[i].Percent).Div(decimal100), 2)
			if !inv.TradeTaxes[i].CalculatedAmount.Equal(expectedVAT) {
				inv.addViolation(rules.BRS9, fmt.Sprintf("Standard rated VAT amount must equal basis * rate (expected %s, got %s)", expectedVAT.String(), inv.TradeTaxes[i].CalculatedAmount.String()))
			}
		}
	}

	// BR-S-10 Umsatzsteuer mit Normalsatz
	// Standard rated breakdown must not have exemption reason or code
	for i := range inv.TradeTaxes {
		if inv.TradeTaxes[i].CategoryCode == "S" && (inv.TradeTaxes[i].ExemptionReason != "" || inv.TradeTaxes[i].ExemptionReasonCode != "") {
			inv.addViolation(rules.BRS10, "Standard rated VAT breakdown must not have exemption reason")
		}
	}
}
