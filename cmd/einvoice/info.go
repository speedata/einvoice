package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/speedata/einvoice"
	"github.com/speedata/einvoice/cmd/codelists"
)

// InvoiceInfo represents the complete invoice information for display
type InvoiceInfo struct {
	File    string          `json:"file"`
	Invoice *InvoiceDetails `json:"invoice,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// InvoiceDetails contains detailed invoice information
type InvoiceDetails struct {
	Number          string          `json:"number"`
	Date            string          `json:"date"`
	Type            string          `json:"type"`
	Profile         string          `json:"profile"`
	ProfileURN      string          `json:"profile_urn,omitempty"`
	BusinessProcess string          `json:"business_process,omitempty"`
	Currency        string          `json:"currency"`
	Seller          *PartyInfo      `json:"seller"`
	Buyer           *PartyInfo      `json:"buyer"`
	Lines           []LineInfo      `json:"lines,omitempty"`
	LineCount       int             `json:"line_count"`
	Totals          *TotalsInfo     `json:"totals"`
	PaymentTerms    []string        `json:"payment_terms,omitempty"`
	Notes           []string        `json:"notes,omitempty"`
}

// PartyInfo contains party details
type PartyInfo struct {
	Name    string  `json:"name"`
	VATNumber string `json:"vat_number,omitempty"`
	Address *AddressInfo `json:"address,omitempty"`
}

// AddressInfo contains address details
type AddressInfo struct {
	Street      string `json:"street,omitempty"`
	City        string `json:"city"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
}

