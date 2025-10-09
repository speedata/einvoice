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
		{"XPP", "package"},
		{"H87", "piece"},
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
