package einvoice

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice/rules"
)

// TestGermanValidation_BRDE1_PaymentInstructions tests BR-DE-1:
// An invoice must contain information on PAYMENT INSTRUCTIONS (BG-16).
func TestGermanValidation_BRDE1_PaymentInstructions(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Invoice)
		wantViol bool
	}{
		{
			name: "valid: has payment means",
			setup: func(inv *Invoice) {
				inv.PaymentMeans = []PaymentMeans{{TypeCode: 58}}
			},
			wantViol: false,
		},
		{
			name: "invalid: no payment means",
			setup: func(inv *Invoice) {
				inv.PaymentMeans = []PaymentMeans{}
			},
			wantViol: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			tt.setup(inv)

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE1)

			if hasViolation != tt.wantViol {
				t.Errorf("BR-DE-1 violation = %v, want %v", hasViolation, tt.wantViol)
			}
		})
	}
}

// TestGermanValidation_BRDE2_SellerContact tests BR-DE-2:
// The element group SELLER CONTACT (BG-6) must be transmitted.
func TestGermanValidation_BRDE2_SellerContact(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Invoice)
		wantViol bool
	}{
		{
			name: "valid: has seller contact",
			setup: func(inv *Invoice) {
				inv.Seller.DefinedTradeContact = []DefinedTradeContact{{PersonName: "John Doe"}}
			},
			wantViol: false,
		},
		{
			name: "invalid: no seller contact",
			setup: func(inv *Invoice) {
				inv.Seller.DefinedTradeContact = []DefinedTradeContact{}
			},
			wantViol: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			tt.setup(inv)

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE2)

			if hasViolation != tt.wantViol {
				t.Errorf("BR-DE-2 violation = %v, want %v", hasViolation, tt.wantViol)
			}
		})
	}
}

// TestGermanValidation_BRDE3_4_SellerAddress tests BR-DE-3 and BR-DE-4:
// Seller city (BT-37) and post code (BT-38) must be transmitted.
func TestGermanValidation_BRDE3_4_SellerAddress(t *testing.T) {
	tests := []struct {
		name         string
		city         string
		postcode     string
		wantBRDE3    bool
		wantBRDE4    bool
	}{
		{
			name:      "valid: has city and postcode",
			city:      "Berlin",
			postcode:  "10115",
			wantBRDE3: false,
			wantBRDE4: false,
		},
		{
			name:      "invalid: missing city",
			city:      "",
			postcode:  "10115",
			wantBRDE3: true,
			wantBRDE4: false,
		},
		{
			name:      "invalid: missing postcode",
			city:      "Berlin",
			postcode:  "",
			wantBRDE3: false,
			wantBRDE4: true,
		},
		{
			name:      "invalid: missing both",
			city:      "",
			postcode:  "",
			wantBRDE3: true,
			wantBRDE4: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			inv.Seller.PostalAddress.City = tt.city
			inv.Seller.PostalAddress.PostcodeCode = tt.postcode

			err := inv.Validate()
			hasBRDE3 := hasRuleViolation(err, rules.BRDE3)
			hasBRDE4 := hasRuleViolation(err, rules.BRDE4)

			if hasBRDE3 != tt.wantBRDE3 {
				t.Errorf("BR-DE-3 violation = %v, want %v", hasBRDE3, tt.wantBRDE3)
			}
			if hasBRDE4 != tt.wantBRDE4 {
				t.Errorf("BR-DE-4 violation = %v, want %v", hasBRDE4, tt.wantBRDE4)
			}
		})
	}
}

