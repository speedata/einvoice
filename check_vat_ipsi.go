package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATIPSI validates BR-IP-1 through BR-IP-10.
//
// These rules apply to invoices with IPSI tax (category code 'M').
// IPSI (Impuesto sobre la Producción, los Servicios y la Importación) is
// the production, services, and import tax applicable in Ceuta and Melilla
// instead of VAT. Similar to IGIC, it operates as a regional replacement
// for VAT with its own rates and rules.
//
// Key requirements for IPSI:
//   - Must have at least one VAT breakdown entry with category 'M'
//   - Seller must have a tax identifier
//   - IPSI rate can be 0 or greater (various rates apply)
//   - IPSI amount is calculated as basis × rate
//   - Must NOT have exemption reason (not an exemption, it's a different tax)
//   - Seller must have tax ID but buyer must NOT have VAT ID
func (inv *Invoice) checkVATIPSI() {
	// BR-IP-1 IPSI (Ceuta/Melilla)
	// Invoice with category M must have seller VAT ID
	hasIPSI := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "M" {
			hasIPSI = true
			break
		}
	}
	if !hasIPSI {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "M" {
				hasIPSI = true
				break
			}
		}
	}
	if hasIPSI {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		if !hasSellerTaxID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-1", InvFields: []string{"BT-31", "BT-32", "BT-63"}, Text: "IPSI requires seller VAT identifier"})
		}
	}

	// BR-IP-2 IPSI
	// VAT rate must be 0 or greater for lines with category M (no validation needed - rate >= 0 is implicit)

	// BR-IP-3 IPSI
	// VAT rate must be 0 or greater for allowances with category M (no validation needed - rate >= 0 is implicit)

	// BR-IP-4 IPSI
	// VAT rate must be 0 or greater for charges with category M (no validation needed - rate >= 0 is implicit)

	// BR-IP-5 IPSI
	// Verify taxable amount calculation for category M
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "M" {
			var lineTotal decimal.Decimal
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "M" {
					lineTotal = lineTotal.Add(line.Total)
				}
			}
			var allowanceTotal decimal.Decimal
			var chargeTotal decimal.Decimal
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "M" {
					if ac.ChargeIndicator {
						chargeTotal = chargeTotal.Add(ac.ActualAmount)
					} else {
						allowanceTotal = allowanceTotal.Add(ac.ActualAmount)
					}
				}
			}
			expectedBasis := lineTotal.Sub(allowanceTotal).Add(chargeTotal)
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-5", InvFields: []string{"BT-116"}, Text: fmt.Sprintf("IPSI taxable amount mismatch: got %s, expected %s", tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2))})
			}
		}
	}

	// BR-IP-6 IPSI
	// VAT amount must equal basis * rate
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "M" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-6", InvFields: []string{"BT-117"}, Text: fmt.Sprintf("IPSI VAT amount must equal basis * rate: got %s, expected %s", tt.CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2))})
			}
		}
	}

	// BR-IP-7 IPSI
	// For each different VAT rate, verify taxable amount calculation
	ipsiRateMap := make(map[string]decimal.Decimal)
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "M" {
			key := line.TaxRateApplicablePercent.String()
			ipsiRateMap[key] = ipsiRateMap[key].Add(line.Total)
		}
	}
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.CategoryTradeTaxCategoryCode == "M" {
			key := ac.CategoryTradeTaxRateApplicablePercent.String()
			if ac.ChargeIndicator {
				ipsiRateMap[key] = ipsiRateMap[key].Add(ac.ActualAmount)
			} else {
				ipsiRateMap[key] = ipsiRateMap[key].Sub(ac.ActualAmount)
			}
		}
	}
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "M" {
			key := tt.Percent.String()
			expectedBasis := ipsiRateMap[key]
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-7", InvFields: []string{"BT-116"}, Text: fmt.Sprintf("IPSI taxable amount for rate %s: got %s, expected %s", tt.Percent.StringFixed(2), tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2))})
			}
		}
	}

	// BR-IP-8 IPSI
	// For each different VAT rate, verify VAT amount calculation
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "M" {
			expectedVAT := tt.BasisAmount.Mul(tt.Percent).Div(decimal.NewFromInt(100)).Round(2)
			if !tt.CalculatedAmount.Equal(expectedVAT) {
				inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-8", InvFields: []string{"BT-117"}, Text: fmt.Sprintf("IPSI VAT amount for rate %s must equal basis * rate: got %s, expected %s", tt.Percent.StringFixed(2), tt.CalculatedAmount.StringFixed(2), expectedVAT.StringFixed(2))})
			}
		}
	}

	// BR-IP-9 IPSI
	// IPSI breakdown must NOT have exemption reason
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "M" && (tt.ExemptionReason != "" || tt.ExemptionReasonCode != "") {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-9", InvFields: []string{"BG-23", "BT-120", "BT-121"}, Text: "IPSI VAT breakdown must not have exemption reason"})
		}
	}

	// BR-IP-10 IPSI
	// Must have seller tax ID and must NOT have buyer VAT ID
	hasIPSIInVATBreakdown := false
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "M" {
			hasIPSIInVATBreakdown = true
			break
		}
	}
	if hasIPSIInVATBreakdown {
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" || inv.Seller.FCTaxRegistration != ""
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""

		if !hasSellerTaxID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-10", InvFields: []string{"BT-31", "BT-32"}, Text: "IPSI requires seller VAT or tax registration identifier"})
		}
		if hasBuyerVATID {
			inv.violations = append(inv.violations, SemanticError{Rule: "BR-IP-10", InvFields: []string{"BT-48"}, Text: "IPSI must not have buyer VAT identifier"})
		}
	}
}
