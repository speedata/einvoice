package einvoice

import "testing"

func TestSimple(t *testing.T) {
	inv, err := ParseXMLFile("testcases/zugferd_2p0_EN16931_1_Teilrechnung.xml")
	if err != nil {
		t.Error(err)
	}
	expected := Invoice{
		InvoiceNumber: "471102",
	}

	if got := inv.InvoiceNumber; got != expected.InvoiceNumber {
		t.Errorf("invoice number got %s, expected %s\n", got, expected.InvoiceNumber)
	}
}