// TestGermanValidation_BRDE5_6_7_SellerContactDetails tests BR-DE-5, BR-DE-6, BR-DE-7:
// Seller contact point (BT-41), telephone (BT-42), and email (BT-43) must be transmitted.
func TestGermanValidation_BRDE5_6_7_SellerContactDetails(t *testing.T) {
	tests := []struct {
		name      string
		personName string
		phone     string
		email     string
		wantBRDE5 bool
		wantBRDE6 bool
		wantBRDE7 bool
	}{
		{
			name:       "valid: all contact details present",
			personName: "John Doe",
			phone:      "+49 30 1234567",
			email:      "john@example.com",
			wantBRDE5:  false,
			wantBRDE6:  false,
			wantBRDE7:  false,
		},
		{
			name:       "invalid: missing contact point",
			personName: "",
			phone:      "+49 30 1234567",
			email:      "john@example.com",
			wantBRDE5:  true,
			wantBRDE6:  false,
			wantBRDE7:  false,
		},
		{
			name:       "invalid: missing phone",
			personName: "John Doe",
			phone:      "",
			email:      "john@example.com",
			wantBRDE5:  false,
			wantBRDE6:  true,
			wantBRDE7:  false,
		},
		{
			name:       "invalid: missing email",
			personName: "John Doe",
			phone:      "+49 30 1234567",
			email:      "",
			wantBRDE5:  false,
			wantBRDE6:  false,
			wantBRDE7:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			if len(inv.Seller.DefinedTradeContact) > 0 {
				inv.Seller.DefinedTradeContact[0].PersonName = tt.personName
				inv.Seller.DefinedTradeContact[0].PhoneNumber = tt.phone
				inv.Seller.DefinedTradeContact[0].EMail = tt.email
			}

			err := inv.Validate()
			hasBRDE5 := hasRuleViolation(err, rules.BRDE5)
			hasBRDE6 := hasRuleViolation(err, rules.BRDE6)
			hasBRDE7 := hasRuleViolation(err, rules.BRDE7)

			if hasBRDE5 != tt.wantBRDE5 {
				t.Errorf("BR-DE-5 violation = %v, want %v", hasBRDE5, tt.wantBRDE5)
			}
			if hasBRDE6 != tt.wantBRDE6 {
				t.Errorf("BR-DE-6 violation = %v, want %v", hasBRDE6, tt.wantBRDE6)
			}
			if hasBRDE7 != tt.wantBRDE7 {
				t.Errorf("BR-DE-7 violation = %v, want %v", hasBRDE7, tt.wantBRDE7)
			}
		})
	}
}

// TestGermanValidation_BRDE8_9_BuyerAddress tests BR-DE-8 and BR-DE-9:
// Buyer city (BT-52) and post code (BT-53) must be transmitted.
func TestGermanValidation_BRDE8_9_BuyerAddress(t *testing.T) {
	tests := []struct {
		name      string
		city      string
		postcode  string
		wantBRDE8 bool
		wantBRDE9 bool
	}{
		{
			name:      "valid: has city and postcode",
			city:      "Munich",
			postcode:  "80331",
			wantBRDE8: false,
			wantBRDE9: false,
		},
		{
			name:      "invalid: missing city",
			city:      "",
			postcode:  "80331",
			wantBRDE8: true,
			wantBRDE9: false,
		},
		{
			name:      "invalid: missing postcode",
			city:      "Munich",
			postcode:  "",
			wantBRDE8: false,
			wantBRDE9: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			inv.Buyer.PostalAddress.City = tt.city
			inv.Buyer.PostalAddress.PostcodeCode = tt.postcode

			err := inv.Validate()
			hasBRDE8 := hasRuleViolation(err, rules.BRDE8)
			hasBRDE9 := hasRuleViolation(err, rules.BRDE9)

			if hasBRDE8 != tt.wantBRDE8 {
				t.Errorf("BR-DE-8 violation = %v, want %v", hasBRDE8, tt.wantBRDE8)
			}
			if hasBRDE9 != tt.wantBRDE9 {
				t.Errorf("BR-DE-9 violation = %v, want %v", hasBRDE9, tt.wantBRDE9)
			}
		})
	}
}

