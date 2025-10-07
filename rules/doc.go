// Package rules contains auto-generated business rules for EN 16931 electronic invoicing validation.
//
// Rules are generated from official schematron specifications maintained by CEN/TC 434.
// See https://github.com/ConnectingEurope/eInvoicing-EN16931
//
// # Generation
//
// Rules are generated using the genrules tool:
//
//	go run ./cmd/genrules \
//	  --source https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/master/cii/schematron/abstract/EN16931-CII-model.sch \
//	  --version v1.3.14.1 \
//	  --package rules \
//	  --output rules/en16931.go
//
// Or using go generate:
//
//	go generate ./rules
//
// # Usage
//
// Import the rules package and reference rule constants:
//
//	import "github.com/speedata/einvoice/rules"
//
//	func (inv *Invoice) checkBR1() {
//	    if inv.SpecificationIdentifier == "" {
//	        inv.addViolation(rules.BR1, "Missing specification identifier")
//	    }
//	}
//
// # Rule Structure
//
// Each rule contains:
//   - Code: The official EN 16931 rule identifier (e.g., "BR-1", "BR-S-08")
//   - Fields: BT-/BG- identifiers from the semantic model
//   - Description: The official specification requirement text
//
// # Rule Categories
//
// Rules are organized by category:
//   - BR-1 to BR-67: Core mandatory business rules
//   - BR-CO-*: Calculation and consistency rules
//   - BR-DEC-*: Decimal precision rules
//   - BR-S-*: Standard rated VAT rules
//   - BR-AE-*: Reverse charge VAT rules
//   - BR-E-*: Exempt from VAT rules
//   - BR-Z-*: Zero rated VAT rules
//   - BR-G-*: Export outside EU rules
//   - BR-IC-*: Intra-community supply rules
//   - BR-AF-*: IGIC (Canary Islands) rules
//   - BR-AG-*: IPSI (Ceuta/Melilla) rules
//   - BR-O-*: Not subject to VAT rules
//   - BR-B-*: Split payment rules
package rules
