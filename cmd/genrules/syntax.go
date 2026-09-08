package main

import (
	"encoding/xml"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// SyntaxRule is one schematron rule of a syntax-binding pattern, ready for
// code generation. The context has already been rewritten from an XSLT match
// pattern to an equivalent XPath expression.
type SyntaxRule struct {
	Context string
	Asserts []SyntaxAssert
}

// SyntaxAssert is one assert of a SyntaxRule.
type SyntaxAssert struct {
	Code        string
	Test        string
	Severity    string // Go identifier: SeverityError or SeverityWarning
	Description string
}

// goxpathTestOverrides replaces assert tests that trip known goxpath bugs
// with semantically equivalent expressions. Remove the entries once the
// referenced upstream issues are fixed and the dependency is updated.
var goxpathTestOverrides = map[string]string{
	// speedata/goxpath#3: the preceding:: axis only considers siblings of the
	// context node. The original tests count the nodes that have no equal
	// preceding node - i.e. the number of distinct values.
	"UBL-SR-44": "count(distinct-values(//cbc:PaymentID)) <= 1",
	"UBL-SR-47": "count(distinct-values(//cbc:PaymentMeansCode)) <= 1",
}

// selfAxisRe matches a self:: axis step with a prefixed name test.
var selfAxisRe = regexp.MustCompile(`self::[a-z]+:([A-Za-z]+)`)

// prefixWildcardRe matches a name test with prefix and wildcard local name.
var prefixWildcardRe = regexp.MustCompile(`([a-z]+):\*`)

// workaroundGoxpath rewrites XPath constructs that goxpath evaluates
// incorrectly into equivalent supported ones.
//
// speedata/goxpath#2: the self:: axis ignores the node test. Every self::
// occurrence in the syntax-binding rules is a boolean element-name check on
// unambiguous vocabularies, so a local-name() comparison is equivalent.
//
// speedata/goxpath#4: the prefixed wildcard (e.g. ram:*) matches nothing;
// a namespace-uri() predicate is equivalent. The prefix/URI pairs come from
// the schematron's own ns declarations.
func workaroundGoxpath(expr string, namespaces map[string]string) string {
	expr = selfAxisRe.ReplaceAllString(expr, "(local-name() = '$1')")
	expr = prefixWildcardRe.ReplaceAllStringFunc(expr, func(match string) string {
		prefix := strings.TrimSuffix(match, ":*")
		uri, ok := namespaces[prefix]
		if !ok {
			return match
		}
		return "*[namespace-uri() = '" + uri + "']"
	})
	return expr
}

// runSyntaxMode generates a []rules.SyntaxRule table from the schematron
// pattern named by --syntax-pattern.
func runSyntaxMode(data []byte) error {
	var schema SchematronSchema
	if err := xml.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	var pattern *SchematronPattern
	for i := range schema.Patterns {
		if schema.Patterns[i].ID == *syntaxFlag {
			pattern = &schema.Patterns[i]
			break
		}
	}
	if pattern == nil {
		return fmt.Errorf("pattern %q not found in schematron source", *syntaxFlag)
	}

	namespaces := make(map[string]string)
	for _, ns := range schema.Namespaces {
		namespaces[ns.Prefix] = ns.URI
	}

	var syntaxRules []SyntaxRule
	assertCount := 0
	for _, rule := range pattern.Rules {
		sr := SyntaxRule{Context: workaroundGoxpath(matchPatternToXPath(rule.Context), namespaces)}
		for _, assert := range rule.Asserts {
			if assert.ID == "" {
				continue
			}
			// The syntax asserts read "[CII-DT-031] - text"; cleanDescription
			// strips the bracketed ID but keeps the separating dash.
			desc := strings.TrimSpace(strings.TrimPrefix(cleanDescription(assert.Description), "- "))
			test := workaroundGoxpath(strings.Join(strings.Fields(assert.Test), " "), namespaces)
			if override, ok := goxpathTestOverrides[assert.ID]; ok {
				test = override
			}
			sr.Asserts = append(sr.Asserts, SyntaxAssert{
				Code:        assert.ID,
				Test:        test,
				Severity:    flagToSeverity(assert.Flag),
				Description: desc,
			})
		}
		if len(sr.Asserts) > 0 {
			syntaxRules = append(syntaxRules, sr)
			assertCount += len(sr.Asserts)
		}
	}

	output, err := generateSyntaxCode(syntaxRules, *packageFlag, *versionFlag, *varnameFlag, *syntaxFlag)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	formatted, err := formatCode(output)
	if err != nil {
		formatted = output
	}

	if err := writeOutput(*outputFlag, formatted); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("Generated %d syntax rules (%d asserts) to %s\n", len(syntaxRules), assertCount, *outputFlag)
	return nil
}

// matchPatternToXPath rewrites a schematron rule context (an XSLT match
// pattern) into an equivalent XPath expression: relative location paths match
// anywhere in the document, so each top-level union branch that does not start
// with "/" is prefixed with "//".
func matchPatternToXPath(context string) string {
	branches := splitTopLevel(context, '|')
	for i, b := range branches {
		b = strings.Join(strings.Fields(b), " ")
		if !strings.HasPrefix(b, "/") {
			b = "//" + b
		}
		branches[i] = b
	}
	return strings.Join(branches, " | ")
}

// splitTopLevel splits s on sep, ignoring occurrences inside [], () or
// quoted strings.
func splitTopLevel(s string, sep byte) []string {
	var parts []string
	depth := 0
	var quote byte
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"':
			quote = c
		case c == '[' || c == '(':
			depth++
		case c == ']' || c == ')':
			depth--
		case c == sep && depth == 0:
			parts = append(parts, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	parts = append(parts, strings.TrimSpace(s[start:]))
	return parts
}

// flagToSeverity maps the schematron flag attribute to a rules.Severity
// identifier. Schematron reports are errors unless flagged as warning.
func flagToSeverity(schFlag string) string {
	if schFlag == "warning" {
		return "SeverityWarning"
	}
	return "SeverityError"
}

// generateSyntaxCode generates the Go source for the syntax rule table.
func generateSyntaxCode(syntaxRules []SyntaxRule, packageName, sourceVersion, varname, patternID string) ([]byte, error) {
	tmpl := template.Must(template.New("syntax").Funcs(template.FuncMap{
		"quote": strconv.Quote,
	}).Parse(syntaxTemplate))

	data := struct {
		Package       string
		SourceFile    string
		SourceVersion string
		Generated     string
		Varname       string
		PatternID     string
		Rules         []SyntaxRule
	}{
		Package:       packageName,
		SourceFile:    path.Base(*sourceFlag),
		SourceVersion: sourceVersion,
		Generated:     time.Now().UTC().Format(time.RFC3339),
		Varname:       varname,
		PatternID:     patternID,
		Rules:         syntaxRules,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return []byte(buf.String()), nil
}

// syntaxTemplate is the Go code generation template for syntax rule tables.
const syntaxTemplate = `// Code generated by genrules from {{.SourceFile}} (pattern {{.PatternID}}); DO NOT EDIT.
// Source: https://github.com/ConnectingEurope/eInvoicing-EN16931
{{- if .SourceVersion}}
// Version: {{.SourceVersion}}
{{- end}}
// Generated: {{.Generated}}

package {{.Package}}

// {{.Varname}} contains the EN 16931 syntax-binding rules of schematron
// pattern {{.PatternID}}. Rule order matters: per context node, only the
// first matching rule is evaluated (schematron "first match wins").
var {{.Varname}} = []SyntaxRule{
{{- range .Rules}}
	{
		Context: {{quote .Context}},
		Asserts: []SyntaxAssert{
{{- range .Asserts}}
			{Code: {{quote .Code}}, Severity: {{.Severity}}, Test: {{quote .Test}}, Description: {{quote .Description}}},
{{- end}}
		},
	},
{{- end}}
}
`