// TestGermanValidation_BRDE10_11_DeliveryAddress tests BR-DE-10 and BR-DE-11:
// Deliver to city (BT-77) and post code (BT-78) must be transmitted if delivery address is provided.
func TestGermanValidation_BRDE10_11_DeliveryAddress(t *testing.T) {
	tests := []struct {
		name       string
		hasShipTo  bool
		city       string
		postcode   string
		wantBRDE10 bool
		wantBRDE11 bool
	}{
		{
			name:       "valid: no delivery address",
			hasShipTo:  false,
			wantBRDE10: false,
			wantBRDE11: false,
		},
		{
			name:       "valid: delivery address with city and postcode",
			hasShipTo:  true,
			city:       "Hamburg",
			postcode:   "20095",
			wantBRDE10: false,
			wantBRDE11: false,
		},
		{
			name:       "invalid: delivery address missing city",
			hasShipTo:  true,
			city:       "",
			postcode:   "20095",
			wantBRDE10: true,
			wantBRDE11: false,
		},
		{
			name:       "invalid: delivery address missing postcode",
			hasShipTo:  true,
			city:       "Hamburg",
			postcode:   "",
			wantBRDE10: false,
			wantBRDE11: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			if tt.hasShipTo {
				inv.ShipTo = &Party{
					PostalAddress: &PostalAddress{
						City:         tt.city,
						PostcodeCode: tt.postcode,
						CountryID:    "DE",
					},
				}
			} else {
				inv.ShipTo = nil
			}

			err := inv.Validate()
			hasBRDE10 := hasRuleViolation(err, rules.BRDE10)
			hasBRDE11 := hasRuleViolation(err, rules.BRDE11)

			if hasBRDE10 != tt.wantBRDE10 {
				t.Errorf("BR-DE-10 violation = %v, want %v", hasBRDE10, tt.wantBRDE10)
			}
			if hasBRDE11 != tt.wantBRDE11 {
				t.Errorf("BR-DE-11 violation = %v, want %v", hasBRDE11, tt.wantBRDE11)
			}
		})
	}
}

// TestGermanValidation_BRDE15_BuyerReference tests BR-DE-15:
// Buyer reference (BT-10) must be transmitted (Leitweg-ID).
func TestGermanValidation_BRDE15_BuyerReference(t *testing.T) {
	tests := []struct {
		name          string
		buyerRef      string
		wantViolation bool
	}{
		{
			name:          "valid: has buyer reference",
			buyerRef:      "04011000-12345-35",
			wantViolation: false,
		},
		{
			name:          "invalid: missing buyer reference",
			buyerRef:      "",
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			inv.BuyerReference = tt.buyerRef

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE15)

			if hasViolation != tt.wantViolation {
				t.Errorf("BR-DE-15 violation = %v, want %v", hasViolation, tt.wantViolation)
			}
		})
	}
}

