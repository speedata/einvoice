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

// Write writes the invoice to the given writer in the format specified by SchemaType.
// For CII (ZUGFeRD/Factur-X), it outputs XML with rsm/ram/udt/qdt namespaces.
// For UBL, it outputs XML with Invoice or CreditNote root based on InvoiceTypeCode.
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
