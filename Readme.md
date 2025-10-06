[![Go Reference](https://pkg.go.dev/badge/github.com/speedata/einvoice.svg)](https://pkg.go.dev/github.com/speedata/einvoice)

# einvoice - a Go library to read, write and verify electronic invoices

**Work in progress**

This library will be used to read, write and verify electronic invoices (XML) which conform to the EN 16931 standard.

## Installation

    go get github.com/speedata/einvoice

## Usage

```go
invoice, err := einvoice.ParseXMLFile(filename)
if err != nil {
	...
}
// now invoice contains all the information from the XML file
// check for validation violations
err = invoice.Validate()
if err != nil {
	var valErr *einvoice.ValidationError
	if errors.As(err, &valErr) {
		for _, v := range valErr.Violations() {
			fmt.Printf("Rule %s: %s\n", v.Rule, v.Text)
		}
	}
}
```

Building and validating an invoice programmatically:

```go
import "errors"

func buildInvoice() error {
	inv := &einvoice.Invoice{
		InvoiceNumber: "INV-001",
		// ... set all required fields
	}

	// Validate before writing
	if err := inv.Validate(); err != nil {
		var valErr *einvoice.ValidationError
		if errors.As(err, &valErr) {
			for _, v := range valErr.Violations() {
				fmt.Printf("Rule %s: %s\n", v.Rule, v.Text)
			}
		}
		return err
	}

	return inv.Write(os.Stdout)
}
```

Writing an invoice (parsing):

```go
func dothings() error {
	inv, err := einvoice.ParseXMLFile("...")
	if err != nil {
		return err
	}

	return inv.Write(os.Stdout)
}
```

There is a [dedicated example](https://pkg.go.dev/github.com/speedata/einvoice#example-Invoice.Write) in [the documentation](https://pkg.go.dev/github.com/speedata/einvoice).


## Current status

* Reading and writing of EN 16931 ZUGFeRD XML files is possible
* Comprehensive validation of business rules when reading ZUGFeRD files:
  - BR-1 to BR-65 (all core business rules)
  - BR-CO-3 to BR-CO-26 (calculation rules)
  - BR-S-1 to BR-S-10 (Standard rate VAT)
  - BR-AE-1 to BR-AE-10 (Reverse charge)
  - BR-E-1 to BR-E-10 (Exempt from VAT)
  - BR-Z-1 to BR-Z-10 (Zero rated VAT)
  - BR-G-1 to BR-G-10 (Export outside EU)
  - BR-IC-1 to BR-IC-12 (Intra-community supply)
  - BR-IG-1 to BR-IG-10 (IGIC - Canary Islands)
  - BR-IP-1 to BR-IP-10 (IPSI - Ceuta/Melilla)
  - BR-O-1 to BR-O-14 (Not subject to VAT)
* XML output for minimum and EN16931 ZUGFeRD profile

## Limitations

* No UBL based XML

These points will be addressed. Stay tuned for updates!