// TestGermanValidation_BRDE16_SellerIdentification tests BR-DE-16:
// When tax codes S, Z, E, AE, K, G, L or M are used, at least one of
// Seller VAT identifier (BT-31), Seller tax registration identifier (BT-32)
// or SELLER TAX REPRESENTATIVE PARTY (BG-11) must be provided.
func TestGermanValidation_BRDE16_SellerIdentification(t *testing.T) {
	tests := []struct {
		name          string
		taxCode       string
		hasVATID      bool
		hasTaxReg     bool
		hasTaxRep     bool
		wantViolation bool
	}{
		{
			name:          "valid: has VAT ID with tax code S",
			taxCode:       "S",
			hasVATID:      true,
			hasTaxReg:     false,
			hasTaxRep:     false,
			wantViolation: false,
		},
		{
			name:          "valid: has tax registration with tax code E",
			taxCode:       "E",
			hasVATID:      false,
			hasTaxReg:     true,
			hasTaxRep:     false,
			wantViolation: false,
		},
		{
			name:          "valid: has tax representative with tax code Z",
			taxCode:       "Z",
			hasVATID:      false,
			hasTaxReg:     false,
			hasTaxRep:     true,
			wantViolation: false,
		},
		{
			name:          "invalid: no identification with tax code S",
			taxCode:       "S",
			hasVATID:      false,
			hasTaxReg:     false,
			hasTaxRep:     false,
			wantViolation: true,
		},
		{
			name:          "invalid: no identification with tax code AE",
			taxCode:       "AE",
			hasVATID:      false,
			hasTaxReg:     false,
			hasTaxRep:     false,
			wantViolation: true,
		},
		{
			name:          "valid: no identification with tax code O (not relevant)",
			taxCode:       "O",
			hasVATID:      false,
			hasTaxReg:     false,
			hasTaxRep:     false,
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()

			// Add invoice lines and trade taxes with the specified tax code
			inv.InvoiceLines = []InvoiceLine{
				{
					LineID:                   "1",
					ItemName:                 "Test Item",
					BilledQuantity:           decimal.NewFromInt(1),
					BilledQuantityUnit:       "C62",
					NetPrice:                 decimal.NewFromInt(100),
					Total:                    decimal.NewFromInt(100),
					TaxCategoryCode:          tt.taxCode,
					TaxRateApplicablePercent: decimal.NewFromInt(19),
				},
			}
			inv.TradeTaxes = []TradeTax{
				{
					CategoryCode:     tt.taxCode,
					Percent:          decimal.NewFromInt(19),
					BasisAmount:      decimal.NewFromInt(100),
					CalculatedAmount: decimal.NewFromInt(19),
				},
			}
			inv.LineTotal = decimal.NewFromInt(100)
			inv.TaxBasisTotal = decimal.NewFromInt(100)
			inv.TaxTotal = decimal.NewFromInt(19)
			inv.GrandTotal = decimal.NewFromInt(119)
			inv.DuePayableAmount = decimal.NewFromInt(119)

			// Set seller identifications based on test case
			if tt.hasVATID {
				inv.Seller.VATaxRegistration = "DE123456789"
			} else {
				inv.Seller.VATaxRegistration = ""
			}

			if tt.hasTaxReg {
				inv.Seller.FCTaxRegistration = "12345678"
			} else {
				inv.Seller.FCTaxRegistration = ""
			}

			if tt.hasTaxRep {
				inv.SellerTaxRepresentativeTradeParty = &Party{
					Name: "Tax Rep",
					VATaxRegistration: "FR12345678901",
					PostalAddress: &PostalAddress{
						CountryID: "FR",
					},
				}
			} else {
				inv.SellerTaxRepresentativeTradeParty = nil
			}

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE16)

			if hasViolation != tt.wantViolation {
				t.Errorf("BR-DE-16 violation = %v, want %v", hasViolation, tt.wantViolation)
				if err != nil {
					t.Logf("Validation error: %v", err)
				}
			}
		})
	}
}

// TestGermanValidation_BRDE21_SpecificationIdentifier tests BR-DE-21:
// Specification identifier must match XRechnung standard for German sellers.
func TestGermanValidation_BRDE21_SpecificationIdentifier(t *testing.T) {
	tests := []struct {
		name          string
		specID        string
		sellerCountry string
		wantViolation bool
	}{
		{
			name:          "valid: XRechnung 3.0 for DE seller",
			specID:        SpecXRechnung30,
			sellerCountry: "DE",
			wantViolation: false,
		},
		{
			name:          "valid: Factur-X Extended for DE seller (BR-DE rules don't apply)",
			specID:        SpecFacturXExtended,
			sellerCountry: "DE",
			wantViolation: false, // BR-DE-21 only applies to XRechnung invoices
		},
		{
			name:          "valid: Factur-X Extended for FR seller",
			specID:        SpecFacturXExtended,
			sellerCountry: "FR",
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			inv.GuidelineSpecifiedDocumentContextParameter = tt.specID
			inv.Seller.PostalAddress.CountryID = tt.sellerCountry

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE21)

			if hasViolation != tt.wantViolation {
				t.Errorf("BR-DE-21 violation = %v, want %v", hasViolation, tt.wantViolation)
			}
		})
	}
}

