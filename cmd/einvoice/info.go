package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/speedata/einvoice"
	"github.com/speedata/einvoice/pkg/codelists"
	"golang.org/x/term"
)

//go:embed templates/text-default.gotmpl
var embeddedTemplates embed.FS

// InvoiceInfo represents the complete invoice information for display
type InvoiceInfo struct {
	File    string          `json:"file"`
	Invoice *InvoiceDetails `json:"invoice,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// NoteInfo holds an invoice note and its subject qualifier (e.g., UNCL 4451
// 3-digit code).
type NoteInfo struct {
	Text             string `json:"text"`
	SubjectQualifier string `json:"subject_qualifier,omitempty"`
}

// InvoiceDetails contains detailed invoice information
type InvoiceDetails struct {
	Number          string      `json:"number"`
	Date            string      `json:"date"`
	Type            string      `json:"type"`
	Profile         string      `json:"profile"`
	ProfileURN      string      `json:"profile_urn,omitempty"`
	BusinessProcess string      `json:"business_process,omitempty"`
	Currency        string      `json:"currency"`
	Seller          *PartyInfo  `json:"seller"`
	Buyer           *PartyInfo  `json:"buyer"`
	Lines           []LineInfo  `json:"lines,omitempty"`
	LineCount       int         `json:"line_count"`
	Totals          *TotalsInfo `json:"totals"`
	PaymentTerms    []string    `json:"payment_terms,omitempty"`
	Notes           []NoteInfo  `json:"notes,omitempty"`
	TermWidth       int         `json:"-"`
}

// PartyInfo contains party details
type PartyInfo struct {
	Name      string       `json:"name"`
	VATNumber string       `json:"vat_number,omitempty"`
	Address   *AddressInfo `json:"address,omitempty"`
}

// AddressInfo contains address details
type AddressInfo struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
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
	LineTotal        string `json:"line_total"`
	AllowanceTotal   string `json:"allowance_total,omitempty"`
	ChargeTotal      string `json:"charge_total,omitempty"`
	TaxBasisTotal    string `json:"tax_basis_total"`
	TaxTotal         string `json:"tax_total"`
	GrandTotal       string `json:"grand_total"`
	TotalPrepaid     string `json:"total_prepaid,omitempty"`
	RoundingAmount   string `json:"rounding_amount,omitempty"`
	DuePayableAmount string `json:"due_payable_amount"`
}

func detectTerminalWidth() int {
	// 1) Try real terminal size
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	// 2) Fall back to $COLUMNS if present
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 {
			return n
		}
	}
	// 3) Sensible default
	return 80
}
func runInfo(args []string) int {
	// Parse flags for the info subcommand
	infoFlags := flag.NewFlagSet("info", flag.ExitOnError)
	var format string
	var showCodes bool
	var verbose bool
	var templatePath string

	infoFlags.StringVar(&format, "format", "text", "Output format: text, json")
	infoFlags.BoolVar(&showCodes, "show-codes", false, "Show raw codes instead of descriptions")
	infoFlags.BoolVar(&verbose, "vv", false, "Show both codes and descriptions")
	infoFlags.StringVar(&templatePath, "template", "", "Path to a custom Go text template file")
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
		if err := outputInfoTextTemplate(info, templatePath); err != nil {
			fmt.Fprintf(os.Stderr, "Template error: %v\n", err)
			return exitError
		}
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

// outputInfoTextTemplate renders the text output using either a user-supplied or embedded template.
func outputInfoTextTemplate(info InvoiceInfo, templatePath string) error {
	if info.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", info.Error)
		return nil
	}

	var tplBytes []byte
	var err error

	// Load template: use embedded one if none provided
	switch {
	case templatePath == "":
		tplBytes, err = embeddedTemplates.ReadFile("templates/text-default.gotmpl")
	case strings.HasPrefix(templatePath, "builtin:"):
		// Reserved for possible future builtins (e.g., "builtin:compact")
		name := strings.TrimPrefix(templatePath, "builtin:")
		tplBytes, err = embeddedTemplates.ReadFile("templates/" + name + ".gotmpl")
	default:
		tplBytes, err = os.ReadFile(templatePath)
	}
	if err != nil {
		return fmt.Errorf("cannot load template: %w", err)
	}

	funcMap := template.FuncMap{
		"wrap": wrapText,
		"pad":  padRight,
		"hr":   hr,
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		"nonempty": func(s string) bool { return strings.TrimSpace(s) != "" },
		"sub1":     func(i int) int { return i - 1 },
	}

	tpl, err := template.New("invoice").Funcs(funcMap).Parse(string(tplBytes))
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	return tpl.Execute(os.Stdout, info.Invoice)
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

	// Parse the invoice (XML or PDF)
	invoice, err := parseInvoiceFile(filename)
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
	for _, n := range invoice.Notes {
		// Skip empty notes
		if n.Text == "" {
			continue
		}

		details.Notes = append(details.Notes, NoteInfo{
			Text:             n.Text,
			SubjectQualifier: n.SubjectCode,
		})
	}

	details.TermWidth = detectTerminalWidth()
	info.Invoice = details

	return info
}

func outputInfoJSON(info InvoiceInfo) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(info); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

func infoUsage() {
	fmt.Fprintf(os.Stderr, `Usage: einvoice info [options] <file>

Display detailed information about an electronic invoice.

Supports both XML and ZUGFeRD/Factur-X PDF formats.

Options:
  --format string   Output format: text, json (default "text")
  --show-codes      Show raw codes instead of descriptions
  --template string Path to a custom Go text template file
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
  einvoice info invoice.pdf
  einvoice info --show-codes invoice.xml
  einvoice info -vv invoice.pdf
  einvoice info --format json invoice.xml
  einvoice info --template custom-template.gotmpl invoice.pdf
`)
}
