package main

import "testing"

func TestFormatDocumentType(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		showCodes bool
		verbose   bool
		want      string
	}{
		// Default mode: description only
		{
			name:      "default known code",
			code:      "380",
			showCodes: false,
			verbose:   false,
			want:      "Standard Invoice",
		},
		{
			name:      "default unknown code",
			code:      "999",
			showCodes: false,
			verbose:   false,
			want:      "Unknown (999)",
		},
		// --show-codes mode: code only
		{
			name:      "show-codes known code",
			code:      "380",
			showCodes: true,
			verbose:   false,
			want:      "380",
		},
		{
			name:      "show-codes unknown code",
			code:      "999",
			showCodes: true,
			verbose:   false,
			want:      "999",
		},
		// -vv mode: both code and description
		{
			name:      "verbose known code",
			code:      "380",
			showCodes: false,
			verbose:   true,
			want:      "380 (Standard Invoice)",
		},
		{
			name:      "verbose unknown code",
			code:      "999",
			showCodes: false,
			verbose:   true,
			want:      "999 (Unknown)",
		},
		// Precedence: --show-codes takes precedence over -vv
		{
			name:      "both flags known code",
			code:      "380",
			showCodes: true,
			verbose:   true,
			want:      "380",
		},
		{
			name:      "both flags unknown code",
			code:      "999",
			showCodes: true,
			verbose:   true,
			want:      "999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDocumentType(tt.code, tt.showCodes, tt.verbose)
			if got != tt.want {
				t.Errorf("formatDocumentType(%q, showCodes=%v, verbose=%v) = %q, want %q",
					tt.code, tt.showCodes, tt.verbose, got, tt.want)
			}
		})
	}
}

func TestFormatUnitCode(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		showCodes bool
		verbose   bool
		want      string
	}{
		// Default mode: description if found, code if not
		{
			name:      "default known code",
			code:      "XPP",
			showCodes: false,
			verbose:   false,
			want:      "package",
		},
		{
			name:      "default unknown code",
			code:      "UNKNOWN",
			showCodes: false,
			verbose:   false,
			want:      "UNKNOWN",
		},
		// --show-codes mode: code only
		{
			name:      "show-codes known code",
			code:      "XPP",
			showCodes: true,
			verbose:   false,
			want:      "XPP",
		},
		{
			name:      "show-codes unknown code",
			code:      "UNKNOWN",
			showCodes: true,
			verbose:   false,
			want:      "UNKNOWN",
		},
		// -vv mode: both code and description
		{
			name:      "verbose known code",
			code:      "XPP",
			showCodes: false,
			verbose:   true,
			want:      "XPP (package)",
		},
		{
			name:      "verbose unknown code",
			code:      "UNKNOWN",
			showCodes: false,
			verbose:   true,
			want:      "UNKNOWN",
		},
		// Precedence: --show-codes takes precedence over -vv
		{
			name:      "both flags known code",
			code:      "XPP",
			showCodes: true,
			verbose:   true,
			want:      "XPP",
		},
		{
			name:      "both flags unknown code",
			code:      "UNKNOWN",
			showCodes: true,
			verbose:   true,
			want:      "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUnitCode(tt.code, tt.showCodes, tt.verbose)
			if got != tt.want {
				t.Errorf("formatUnitCode(%q, showCodes=%v, verbose=%v) = %q, want %q",
					tt.code, tt.showCodes, tt.verbose, got, tt.want)
			}
		})
	}
}