// TestGermanValidation_BRDE19_20_IBANValidation tests BR-DE-19 and BR-DE-20:
// IBAN validation for SEPA credit transfer and direct debit.
func TestGermanValidation_BRDE19_20_IBANValidation(t *testing.T) {
	tests := []struct {
		name          string
		typeCode      int
		iban          string
		wantViolation bool
		ruleCode      rules.Rule
	}{
		{
			name:          "valid: SEPA credit transfer with valid IBAN",
			typeCode:      58,
			iban:          "DE89370400440532013000",
			wantViolation: false,
			ruleCode:      rules.BRDE19,
		},
		{
			name:          "invalid: SEPA credit transfer with invalid IBAN",
			typeCode:      58,
			iban:          "INVALID",
			wantViolation: true,
			ruleCode:      rules.BRDE19,
		},
		{
			name:          "valid: SEPA direct debit with valid IBAN",
			typeCode:      59,
			iban:          "DE89370400440532013000",
			wantViolation: false,
			ruleCode:      rules.BRDE20,
		},
		{
			name:          "invalid: SEPA direct debit with invalid IBAN",
			typeCode:      59,
			iban:          "123",
			wantViolation: true,
			ruleCode:      rules.BRDE20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			if tt.typeCode == 58 {
				inv.PaymentMeans = []PaymentMeans{{
					TypeCode: tt.typeCode,
					PayeePartyCreditorFinancialAccountIBAN: tt.iban,
				}}
			} else {
				inv.PaymentMeans = []PaymentMeans{{
					TypeCode:                          tt.typeCode,
					PayerPartyDebtorFinancialAccountIBAN: tt.iban,
				}}
				inv.CreditorReferenceID = "DE98ZZZ09999999999"
			}

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, tt.ruleCode)

			if hasViolation != tt.wantViolation {
				t.Errorf("%s violation = %v, want %v", tt.ruleCode.Code, hasViolation, tt.wantViolation)
			}
		})
	}
}

// TestGermanValidation_BRDE26_CorrectedInvoice tests BR-DE-26:
// Corrected invoice must reference preceding invoice.
func TestGermanValidation_BRDE26_CorrectedInvoice(t *testing.T) {
	tests := []struct {
		name          string
		typeCode      CodeDocument
		hasReference  bool
		wantViolation bool
	}{
		{
			name:          "valid: corrected invoice with reference",
			typeCode:      384,
			hasReference:  true,
			wantViolation: false,
		},
		{
			name:          "invalid: corrected invoice without reference",
			typeCode:      384,
			hasReference:  false,
			wantViolation: true,
		},
		{
			name:          "valid: regular invoice without reference",
			typeCode:      380,
			hasReference:  false,
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			inv.InvoiceTypeCode = tt.typeCode
			if tt.hasReference {
				inv.InvoiceReferencedDocument = []ReferencedDocument{{ID: "INV-2024-001"}}
			} else {
				inv.InvoiceReferencedDocument = []ReferencedDocument{}
			}

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE26)

			if hasViolation != tt.wantViolation {
				t.Errorf("BR-DE-26 violation = %v, want %v", hasViolation, tt.wantViolation)
			}
		})
	}
}

// TestGermanValidation_BRDE27_PhoneDigits tests BR-DE-27:
// Seller contact telephone must contain at least 3 digits.
func TestGermanValidation_BRDE27_PhoneDigits(t *testing.T) {
	tests := []struct {
		name          string
		phone         string
		wantViolation bool
	}{
		{
			name:          "valid: phone with many digits",
			phone:         "+49 30 1234567",
			wantViolation: false,
		},
		{
			name:          "valid: phone with exactly 3 digits",
			phone:         "123",
			wantViolation: false,
		},
		{
			name:          "invalid: phone with 2 digits",
			phone:         "12",
			wantViolation: true,
		},
		{
			name:          "invalid: no digits",
			phone:         "are known",
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			if len(inv.Seller.DefinedTradeContact) > 0 {
				inv.Seller.DefinedTradeContact[0].PhoneNumber = tt.phone
			}

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE27)

			if hasViolation != tt.wantViolation {
				t.Errorf("BR-DE-27 violation = %v, want %v", hasViolation, tt.wantViolation)
			}
		})
	}
}

