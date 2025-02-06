package einvoice_test

import (
	"testing"

	"github.com/speedata/einvoice"
)

func TestSimple(t *testing.T) {
	t.Parallel()

	inv, err := einvoice.ParseXMLFile("testcases/zugferd_2p0_EN16931_1_Teilrechnung.xml")
	if err != nil {
		t.Error(err)
	}

	expected := einvoice.Invoice{
		InvoiceNumber: "471102",
	}
	if got := inv.InvoiceNumber; got != expected.InvoiceNumber {
		t.Errorf("invoice number got %s, expected %s\n", got, expected.InvoiceNumber)
	}
}
