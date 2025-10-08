package einvoice

import (
	"testing"
)

func TestValidateBusinessProcessID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid standard billing process",
			id:      BPPEPPOLBilling01,
			wantErr: false,
		},
		{
			name:    "valid process with different number",
			id:      "urn:fdc:peppol.eu:2017:poacc:billing:02:1.0",
			wantErr: false,
		},
		{
			name:    "valid process number 99",
			id:      "urn:fdc:peppol.eu:2017:poacc:billing:99:1.0",
			wantErr: false,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "wrong domain",
			id:      "urn:fdc:example.com:2017:poacc:billing:01:1.0",
			wantErr: true,
		},
		{
			name:    "single digit process number",
			id:      "urn:fdc:peppol.eu:2017:poacc:billing:1:1.0",
			wantErr: true,
		},
		{
			name:    "three digit process number",
			id:      "urn:fdc:peppol.eu:2017:poacc:billing:001:1.0",
			wantErr: true,
		},
		{
			name:    "wrong version",
			id:      "urn:fdc:peppol.eu:2017:poacc:billing:01:2.0",
			wantErr: true,
		},
		{
			name:    "wrong year",
			id:      "urn:fdc:peppol.eu:2018:poacc:billing:01:1.0",
			wantErr: true,
		},
		{
			name:    "wrong process type",
			id:      "urn:fdc:peppol.eu:2017:poacc:ordering:01:1.0",
			wantErr: true,
		},
		{
			name:    "missing trailing version",
			id:      "urn:fdc:peppol.eu:2017:poacc:billing:01",
			wantErr: true,
		},
		{
			name:    "random string",
			id:      "not-a-valid-urn",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBusinessProcessID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBusinessProcessID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePEPPOLSpecificationID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid PEPPOL BIS Billing 3.0",
			id:      SpecPEPPOLBilling30,
			wantErr: false,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "wrong specification",
			id:      "urn:cen.eu:en16931:2017",
			wantErr: true,
		},
		{
			name:    "PEPPOL 2.0 (old version)",
			id:      "urn:cen.eu:en16931:2017#compliant#urn:fdc:peppol.eu:2017:poacc:billing:2.0",
			wantErr: true,
		},
		{
			name:    "XRechnung specification",
			id:      "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
			wantErr: true,
		},
		{
			name:    "Factur-X basic",
			id:      "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic",
			wantErr: true,
		},
		{
			name:    "random string",
			id:      "invalid-spec-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePEPPOLSpecificationID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePEPPOLSpecificationID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEASCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid 4-digit code starting with 0",
			code:    EAS0088,
			wantErr: false,
		},
		{
			name:    "valid 4-digit code starting with 9",
			code:    EAS9906,
			wantErr: false,
		},
		{
			name:    "valid code 0002",
			code:    "0002",
			wantErr: false,
		},
		{
			name:    "valid code 9999",
			code:    "9999",
			wantErr: false,
		},
		{
			name:    "empty string",
			code:    "",
			wantErr: true,
		},
		{
			name:    "3-digit code",
			code:    "088",
			wantErr: true,
		},
		{
			name:    "5-digit code",
			code:    "00088",
			wantErr: true,
		},
		{
			name:    "non-numeric code",
			code:    "abcd",
			wantErr: true,
		},
		{
			name:    "mixed alphanumeric",
			code:    "0a88",
			wantErr: true,
		},
		{
			name:    "code with spaces",
			code:    "0 88",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEASCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEASCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsesPEPPOLBusinessProcess(t *testing.T) {
	tests := []struct {
		name    string
		invoice *Invoice
		want    bool
	}{
		{
			name: "valid PEPPOL business process with Factur-X Extended",
			invoice: &Invoice{
				BPSpecifiedDocumentContextParameter: BPPEPPOLBilling01,
				Profile:                             CProfileExtended,
			},
			want: true,
		},
		{
			name: "valid PEPPOL business process with EN16931",
			invoice: &Invoice{
				BPSpecifiedDocumentContextParameter: BPPEPPOLBilling01,
				Profile:                             CProfileEN16931,
			},
			want: true,
		},
		{
			name: "valid PEPPOL business process with XRechnung",
			invoice: &Invoice{
				BPSpecifiedDocumentContextParameter: BPPEPPOLBilling01,
				Profile:                             CProfileXRechnung,
			},
			want: true,
		},
		{
			name: "missing business process",
			invoice: &Invoice{
				BPSpecifiedDocumentContextParameter: "",
				Profile:                             CProfileEN16931,
			},
			want: false,
		},
		{
			name: "invalid business process format",
			invoice: &Invoice{
				BPSpecifiedDocumentContextParameter: "invalid",
				Profile:                             CProfileEN16931,
			},
			want: false,
		},
		{
			name: "different valid PEPPOL process number",
			invoice: &Invoice{
				BPSpecifiedDocumentContextParameter: "urn:fdc:peppol.eu:2017:poacc:billing:99:1.0",
				Profile:                             CProfileBasic,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.invoice.UsesPEPPOLBusinessProcess(); got != tt.want {
				t.Errorf("Invoice.UsesPEPPOLBusinessProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEASConstants(t *testing.T) {
	// Test that all EAS constants are in the correct format
	easCodes := []string{
		EAS0002, EAS0007, EAS0009, EAS0037, EAS0060,
		EAS0088, EAS0096, EAS0130, EAS0135, EAS0183,
		EAS0184, EAS0188, EAS0190, EAS0192, EAS0195,
		EAS0196, EAS0198, EAS0204, EAS0208, EAS0209,
		EAS0210, EAS0211, EAS0212, EAS0213,
		EAS9901, EAS9906, EAS9907, EAS9910, EAS9913,
		EAS9914, EAS9915, EAS9918, EAS9919, EAS9920,
		EAS9922, EAS9923, EAS9925, EAS9926, EAS9927,
		EAS9928, EAS9929, EAS9930, EAS9931, EAS9933,
		EAS9934, EAS9935, EAS9936, EAS9937, EAS9938,
		EAS9939, EAS9940, EAS9941, EAS9942, EAS9943,
		EAS9944, EAS9945, EAS9946, EAS9947, EAS9948,
		EAS9949, EAS9950, EAS9951, EAS9952, EAS9953,
		EAS9955, EAS9956, EAS9957, EAS9958,
	}

	for _, code := range easCodes {
		t.Run("validate_"+code, func(t *testing.T) {
			if err := ValidateEASCode(code); err != nil {
				t.Errorf("EAS constant %q failed validation: %v", code, err)
			}
		})
	}
}

func TestPEPPOLConstants(t *testing.T) {
	// Test that PEPPOL constants are valid
	t.Run("BPPEPPOLBilling01", func(t *testing.T) {
		if err := ValidateBusinessProcessID(BPPEPPOLBilling01); err != nil {
			t.Errorf("BPPEPPOLBilling01 constant is invalid: %v", err)
		}
	})

	t.Run("SpecPEPPOLBilling30", func(t *testing.T) {
		if err := ValidatePEPPOLSpecificationID(SpecPEPPOLBilling30); err != nil {
			t.Errorf("SpecPEPPOLBilling30 constant is invalid: %v", err)
		}
	})
}
