package einvoice

import (
	"bytes"
	"testing"

	"github.com/shopspring/decimal"
)

// subInvoiceLineFixtures are official ZUGFeRD 2.5 / Factur-X 1.09 EXTENDED
// examples that exercise sub invoice lines (chapter 7.6.2): GROUP subtotal
// containers and INFORMATION breakdown lines, identified by the subtype
// LineStatusReasonCode (BT-X-8).
var subInvoiceLineFixtures = []struct {
	name string
	path string
}{
	{"group-hardware", "testdata/cii/extended/zf25-subline-group-hardware.xml"},
	{"group-bundle", "testdata/cii/extended/zf25-subline-group-bundle.xml"},
	{"information", "testdata/cii/extended/zf25-subline-information.xml"},
	{"nested", "testdata/cii/extended/zf25-subline-nested.xml"},
}

// TestSubInvoiceLines_Validate verifies that the official ZUGFeRD 2.5 sub
// invoice line examples parse and pass validation. Only detail lines (subtype
// "DETAIL" or unspecified) must contribute to the totals; aggregation lines
// (GROUP / INFORMATION) are excluded and exempt from the per-line detail rules.
func TestSubInvoiceLines_Validate(t *testing.T) {
	for _, fx := range subInvoiceLineFixtures {
		t.Run(fx.name, func(t *testing.T) {
			inv, err := ParseXMLFile(fx.path)
			if err != nil {
				t.Fatalf("parse %s: %v", fx.path, err)
			}

			if err := inv.Validate(); err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}

			// The fixture must actually contain at least one aggregation line,
			// otherwise the sub invoice line handling would not be exercised.
			details, aggregations := 0, 0
			for i := range inv.InvoiceLines {
				if inv.InvoiceLines[i].isDetailLine() {
					details++
				} else {
					aggregations++
				}
			}
			if aggregations == 0 {
				t.Fatalf("fixture %s has no GROUP/INFORMATION lines; not a sub invoice line case", fx.path)
			}

			// BR-FXEXT-CO-10: the line total (BT-106) is the sum of the detail
			// lines only; including aggregation lines would double count.
			sum := decimal.Zero
			for i := range inv.InvoiceLines {
				if inv.InvoiceLines[i].isDetailLine() {
					sum = sum.Add(inv.InvoiceLines[i].Total)
				}
			}
			if !inv.LineTotal.Equal(sum) {
				t.Errorf("LineTotal %s != sum of detail lines %s", inv.LineTotal, sum)
			}
		})
	}
}

// TestSubInvoiceLines_RoundTrip verifies that the sub invoice line metadata
// (ParentLineID BT-X-304, LineStatusReasonCode BT-X-8) survives a write/parse
// round-trip and that the re-parsed invoice still validates.
//
// LineStatusCode (BT-X-7) is covered separately in
// TestSubInvoiceLines_LineStatusCodeRoundTrip because the official fixtures do
// not populate it, so asserting it here would compare "" against "".
func TestSubInvoiceLines_RoundTrip(t *testing.T) {
	for _, fx := range subInvoiceLineFixtures {
		t.Run(fx.name, func(t *testing.T) {
			inv, err := ParseXMLFile(fx.path)
			if err != nil {
				t.Fatalf("parse %s: %v", fx.path, err)
			}

			var buf bytes.Buffer
			if err := inv.Write(&buf); err != nil {
				t.Fatalf("Write: %v", err)
			}

			rt, err := ParseReader(&buf)
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}

			if err := rt.Validate(); err != nil {
				t.Errorf("Validate() after round-trip = %v, want nil", err)
			}

			if len(rt.InvoiceLines) != len(inv.InvoiceLines) {
				t.Fatalf("round-trip line count = %d, want %d", len(rt.InvoiceLines), len(inv.InvoiceLines))
			}
			for i := range inv.InvoiceLines {
				if got, want := rt.InvoiceLines[i].ParentLineID, inv.InvoiceLines[i].ParentLineID; got != want {
					t.Errorf("line %s ParentLineID = %q, want %q", inv.InvoiceLines[i].LineID, got, want)
				}
				if got, want := rt.InvoiceLines[i].LineStatusReasonCode, inv.InvoiceLines[i].LineStatusReasonCode; got != want {
					t.Errorf("line %s LineStatusReasonCode = %q, want %q", inv.InvoiceLines[i].LineID, got, want)
				}
				if got, want := rt.InvoiceLines[i].LineStatusCode, inv.InvoiceLines[i].LineStatusCode; got != want {
					t.Errorf("line %s LineStatusCode = %q, want %q", inv.InvoiceLines[i].LineID, got, want)
				}
			}
		})
	}
}

// TestSubInvoiceLines_LineStatusCodeRoundTrip covers the third new sub invoice
// line field, LineStatusCode (BT-X-7), end to end. The official fixtures leave it
// empty, so it is seeded onto an aggregation line before writing to prove that
// the writer emits ram:LineStatusCode and the parser reads it back.
func TestSubInvoiceLines_LineStatusCodeRoundTrip(t *testing.T) {
	const path = "testdata/cii/extended/zf25-subline-nested.xml"
	inv, err := ParseXMLFile(path)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	// Seed BT-X-7 on the first aggregation line with a UNTDID 4405 status code.
	const wantCode = "39"
	seeded := -1
	for i := range inv.InvoiceLines {
		if !inv.InvoiceLines[i].isDetailLine() {
			inv.InvoiceLines[i].LineStatusCode = wantCode
			seeded = i
			break
		}
	}
	if seeded == -1 {
		t.Fatalf("fixture %s has no aggregation line to seed", path)
	}
	wantLineID := inv.InvoiceLines[seeded].LineID

	var buf bytes.Buffer
	if err := inv.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}

	rt, err := ParseReader(&buf)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	found := false
	for i := range rt.InvoiceLines {
		if rt.InvoiceLines[i].LineID != wantLineID {
			continue
		}
		found = true
		if got := rt.InvoiceLines[i].LineStatusCode; got != wantCode {
			t.Errorf("line %s LineStatusCode = %q, want %q", wantLineID, got, wantCode)
		}
	}
	if !found {
		t.Fatalf("seeded line %s not found after round-trip", wantLineID)
	}
}
