package einvoice

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/speedata/cxpath"
	"github.com/speedata/einvoice/rules"
)

// parseFixtureWithReplacement reads a fixture, applies a single string
// replacement to inject a syntax violation, and parses the result.
func parseFixtureWithReplacement(t *testing.T, filename, old, new string, n int) *Invoice {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	modified := strings.Replace(string(data), old, new, n)
	if modified == string(data) {
		t.Fatalf("replacement %q not applied - fixture %s changed?", old, filename)
	}
	inv, err := ParseReader(strings.NewReader(modified))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return inv
}

// hasCode reports whether a finding with the given rule code exists.
func hasCode(findings []SemanticError, code string) bool {
	for _, f := range findings {
		if f.Rule.Code == code {
			return true
		}
	}
	return false
}

// TestSyntaxCIIDT31 checks that a currencyID attribute on an amount other
// than TaxTotalAmount is reported as CII-DT-031 (fatal in the CEN artifacts).
func TestSyntaxCIIDT31(t *testing.T) {
	inv := parseFixtureWithReplacement(t,
		"testdata/cii/en16931/zugferd-en16931-einfach.xml",
		"<ram:LineTotalAmount>", `<ram:LineTotalAmount currencyID="EUR">`, 1)

	err := inv.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want CII-DT-031 violation")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Validate() = %v, want *ValidationError", err)
	}
	if !valErr.HasRuleCode("CII-DT-031") {
		t.Errorf("violations = %v, want CII-DT-031", valErr.Violations())
	}

	// The finding must point at the offending node.
	for _, v := range valErr.Violations() {
		if v.Rule.Code != "CII-DT-031" {
			continue
		}
		if !strings.Contains(v.Location, "/ram:LineTotalAmount") || !strings.Contains(v.Location, ", line ") {
			t.Errorf("Location = %q, want XPath-like path to ram:LineTotalAmount with line number", v.Location)
		}
		if !strings.Contains(v.Text, v.Location) {
			t.Errorf("Text = %q, want it to include the location %q", v.Text, v.Location)
		}
	}
}

// TestSyntaxCIIDT101 checks that a schemeName attribute on a party ID is
// reported as CII-DT-101. The rule context uses the prefixed wildcard form
// (//ram:*[...]), so this also guards goxpath's support for it
// (speedata/goxpath#4).
func TestSyntaxCIIDT101(t *testing.T) {
	inv := parseFixtureWithReplacement(t,
		"testdata/cii/en16931/zugferd-en16931-einfach.xml",
		"<ram:ID>549910</ram:ID>", `<ram:ID schemeName="x">549910</ram:ID>`, 1)

	err := inv.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want CII-DT-101 violation")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Validate() = %v, want *ValidationError", err)
	}
	if !valErr.HasRuleCode("CII-DT-101") {
		t.Errorf("violations = %v, want CII-DT-101", valErr.Violations())
	}
}

// TestSyntaxCIIWarning checks that a "warning" flagged syntax rule ends up in
// the warnings, not the violations. CII-SR-04 (Value should not be present in
// a DocumentContextParameter) is such a rule.
func TestSyntaxCIIWarning(t *testing.T) {
	inv := parseFixtureWithReplacement(t,
		"testdata/cii/en16931/zugferd-en16931-einfach.xml",
		"</rsm:ExchangedDocumentContext>",
		"<ram:BusinessProcessSpecifiedDocumentContextParameter><ram:Value>x</ram:Value></ram:BusinessProcessSpecifiedDocumentContextParameter></rsm:ExchangedDocumentContext>", 1)

	if err := inv.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil (warnings must not fail validation)", err)
	}
	if !hasCode(inv.Warnings(), "CII-SR-04") {
		t.Errorf("warnings = %v, want CII-SR-04", inv.Warnings())
	}
}

