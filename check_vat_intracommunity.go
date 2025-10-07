package einvoice

import (
	"github.com/speedata/einvoice/rules"
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATIntracommunity validates BR-IC-1 through BR-IC-12.
//
// These rules apply to invoices with Intra-community supply VAT (category code 'K').
// This category is for goods or services traded between EU member states.
// Under EU VAT rules, the supplier typically does not charge VAT, and the buyer
// accounts for VAT through the reverse charge mechanism.
//
// Key requirements for Intra-community supply:
//   - Both seller and buyer must have VAT identifiers
//   - VAT rate must be 0 (reverse charge applies)
//   - VAT amount must be 0 in the invoice
//   - Must have exemption reason explaining the intra-community nature
//   - Must have actual delivery date or invoicing period
//   - Must have deliver to country code
func (inv *Invoice) checkVATIntracommunity() {
	// BR-IC-1 Innergemeinschaftliche Lieferung (Intra-community supply)
	// Invoice with category K must have both seller and buyer VAT IDs
	hasIntracommunitySupply := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "K" {
			hasIntracommunitySupply = true
			break
		}
	}
	if !hasIntracommunitySupply {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "K" {
				hasIntracommunitySupply = true
				break
			}
		}
	}
	if hasIntracommunitySupply {
		// Check seller VAT ID
		hasSellerTaxID := inv.Seller.VATaxRegistration != "" ||
			inv.Seller.FCTaxRegistration != "" ||
			(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
		// Check buyer VAT ID
		hasBuyerVATID := inv.Buyer.VATaxRegistration != ""
		// Also check for buyer legal registration
		hasBuyerLegalID := inv.Buyer.SpecifiedLegalOrganization != nil && inv.Buyer.SpecifiedLegalOrganization.ID != ""

		if !hasSellerTaxID {
			inv.addViolation(rules.BRIC1, "Intra-community supply requires seller VAT identifier")
		}
		if !hasBuyerVATID && !hasBuyerLegalID {
			inv.addViolation(rules.BRIC1, "Intra-community supply requires buyer VAT identifier or legal registration identifier")
		}
	}

	// BR-IC-2 Innergemeinschaftliche Lieferung
	// Lines with category K require seller and buyer VAT IDs
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "K" {
			hasSellerVATID := inv.Seller.VATaxRegistration != "" ||
				inv.Seller.FCTaxRegistration != "" ||
				(inv.SellerTaxRepresentativeTradeParty != nil && inv.SellerTaxRepresentativeTradeParty.VATaxRegistration != "")
			hasBuyerVATID := inv.Buyer.VATaxRegistration != ""

			if !hasSellerVATID {
				inv.addViolation(rules.BRIC2, "Intra-community supply line requires seller VAT identifier")
			}
			if !hasBuyerVATID {
				inv.addViolation(rules.BRIC2, "Intra-community supply line requires buyer VAT identifier")
			}
			break
		}
	}

	// BR-IC-3 Innergemeinschaftliche Lieferung
	// VAT rate must be 0 for lines with category K
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "K" && !line.TaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRIC3, "Intra-community supply VAT rate must be 0")
		}
	}

	// BR-IC-4 Innergemeinschaftliche Lieferung
	// VAT rate must be 0 for allowances with category K
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "K" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRIC4, "Intra-community supply allowance VAT rate must be 0")
		}
	}

	// BR-IC-5 Innergemeinschaftliche Lieferung
	// VAT rate must be 0 for charges with category K
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "K" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.addViolation(rules.BRIC5, "Intra-community supply charge VAT rate must be 0")
		}
	}

	// BR-IC-6 Innergemeinschaftliche Lieferung
	// Verify taxable amount calculation for category K
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "K" {
			var lineTotal decimal.Decimal
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "K" {
					lineTotal = lineTotal.Add(line.Total)
				}
			}
			var allowanceTotal decimal.Decimal
			var chargeTotal decimal.Decimal
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "K" {
					if ac.ChargeIndicator {
						chargeTotal = chargeTotal.Add(ac.ActualAmount)
					} else {
						allowanceTotal = allowanceTotal.Add(ac.ActualAmount)
					}
				}
			}
			expectedBasis := lineTotal.Sub(allowanceTotal).Add(chargeTotal)
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.addViolation(rules.BRIC6, fmt.Sprintf("Intra-community supply taxable amount mismatch: got %s, expected %s", tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
			}
		}
	}

	// BR-IC-7 Innergemeinschaftliche Lieferung
	// VAT amount must be 0 for category K
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "K" && !tt.CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRIC7, "Intra-community supply VAT amount must be 0")
		}
	}

	// BR-IC-8 Innergemeinschaftliche Lieferung
	// For each different VAT rate, verify taxable amount calculation
	taxRateMap := make(map[string]decimal.Decimal)
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "K" {
			key := line.TaxRateApplicablePercent.String()
			taxRateMap[key] = taxRateMap[key].Add(line.Total)
		}
	}
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.CategoryTradeTaxCategoryCode == "K" {
			key := ac.CategoryTradeTaxRateApplicablePercent.String()
			if ac.ChargeIndicator {
				taxRateMap[key] = taxRateMap[key].Add(ac.ActualAmount)
			} else {
				taxRateMap[key] = taxRateMap[key].Sub(ac.ActualAmount)
			}
		}
	}
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "K" {
			key := tt.Percent.String()
			expectedBasis := taxRateMap[key]
			if !tt.BasisAmount.Equal(expectedBasis) {
				inv.addViolation(rules.BRIC8, fmt.Sprintf("Intra-community supply taxable amount for rate %s: got %s, expected %s", tt.Percent.StringFixed(2), tt.BasisAmount.StringFixed(2), expectedBasis.StringFixed(2)))
			}
		}
	}

	// BR-IC-9 Innergemeinschaftliche Lieferung
	// VAT amount must be 0 for category K (duplicate of BR-IC-7, but specified separately in spec)
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "K" && !tt.CalculatedAmount.IsZero() {
			inv.addViolation(rules.BRIC9, "Intra-community supply VAT amount must be 0")
		}
	}

	// BR-IC-10 Innergemeinschaftliche Lieferung
	// Intra-community supply breakdown must have exemption reason code or text
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "K" && tt.ExemptionReason == "" && tt.ExemptionReasonCode == "" {
			inv.addViolation(rules.BRIC10, "Intra-community supply VAT breakdown must have exemption reason")
		}
	}

	// BR-IC-11 Innergemeinschaftliche Lieferung
	// Must have actual delivery date or invoicing period
	hasIntraCommunityInVATBreakdown := false
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "K" {
			hasIntraCommunityInVATBreakdown = true
			break
		}
	}
	if hasIntraCommunityInVATBreakdown {
		hasDeliveryDate := !inv.OccurrenceDateTime.IsZero()
		hasBillingPeriod := !inv.BillingSpecifiedPeriodStart.IsZero() || !inv.BillingSpecifiedPeriodEnd.IsZero()
		if !hasDeliveryDate && !hasBillingPeriod {
			inv.addViolation(rules.BRIC11, "Intra-community supply requires actual delivery date or invoicing period")
		}
	}

	// BR-IC-12 Innergemeinschaftliche Lieferung
	// Must have deliver to country code
	if hasIntraCommunityInVATBreakdown {
		hasDeliverToCountry := inv.ShipTo != nil && inv.ShipTo.PostalAddress != nil && inv.ShipTo.PostalAddress.CountryID != ""
		if !hasDeliverToCountry {
			inv.addViolation(rules.BRIC12, "Intra-community supply requires deliver to country code")
		}
	}
}