// TestGermanValidation_BRDE28_EmailFormat tests BR-DE-28:
// Email address must have valid format.
func TestGermanValidation_BRDE28_EmailFormat(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		wantViolation bool
	}{
		{
			name:          "valid: normal email",
			email:         "test@example.com",
			wantViolation: false,
		},
		{
			name:          "invalid: no @",
			email:         "testexample.com",
			wantViolation: true,
		},
		{
			name:          "invalid: multiple @",
			email:         "test@@example.com",
			wantViolation: true,
		},
		{
			name:          "invalid: starts with dot",
			email:         ".test@example.com",
			wantViolation: true,
		},
		{
			name:          "invalid: ends with dot",
			email:         "test@example.com.",
			wantViolation: true,
		},
		{
			name:          "invalid: dot before @",
			email:         "test.@example.com",
			wantViolation: true,
		},
		{
			name:          "invalid: dot after @",
			email:         "test@.example.com",
			wantViolation: true,
		},
		{
			name:          "invalid: local part too short",
			email:         "t@example.com",
			wantViolation: true,
		},
		{
			name:          "invalid: domain part too short",
			email:         "test@e",
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			if len(inv.Seller.DefinedTradeContact) > 0 {
				inv.Seller.DefinedTradeContact[0].EMail = tt.email
			}

			err := inv.Validate()
			hasViolation := hasRuleViolation(err, rules.BRDE28)

			if hasViolation != tt.wantViolation {
				t.Errorf("BR-DE-28 violation = %v, want %v for email %q", hasViolation, tt.wantViolation, tt.email)
			}
		})
	}
}

// createGermanTestInvoice creates a minimal valid German XRechnung invoice for testing.
func createGermanTestInvoice() *Invoice {
	return &Invoice{
		SchemaType:                                 SchemaTypeUnknown, // Programmatically created
		GuidelineSpecifiedDocumentContextParameter: SpecXRechnung30,
		InvoiceNumber:                              "INV-2025-001",
		InvoiceTypeCode:                            380,
		InvoiceCurrencyCode:                        "EUR",
		BuyerReference:                             "04011000-12345-35", // Leitweg-ID
		Seller: Party{
			Name: "Test Seller GmbH",
			PostalAddress: &PostalAddress{
				CountryID:    "DE",
				City:         "Berlin",
				PostcodeCode: "10115",
				Line1:        "Musterstraße 1",
			},
			VATaxRegistration: "DE123456789",
			DefinedTradeContact: []DefinedTradeContact{
				{
					PersonName:  "John Doe",
					PhoneNumber: "+49 30 1234567",
					EMail:       "test@example.com",
				},
			},
		},
		Buyer: Party{
			Name: "Test Buyer AG",
			PostalAddress: &PostalAddress{
				CountryID:    "DE",
				City:         "Munich",
				PostcodeCode: "80331",
				Line1:        "Kaufingerstraße 1",
			},
			VATaxRegistration: "DE987654321",
		},
		PaymentMeans: []PaymentMeans{
			{
				TypeCode: 58,
				PayeePartyCreditorFinancialAccountIBAN: "DE89370400440532013000",
			},
		},
	}
}

