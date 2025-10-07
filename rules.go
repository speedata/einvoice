package einvoice

// Rule defines a business rule with its code, affected fields, and specification.
// Rules are used to create validation violations in a type-safe, maintainable way.
type Rule struct {
	Code        string   // Rule identifier (e.g., "BR-CO-14", "BR-S-1")
	Fields      []string // EN 16931 field identifiers (e.g., "BT-110", "BG-23")
	Description string   // Rule specification or requirement description
}

// Business rule definitions for EN 16931 electronic invoice validation.
// These rules are grouped by category and organized according to the specification.
//
// Rule naming convention:
// - BR-1 to BR-65: Core business rules
// - BR-CO-*: Calculation and totals rules
// - BR-S-*: Standard rated VAT rules
// - BR-AE-*: Reverse charge VAT rules
// - BR-E-*: Exempt from VAT rules
// - BR-Z-*: Zero rated VAT rules
// - BR-G-*: Export outside EU rules
// - BR-IC-*: Intra-community supply rules
// - BR-IG-*: IGIC (Canary Islands) rules
// - BR-IP-*: IPSI (Ceuta/Melilla) rules
// - BR-O-*: Not subject to VAT rules
var (
	// Core business rules (BR-1 to BR-65)

	BR1 = Rule{
		Code:   "BR-1",
		Fields: []string{"BT-24"},
		Description: "Invoice must contain specification identifier (BT-24)",
	}

	BR2 = Rule{
		Code:   "BR-2",
		Fields: []string{"BT-1"},
		Description: "Invoice must contain invoice number (BT-1)",
	}

	BR3 = Rule{
		Code:   "BR-3",
		Fields: []string{"BT-2"},
		Description: "Invoice must contain invoice issue date (BT-2)",
	}

	BR4 = Rule{
		Code:   "BR-4",
		Fields: []string{"BT-3"},
		Description: "Invoice must contain invoice type code (BT-3)",
	}

	BR5 = Rule{
		Code:   "BR-5",
		Fields: []string{"BT-5"},
		Description: "Invoice must contain invoice currency code (BT-5)",
	}

	BR6 = Rule{
		Code:   "BR-6",
		Fields: []string{"BT-27"},
		Description: "Invoice must contain seller name (BT-27)",
	}

	BR7 = Rule{
		Code:   "BR-7",
		Fields: []string{"BT-44"},
		Description: "Invoice must contain buyer name (BT-44)",
	}

	BR8 = Rule{
		Code:   "BR-8",
		Fields: []string{"BG-5"},
		Description: "Invoice must contain seller postal address (BG-5)",
	}

	BR9 = Rule{
		Code:   "BR-9",
		Fields: []string{"BT-40"},
		Description: "Seller postal address must contain country code (BT-40)",
	}

	BR10 = Rule{
		Code:   "BR-10",
		Fields: []string{"BG-8"},
		Description: "Invoice must contain buyer postal address (BG-8)",
	}

	BR11 = Rule{
		Code:   "BR-11",
		Fields: []string{"BT-55"},
		Description: "Buyer postal address must contain country code (BT-55)",
	}

	BR12 = Rule{
		Code:   "BR-12",
		Fields: []string{"BT-106"},
		Description: "Invoice total amount without VAT must be provided (BT-106)",
	}

	BR13 = Rule{
		Code:   "BR-13",
		Fields: []string{"BT-109"},
		Description: "Invoice total VAT amount must be provided (BT-109)",
	}

	BR14 = Rule{
		Code:   "BR-14",
		Fields: []string{"BT-112"},
		Description: "Invoice total amount with VAT must be provided (BT-112)",
	}

	BR15 = Rule{
		Code:   "BR-15",
		Fields: []string{"BT-115"},
		Description: "Amount due for payment must be provided (BT-115)",
	}

	BR16 = Rule{
		Code:   "BR-16",
		Fields: []string{"BG-25"},
		Description: "Invoice must have at least one invoice line (BG-25)",
	}

	BR17 = Rule{
		Code:   "BR-17",
		Fields: []string{"BT-59", "BG-10", "BG-4"},
		Description: "Payee name must be provided if payee differs from seller (BT-59)",
	}

	BR18 = Rule{
		Code:   "BR-18",
		Fields: []string{"BT-62", "BG-4", "BG-11"},
		Description: "Seller tax representative name must be provided (BT-62)",
	}

	BR19 = Rule{
		Code:   "BR-19",
		Fields: []string{"BG-4", "BG-12"},
		Description: "Seller tax representative postal address must be provided (BG-12)",
	}

	BR20 = Rule{
		Code:   "BR-20",
		Fields: []string{"BG-4", "BG-12"},
		Description: "Seller tax representative postal address must contain country code",
	}

	BR21 = Rule{
		Code:   "BR-21",
		Fields: []string{"BG-25", "BT-126"},
		Description: "Each invoice line must have invoice line identifier (BT-126)",
	}

	BR22 = Rule{
		Code:   "BR-22",
		Fields: []string{"BG-25", "BT-129"},
		Description: "Each invoice line must have invoiced quantity (BT-129)",
	}

	BR23 = Rule{
		Code:   "BR-23",
		Fields: []string{"BG-25", "BT-130"},
		Description: "Invoiced quantity must have unit of measure (BT-130)",
	}

	BR24 = Rule{
		Code:   "BR-24",
		Fields: []string{"BG-25", "BT-131"},
		Description: "Each invoice line must have invoice line net amount (BT-131)",
	}

	BR25 = Rule{
		Code:   "BR-25",
		Fields: []string{"BG-25", "BT-153"},
		Description: "Each invoice line must have item name (BT-153)",
	}

	BR26 = Rule{
		Code:   "BR-26",
		Fields: []string{"BG-25", "BT-146"},
		Description: "Each invoice line must have item net price (BT-146)",
	}

	BR27 = Rule{
		Code:   "BR-27",
		Fields: []string{"BG-25", "BT-146"},
		Description: "Item net price must not be negative (BT-146)",
	}

	BR28 = Rule{
		Code:   "BR-28",
		Fields: []string{"BG-25", "BT-148"},
		Description: "Item gross price must not be negative (BT-148)",
	}

	BR29 = Rule{
		Code:   "BR-29",
		Fields: []string{"BT-73", "BT-74"},
		Description: "Invoicing period end date must be later than or equal to start date",
	}

	BR30 = Rule{
		Code:   "BR-30",
		Fields: []string{"BG-25", "BT-135", "BT-134"},
		Description: "Line period end date must be later than or equal to start date",
	}

	BR31 = Rule{
		Code:   "BR-31",
		Fields: []string{"BG-20", "BT-92"},
		Description: "Document level allowance amount must not be zero (BT-92)",
	}

	BR32 = Rule{
		Code:   "BR-32",
		Fields: []string{"BG-20", "BT-95"},
		Description: "Document level allowance must have tax category code (BT-95)",
	}

	BR33 = Rule{
		Code:   "BR-33",
		Fields: []string{"BG-20", "BT-97", "BT-98"},
		Description: "Document level allowance must have reason or reason code (BT-97, BT-98)",
	}

	BR34 = Rule{
		Code:   "BR-34",
		Fields: []string{"BG-20", "BT-92"},
		Description: "Document level allowance amount must not be negative (BT-92)",
	}

	BR35 = Rule{
		Code:   "BR-35",
		Fields: []string{"BG-20", "BT-93"},
		Description: "Document level allowance base amount must not be negative (BT-93)",
	}

	BR36 = Rule{
		Code:   "BR-36",
		Fields: []string{"BG-21", "BT-99"},
		Description: "Document level charge amount must not be zero (BT-99)",
	}

	BR37 = Rule{
		Code:   "BR-37",
		Fields: []string{"BG-21", "BT-102"},
		Description: "Document level charge must have tax category code (BT-102)",
	}

	BR38 = Rule{
		Code:   "BR-38",
		Fields: []string{"BG-21", "BT-104", "BT-105"},
		Description: "Document level charge must have reason or reason code (BT-104, BT-105)",
	}

	BR39 = Rule{
		Code:   "BR-39",
		Fields: []string{"BG-21", "BT-99"},
		Description: "Document level charge amount must not be negative (BT-99)",
	}

	BR40 = Rule{
		Code:   "BR-40",
		Fields: []string{"BG-21", "BT-100"},
		Description: "Document level charge base amount must not be negative (BT-100)",
	}

	BR41 = Rule{
		Code:   "BR-41",
		Fields: []string{"BG-27", "BT-136"},
		Description: "Invoice line allowance amount must not be zero (BT-136)",
	}

	BR42 = Rule{
		Code:   "BR-42",
		Fields: []string{"BG-27", "BT-139", "BT-140"},
		Description: "Invoice line allowance must have reason or reason code (BT-139, BT-140)",
	}

	BR43 = Rule{
		Code:   "BR-43",
		Fields: []string{"BG-28", "BT-141"},
		Description: "Invoice line charge amount must not be zero (BT-141)",
	}

	BR44 = Rule{
		Code:   "BR-44",
		Fields: []string{"BG-28", "BT-144", "BT-145"},
		Description: "Invoice line charge must have reason or reason code (BT-144, BT-145)",
	}

	BR45 = Rule{
		Code:   "BR-45",
		Fields: []string{"BG-23", "BT-116"},
		Description: "VAT category tax amount must equal sum of line net amounts minus allowances plus charges (BT-116)",
	}

	BR47 = Rule{
		Code:   "BR-47",
		Fields: []string{"BG-23", "BT-118"},
		Description: "VAT breakdown must have VAT category code (BT-118)",
	}

	BR49 = Rule{
		Code:   "BR-49",
		Fields: []string{"BT-81"},
		Description: "Payment means type code must be provided (BT-81)",
	}

	BR52 = Rule{
		Code:   "BR-52",
		Fields: []string{"BG-24", "BT-122"},
		Description: "Supporting document must have reference (BT-122)",
	}

	BR53 = Rule{
		Code:   "BR-53",
		Fields: []string{"BT-6", "BT-111"},
		Description: "If tax currency code differs from invoice currency, VAT total in accounting currency must be provided (BT-6, BT-111)",
	}

	BR54 = Rule{
		Code:   "BR-54",
		Fields: []string{"BG-32", "BT-160", "BT-161"},
		Description: "Item attribute must have both name and value (BT-160, BT-161)",
	}

	BR55 = Rule{
		Code:   "BR-55",
		Fields: []string{"BG-3", "BT-25"},
		Description: "Preceding invoice reference must contain invoice number (BT-25)",
	}

	BR56 = Rule{
		Code:   "BR-56",
		Fields: []string{"BG-11", "BT-63"},
		Description: "Seller tax representative must have VAT identifier (BT-63)",
	}

	BR57 = Rule{
		Code:   "BR-57",
		Fields: []string{"BG-15", "BT-80"},
		Description: "Deliver to address must have country code (BT-80)",
	}

	BR61 = Rule{
		Code:   "BR-61",
		Fields: []string{"BT-31", "BT-32"},
		Description: "Seller VAT identifier or tax registration identifier must be provided",
	}

	BR62 = Rule{
		Code:   "BR-62",
		Fields: []string{"BT-48", "BT-46"},
		Description: "Buyer VAT identifier or tax registration identifier must be provided",
	}

	BR63 = Rule{
		Code:   "BR-63",
		Fields: []string{"BT-31"},
		Description: "Seller VAT identifier must be provided",
	}

	BR64 = Rule{
		Code:   "BR-64",
		Fields: []string{"BT-151"},
		Description: "Invoice line VAT category code must match document level VAT breakdown",
	}

	BR65 = Rule{
		Code:   "BR-65",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge VAT category code must match document level VAT breakdown",
	}

	// Calculation rules (BR-CO-*)

	BRCO3 = Rule{
		Code:   "BR-CO-3",
		Fields: []string{"BT-7", "BT-8"},
		Description: "Value added tax point date (BT-7) and value added tax point date code (BT-8) are mutually exclusive",
	}

	BRCO4 = Rule{
		Code:   "BR-CO-4",
		Fields: []string{"BG-25", "BT-151"},
		Description: "Each invoice line must have invoiced item VAT category code (BT-151)",
	}

	BRCO10 = Rule{
		Code:   "BR-CO-10",
		Fields: []string{"BT-106", "BT-131"},
		Description: "Sum of invoice line net amount (BT-131) = Invoice line net amount (BT-106)",
	}

	BRCO11 = Rule{
		Code:   "BR-CO-11",
		Fields: []string{"BT-107", "BT-92"},
		Description: "Sum of allowances on document level (BT-92) = Sum of document level allowance amounts (BT-107)",
	}

	BRCO12 = Rule{
		Code:   "BR-CO-12",
		Fields: []string{"BT-108", "BT-99"},
		Description: "Sum of charges on document level (BT-99) = Sum of document level charge amounts (BT-108)",
	}

	BRCO13 = Rule{
		Code:   "BR-CO-13",
		Fields: []string{"BT-109", "BT-106", "BT-107", "BT-108"},
		Description: "Invoice total amount without VAT (BT-109) = Σ Invoice line net amount (BT-106) - Sum of allowances on document level (BT-107) + Sum of charges on document level (BT-108)",
	}

	BRCO14 = Rule{
		Code:   "BR-CO-14",
		Fields: []string{"BT-110", "BT-117"},
		Description: "Invoice total VAT amount (BT-110) = Σ VAT category tax amount (BT-117)",
	}

	BRCO15 = Rule{
		Code:   "BR-CO-15",
		Fields: []string{"BT-112", "BT-109", "BT-110"},
		Description: "Invoice total amount with VAT (BT-112) = Invoice total amount without VAT (BT-109) + Invoice total VAT amount (BT-110)",
	}

	BRCO16 = Rule{
		Code:   "BR-CO-16",
		Fields: []string{"BT-115", "BT-112", "BT-113", "BT-114"},
		Description: "Amount due for payment (BT-115) = Invoice total amount with VAT (BT-112) - Paid amount (BT-113) + Rounding amount (BT-114)",
	}

	BRCO17 = Rule{
		Code:   "BR-CO-17",
		Fields: []string{"BT-116", "BT-117", "BT-119"},
		Description: "VAT category tax amount (BT-117) = VAT category taxable amount (BT-116) × (VAT category rate (BT-119) / 100), rounded to two decimals",
	}

	BRCO18 = Rule{
		Code:   "BR-CO-18",
		Fields: []string{"BG-23"},
		Description: "Invoice must have at least one VAT breakdown group (BG-23)",
	}

	BRCO19 = Rule{
		Code:   "BR-CO-19",
		Fields: []string{"BG-14", "BT-73", "BT-74"},
		Description: "If invoicing period (BG-14) is used, invoicing period start date (BT-73) or invoicing period end date (BT-74) must be provided",
	}

	BRCO20 = Rule{
		Code:   "BR-CO-20",
		Fields: []string{"BG-26", "BT-134", "BT-135"},
		Description: "If invoice line period (BG-26) is used, invoice line period start date (BT-134) or invoice line period end date (BT-135) must be provided",
	}

	BRCO25 = Rule{
		Code:   "BR-CO-25",
		Fields: []string{"BT-9", "BT-20", "BT-115"},
		Description: "If amount due for payment (BT-115) is positive, either payment due date (BT-9) or payment terms (BT-20) must be provided",
	}

	// Standard rated VAT rules (BR-S-*)

	BRS1 = Rule{
		Code:   "BR-S-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with standard rated VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'S'",
	}

	BRS2 = Rule{
		Code:   "BR-S-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with standard rated VAT must have invoiced item VAT category code (BT-151) = 'S'",
	}

	BRS3 = Rule{
		Code:   "BR-S-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with standard rated VAT must have VAT category code (BT-95, BT-102) = 'S'",
	}

	BRS4 = Rule{
		Code:   "BR-S-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with standard rated VAT must contain seller VAT identifier (BT-31)",
	}

	BRS5 = Rule{
		Code:   "BR-S-5",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for standard rated VAT",
	}

	BRS6 = Rule{
		Code:   "BR-S-6",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be provided for standard rated VAT",
	}

	BRS7 = Rule{
		Code:   "BR-S-7",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'S' for standard rated VAT",
	}

	BRS8 = Rule{
		Code:   "BR-S-8",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must be provided for standard rated VAT and must not be zero",
	}

	BRS9 = Rule{
		Code:   "BR-S-9",
		Fields: []string{"BT-120"},
		Description: "VAT exemption reason code (BT-120) must not be provided for standard rated VAT",
	}

	BRS10 = Rule{
		Code:   "BR-S-10",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must not be provided for standard rated VAT",
	}

	// Reverse charge VAT rules (BR-AE-*)

	BRAE1 = Rule{
		Code:   "BR-AE-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with reverse charge VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'AE'",
	}

	BRAE2 = Rule{
		Code:   "BR-AE-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with reverse charge VAT must have invoiced item VAT category code (BT-151) = 'AE'",
	}

	BRAE3 = Rule{
		Code:   "BR-AE-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with reverse charge VAT must have VAT category code (BT-95, BT-102) = 'AE'",
	}

	BRAE4 = Rule{
		Code:   "BR-AE-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with reverse charge VAT must contain seller VAT identifier (BT-31) or seller tax registration identifier (BT-32)",
	}

	BRAE5 = Rule{
		Code:   "BR-AE-5",
		Fields: []string{"BT-48"},
		Description: "Invoice with reverse charge VAT must contain buyer VAT identifier (BT-48)",
	}

	BRAE6 = Rule{
		Code:   "BR-AE-6",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for reverse charge VAT",
	}

	BRAE7 = Rule{
		Code:   "BR-AE-7",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be zero for reverse charge VAT",
	}

	BRAE8 = Rule{
		Code:   "BR-AE-8",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'AE' for reverse charge VAT",
	}

	BRAE9 = Rule{
		Code:   "BR-AE-9",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must not be provided for reverse charge VAT",
	}

	BRAE10 = Rule{
		Code:   "BR-AE-10",
		Fields: []string{"BT-120", "BT-121"},
		Description: "VAT exemption reason code (BT-120) or VAT exemption reason text (BT-121) must be provided for reverse charge VAT",
	}

	// Exempt from VAT rules (BR-E-*)

	BRE1 = Rule{
		Code:   "BR-E-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with exempt from VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'E'",
	}

	BRE2 = Rule{
		Code:   "BR-E-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with exempt from VAT must have invoiced item VAT category code (BT-151) = 'E'",
	}

	BRE3 = Rule{
		Code:   "BR-E-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with exempt from VAT must have VAT category code (BT-95, BT-102) = 'E'",
	}

	BRE4 = Rule{
		Code:   "BR-E-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with exempt from VAT must contain seller VAT identifier (BT-31) or seller tax registration identifier (BT-32)",
	}

	BRE5 = Rule{
		Code:   "BR-E-5",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for exempt from VAT",
	}

	BRE6 = Rule{
		Code:   "BR-E-6",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be zero for exempt from VAT",
	}

	BRE7 = Rule{
		Code:   "BR-E-7",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'E' for exempt from VAT",
	}

	BRE8 = Rule{
		Code:   "BR-E-8",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must not be provided for exempt from VAT",
	}

	BRE9 = Rule{
		Code:   "BR-E-9",
		Fields: []string{"BT-120", "BT-121"},
		Description: "VAT exemption reason code (BT-120) or VAT exemption reason text (BT-121) must be provided for exempt from VAT",
	}

	BRE10 = Rule{
		Code:   "BR-E-10",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must be provided if VAT exemption reason code (BT-120) is not provided",
	}

	// Zero rated VAT rules (BR-Z-*)

	BRZ1 = Rule{
		Code:   "BR-Z-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with zero rated VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'Z'",
	}

	BRZ2 = Rule{
		Code:   "BR-Z-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with zero rated VAT must have invoiced item VAT category code (BT-151) = 'Z'",
	}

	BRZ3 = Rule{
		Code:   "BR-Z-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with zero rated VAT must have VAT category code (BT-95, BT-102) = 'Z'",
	}

	BRZ4 = Rule{
		Code:   "BR-Z-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with zero rated VAT must contain seller VAT identifier (BT-31) or seller tax registration identifier (BT-32)",
	}

	BRZ5 = Rule{
		Code:   "BR-Z-5",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for zero rated VAT",
	}

	BRZ6 = Rule{
		Code:   "BR-Z-6",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be zero for zero rated VAT",
	}

	BRZ7 = Rule{
		Code:   "BR-Z-7",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'Z' for zero rated VAT",
	}

	BRZ8 = Rule{
		Code:   "BR-Z-8",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must be zero for zero rated VAT",
	}

	BRZ9 = Rule{
		Code:   "BR-Z-9",
		Fields: []string{"BT-120", "BT-121"},
		Description: "VAT exemption reason code (BT-120) or VAT exemption reason text (BT-121) must be provided for zero rated VAT",
	}

	BRZ10 = Rule{
		Code:   "BR-Z-10",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must be provided if VAT exemption reason code (BT-120) is not provided",
	}

	// Export outside EU rules (BR-G-*)

	BRG1 = Rule{
		Code:   "BR-G-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with export outside EU VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'G'",
	}

	BRG2 = Rule{
		Code:   "BR-G-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with export outside EU VAT must have invoiced item VAT category code (BT-151) = 'G'",
	}

	BRG3 = Rule{
		Code:   "BR-G-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with export outside EU VAT must have VAT category code (BT-95, BT-102) = 'G'",
	}

	BRG4 = Rule{
		Code:   "BR-G-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with export outside EU VAT must contain seller VAT identifier (BT-31) or seller tax registration identifier (BT-32)",
	}

	BRG5 = Rule{
		Code:   "BR-G-5",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for export outside EU VAT",
	}

	BRG6 = Rule{
		Code:   "BR-G-6",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be zero for export outside EU VAT",
	}

	BRG7 = Rule{
		Code:   "BR-G-7",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'G' for export outside EU VAT",
	}

	BRG8 = Rule{
		Code:   "BR-G-8",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must not be provided for export outside EU VAT",
	}

	BRG9 = Rule{
		Code:   "BR-G-9",
		Fields: []string{"BT-120", "BT-121"},
		Description: "VAT exemption reason code (BT-120) or VAT exemption reason text (BT-121) must be provided for export outside EU VAT",
	}

	BRG10 = Rule{
		Code:   "BR-G-10",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must be provided if VAT exemption reason code (BT-120) is not provided",
	}

	// Intra-community supply rules (BR-IC-*)

	BRIC1 = Rule{
		Code:   "BR-IC-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with intra-community supply VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'K'",
	}

	BRIC2 = Rule{
		Code:   "BR-IC-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with intra-community supply VAT must have invoiced item VAT category code (BT-151) = 'K'",
	}

	BRIC3 = Rule{
		Code:   "BR-IC-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with intra-community supply VAT must have VAT category code (BT-95, BT-102) = 'K'",
	}

	BRIC4 = Rule{
		Code:   "BR-IC-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with intra-community supply VAT must contain seller VAT identifier (BT-31)",
	}

	BRIC5 = Rule{
		Code:   "BR-IC-5",
		Fields: []string{"BT-48"},
		Description: "Invoice with intra-community supply VAT must contain buyer VAT identifier (BT-48)",
	}

	BRIC6 = Rule{
		Code:   "BR-IC-6",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for intra-community supply VAT",
	}

	BRIC7 = Rule{
		Code:   "BR-IC-7",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be zero for intra-community supply VAT",
	}

	BRIC8 = Rule{
		Code:   "BR-IC-8",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'K' for intra-community supply VAT",
	}

	BRIC9 = Rule{
		Code:   "BR-IC-9",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must not be provided for intra-community supply VAT",
	}

	BRIC10 = Rule{
		Code:   "BR-IC-10",
		Fields: []string{"BT-120", "BT-121"},
		Description: "VAT exemption reason code (BT-120) or VAT exemption reason text (BT-121) must be provided for intra-community supply VAT",
	}

	BRIC11 = Rule{
		Code:   "BR-IC-11",
		Fields: []string{"BT-40", "BT-55"},
		Description: "Seller country code (BT-40) and buyer country code (BT-55) must differ for intra-community supply",
	}

	BRIC12 = Rule{
		Code:   "BR-IC-12",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must be provided if VAT exemption reason code (BT-120) is not provided",
	}

	// IGIC (Canary Islands) rules (BR-IG-*)

	BRIG1 = Rule{
		Code:   "BR-IG-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with IGIC VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'L'",
	}

	BRIG5 = Rule{
		Code:   "BR-IG-5",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for IGIC VAT",
	}

	BRIG6 = Rule{
		Code:   "BR-IG-6",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be provided for IGIC VAT",
	}

	BRIG7 = Rule{
		Code:   "BR-IG-7",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'L' for IGIC VAT",
	}

	BRIG8 = Rule{
		Code:   "BR-IG-8",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must be provided for IGIC VAT",
	}

	BRIG9 = Rule{
		Code:   "BR-IG-9",
		Fields: []string{"BT-120"},
		Description: "VAT exemption reason code (BT-120) must not be provided for IGIC VAT",
	}

	BRIG10 = Rule{
		Code:   "BR-IG-10",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must not be provided for IGIC VAT",
	}

	// IPSI (Ceuta/Melilla) rules (BR-IP-*)

	BRIP1 = Rule{
		Code:   "BR-IP-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with IPSI VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'M'",
	}

	BRIP5 = Rule{
		Code:   "BR-IP-5",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for IPSI VAT",
	}

	BRIP6 = Rule{
		Code:   "BR-IP-6",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be provided for IPSI VAT",
	}

	BRIP7 = Rule{
		Code:   "BR-IP-7",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'M' for IPSI VAT",
	}

	BRIP8 = Rule{
		Code:   "BR-IP-8",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must be provided for IPSI VAT",
	}

	BRIP9 = Rule{
		Code:   "BR-IP-9",
		Fields: []string{"BT-120"},
		Description: "VAT exemption reason code (BT-120) must not be provided for IPSI VAT",
	}

	BRIP10 = Rule{
		Code:   "BR-IP-10",
		Fields: []string{"BT-121"},
		Description: "VAT exemption reason text (BT-121) must not be provided for IPSI VAT",
	}

	// Not subject to VAT rules (BR-O-*)

	BRO1 = Rule{
		Code:   "BR-O-1",
		Fields: []string{"BG-23", "BT-118"},
		Description: "Invoice with not subject to VAT must have VAT breakdown (BG-23) with VAT category code (BT-118) = 'O'",
	}

	BRO2 = Rule{
		Code:   "BR-O-2",
		Fields: []string{"BT-151"},
		Description: "Invoice line with not subject to VAT must have invoiced item VAT category code (BT-151) = 'O'",
	}

	BRO3 = Rule{
		Code:   "BR-O-3",
		Fields: []string{"BT-95", "BT-102"},
		Description: "Document level allowance/charge with not subject to VAT must have VAT category code (BT-95, BT-102) = 'O'",
	}

	BRO4 = Rule{
		Code:   "BR-O-4",
		Fields: []string{"BT-31"},
		Description: "Invoice with not subject to VAT must not contain seller VAT identifier (BT-31)",
	}

	BRO5 = Rule{
		Code:   "BR-O-5",
		Fields: []string{"BT-48"},
		Description: "Invoice with not subject to VAT must not contain buyer VAT identifier (BT-48)",
	}

	BRO6 = Rule{
		Code:   "BR-O-6",
		Fields: []string{"BT-116"},
		Description: "VAT category taxable amount (BT-116) must be provided for not subject to VAT",
	}

	BRO7 = Rule{
		Code:   "BR-O-7",
		Fields: []string{"BT-117"},
		Description: "VAT category tax amount (BT-117) must be zero for not subject to VAT",
	}

	BRO8 = Rule{
		Code:   "BR-O-8",
		Fields: []string{"BT-118"},
		Description: "VAT category code (BT-118) must be 'O' for not subject to VAT",
	}

	BRO9 = Rule{
		Code:   "BR-O-9",
		Fields: []string{"BT-119"},
		Description: "VAT category rate (BT-119) must not be provided for not subject to VAT",
	}

	BRO10 = Rule{
		Code:   "BR-O-10",
		Fields: []string{"BT-120", "BT-121"},
		Description: "VAT exemption reason code (BT-120) or VAT exemption reason text (BT-121) must be provided for not subject to VAT",
	}

	BRO11 = Rule{
		Code:   "BR-O-11",
		Fields: []string{"BT-151"},
		Description: "Invoice line VAT category code (BT-151) must be 'O' when not subject to VAT",
	}

	BRO12 = Rule{
		Code:   "BR-O-12",
		Fields: []string{"BT-152"},
		Description: "Invoice line VAT rate (BT-152) must not be provided when not subject to VAT",
	}

	BRO13 = Rule{
		Code:   "BR-O-13",
		Fields: []string{"BT-95"},
		Description: "Document level allowance VAT category code (BT-95) must be 'O' when not subject to VAT",
	}

	BRO14 = Rule{
		Code:   "BR-O-14",
		Fields: []string{"BT-102"},
		Description: "Document level charge VAT category code (BT-102) must be 'O' when not subject to VAT",
	}

	// Additional check for line total calculation
	Check = Rule{
		Code:   "Check",
		Fields: []string{"BT-146", "BT-149", "BT-131"},
		Description: "Invoice line net amount (BT-131) = invoiced quantity (BT-129) × item net price (BT-146) / item price base quantity (BT-149)",
	}
)
