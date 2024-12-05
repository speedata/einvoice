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


## Limitation, current status

Coding just started, only the basic parts are implemented.

* Reading of EN 16931 ZUGFeRD XML files is possible
* No UBL based XML
* No checks done
* Not all possible fields are read
* No (XML) output

All of these points will be addressed. Stay tuned for updates!
