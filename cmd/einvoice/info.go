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
	Number                       string                `json:"number"`
	Date                         string                `json:"date"`
	Type                         string                `json:"type"`
	Profile                      string                `json:"profile"`
	ProfileURN                   string                `json:"profile_urn,omitempty"`
	BusinessProcess              string                `json:"business_process,omitempty"`
	Currency                     string                `json:"currency"`
	BuyerOrderReferencedDocument string                `json:"buyer_order_referenced_document,omitempty"`
	Seller                       *PartyInfo            `json:"seller"`
	Buyer                        *PartyInfo            `json:"buyer"`
	Lines                        []LineInfo            `json:"lines,omitempty"`
	LineCount                    int                   `json:"line_count"`
	Totals                       *TotalsInfo           `json:"totals"`
	PaymentTerms                 []string              `json:"payment_terms,omitempty"`
	TradeTax                     []TaxInfo             `json:"trade_tax,omitempty"`
	ChargeAllowances             []ChargeAllowanceInfo `json:"charge_allowances,omitempty"`
	Notes                        []NoteInfo            `json:"notes,omitempty"`
	TermWidth                    int                   `json:"-"`
}

// PartyInfo contains party details
type PartyInfo struct {
	Name      string       `json:"name"`
	ID        string       `json:"id,omitempty"`
	GlobalID  string       `json:"global_id,omitempty"`
	VATNumber string       `json:"vat_number,omitempty"`
	TaxNumber string       `json:"tax_number,omitempty"`
	Address   *AddressInfo `json:"address,omitempty"`
}