// TestSyntaxUBLSR47 checks that differing payment means codes are reported
// as UBL-SR-47.
func TestSyntaxUBLSR47(t *testing.T) {
	inv := parseFixtureWithReplacement(t,
		"testdata/ubl/invoice/ubl-tc434-example1.xml",
		"<cbc:PaymentMeansCode>30</cbc:PaymentMeansCode>",
		"<cbc:PaymentMeansCode>58</cbc:PaymentMeansCode>", 1)

	err := inv.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want UBL-SR-47 violation")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("Validate() = %v, want *ValidationError", err)
	}
	if !valErr.HasRuleCode("UBL-SR-47") {
		t.Errorf("violations = %v, want UBL-SR-47", valErr.Violations())
	}
}

// TestSyntaxFirstMatchWins checks the schematron rule precedence: a node is
// only checked against the first rule whose context matches. The document
// level TypeCode is claimed by the generic //ram:TypeCode rule (CII-DT-008,
// CII-DT-009), so the later rule for the ExchangedDocument TypeCode
// (CII-DT-010: listID should not be present) must not fire.
func TestSyntaxFirstMatchWins(t *testing.T) {
	inv := parseFixtureWithReplacement(t,
		"testdata/cii/en16931/zugferd-en16931-einfach.xml",
		"<ram:TypeCode>380</ram:TypeCode>", `<ram:TypeCode listID="x">380</ram:TypeCode>`, 1)

	if err := inv.Validate(); err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) && valErr.HasRuleCode("CII-DT-010") {
			t.Errorf("CII-DT-010 fired although the //ram:TypeCode rule claims the node first: %v", valErr.Violations())
		}
	}
}

// TestSyntaxExtendedProfileSkipped checks that the CEN syntax-binding rules
// are not applied to Extended profile invoices: the profile is only
// "conformant" to EN 16931 and deliberately allows additional elements
// (e.g. nested sub-lines).
func TestSyntaxExtendedProfileSkipped(t *testing.T) {
	inv, err := ParseXMLFile("testdata/cii/extended/zf25-subline-nested.xml")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := inv.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
	if len(inv.syntaxViolations) != 0 || len(inv.syntaxWarnings) != 0 {
		t.Errorf("syntax findings on Extended profile: %v %v", inv.syntaxViolations, inv.syntaxWarnings)
	}
}

// TestSyntaxRulesEvaluate guards the generated rule tables against XPath
// expressions the evaluator cannot handle: every context and every test must
// evaluate without error. Silent evaluation errors would turn into silent
// false negatives at runtime.
func TestSyntaxRulesEvaluate(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		ns       map[string]string
		rules    []rules.SyntaxRule
	}{
		{
			name:     "CII",
			filename: "testdata/cii/en16931/CII_example1.xml",
			ns: map[string]string{
				"rsm": "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100",
				"ram": "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100",
				"udt": "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100",
				"qdt": "urn:un:unece:uncefact:data:standard:QualifiedDataType:100",
			},
			rules: rules.CIISyntaxRules,
		},
		{
			name:     "UBL",
			filename: "testdata/ubl/invoice/ubl-tc434-example1.xml",
			ns: map[string]string{
				"ubl": nsUBLInvoice,
				"cn":  nsUBLCreditNote,
				"cac": nsUBLCAC,
				"cbc": nsUBLCBC,
			},
			rules: rules.UBLSyntaxRules,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, err := cxpath.NewFromFile(tc.filename)
			if err != nil {
				t.Fatalf("read %s: %v", tc.filename, err)
			}
			for prefix, uri := range tc.ns {
				ctx.SetNamespace(prefix, uri)
			}
			root := ctx.Root()

			for _, rule := range tc.rules {
				for node := range root.Each(rule.Context) {
					if node.Error != nil {
						t.Errorf("context %q: %v", rule.Context, node.Error)
						break
					}
				}
				// Evaluate every test against the root node too, so that
				// tests of rules whose context never matches a valid
				// document are still exercised.
				for _, assert := range rule.Asserts {
					res := root.Eval(assert.Test)
					res.Bool()
					if res.Error != nil {
						t.Errorf("%s test %q: %v", assert.Code, assert.Test, res.Error)
					}
				}
			}
		})
	}
}