// LineInfo contains invoice line details
type LineInfo struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Quantity    string `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	NetAmount   string `json:"net_amount"`
}

// TotalsInfo contains all monetary totals
type TotalsInfo struct {
	LineTotal       string `json:"line_total"`
	AllowanceTotal  string `json:"allowance_total,omitempty"`
	ChargeTotal     string `json:"charge_total,omitempty"`
	TaxBasisTotal   string `json:"tax_basis_total"`
	TaxTotal        string `json:"tax_total"`
	GrandTotal      string `json:"grand_total"`
	TotalPrepaid    string `json:"total_prepaid,omitempty"`
	RoundingAmount  string `json:"rounding_amount,omitempty"`
	DuePayableAmount string `json:"due_payable_amount"`
}


func runInfo(args []string) int {
	// Parse flags for the info subcommand
	infoFlags := flag.NewFlagSet("info", flag.ExitOnError)
	var format string
	var showCodes bool
	var verbose bool
	infoFlags.StringVar(&format, "format", "text", "Output format: text, json")
	infoFlags.BoolVar(&showCodes, "show-codes", false, "Show raw codes instead of descriptions")
	infoFlags.BoolVar(&verbose, "vv", false, "Show both codes and descriptions")
	infoFlags.Usage = infoUsage
	_ = infoFlags.Parse(args)

	// Require exactly one file argument
	if infoFlags.NArg() != 1 {
		infoUsage()
		return exitError
	}

	filename := infoFlags.Arg(0)

	// Get invoice information
	info := getInvoiceInfo(filename, showCodes, verbose)

	// Output results
	switch format {
	case "json":
		outputInfoJSON(info)
	case "text":
		outputInfoText(info)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format %q (use 'text' or 'json')\n", format)
		return exitError
	}

	// Return appropriate exit code
	if info.Error != "" {
		return exitError
	}
	return exitOK
}

// formatDocumentType formats a document type code based on display flags.
// - showCodes=true: returns only the code (e.g., "380")
// - verbose=true: returns code with description (e.g., "380 (Commercial invoice)")
// - default: returns description only, or "Unknown (code)" if not found
func formatDocumentType(code string, showCodes bool, verbose bool) string {
	if showCodes {
		return code
	}

	description := codelists.DocumentType(code)

	if verbose {
		return code + " (" + description + ")"
	}

	// Default mode: description only, but show code for unknown types
	if description == "Unknown" {
		return "Unknown (" + code + ")"
	}
	return description
}

// formatUnitCode formats a unit code based on display flags.
// - showCodes=true: returns only the code (e.g., "XPP")
// - verbose=true: returns code with description (e.g., "XPP (package)")
// - default: returns description if found, code if not found
func formatUnitCode(code string, showCodes bool, verbose bool) string {
	if showCodes {
		return code
	}

	description := codelists.UnitCode(code)

	if verbose {
		// If description equals code, it means it wasn't found
		if description == code {
			return code
		}
		return code + " (" + description + ")"
	}

	// Default mode: return whatever UnitCode returns (description or code)
	return description
}

func getInvoiceInfo(filename string, showCodes bool, verbose bool) InvoiceInfo {
	info := InvoiceInfo{
		File: filename,
	}

	// Parse the invoice XML
	invoice, err := einvoice.ParseXMLFile(filename)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to parse invoice: %v", err)
		return info
	}

	// Extract invoice details
	typeCode := invoice.InvoiceTypeCode.String()
	details := &InvoiceDetails{
		Number:          invoice.InvoiceNumber,
		Type:            formatDocumentType(typeCode, showCodes, verbose),
		Profile:         einvoice.GetProfileName(invoice.GuidelineSpecifiedDocumentContextParameter),
		ProfileURN:      invoice.GuidelineSpecifiedDocumentContextParameter,
		BusinessProcess: invoice.BPSpecifiedDocumentContextParameter,
		Currency:        invoice.InvoiceCurrencyCode,
		LineCount:       len(invoice.InvoiceLines),
	}

	// Format date
	if !invoice.InvoiceDate.IsZero() {
		details.Date = invoice.InvoiceDate.Format("2006-01-02")
	}

	// Extract seller information
	details.Seller = &PartyInfo{
		Name:      invoice.Seller.Name,
		VATNumber: invoice.Seller.VATaxRegistration,
	}
	if invoice.Seller.PostalAddress != nil {
		street := invoice.Seller.PostalAddress.Line1
		if invoice.Seller.PostalAddress.Line2 != "" {
			street += ", " + invoice.Seller.PostalAddress.Line2
		}
		if invoice.Seller.PostalAddress.Line3 != "" {
			street += ", " + invoice.Seller.PostalAddress.Line3
		}
		details.Seller.Address = &AddressInfo{
			Street:     street,
			City:       invoice.Seller.PostalAddress.City,
			PostalCode: invoice.Seller.PostalAddress.PostcodeCode,
			Country:    invoice.Seller.PostalAddress.CountryID,
		}
	}

	// Extract buyer information
	details.Buyer = &PartyInfo{
		Name:      invoice.Buyer.Name,
		VATNumber: invoice.Buyer.VATaxRegistration,
	}
	if invoice.Buyer.PostalAddress != nil {
		street := invoice.Buyer.PostalAddress.Line1
		if invoice.Buyer.PostalAddress.Line2 != "" {
			street += ", " + invoice.Buyer.PostalAddress.Line2
		}
		if invoice.Buyer.PostalAddress.Line3 != "" {
			street += ", " + invoice.Buyer.PostalAddress.Line3
		}
		details.Buyer.Address = &AddressInfo{
			Street:     street,
			City:       invoice.Buyer.PostalAddress.City,
			PostalCode: invoice.Buyer.PostalAddress.PostcodeCode,
			Country:    invoice.Buyer.PostalAddress.CountryID,
		}
	}

	// Extract invoice lines
	details.Lines = make([]LineInfo, 0, len(invoice.InvoiceLines))
	for _, line := range invoice.InvoiceLines {
		lineInfo := LineInfo{
			ID:        line.LineID,
			NetAmount: line.Total.String(),
		}

		if line.ItemName != "" {
			lineInfo.Description = line.ItemName
		}

		if !line.BilledQuantity.IsZero() {
			lineInfo.Quantity = line.BilledQuantity.String()
			if line.BilledQuantityUnit != "" {
				unitName := formatUnitCode(line.BilledQuantityUnit, showCodes, verbose)
				lineInfo.Quantity += " " + unitName
			}
		}

		// Prefer gross price, fall back to net price
		if !line.GrossPrice.IsZero() {
			lineInfo.UnitPrice = line.GrossPrice.String()
		} else if !line.NetPrice.IsZero() {
			lineInfo.UnitPrice = line.NetPrice.String()
		}

		details.Lines = append(details.Lines, lineInfo)
	}

	// Extract totals
	details.Totals = &TotalsInfo{
		LineTotal:        invoice.LineTotal.String(),
		TaxBasisTotal:    invoice.TaxBasisTotal.String(),
		TaxTotal:         invoice.TaxTotal.String(),
		GrandTotal:       invoice.GrandTotal.String(),
		DuePayableAmount: invoice.DuePayableAmount.String(),
	}

	if !invoice.AllowanceTotal.IsZero() {
		details.Totals.AllowanceTotal = invoice.AllowanceTotal.String()
	}
	if !invoice.ChargeTotal.IsZero() {
		details.Totals.ChargeTotal = invoice.ChargeTotal.String()
	}
	if !invoice.TotalPrepaid.IsZero() {
		details.Totals.TotalPrepaid = invoice.TotalPrepaid.String()
	}
	if !invoice.RoundingAmount.IsZero() {
		details.Totals.RoundingAmount = invoice.RoundingAmount.String()
	}

	// Extract payment terms
	details.PaymentTerms = make([]string, 0, len(invoice.SpecifiedTradePaymentTerms))
	for _, term := range invoice.SpecifiedTradePaymentTerms {
		if term.Description != "" {
			details.PaymentTerms = append(details.PaymentTerms, term.Description)
		}
	}

	// Extract notes
	details.Notes = make([]string, 0, len(invoice.Notes))
	for _, note := range invoice.Notes {
		if note.Text != "" {
			details.Notes = append(details.Notes, note.Text)
		}
	}

	info.Invoice = details

	return info
}

func outputInfoText(info InvoiceInfo) {
	if info.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", info.Error)
		return
	}

	inv := info.Invoice

	// Header section
	fmt.Printf("Invoice Information\n")
	fmt.Printf("===================\n\n")

	// Basic invoice details
	fmt.Printf("Invoice Number:  %s\n", inv.Number)
	fmt.Printf("Date:            %s\n", inv.Date)
	fmt.Printf("Type:            %s\n", inv.Type)
	fmt.Printf("Profile:         %s\n", inv.Profile)
	if inv.BusinessProcess != "" {
		fmt.Printf("Business Process: %s\n", inv.BusinessProcess)
	}
	fmt.Printf("Currency:        %s\n", inv.Currency)
	fmt.Printf("\n")

	// Seller information
	fmt.Printf("Seller\n")
	fmt.Printf("------\n")
	fmt.Printf("Name:            %s\n", inv.Seller.Name)
	if inv.Seller.VATNumber != "" {
		fmt.Printf("VAT Number:      %s\n", inv.Seller.VATNumber)
	}
	if inv.Seller.Address != nil {
		if inv.Seller.Address.Street != "" {
			fmt.Printf("Address:         %s\n", inv.Seller.Address.Street)
		}
		fmt.Printf("                 %s %s\n", inv.Seller.Address.PostalCode, inv.Seller.Address.City)
		fmt.Printf("                 %s\n", inv.Seller.Address.Country)
	}
	fmt.Printf("\n")

	// Buyer information
	fmt.Printf("Buyer\n")
	fmt.Printf("-----\n")
	fmt.Printf("Name:            %s\n", inv.Buyer.Name)
	if inv.Buyer.VATNumber != "" {
		fmt.Printf("VAT Number:      %s\n", inv.Buyer.VATNumber)
	}
	if inv.Buyer.Address != nil {
		if inv.Buyer.Address.Street != "" {
			fmt.Printf("Address:         %s\n", inv.Buyer.Address.Street)
		}
		fmt.Printf("                 %s %s\n", inv.Buyer.Address.PostalCode, inv.Buyer.Address.City)
		fmt.Printf("                 %s\n", inv.Buyer.Address.Country)
	}
	fmt.Printf("\n")

	// Invoice lines
	fmt.Printf("Invoice Lines (%d items)\n", inv.LineCount)
	fmt.Printf("---------------------\n")
	for _, line := range inv.Lines {
		fmt.Printf("Line %s:\n", line.ID)
		if line.Description != "" {
			fmt.Printf("  Description:   %s\n", line.Description)
		}
		if line.Quantity != "" {
			fmt.Printf("  Quantity:      %s\n", line.Quantity)
		}
		if line.UnitPrice != "" {
			fmt.Printf("  Unit Price:    %s %s\n", line.UnitPrice, inv.Currency)
		}
		fmt.Printf("  Net Amount:    %s %s\n", line.NetAmount, inv.Currency)
		fmt.Printf("\n")
	}

	// Totals
	fmt.Printf("Totals\n")
	fmt.Printf("------\n")
	fmt.Printf("Line Total:      %s %s\n", inv.Totals.LineTotal, inv.Currency)
	if inv.Totals.AllowanceTotal != "" {
		fmt.Printf("Allowances:      -%s %s\n", inv.Totals.AllowanceTotal, inv.Currency)
	}
	if inv.Totals.ChargeTotal != "" {
		fmt.Printf("Charges:         +%s %s\n", inv.Totals.ChargeTotal, inv.Currency)
	}
	fmt.Printf("Tax Basis:       %s %s\n", inv.Totals.TaxBasisTotal, inv.Currency)
	fmt.Printf("Tax Total:       %s %s\n", inv.Totals.TaxTotal, inv.Currency)
	fmt.Printf("Grand Total:     %s %s\n", inv.Totals.GrandTotal, inv.Currency)
	if inv.Totals.TotalPrepaid != "" {
		fmt.Printf("Prepaid:         -%s %s\n", inv.Totals.TotalPrepaid, inv.Currency)
	}
	if inv.Totals.RoundingAmount != "" {
		fmt.Printf("Rounding:        %s %s\n", inv.Totals.RoundingAmount, inv.Currency)
	}
	fmt.Printf("Due Amount:      %s %s\n", inv.Totals.DuePayableAmount, inv.Currency)
	fmt.Printf("\n")

	// Payment terms
	if len(inv.PaymentTerms) > 0 {
		fmt.Printf("Payment Terms\n")
		fmt.Printf("-------------\n")
		for _, term := range inv.PaymentTerms {
			fmt.Printf("%s\n", term)
		}
		fmt.Printf("\n")
	}

	// Notes
	if len(inv.Notes) > 0 {
		fmt.Printf("Notes\n")
		fmt.Printf("-----\n")
		for _, note := range inv.Notes {
			fmt.Printf("%s\n", note)
		}
		fmt.Printf("\n")
	}

}

func outputInfoJSON(info InvoiceInfo) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(info); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

func infoUsage() {
	fmt.Fprintf(os.Stderr, `Usage: einvoice info [options] <file.xml>

Display detailed information about an electronic invoice.

Options:
  --format string   Output format: text, json (default "text")
  --show-codes      Show raw codes instead of descriptions
  -vv               Show both codes and descriptions
  --help            Show this help message

Display modes:
  Default:      Shows human-readable descriptions (e.g., Type: Commercial invoice)
  --show-codes: Shows only raw codes (e.g., Type: 380)
  -vv:          Shows both codes and descriptions (e.g., Type: 380 (Commercial invoice))

Note: If both --show-codes and -vv are provided, --show-codes takes precedence.

Exit codes:
  0  Success
  1  Error occurred (file not found, parse error, etc.)

Examples:
  einvoice info invoice.xml
  einvoice info --show-codes invoice.xml
  einvoice info -vv invoice.xml
  einvoice info --format json invoice.xml
`)
}
