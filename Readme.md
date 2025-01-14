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



## Limitation, current status

Coding just started, only the basic parts are implemented.

* Reading and writing of EN 16931 ZUGFeRD XML files is possible
* Some checks can be performed (BR-1 to BR-41)
* XML output for minimum and EN16931 ZUGFeRD profile

* No UBL based XML

These points will be addressed. Stay tuned for updates!
