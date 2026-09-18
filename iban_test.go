package einvoice

import "testing"

func TestIsValidIBAN(t *testing.T) {
	tests := []struct {
		name string
		iban string
		want bool
	}{
		// Valid IBANs from various countries (checksums verified)
		{"valid DE", "DE89370400440532013000", true},
		{"valid AT", "AT611904300234573201", true},
		{"valid BE", "BE68539007547034", true},
		{"valid CH", "CH9300762011623852957", true},
		{"valid ES", "ES9121000418450200051332", true},
		{"valid FR with letter in BBAN", "FR1420041010050500013M02606", true},
		{"valid GB with letters in BBAN", "GB82WEST12345698765432", true},
		{"valid IT with letter in BBAN", "IT60X0542811101000000123456", true},
		{"valid NL", "NL91ABNA0417164300", true},
		{"valid NO (shortest, 15 chars)", "NO9386011117947", true},
		{"valid MT (31 chars)", "MT84MALT011000012345MTLCAST001S", true},

		// Formatting variants
		{"valid print format with spaces", "DE89 3704 0044 0532 0130 00", true},
		{"valid lowercase", "de89370400440532013000", true},

		// Unknown country code: only generic checks apply
		{"unknown country with valid checksum", "XX57123456789012345", true},
		{"unknown country with invalid checksum", "XX58123456789012345", false},

		// Checksum errors
		{"wrong check digits", "DE00370400440532013000", false},
		{"check digits 01", "DE01370400440532013000", false},
		{"check digits 99", "DE99370400440532013000", false},
		{"single digit transposed", "DE89370400440532031000", false},
		{"single digit changed", "DE89370400440532013001", false},

		// Length errors
		{"DE too short by one", "DE8937040044053201300", false},
		{"DE too long by one", "DE893704004405320130000", false},
		{"BE with DE length", "BE68539007547034000000", false},
		{"below minimum length", "DE8937040044053", false},
		{"above maximum length", "DE89370400440532013000123456789012345", false},

		// Structural errors
		{"empty", "", false},
		{"garbage", "INVALID", false},
		{"digits only", "123", false},
		{"digit in country code", "D189370400440532013000", false},
		{"letter in check digits", "DEA9370400440532013000", false},
		{"special character in BBAN", "DE89370400440532013-00", false},
		{"non-ASCII character", "DE89370400440532013Ö00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidIBAN(tt.iban); got != tt.want {
				t.Errorf("isValidIBAN(%q) = %v, want %v", tt.iban, got, tt.want)
			}
		})
	}
}

func TestIBANMod97(t *testing.T) {
	tests := []struct {
		iban string
		want int
	}{
		{"DE89370400440532013000", 1},
		{"GB82WEST12345698765432", 1},
		{"DE00370400440532013000", 9},
	}
	for _, tt := range tests {
		if got := ibanMod97(tt.iban); got != tt.want {
			t.Errorf("ibanMod97(%q) = %d, want %d", tt.iban, got, tt.want)
		}
	}
}
