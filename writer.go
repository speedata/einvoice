package einvoice

import (
	"errors"
	"io"
	"regexp"

	"github.com/shopspring/decimal"
)

var percentageRE = regexp.MustCompile(`^(.*?)\.?0+$`)
var ErrWrite = errors.New("creating the XML failed")

// Profile level constants for writer
// These match the ProfileLevel() return values in model.go
const (
	levelMinimum  = 1
	levelBasicWL  = 2
	levelBasic    = 3
	levelEN16931  = 4 // Also covers PEPPOL and XRechnung
	levelExtended = 5
)

// is returns true if the invoice profile meets or exceeds the minimum profile level.
// This replaces the old enum-based comparison.
func is(minLevel int, inv *Invoice) bool {
	return inv.ProfileLevel() >= minLevel
}

// formatPercent removes trailing zeros and the decimal point, if possible.
func formatPercent(d decimal.Decimal) string {
	str := d.StringFixed(4)

	return percentageRE.ReplaceAllString(str, "$1")
}

// ErrUnsupportedSchema is returned when the library does not recognize the schema.
var ErrUnsupportedSchema = errors.New("unsupported schema")

// Write writes the invoice as XML to the given writer.
//
// The output format is determined by the SchemaType field:
//   - UBL: Outputs UBL 2.1 XML with <Invoice> root (InvoiceTypeCode 380) or
//     <CreditNote> root (InvoiceTypeCode 381)
//   - CII or SchemaTypeUnknown: Outputs ZUGFeRD/Factur-X CII format with
//     rsm/ram/udt/qdt namespaces
//
// Programmatically created invoices have SchemaTypeUnknown (zero value) and
// default to CII format for backwards compatibility.
//
// Write does not perform validation. Consider calling Validate() before Write()
// to ensure the invoice meets EN 16931 requirements. You should also call
// UpdateApplicableTradeTax() and UpdateTotals() to calculate monetary values.
//
// Returns ErrUnsupportedSchema if SchemaType is not recognized.
//
// Example for CII format:
//
//	inv := &einvoice.Invoice{
//		SchemaType: einvoice.CII, // or omit for default
//		InvoiceNumber: "INV-001",
//		// ... set required fields
//	}
//	inv.UpdateApplicableTradeTax(nil)
//	inv.UpdateTotals()
//	if err := inv.Validate(); err != nil {
//		// handle validation errors
//	}
//	err := inv.Write(os.Stdout)
//
// Example for UBL format:
//
//	inv := &einvoice.Invoice{
//		SchemaType: einvoice.UBL,
//		InvoiceNumber: "INV-001",
//		InvoiceTypeCode: 380, // 380 = Invoice, 381 = CreditNote
//		// ... set required fields
//	}
//	inv.UpdateApplicableTradeTax(nil)
//	inv.UpdateTotals()
//	err := inv.Write(os.Stdout)
func (inv *Invoice) Write(w io.Writer) error {
	switch inv.SchemaType {
	case UBL:
		return writeUBL(inv, w)
	case CII, SchemaTypeUnknown:
		// Programmatically created invoices have SchemaTypeUnknown (zero value)
		// Treat them the same as CII for writing
		return writeCII(inv, w)
	default:
		return ErrUnsupportedSchema
	}
}
