package codelists

import "testing"

func TestDocumentType(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"380", "Standard Invoice"},
		{"381", "Credit note"},
		{"383", "Debit note"},
		{"326", "Partial invoice"},
		{"384", "Corrected invoice"},
		{"999", "Unknown"}, // Unknown code
		{"", "Unknown"},    // Empty code
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := DocumentType(tt.code)
			if got != tt.want {
				t.Errorf("DocumentType(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestDocumentTypeConsistency(t *testing.T) {
	// Ensure multiple calls return the same result (test lazy loading)
	first := DocumentType("380")
	second := DocumentType("380")

	if first != second {
		t.Errorf("DocumentType returned inconsistent results: %q vs %q", first, second)
	}

	if first != "Standard Invoice" {
		t.Errorf("DocumentType(\"380\") = %q, want \"Standard Invoice\"", first)
	}
}

func TestUnitCode(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"C62", "one"},
		{"XPP", "piece"},  // UNECE Rec 21 code for piece
		{"H87", "piece"},  // UNECE Rec 20 code for piece
		{"MTR", "metre"},
		{"KGM", "kilogram"},
		{"UNKNOWN", "UNKNOWN"}, // Unknown code returns itself
		{"", ""},               // Empty code returns empty
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := UnitCode(tt.code)
			if got != tt.want {
				t.Errorf("UnitCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestUnitCodeConsistency(t *testing.T) {
	// Ensure multiple calls return the same result (test lazy loading)
	first := UnitCode("C62")
	second := UnitCode("C62")

	if first != second {
		t.Errorf("UnitCode returned inconsistent results: %q vs %q", first, second)
	}

	if first != "one" {
		t.Errorf("UnitCode(\"C62\") = %q, want \"one\"", first)
	}
}

func TestTextSubjectQualifier(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"AAA", "Goods item description"},
		{"AAB", "Payment term"},
		{"AUT", "Authentication"},
		{"BLC", "Transport contract document clause"},
		{"999", "Unknown"}, // Unknown code
		{"", "Unknown"},    // Empty code
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := TextSubjectQualifier(tt.code)
			if got != tt.want {
				t.Errorf("TextSubjectQualifier(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestTextSubjectQualifierConsistency(t *testing.T) {
	// Ensure multiple calls return the same result
	first := TextSubjectQualifier("AAA")
	second := TextSubjectQualifier("AAA")

	if first != second {
		t.Errorf("TextSubjectQualifier returned inconsistent results: %q vs %q", first, second)
	}

	if first != "Goods item description" {
		t.Errorf("TextSubjectQualifier(\"AAA\") = %q, want \"Goods item description\"", first)
	}
}
