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
	Seller                            *PartyInfo            `json:"seller"`
	Buyer                             *PartyInfo            `json:"buyer"`
	Payee                             *PartyInfo            `json:"payee,omitempty"`
	SellerTaxRepresentative           *PartyInfo            `json:"seller_tax_representative,omitempty"`
	ShipTo                            *PartyInfo            `json:"ship_to,omitempty"`
	Totals                            *TotalsInfo           `json:"totals"`
	Number                            string                `json:"number"`
	Date                              string                `json:"date"`
	BillingPeriodStart                string                `json:"billing_period_start,omitempty"`
	BillingPeriodEnd                  string                `json:"billing_period_end,omitempty"`
	OccurrenceDate                    string                `json:"occurrence_date,omitempty"`
	Type                              string                `json:"type"`
	Profile                           string                `json:"profile"`
	ProfileURN                        string                `json:"profile_urn,omitempty"`
	BusinessProcess                   string                `json:"business_process,omitempty"`
	Currency                          string                `json:"currency"`
	TaxCurrency                       string                `json:"tax_currency,omitempty"`
	TaxTotalCurrency                  string                `json:"tax_total_currency,omitempty"`
	TaxTotalAccountingCurrency        string                `json:"tax_total_accounting_currency,omitempty"`
	BuyerReference                    string                `json:"buyer_reference,omitempty"`
	BuyerOrderReferencedDocument      string                `json:"buyer_order_referenced_document,omitempty"`
	SellerOrderReferencedDocument     string                `json:"seller_order_referenced_document,omitempty"`
	ContractReferencedDocument        string                `json:"contract_referenced_document,omitempty"`
	DespatchAdviceReferencedDocument  string                `json:"despatch_advice_referenced_document,omitempty"`
	ReceivingAdviceReferencedDocument string                `json:"receiving_advice_referenced_document,omitempty"`
	ProcuringProjectID                string                `json:"procuring_project_id,omitempty"`
	ProcuringProjectName              string                `json:"procuring_project_name,omitempty"`
	ReceivableAccountingAccount       string                `json:"receivable_accounting_account,omitempty"`
	InvoiceReferences                 []ReferenceInfo       `json:"invoice_references,omitempty"`
	AdditionalReferences              []DocumentInfo        `json:"additional_references,omitempty"`
	PaymentMeans                      []PaymentMeansInfo    `json:"payment_means,omitempty"`
	Lines                             []LineInfo            `json:"lines,omitempty"`
	PaymentTerms                      []string              `json:"payment_terms,omitempty"`
	PaymentTermsDetailed              []PaymentTermInfo     `json:"payment_terms_detailed,omitempty"`
	TradeTax                          []TaxInfo             `json:"trade_tax,omitempty"`
	ChargeAllowances                  []ChargeAllowanceInfo `json:"charge_allowances,omitempty"`
	Notes                             []NoteInfo            `json:"notes,omitempty"`
	LineCount                         int                   `json:"line_count"`
	TermWidth                         int                   `json:"-"`
}

