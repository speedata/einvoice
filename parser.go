package einvoice

import (
	"fmt"
	"io"
	"os"

	"github.com/shopspring/decimal"
	"github.com/speedata/cxpath"
)

// getDecimal parses a decimal value from an XPath evaluation result.
// Shared by both CII and UBL parsers.
func getDecimal(ctx *cxpath.Context, eval string) (decimal.Decimal, error) {
	a := ctx.Eval(eval).String()
	if a == "" {
		return decimal.Zero, nil
	}
	str, err := decimal.NewFromString(a)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid decimal value '%s' at %s: %w", a, eval, err)
	}
	return str, nil
}

// ParseReader reads the XML from the reader and auto-detects the format (CII or UBL).
// It detects the format by examining the root element namespace and routes to the
// appropriate parser. Each parser handles its own namespace setup.
func ParseReader(r io.Reader) (*Invoice, error) {
	ctx, err := cxpath.NewFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read from reader: %w", err)
	}

	// Detect format by checking root element namespace
	root := ctx.Root()
	rootns := root.Eval("namespace-uri()").String()

	var inv *Invoice

	switch rootns {
	case "":
		return nil, fmt.Errorf("empty root element namespace")

	// CII format (ZUGFeRD/Factur-X)
	case nsCIIRootInvoice:
		inv, err = parseCII(ctx)
		if err != nil {
			return nil, fmt.Errorf("parse CII: %w", err)
		}

	// UBL format (Invoice or CreditNote)
	case nsUBLInvoice, nsUBLCreditNote:
		inv, err = parseUBL(ctx)
		if err != nil {
			return nil, fmt.Errorf("parse UBL: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown root element namespace: %s", rootns)
	}

	return inv, nil
}

// ParseXMLFile reads the XML file at filename.
func ParseXMLFile(filename string) (*Invoice, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("einvoice: cannot open file (%w)", err)
	}
	defer func() { _ = r.Close() }()

	return ParseReader(r)
}
