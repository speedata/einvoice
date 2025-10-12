package einvoice

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/cxpath"
)

// parseTime parses CII format dates (YYYYMMDD) into time.Time.
// Shared by CII parser.
func parseTime(ctx *cxpath.Context, path string) (time.Time, error) {
	timestring := ctx.Eval(path).String()
	if timestring == "" {
		return time.Time{}, nil
	}

	parsedDate, err := time.Parse("20060102", timestring)
	if err != nil {
		return parsedDate, fmt.Errorf("%w", err)
	}

	return parsedDate, nil
}

// parseParty parses a party (buyer, seller, payee, etc.) from CII format.
// Shared by CII parser.
func parseParty(tradeParty *cxpath.Context) Party {
	adr := Party{}
	for id := range tradeParty.Each("ram:ID") {
		adr.ID = append(adr.ID, id.String())
	}

	for gid := range tradeParty.Each("ram:GlobalID") {
		scheme := GlobalID{
			Scheme: gid.Eval("@schemeID").String(),
			ID:     gid.String(),
		}
		adr.GlobalID = append(adr.GlobalID, scheme)
	}

	adr.Name = tradeParty.Eval("ram:Name").String()

	if tradeParty.Eval("count(ram:SpecifiedLegalOrganization) > 0").Bool() {
		slo := SpecifiedLegalOrganization{}
		slo.ID = tradeParty.Eval("ram:SpecifiedLegalOrganization/ram:ID").String()
		slo.Scheme = tradeParty.Eval("ram:SpecifiedLegalOrganization/ram:ID/@schemeID").String()
		slo.TradingBusinessName = tradeParty.Eval("ram:SpecifiedLegalOrganization/ram:TradingBusinessName").String()
		adr.SpecifiedLegalOrganization = &slo
	}

	for dtc := range tradeParty.Each("ram:DefinedTradeContact") {
		contact := DefinedTradeContact{}
		contact.EMail = dtc.Eval("ram:EmailURIUniversalCommunication/ram:URIID").String()
		contact.PhoneNumber = dtc.Eval("ram:TelephoneUniversalCommunication/ram:CompleteNumber").String()
		contact.PersonName = dtc.Eval("ram:PersonName").String()
		adr.DefinedTradeContact = append(adr.DefinedTradeContact, contact)
	}

	if tradeParty.Eval("count(ram:PostalTradeAddress)").Int() > 0 {
		postalAddress := &PostalAddress{
			PostcodeCode:           tradeParty.Eval("ram:PostalTradeAddress/ram:PostcodeCode").String(),
			Line1:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineOne").String(),
			Line2:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineTwo").String(),
			Line3:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineThree").String(),
			City:                   tradeParty.Eval("ram:PostalTradeAddress/ram:CityName").String(),
			CountryID:              tradeParty.Eval("ram:PostalTradeAddress/ram:CountryID").String(),
			CountrySubDivisionName: tradeParty.Eval("ram:PostalTradeAddress/ram:CountrySubDivisionName").String(),
		}
		adr.PostalAddress = postalAddress
	}

	adr.FCTaxRegistration = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='FC']").String()
	adr.VATaxRegistration = tradeParty.Eval("ram:SpecifiedTaxRegistration/ram:ID[@schemeID='VA']").String()

	return adr
}

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
	case "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100":
		inv, err = parseCII(ctx)
		if err != nil {
			return nil, fmt.Errorf("parse CII: %w", err)
		}

	// UBL format (Invoice or CreditNote)
	case "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		"urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2":
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