// PartyInfo contains party details
type PartyInfo struct {
	Address   *AddressInfo `json:"address,omitempty"`
	Name      string       `json:"name"`
	ID        string       `json:"id,omitempty"`
	GlobalID  string       `json:"global_id,omitempty"`
	VATNumber string       `json:"vat_number,omitempty"`
	TaxNumber string       `json:"tax_number,omitempty"`
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

// ReferenceInfo contains invoice reference details (BG-3).
type ReferenceInfo struct {
	ID   string `json:"id"`
	Date string `json:"date,omitempty"`
}

// DocumentInfo contains additional reference details (BG-24).
type DocumentInfo struct {
	ID                string `json:"id,omitempty"`
	URI               string `json:"uri,omitempty"`
	TypeCode          string `json:"type_code,omitempty"`
	ReferenceTypeCode string `json:"reference_type_code,omitempty"`
	Name              string `json:"name,omitempty"`
	Filename          string `json:"filename,omitempty"`
}

// PaymentMeansInfo contains payment means details.
type PaymentMeansInfo struct {
	Information        string `json:"information,omitempty"`
	PayeeIBAN          string `json:"payee_iban,omitempty"`
	PayeeAccountName   string `json:"payee_account_name,omitempty"`
	PayeeProprietaryID string `json:"payee_proprietary_id,omitempty"`
	PayeeBIC           string `json:"payee_bic,omitempty"`
	PayerIBAN          string `json:"payer_iban,omitempty"`
	CardID             string `json:"card_id,omitempty"`
	CardholderName     string `json:"cardholder_name,omitempty"`
	TypeCode           int    `json:"type_code,omitempty"`
}

// PaymentTermInfo contains detailed payment term data.
type PaymentTermInfo struct {
	Description          string `json:"description,omitempty"`
	DueDate              string `json:"due_date,omitempty"`
	DirectDebitMandateID string `json:"direct_debit_mandate_id,omitempty"`
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

// ChargeAllowanceInfo contains charge/allowance details
type ChargeAllowanceInfo struct {
	Amount          string `json:"amount"`
	Reason          string `json:"reason,omitempty"`
	Type            string `json:"type,omitempty"`
	Category        string `json:"category,omitempty"`
	Percent         string `json:"percent,omitempty"`
	BasisAmount     string `json:"basis_amount,omitempty"`
	ChargeIndicator bool   `json:"charge_indicator"`
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

// formatTextSubjectQualifier formats a text subject qualifier code based on display flags.
// - showCodes=true: returns only the code (e.g., "AAI")
// - verbose=true: returns code with description (e.g., "AAI (Additional information)")
// - default: returns description only, or "Unknown (code)" if not found
func formatTextSubjectQualifier(code string, showCodes bool, verbose bool) string {
	if showCodes {
		return code
	}
	if code == "" {
		return ""
	}
	description := codelists.TextSubjectQualifier(code)
	if verbose {
		return code + " (" + description + ")"
	}
	return description
}

func partyInfoFromParty(p *einvoice.Party, showCodes bool, verbose bool) *PartyInfo {
	if p == nil {
		return nil
	}

	info := &PartyInfo{
		Name:      p.Name,
		VATNumber: p.VATaxRegistration,
		TaxNumber: p.FCTaxRegistration,
		ID:        strings.Join(p.ID, ", "),
	}

	if len(p.GlobalID) > 0 {
		var gids []string
		for i := range p.GlobalID {
			if p.GlobalID[i].ID == "" {
				continue
			}
			if p.GlobalID[i].Scheme != "" {
				gids = append(gids, fmt.Sprintf("%s:%s", p.GlobalID[i].Scheme, p.GlobalID[i].ID))
			} else {
				gids = append(gids, p.GlobalID[i].ID)
			}
		}
		info.GlobalID = strings.Join(gids, ", ")
	}

	if p.PostalAddress != nil {
		info.Address = &AddressInfo{
			Line1:      p.PostalAddress.Line1,
			Line2:      p.PostalAddress.Line2,
			Line3:      p.PostalAddress.Line3,
			City:       p.PostalAddress.City,
			PostalCode: p.PostalAddress.PostcodeCode,
			Country:    p.PostalAddress.CountryID,
		}
	}

	return info
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
		Number:                     invoice.InvoiceNumber,
		Type:                       formatDocumentType(typeCode, showCodes, verbose),
		Profile:                    einvoice.GetProfileName(invoice.GuidelineSpecifiedDocumentContextParameter),
		ProfileURN:                 invoice.GuidelineSpecifiedDocumentContextParameter,
		BusinessProcess:            invoice.BPSpecifiedDocumentContextParameter,
		Currency:                   invoice.InvoiceCurrencyCode,
		TaxCurrency:                invoice.TaxCurrencyCode,
		TaxTotalCurrency:           invoice.TaxTotalCurrency,
		TaxTotalAccountingCurrency: invoice.TaxTotalAccountingCurrency,
		BuyerReference:             invoice.BuyerReference,
		LineCount:                  len(invoice.InvoiceLines),
	}

	// Format date
	if !invoice.InvoiceDate.IsZero() {
		details.Date = invoice.InvoiceDate.Format("2006-01-02")
	}

	if !invoice.BillingSpecifiedPeriodStart.IsZero() {
		details.BillingPeriodStart = invoice.BillingSpecifiedPeriodStart.Format("2006-01-02")
	}
	if !invoice.BillingSpecifiedPeriodEnd.IsZero() {
		details.BillingPeriodEnd = invoice.BillingSpecifiedPeriodEnd.Format("2006-01-02")
	}
	if !invoice.OccurrenceDateTime.IsZero() {
		details.OccurrenceDate = invoice.OccurrenceDateTime.Format("2006-01-02")
	}

	details.DespatchAdviceReferencedDocument = invoice.DespatchAdviceReferencedDocument
	details.ReceivingAdviceReferencedDocument = invoice.ReceivingAdviceReferencedDocument
	details.BuyerOrderReferencedDocument = invoice.BuyerOrderReferencedDocument
	details.SellerOrderReferencedDocument = invoice.SellerOrderReferencedDocument
	details.ContractReferencedDocument = invoice.ContractReferencedDocument
	details.BuyerReference = invoice.BuyerReference
	details.ProcuringProjectID = invoice.SpecifiedProcuringProjectID
	details.ProcuringProjectName = invoice.SpecifiedProcuringProjectName
	details.ReceivableAccountingAccount = invoice.ReceivableSpecifiedTradeAccountingAccount

	// Extract seller information
	details.Seller = partyInfoFromParty(&invoice.Seller, showCodes, verbose)

	// Extract buyer information
	details.Buyer = partyInfoFromParty(&invoice.Buyer, showCodes, verbose)
	details.Payee = partyInfoFromParty(invoice.PayeeTradeParty, showCodes, verbose)
	details.SellerTaxRepresentative = partyInfoFromParty(invoice.SellerTaxRepresentativeTradeParty, showCodes, verbose)
	details.ShipTo = partyInfoFromParty(invoice.ShipTo, showCodes, verbose)

	// Extract invoice lines
	details.Lines = make([]LineInfo, 0, len(invoice.InvoiceLines))
	for i := range invoice.InvoiceLines {
		lineInfo := LineInfo{
			ID:          invoice.InvoiceLines[i].LineID,
			NetAmount:   invoice.InvoiceLines[i].Total.String(),
			Description: invoice.InvoiceLines[i].Description,
			Name:        invoice.InvoiceLines[i].ItemName,
		}

		if !invoice.InvoiceLines[i].BilledQuantity.IsZero() {
			lineInfo.Quantity = invoice.InvoiceLines[i].BilledQuantity.String()
			if invoice.InvoiceLines[i].BilledQuantityUnit != "" {
				unitName := formatUnitCode(invoice.InvoiceLines[i].BilledQuantityUnit, showCodes, verbose)
				lineInfo.Quantity += " " + unitName
			}
		}

		lineInfo.NetPrice = invoice.InvoiceLines[i].NetPrice.String()
		lineInfo.GrossPrice = invoice.InvoiceLines[i].GrossPrice.String()
		lineInfo.Note = invoice.InvoiceLines[i].Note
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
	details.PaymentTermsDetailed = make([]PaymentTermInfo, 0, len(invoice.SpecifiedTradePaymentTerms))
	for i := range invoice.SpecifiedTradePaymentTerms {
		if invoice.SpecifiedTradePaymentTerms[i].Description != "" {
			details.PaymentTerms = append(details.PaymentTerms, invoice.SpecifiedTradePaymentTerms[i].Description)
		}
		pt := PaymentTermInfo{
			Description: invoice.SpecifiedTradePaymentTerms[i].Description,
		}
		if !invoice.SpecifiedTradePaymentTerms[i].DueDate.IsZero() {
			pt.DueDate = invoice.SpecifiedTradePaymentTerms[i].DueDate.Format("2006-01-02")
		}
		if invoice.SpecifiedTradePaymentTerms[i].DirectDebitMandateID != "" {
			pt.DirectDebitMandateID = invoice.SpecifiedTradePaymentTerms[i].DirectDebitMandateID
		}
		if pt.Description != "" || pt.DueDate != "" || pt.DirectDebitMandateID != "" {
			details.PaymentTermsDetailed = append(details.PaymentTermsDetailed, pt)
		}
	}

	// Extract notes
	for i := range invoice.Notes {
		// Skip empty notes
		if invoice.Notes[i].Text == "" {
			continue
		}

		details.Notes = append(details.Notes, NoteInfo{
			Text:             invoice.Notes[i].Text,
			SubjectQualifier: formatTextSubjectQualifier(invoice.Notes[i].SubjectCode, showCodes, verbose),
		})
	}

	// Extract trade tax breakdown
	for i := range invoice.TradeTaxes {
		taxInfo := TaxInfo{
			CalculatedAmount: invoice.TradeTaxes[i].CalculatedAmount.StringFixed(2),
			Percent:          invoice.TradeTaxes[i].Percent.String(),
			Type:             invoice.TradeTaxes[i].TypeCode,
			Category:         invoice.TradeTaxes[i].CategoryCode,
			ExemptionCode:    invoice.TradeTaxes[i].ExemptionReasonCode,
			ExemptionReason:  invoice.TradeTaxes[i].ExemptionReason,
			BasisAmount:      invoice.TradeTaxes[i].BasisAmount.StringFixed(2),
		}
		details.TradeTax = append(details.TradeTax, taxInfo)
	}

	// Extract charge/allowance information
	for i := range invoice.SpecifiedTradeAllowanceCharge {
		caInfo := ChargeAllowanceInfo{
			ChargeIndicator: invoice.SpecifiedTradeAllowanceCharge[i].ChargeIndicator,
			Amount:          invoice.SpecifiedTradeAllowanceCharge[i].ActualAmount.StringFixed(2),
			Reason:          invoice.SpecifiedTradeAllowanceCharge[i].Reason,
			Type:            invoice.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxType,
			Category:        invoice.SpecifiedTradeAllowanceCharge[i].CategoryTradeTaxCategoryCode,
			BasisAmount:     invoice.SpecifiedTradeAllowanceCharge[i].BasisAmount.StringFixed(2),
			Percent:         invoice.SpecifiedTradeAllowanceCharge[i].CalculationPercent.String(),
		}
		details.ChargeAllowances = append(details.ChargeAllowances, caInfo)
	}

	// Extract payment means
	for i := range invoice.PaymentMeans {
		pmInfo := PaymentMeansInfo{
			TypeCode:           invoice.PaymentMeans[i].TypeCode,
			Information:        invoice.PaymentMeans[i].Information,
			PayeeIBAN:          invoice.PaymentMeans[i].PayeePartyCreditorFinancialAccountIBAN,
			PayeeAccountName:   invoice.PaymentMeans[i].PayeePartyCreditorFinancialAccountName,
			PayeeProprietaryID: invoice.PaymentMeans[i].PayeePartyCreditorFinancialAccountProprietaryID,
			PayeeBIC:           invoice.PaymentMeans[i].PayeeSpecifiedCreditorFinancialInstitutionBIC,
			PayerIBAN:          invoice.PaymentMeans[i].PayerPartyDebtorFinancialAccountIBAN,
			CardID:             invoice.PaymentMeans[i].ApplicableTradeSettlementFinancialCardID,
			CardholderName:     invoice.PaymentMeans[i].ApplicableTradeSettlementFinancialCardCardholderName,
		}
		details.PaymentMeans = append(details.PaymentMeans, pmInfo)
	}

	// Extract invoice references (BG-3)
	for i := range invoice.InvoiceReferencedDocument {
		refInfo := ReferenceInfo{
			ID: invoice.InvoiceReferencedDocument[i].ID,
		}
		if !invoice.InvoiceReferencedDocument[i].Date.IsZero() {
			refInfo.Date = invoice.InvoiceReferencedDocument[i].Date.Format("2006-01-02")
		}
		if refInfo.ID != "" || refInfo.Date != "" {
			details.InvoiceReferences = append(details.InvoiceReferences, refInfo)
		}
	}

	// Extract additional referenced documents (BG-24)
	for i := range invoice.AdditionalReferencedDocument {
		docInfo := DocumentInfo{
			ID:                invoice.AdditionalReferencedDocument[i].IssuerAssignedID,
			URI:               invoice.AdditionalReferencedDocument[i].URIID,
			TypeCode:          invoice.AdditionalReferencedDocument[i].TypeCode,
			ReferenceTypeCode: invoice.AdditionalReferencedDocument[i].ReferenceTypeCode,
			Name:              invoice.AdditionalReferencedDocument[i].Name,
			Filename:          invoice.AdditionalReferencedDocument[i].AttachmentFilename,
		}
		if docInfo.ID != "" || docInfo.URI != "" || docInfo.Name != "" || docInfo.Filename != "" {
			details.AdditionalReferences = append(details.AdditionalReferences, docInfo)
		}
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
