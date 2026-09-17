package einvoice

import "strings"

// ibanLengths maps ISO 3166-1 alpha-2 country codes to the IBAN length
// registered for that country in the SWIFT IBAN Registry.
//
// Countries not listed here (e.g. newly added registry entries or partial
// IBAN adopters) fall back to the generic length range of 15-34 characters.
var ibanLengths = map[string]int{
	"AD": 24, "AE": 23, "AL": 28, "AT": 20, "AZ": 28,
	"BA": 20, "BE": 16, "BG": 22, "BH": 22, "BI": 27, "BR": 29, "BY": 28,
	"CH": 21, "CR": 22, "CY": 28, "CZ": 24,
	"DE": 22, "DJ": 27, "DK": 18, "DO": 28,
	"EE": 20, "EG": 29, "ES": 24,
	"FI": 18, "FK": 18, "FO": 18, "FR": 27,
	"GB": 22, "GE": 22, "GI": 23, "GL": 18, "GR": 27, "GT": 28,
	"HN": 28, "HR": 21, "HU": 28,
	"IE": 22, "IL": 23, "IQ": 23, "IS": 26, "IT": 27,
	"JO": 30,
	"KW": 30, "KZ": 20,
	"LB": 28, "LC": 32, "LI": 21, "LT": 20, "LU": 20, "LV": 21, "LY": 25,
	"MC": 27, "MD": 24, "ME": 22, "MK": 19, "MN": 20, "MR": 27, "MT": 31, "MU": 30,
	"NI": 28, "NL": 18, "NO": 15,
	"OM": 23,
	"PK": 24, "PL": 28, "PS": 29, "PT": 25,
	"QA": 29,
	"RO": 24, "RS": 22, "RU": 33,
	"SA": 24, "SC": 31, "SD": 18, "SE": 24, "SI": 19, "SK": 24, "SM": 27, "SO": 23, "ST": 25, "SV": 28,
	"TL": 23, "TN": 24, "TR": 26,
	"UA": 29,
	"VA": 22, "VG": 24,
	"XK": 20,
	"YE": 30,
}

// isValidIBAN validates an IBAN per ISO 13616.
//
// The check covers:
//   - the basic structure: 2-letter country code, 2 check digits, alphanumeric BBAN
//   - the length: the country-specific length from the SWIFT IBAN Registry, or
//     the generic 15-34 character range for unlisted country codes
//   - the modulo-97 checksum of the check digits (ISO 7064 MOD 97-10)
//
// Spaces are ignored and letters are treated case-insensitively, so both the
// electronic format ("DE89370400440532013000") and the print format
// ("DE89 3704 0044 0532 0130 00") are accepted.
func isValidIBAN(iban string) bool {
	iban = strings.ToUpper(strings.ReplaceAll(iban, " ", ""))

	// Generic length range per SWIFT registry (NO has 15, RU has 33; max is 34)
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}

	// Country code (2 letters) followed by check digits (2 digits)
	if !isUppercaseLetter(iban[0]) || !isUppercaseLetter(iban[1]) {
		return false
	}
	if !isDigit(iban[2]) || !isDigit(iban[3]) {
		return false
	}

	// BBAN must be alphanumeric
	for i := 4; i < len(iban); i++ {
		if !isAlphanumeric(iban[i]) {
			return false
		}
	}

	// Country-specific length
	if want, ok := ibanLengths[iban[:2]]; ok && len(iban) != want {
		return false
	}

	// Check digits 00, 01 and 99 are never valid per ISO 13616
	if cd := iban[2:4]; cd == "00" || cd == "01" || cd == "99" {
		return false
	}

	return ibanMod97(iban) == 1
}

// ibanMod97 computes the ISO 7064 MOD 97-10 remainder of an IBAN.
//
// The country code and check digits are moved to the end, letters are
// replaced by their numeric value (A=10 ... Z=35) and the resulting number is
// reduced modulo 97. A valid IBAN yields a remainder of 1.
//
// The input must already be normalized (uppercase, no spaces, alphanumeric).
func ibanMod97(iban string) int {
	rearranged := iban[4:] + iban[:4]
	remainder := 0
	for i := 0; i < len(rearranged); i++ {
		c := rearranged[i]
		if isDigit(c) {
			remainder = (remainder*10 + int(c-'0')) % 97
		} else {
			// Letters expand to two digits (10-35)
			remainder = (remainder*100 + int(c-'A') + 10) % 97
		}
	}
	return remainder
}

// isDigit checks if a byte represents a digit (0-9).
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// isAlphanumeric checks if a byte represents an alphanumeric character (0-9, A-Z).
func isAlphanumeric(b byte) bool {
	return isDigit(b) || isUppercaseLetter(b)
}