// TestGermanValidation_BRDE23AB_PaymentMeansMutualExclusivity tests BR-DE-23-a and BR-DE-23-b:
// Credit transfer payment means must have BG-17 and must NOT have BG-18 or BG-19.
func TestGermanValidation_BRDE23AB_PaymentMeansMutualExclusivity(t *testing.T) {
	tests := []struct {
		name              string
		typeCode          int
		hasIBAN           bool
		hasCardID         bool
		hasDebitAccount   bool
		wantBRDE23A       bool
		wantBRDE23B       bool
	}{
		{
			name:              "valid: credit transfer with BG-17 only",
			typeCode:          58,
			hasIBAN:           true,
			hasCardID:         false,
			hasDebitAccount:   false,
			wantBRDE23A:       false,
			wantBRDE23B:       false,
		},
		{
			name:              "invalid: credit transfer missing BG-17",
			typeCode:          58,
			hasIBAN:           false,
			hasCardID:         false,
			hasDebitAccount:   false,
			wantBRDE23A:       true,
			wantBRDE23B:       false,
		},
		{
			name:              "invalid: credit transfer with BG-18 (payment card)",
			typeCode:          58,
			hasIBAN:           true,
			hasCardID:         true,
			hasDebitAccount:   false,
			wantBRDE23A:       false,
			wantBRDE23B:       true,
		},
		{
			name:              "invalid: credit transfer with BG-19 (direct debit)",
			typeCode:          58,
			hasIBAN:           true,
			hasCardID:         false,
			hasDebitAccount:   true,
			wantBRDE23A:       false,
			wantBRDE23B:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			pm := PaymentMeans{TypeCode: tt.typeCode}
			if tt.hasIBAN {
				pm.PayeePartyCreditorFinancialAccountIBAN = "DE89370400440532013000"
			}
			if tt.hasCardID {
				pm.ApplicableTradeSettlementFinancialCardID = "1234567890123456"
			}
			if tt.hasDebitAccount {
				pm.PayerPartyDebtorFinancialAccountIBAN = "DE89370400440532013000"
			}
			inv.PaymentMeans = []PaymentMeans{pm}

			err := inv.Validate()
			hasBRDE23A := hasRuleViolation(err, rules.BRDE23A)
			hasBRDE23B := hasRuleViolation(err, rules.BRDE23B)

			if hasBRDE23A != tt.wantBRDE23A {
				t.Errorf("BR-DE-23-a violation = %v, want %v", hasBRDE23A, tt.wantBRDE23A)
			}
			if hasBRDE23B != tt.wantBRDE23B {
				t.Errorf("BR-DE-23-b violation = %v, want %v", hasBRDE23B, tt.wantBRDE23B)
			}
		})
	}
}

// TestGermanValidation_BRDE24AB_PaymentCardMutualExclusivity tests BR-DE-24-a and BR-DE-24-b:
// Payment card means must have BG-18 and must NOT have BG-17 or BG-19.
func TestGermanValidation_BRDE24AB_PaymentCardMutualExclusivity(t *testing.T) {
	tests := []struct {
		name              string
		typeCode          int
		hasIBAN           bool
		hasCardID         bool
		hasDebitAccount   bool
		wantBRDE24A       bool
		wantBRDE24B       bool
	}{
		{
			name:              "valid: payment card with BG-18 only",
			typeCode:          48,
			hasIBAN:           false,
			hasCardID:         true,
			hasDebitAccount:   false,
			wantBRDE24A:       false,
			wantBRDE24B:       false,
		},
		{
			name:              "invalid: payment card missing BG-18",
			typeCode:          54,
			hasIBAN:           false,
			hasCardID:         false,
			hasDebitAccount:   false,
			wantBRDE24A:       true,
			wantBRDE24B:       false,
		},
		{
			name:              "invalid: payment card with BG-17 (credit transfer)",
			typeCode:          48,
			hasIBAN:           true,
			hasCardID:         true,
			hasDebitAccount:   false,
			wantBRDE24A:       false,
			wantBRDE24B:       true,
		},
		{
			name:              "invalid: payment card with BG-19 (direct debit)",
			typeCode:          55,
			hasIBAN:           false,
			hasCardID:         true,
			hasDebitAccount:   true,
			wantBRDE24A:       false,
			wantBRDE24B:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			pm := PaymentMeans{TypeCode: tt.typeCode}
			if tt.hasIBAN {
				pm.PayeePartyCreditorFinancialAccountIBAN = "DE89370400440532013000"
			}
			if tt.hasCardID {
				pm.ApplicableTradeSettlementFinancialCardID = "1234567890123456"
			}
			if tt.hasDebitAccount {
				pm.PayerPartyDebtorFinancialAccountIBAN = "DE89370400440532013000"
			}
			inv.PaymentMeans = []PaymentMeans{pm}

			err := inv.Validate()
			hasBRDE24A := hasRuleViolation(err, rules.BRDE24A)
			hasBRDE24B := hasRuleViolation(err, rules.BRDE24B)

			if hasBRDE24A != tt.wantBRDE24A {
				t.Errorf("BR-DE-24-a violation = %v, want %v", hasBRDE24A, tt.wantBRDE24A)
			}
			if hasBRDE24B != tt.wantBRDE24B {
				t.Errorf("BR-DE-24-b violation = %v, want %v", hasBRDE24B, tt.wantBRDE24B)
			}
		})
	}
}