// AddressInfo contains address details
type AddressInfo struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	Line3      string `json:"line3,omitempty"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// LineInfo contains invoice line details
type LineInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	Quantity    string `json:"quantity"`
	NetPrice    string `json:"net_price"`
	GrossPrice  string `json:"gross_price,omitempty"`
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

// TaxInfo contains tax breakdown details
type TaxInfo struct {
	CalculatedAmount string `json:"calculated_amount"`
	Percent          string `json:"percent"`
	Type             string `json:"type,omitempty"`
	Category         string `json:"category,omitempty"`
	ExemptionReason  string `json:"exemption_reason,omitempty"`
	ExemptionCode    string `json:"exemption_code,omitempty"`
	BasisAmount      string `json:"basis_amount,omitempty"`
}

type ChargeAllowanceInfo struct {
	ChargeIndicator bool   `json:"charge_indicator"`
	Amount          string `json:"amount"`
	Reason          string `json:"reason,omitempty"`
	Type            string `json:"type,omitempty"`
	Category        string `json:"category,omitempty"`
	Percent         string `json:"percent,omitempty"`
	BasisAmount     string `json:"basis_amount,omitempty"`
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
		"add":       func(a, b int) int { return a + b },
		"sub":       func(a, b int) int { return a - b },
		"wrap":      wrapTextIndent,
		"padright":  padRight,
		"padmiddle": padMiddle,
		"hr":        hr,
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
		"nonempty":  func(s string) bool { return strings.TrimSpace(s) != "" },
		"sub1":      func(i int) int { return i - 1 },
		"underline": underline,
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
		ID:        strings.Join(invoice.Seller.ID, ", "),
		Name:      invoice.Seller.Name,
		VATNumber: invoice.Seller.VATaxRegistration,
		TaxNumber: invoice.Seller.FCTaxRegistration,
	}
	// Add each global id in the form "scheme:id" to GlobalID
	if len(invoice.Seller.GlobalID) > 0 {
		var gids []string
		for _, gid := range invoice.Seller.GlobalID {
			if gid.ID != "" && gid.Scheme != "" {
				gids = append(gids, fmt.Sprintf("%s:%s", gid.Scheme, gid.ID))
			} else if gid.ID != "" {
				gids = append(gids, gid.ID)
			}
		}
		details.Seller.GlobalID = strings.Join(gids, ", ")
	}

	if invoice.Seller.PostalAddress != nil {
		details.Seller.Address = &AddressInfo{
			Line1:      invoice.Seller.PostalAddress.Line1,
			Line2:      invoice.Seller.PostalAddress.Line2,
			Line3:      invoice.Seller.PostalAddress.Line3,
			City:       invoice.Seller.PostalAddress.City,
			PostalCode: invoice.Seller.PostalAddress.PostcodeCode,
			Country:    invoice.Seller.PostalAddress.CountryID,
		}
	}

	// Extract buyer information
	details.Buyer = &PartyInfo{
		ID:        strings.Join(invoice.Buyer.ID, ", "),
		Name:      invoice.Buyer.Name,
		VATNumber: invoice.Buyer.VATaxRegistration,
	}
	// Add each global id in the form "scheme:id" to GlobalID
	if len(invoice.Buyer.GlobalID) > 0 {
		var gids []string
		for _, gid := range invoice.Buyer.GlobalID {
			if gid.ID != "" && gid.Scheme != "" {
				gids = append(gids, fmt.Sprintf("%s:%s", gid.Scheme, gid.ID))
			} else if gid.ID != "" {
				gids = append(gids, gid.ID)
			}
		}
		details.Buyer.GlobalID = strings.Join(gids, ", ")
	}

	if invoice.Buyer.PostalAddress != nil {
		details.Buyer.Address = &AddressInfo{
			Line1: invoice.Buyer.PostalAddress.Line1,
			Line2: invoice.Buyer.PostalAddress.Line2,
			Line3: invoice.Buyer.PostalAddress.Line3,

			City:       invoice.Buyer.PostalAddress.City,
			PostalCode: invoice.Buyer.PostalAddress.PostcodeCode,
			Country:    invoice.Buyer.PostalAddress.CountryID,
		}
	}

	// Extract invoice lines
	details.Lines = make([]LineInfo, 0, len(invoice.InvoiceLines))
	for _, line := range invoice.InvoiceLines {
		lineInfo := LineInfo{
			ID:          line.LineID,
			NetAmount:   line.Total.String(),
			Description: line.Description,
			Name:        line.ItemName,
		}

		if !line.BilledQuantity.IsZero() {
			lineInfo.Quantity = line.BilledQuantity.String()
			if line.BilledQuantityUnit != "" {
				unitName := formatUnitCode(line.BilledQuantityUnit, showCodes, verbose)
				lineInfo.Quantity += " " + unitName
			}
		}

		lineInfo.NetPrice = line.NetPrice.String()
		lineInfo.GrossPrice = line.GrossPrice.String()
		lineInfo.Note = line.Note
		details.Lines = append(details.Lines, lineInfo)
	}

	// Extract totals
	details.Totals = &TotalsInfo{
		LineTotal:        invoice.LineTotal.StringFixed(2),
		TaxBasisTotal:    invoice.TaxBasisTotal.StringFixed(2),
		TaxTotal:         invoice.TaxTotal.StringFixed(2),
		GrandTotal:       invoice.GrandTotal.StringFixed(2),
		DuePayableAmount: invoice.DuePayableAmount.StringFixed(2),
	}

	if !invoice.AllowanceTotal.IsZero() {
		details.Totals.AllowanceTotal = invoice.AllowanceTotal.StringFixed(2)
	}
	if !invoice.ChargeTotal.IsZero() {
		details.Totals.ChargeTotal = invoice.ChargeTotal.StringFixed(2)
	}
	if !invoice.TotalPrepaid.IsZero() {
		details.Totals.TotalPrepaid = invoice.TotalPrepaid.StringFixed(2)
	}
	if !invoice.RoundingAmount.IsZero() {
		details.Totals.RoundingAmount = invoice.RoundingAmount.StringFixed(2)
	}

	details.BuyerOrderReferencedDocument = invoice.BuyerOrderReferencedDocument

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

	// Extract trade tax breakdown
	for _, tax := range invoice.TradeTaxes {
		taxInfo := TaxInfo{
			CalculatedAmount: tax.CalculatedAmount.StringFixed(2),
			Percent:          tax.Percent.String(),
			Type:             tax.TypeCode,
			Category:         tax.CategoryCode,
			ExemptionCode:    tax.ExemptionReasonCode,
			ExemptionReason:  tax.ExemptionReason,
			BasisAmount:      tax.BasisAmount.StringFixed(2),
		}
		details.TradeTax = append(details.TradeTax, taxInfo)
	}

	// Extract charge/allowance information
	for _, ca := range invoice.SpecifiedTradeAllowanceCharge {
		caInfo := ChargeAllowanceInfo{
			ChargeIndicator: ca.ChargeIndicator,
			Amount:          ca.ActualAmount.StringFixed(2),
			Reason:          ca.Reason,
			Type:            ca.CategoryTradeTaxType,
			Category:        ca.CategoryTradeTaxCategoryCode,
			BasisAmount:     ca.BasisAmount.StringFixed(2),
			Percent:         ca.CalculationPercent.String(),
		}
		details.ChargeAllowances = append(details.ChargeAllowances, caInfo)
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
