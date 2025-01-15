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
// invoice.Violations contains a slice of possible logical errors in the XML file
```

writing an invoice:

```go
func dothings() error {
	inv, err := einvoice.ParseXMLFile("...")
	if err != nil {
		return err
	}
    // or better, create all the necessary fields yourself


	return inv.Write(os.Stdout)
}
```

There is a [dedicated example](https://pkg.go.dev/github.com/speedata/einvoice#example-Invoice.Write) in [the documentation](https://pkg.go.dev/github.com/speedata/einvoice).


## Current status

Coding just started, only the basic parts are implemented.

* Reading and writing of EN 16931 ZUGFeRD XML files is possible
* Some checks are performed when reading the ZUGFeRD file (BR-1 to BR-45)
* XML output for minimum and EN16931 ZUGFeRD profile

## Limitations

* No UBL based XML

These points will be addressed. Stay tuned for updates!