// TestGermanValidation_BRDE25AB_DirectDebitMutualExclusivity tests BR-DE-25-a and BR-DE-25-b:
// Direct debit means must have BG-19 and must NOT have BG-17 or BG-18.
func TestGermanValidation_BRDE25AB_DirectDebitMutualExclusivity(t *testing.T) {
	tests := []struct {
		name              string
		hasIBAN           bool
		hasCardID         bool
		hasDebitAccount   bool
		wantBRDE25A       bool
		wantBRDE25B       bool
	}{
		{
			name:              "valid: direct debit with BG-19 only",
			hasIBAN:           false,
			hasCardID:         false,
			hasDebitAccount:   true,
			wantBRDE25A:       false,
			wantBRDE25B:       false,
		},
		{
			name:              "invalid: direct debit missing BG-19",
			hasIBAN:           false,
			hasCardID:         false,
			hasDebitAccount:   false,
			wantBRDE25A:       true,
			wantBRDE25B:       false,
		},
		{
			name:              "invalid: direct debit with BG-17 (credit transfer)",
			hasIBAN:           true,
			hasCardID:         false,
			hasDebitAccount:   true,
			wantBRDE25A:       false,
			wantBRDE25B:       true,
		},
		{
			name:              "invalid: direct debit with BG-18 (payment card)",
			hasIBAN:           false,
			hasCardID:         true,
			hasDebitAccount:   true,
			wantBRDE25A:       false,
			wantBRDE25B:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createGermanTestInvoice()
			pm := PaymentMeans{TypeCode: 59} // Direct debit
			if tt.hasIBAN {
				pm.PayeePartyCreditorFinancialAccountIBAN = "DE89370400440532013000"
			}
			if tt.hasCardID {
				pm.ApplicableTradeSettlementFinancialCardID = "1234567890123456"
			}
			if tt.hasDebitAccount {
				pm.PayerPartyDebtorFinancialAccountIBAN = "DE89370400440532013000"
			}
			inv.PaymentMeans = []PaymentMeans{pm}
			// Also set creditor reference for BR-DE-30
			inv.CreditorReferenceID = "DE98ZZZ09999999999"

			err := inv.Validate()
			hasBRDE25A := hasRuleViolation(err, rules.BRDE25A)
			hasBRDE25B := hasRuleViolation(err, rules.BRDE25B)

			if hasBRDE25A != tt.wantBRDE25A {
				t.Errorf("BR-DE-25-a violation = %v, want %v", hasBRDE25A, tt.wantBRDE25A)
			}
			if hasBRDE25B != tt.wantBRDE25B {
				t.Errorf("BR-DE-25-b violation = %v, want %v", hasBRDE25B, tt.wantBRDE25B)
			}
		})
	}
}

// hasRuleViolation checks if an error contains a specific rule violation.
func hasRuleViolation(err error, rule rules.Rule) bool {
	if err == nil {
		return false
	}

	valErr, ok := err.(*ValidationError)
	if !ok {
		return false
	}

	return valErr.HasRule(rule)
}
