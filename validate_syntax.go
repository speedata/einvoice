package einvoice

import (
	"fmt"

	"github.com/speedata/cxpath"
	"github.com/speedata/einvoice/rules"
	"github.com/speedata/goxpath"
)

// validateSyntaxRules evaluates the EN 16931 syntax-binding rules
// (CII-SR-*/CII-DT-* or UBL-SR-*/UBL-CR-*/UBL-DT-*) against the XML tree.
//
// These rules restrict the UN/CEFACT and UBL vocabularies to the subset used
// by the EN 16931 semantic model: elements and attributes without a mapping to
// a business term must not be present. They cannot be checked on the parsed
// Invoice struct because the parser drops everything it does not know, so they
// run during parsing while the document tree is still available. Findings are
// stored in inv.syntaxViolations / inv.syntaxWarnings; Validate() surfaces
// them through the regular violations/warnings API.
//
// The rules follow schematron semantics: within the rule set, each node is
// checked only against the first rule whose context matches ("first match
// wins"), in rule order.
//
// A rule or test whose XPath cannot be evaluated is skipped;
// TestSyntaxRulesEvaluate guards against unsupported expressions entering the
// generated tables.
func validateSyntaxRules(root *cxpath.Context, syntaxRules []rules.SyntaxRule, inv *Invoice) {
	// One shared evaluation context: goxpath caches parsed expressions
	// globally, and reusing the context avoids a copy per evaluation. Axis
	// steps mutate the context sequence, so it is reset before every call.
	p := &goxpath.Parser{Ctx: goxpath.CopyContext(root.P.Ctx)}
	rootSeq := root.Seq
	claimed := make(map[goxpath.Item]bool)

	for _, rule := range syntaxRules {
		p.Ctx.SetContextSequence(rootSeq)
		nodes, err := safeEvaluateSyntax(p, rule.Context)
		if err != nil {
			continue
		}
		for _, item := range nodes {
			if claimed[item] {
				continue
			}
			claimed[item] = true

			for _, assert := range rule.Asserts {
				p.Ctx.SetContextSequence(goxpath.Sequence{item})
				result, err := safeEvaluateSyntax(p, assert.Test)
				if err != nil {
					continue
				}
				passed, err := goxpath.BooleanValue(result)
				if err != nil || passed {
					continue
				}
				finding := SemanticError{
					Rule: rules.Rule{Code: assert.Code, Description: assert.Description},
					Text: assert.Description,
				}
				if assert.Severity == rules.SeverityWarning {
					inv.syntaxWarnings = append(inv.syntaxWarnings, finding)
				} else {
					inv.syntaxViolations = append(inv.syntaxViolations, finding)
				}
			}
		}
	}
}

// safeEvaluateSyntax evaluates an XPath expression and converts goxpath
// panics on unsupported expressions into errors.
func safeEvaluateSyntax(p *goxpath.Parser, expr string) (seq goxpath.Sequence, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("xpath evaluation panic: %v", r)
		}
	}()
	return p.Evaluate(expr)
}
