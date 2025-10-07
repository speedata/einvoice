// Package rules contains EN 16931 business rule definitions for electronic invoicing validation.
//
// Rules are auto-generated from official schematron specifications maintained by CEN/TC 434.
// Source: https://github.com/ConnectingEurope/eInvoicing-EN16931
//
// # Generation
//
// Rules are regenerated using go generate:
//
//	cd rules && go generate
//
// See cmd/genrules/README.md for detailed generation instructions.
//
// # Usage
//
//	import "github.com/speedata/einvoice/rules"
//
//	func (inv *Invoice) validate() {
//	    if inv.SpecificationIdentifier == "" {
//	        inv.addViolation(rules.BR1, "Missing specification identifier")
//	    }
//	}
package rules
