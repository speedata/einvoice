package einvoice

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// checkVATReverse validates BR-AE-1 through BR-AE-10.
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
//
// Business rules implemented:
//   - BR-AE-1: VAT breakdown must exist for Reverse charge items
//   - BR-AE-2: Invoice line requires both seller and buyer VAT IDs
//   - BR-AE-3: Document allowance requires both seller and buyer VAT IDs
//   - BR-AE-4: Document charge requires both seller and buyer VAT IDs
//   - BR-AE-5: Line VAT rate must be 0
//   - BR-AE-6: Allowance VAT rate must be 0
//   - BR-AE-7: Charge VAT rate must be 0
//   - BR-AE-8: Taxable amount must match calculated sum
//   - BR-AE-9: VAT amount must be 0
//   - BR-AE-10: Must have exemption reason explaining reverse charge
func (inv *Invoice) checkVATReverse() {
	// BR-AE-1 Umkehrung der Steuerschuldnerschaft (Reverse charge)
	// If invoice has line/allowance/charge with "AE", must have at least one "AE" in VAT breakdown
	hasReverseCharge := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "AE" {
			hasReverseCharge = true
			break
		}
	}
	if !hasReverseCharge {
		for _, ac := range inv.SpecifiedTradeAllowanceCharge {
			if ac.CategoryTradeTaxCategoryCode == "AE" {
				hasReverseCharge = true
				break
			}
		}
	}
	if hasReverseCharge {
		hasAEInBreakdown := false
		for _, tt := range inv.TradeTaxes {
			if tt.CategoryCode == "AE" {
				hasAEInBreakdown = true
				break
			}
		}
		if !hasAEInBreakdown {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-1", InvFields: []string{"BG-23", "BT-118"}, Text: "Invoice with Reverse charge items must have Reverse charge VAT breakdown"})
		}
	}

	// BR-AE-2 Umkehrung der Steuerschuldnerschaft
	// If invoice line has "AE", must have seller VAT ID/tax reg/rep VAT ID AND buyer VAT ID
	hasAELine := false
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "AE" {
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
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-2", InvFields: []string{"BT-31", "BT-32", "BT-63", "BT-48"}, Text: "Invoice with Reverse charge line must have seller and buyer VAT identifiers"})
		}
	}

	// BR-AE-3 Umkehrung der Steuerschuldnerschaft
	// If document level allowance has "AE", must have seller and buyer tax IDs
	hasAEAllowance := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "AE" {
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
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-3", InvFields: []string{"BT-31", "BT-32", "BT-63", "BT-48"}, Text: "Invoice with Reverse charge allowance must have seller and buyer VAT identifiers"})
		}
	}

	// BR-AE-4 Umkehrung der Steuerschuldnerschaft
	// If document level charge has "AE", must have seller and buyer tax IDs
	hasAECharge := false
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "AE" {
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
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-4", InvFields: []string{"BT-31", "BT-32", "BT-63", "BT-48"}, Text: "Invoice with Reverse charge charge must have seller and buyer VAT identifiers"})
		}
	}

	// BR-AE-5 Umkehrung der Steuerschuldnerschaft
	// In invoice line with "AE", VAT rate must be 0
	for _, line := range inv.InvoiceLines {
		if line.TaxCategoryCode == "AE" && !line.TaxRateApplicablePercent.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-5", InvFields: []string{"BG-25", "BT-152"}, Text: "Reverse charge invoice line must have VAT rate of 0"})
		}
	}

	// BR-AE-6 Umkehrung der Steuerschuldnerschaft
	// In document level allowance with "AE", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if !ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "AE" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-6", InvFields: []string{"BG-20", "BT-96"}, Text: "Reverse charge allowance must have VAT rate of 0"})
		}
	}

	// BR-AE-7 Umkehrung der Steuerschuldnerschaft
	// In document level charge with "AE", VAT rate must be 0
	for _, ac := range inv.SpecifiedTradeAllowanceCharge {
		if ac.ChargeIndicator && ac.CategoryTradeTaxCategoryCode == "AE" && !ac.CategoryTradeTaxRateApplicablePercent.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-7", InvFields: []string{"BG-21", "BT-103"}, Text: "Reverse charge charge must have VAT rate of 0"})
		}
	}

	// BR-AE-8 Umkehrung der Steuerschuldnerschaft
	// Taxable amount must match calculated sum for Reverse charge category
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "AE" {
			calculatedBasis := decimal.Zero
			for _, line := range inv.InvoiceLines {
				if line.TaxCategoryCode == "AE" {
					calculatedBasis = calculatedBasis.Add(line.Total)
				}
			}
			for _, ac := range inv.SpecifiedTradeAllowanceCharge {
				if ac.CategoryTradeTaxCategoryCode == "AE" {
					if ac.ChargeIndicator {
						calculatedBasis = calculatedBasis.Add(ac.ActualAmount)
					} else {
						calculatedBasis = calculatedBasis.Sub(ac.ActualAmount)
					}
				}
			}
			calculatedBasis = calculatedBasis.Round(2)
			if !tt.BasisAmount.Equal(calculatedBasis) {
				inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-8", InvFields: []string{"BG-23", "BT-116"}, Text: fmt.Sprintf("Reverse charge taxable amount must equal sum of line amounts (expected %s, got %s)", calculatedBasis.String(), tt.BasisAmount.String())})
			}
		}
	}

	// BR-AE-9 Umkehrung der Steuerschuldnerschaft
	// VAT amount must be 0 for Reverse charge
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "AE" && !tt.CalculatedAmount.IsZero() {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-9", InvFields: []string{"BG-23", "BT-117"}, Text: "Reverse charge VAT amount must be 0"})
		}
	}

	// BR-AE-10 Umkehrung der Steuerschuldnerschaft
	// Reverse charge breakdown must have exemption reason code or text
	for _, tt := range inv.TradeTaxes {
		if tt.CategoryCode == "AE" && tt.ExemptionReason == "" && tt.ExemptionReasonCode == "" {
			inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-AE-10", InvFields: []string{"BG-23", "BT-120", "BT-121"}, Text: "Reverse charge VAT breakdown must have exemption reason"})
		}
	}
}
