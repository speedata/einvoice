# Bug Report and Spec Compliance Issues

This report documents bugs, spec violations, and potential issues found in the einvoice library.

**Last Updated:** 2025-10-05

## Status Summary

### Fixed Bugs ✅
- Bug #1: UpdateTotals() LineTotal calculation (calculate.go:61-63)
- Bug #2: UpdateTotals() ignores allowances/charges (calculate.go:71)
- Bug #3: DuePayableAmount missing RoundingAmount (calculate.go:77)
- Bug #6: Incorrect BT field in BR-11 error message (check.go:737)
- Bug #7: Incorrect rule number for charges (check.go:880)
- Bug #9: Silent error handling in getDecimal (parser.go:79-89)
- Bug #10: Duplicate assignment in parser (parser.go:156)
- Bug #11: Missing CountrySubDivisionName parsing (parser.go:68)

### Active Bugs 🔴
- Bug #4: UpdateApplicableTradeTax doesn't account for document-level allowances/charges (High)
- Bug #5: Tax basis calculation uses wrong key (CategoryCode missing) (Medium)
- Bug #8: Inconsistent decimal formatting in writer (Low)
- Bug #20: Test has duplicate LineID (Medium)
- Bug #25: Typo in profile name string "EN 19631" → "EN 16931" (Trivial)

### Outstanding Spec Compliance Issues 📋
- Multiple missing BR-CO-* validations (#12-19)
- Incomplete category-specific validations (BR-AE-*, BR-E-*, etc.)

---

## Critical Bugs

### 1. ~~**UpdateTotals() calculates LineTotal incorrectly**~~ ✅ FIXED

**Status:** ✅ **FIXED** (calculate.go:55-78)
**Severity:** Critical
**Originally reported:** calculate.go:54-56

The `UpdateTotals()` function has been corrected and now:
- Properly resets totals to zero before recalculation (lines 57-58)
- Correctly calculates `LineTotal` from invoice lines (lines 61-63)
- Properly applies document-level allowances and charges (line 71)
- Includes `RoundingAmount` in `DuePayableAmount` calculation (line 77)

**Fixed implementation:**
```go
func (inv *Invoice) UpdateTotals() {
    // Reset all calculated totals to zero to ensure idempotent behavior
    inv.LineTotal = decimal.Zero
    inv.TaxTotal = decimal.Zero

    // BR-CO-10: Calculate line total from invoice lines (BT-106)
    for _, line := range inv.InvoiceLines {
        inv.LineTotal = inv.LineTotal.Add(line.Total)
    }

    // Calculate tax total from trade taxes (BT-110)
    for _, v := range inv.TradeTaxes {
        inv.TaxTotal = inv.TaxTotal.Add(v.CalculatedAmount)
    }

    // BR-CO-13: Apply document-level allowances and charges
    inv.TaxBasisTotal = inv.LineTotal.Sub(inv.AllowanceTotal).Add(inv.ChargeTotal)

    // BR-CO-15: Calculate grand total
    inv.GrandTotal = inv.TaxBasisTotal.Add(inv.TaxTotal)

    // BR-CO-16: Calculate due payable amount including rounding
    inv.DuePayableAmount = inv.GrandTotal.Sub(inv.TotalPrepaid).Add(inv.RoundingAmount)
}
```

**Spec reference:** BR-CO-10, BR-CO-13, BR-CO-15, BR-CO-16

---

### 2. ~~**UpdateTotals() ignores allowances and charges**~~ ✅ FIXED

**Status:** ✅ **FIXED** (calculate.go:71)
**Severity:** High

See Bug #1 - this has been fixed as part of the same correction.

---

### 3. ~~**DuePayableAmount calculation missing RoundingAmount**~~ ✅ FIXED

**Status:** ✅ **FIXED** (calculate.go:77)
**Severity:** High

See Bug #1 - this has been fixed as part of the same correction.

---

### 4. **UpdateApplicableTradeTax doesn't account for document-level allowances/charges** 🔴 ACTIVE

**Status:** 🔴 **ACTIVE BUG**
**Severity:** High
**Location:** `calculate.go:10-47`

The `UpdateApplicableTradeTax()` function only aggregates taxes from invoice lines (lines 13-34), but ignores document-level allowances and charges that also have tax categories.

According to BR-45 and the various category-specific rules (BR-S-8, BR-AE-8, BR-E-8, etc.), the VAT category taxable amount must include:

- Sum of invoice line net amounts for that category
- **Minus** document level allowance amounts for that category
- **Plus** document level charge amounts for that category

**Current code only processes:**
```go
for _, lineitem := range inv.InvoiceLines {
    // Only aggregates from invoice lines
}
```

**Missing:** Iteration over `inv.SpecifiedTradeAllowanceCharge` to adjust tax basis amounts.

**Impact:** Tax basis amounts will be incorrect when document-level allowances or charges are present.

---

### 5. **Tax basis calculation uses wrong key** 🔴 ACTIVE

**Status:** 🔴 **ACTIVE BUG**
**Severity:** Medium
**Location:** `check.go:852-880` (BR-45 validation)

The BR-45 validation uses only the tax percentage as the map key:

```go
// Line 855
percentString := lineitem.TaxRateApplicablePercent.String()
applicableTradeTaxes[percentString] = applicableTradeTaxes[percentString].Add(lineitem.Total)

// Line 946 - validation
if !applicableTradeTaxes[tt.Percent.String()].Equal(tt.BasisAmount) {
    inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-45", ...})
}
```

**Problem:**
This incorrectly groups tax amounts that have the same percentage but different category codes. For example:
- A 19% standard rate (category "S")
- A 19% reverse charge (category "AE" with 19% in some edge cases)

Would be incorrectly combined into a single tax basis.

**Fix:**
Use a composite key of `CategoryCode + "_" + Percent` instead of just `Percent`.

---

## Medium Severity Bugs

### 6. ~~**Incorrect BT field in error message**~~ ✅ FIXED

**Status:** ✅ **FIXED** (check.go:737)
**Originally reported:** check.go:637

The BR-11 error message now correctly references `"BT-55"` (Buyer country code) instead of the incorrect `"BT-5"`.

**Fixed code:**
```go
inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-11", InvFields: []string{"BT-55"}, Text: "Buyer country code empty"})
```

---

### 7. ~~**Incorrect rule number in error message**~~ ✅ FIXED

**Status:** ✅ **FIXED** (check.go:880)
**Originally reported:** check.go:784

The error message for charge tax category validation now correctly uses rule `"BR-37"` instead of the incorrect `"BR-32"`.

**Fixed code:**
```go
inv.Violations = append(inv.Violations, SemanticError{Rule: "BR-37", InvFields: []string{"BG-21", "BT-102"}, Text: "Charge tax category code not set"})
```

---

### 8. **Inconsistent decimal formatting** 🔴 ACTIVE

**Status:** 🔴 **ACTIVE BUG**
**Severity:** Low
**Location:** `writer.go:83`

GrossPrice uses `StringFixed(12)` while all other monetary amounts use `StringFixed(2)`:

```go
gpptp.CreateElement("ram:ChargeAmount").SetText(invoiceLine.GrossPrice.StringFixed(12))
```

**Issue:** This inconsistency may cause:
- Parsing issues with some invoice readers
- Unnecessarily high precision that doesn't match EN 16931 requirements
- Inconsistent data representation

**Recommendation:** Use `StringFixed(2)` for consistency with other monetary amounts, or document why 12 decimal places are needed for GrossPrice specifically.

---

### 9. ~~**Silent error handling in getDecimal**~~ ✅ FIXED

**Status:** ✅ **FIXED** (parser.go:79-89)
**Originally reported:** parser.go:78-82

The `getDecimal()` function now properly returns errors instead of silently ignoring them:

**Fixed implementation:**
```go
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
```

---

### 10. ~~**Duplicate assignment**~~ ✅ FIXED

**Status:** ✅ **FIXED** (parser.go:156)
**Originally reported:** parser.go:133-140

The duplicate assignment of `spt.Description` has been removed. The code now has a single assignment at line 156.

---

### 11. ~~**Missing parsing of PostalAddress.CountrySubDivisionName**~~ ✅ FIXED

**Status:** ✅ **FIXED** (parser.go:68)
**Originally reported:** parser.go:60-69

The `CountrySubDivisionName` field (BT-39, BT-54, BT-68, BT-79) is now being parsed:

**Fixed code:**
```go
postalAddress := &PostalAddress{
    PostcodeCode:           tradeParty.Eval("ram:PostalTradeAddress/ram:PostcodeCode").String(),
    Line1:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineOne").String(),
    Line2:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineTwo").String(),
    Line3:                  tradeParty.Eval("ram:PostalTradeAddress/ram:LineThree").String(),
    City:                   tradeParty.Eval("ram:PostalTradeAddress/ram:CityName").String(),
    CountryID:              tradeParty.Eval("ram:PostalTradeAddress/ram:CountryID").String(),
    CountrySubDivisionName: tradeParty.Eval("ram:PostalTradeAddress/ram:CountrySubDivisionName").String(),
}
```

---

## Spec Compliance Issues

### 12. **Missing BR-CO-3 validation**

**Severity:** Medium
**Referenced in:** `check.go:27-29` (commented)

BR-CO-3: "Umsatzsteuerdatum (BT-7) und Code für das Umsatzsteuerdatum (BT-8) schließen sich gegenseitig aus."

The rule states that TaxPointDate (BT-7) and DueDateTypeCode (BT-8) are mutually exclusive, but there's no validation for this.

---

### 13. **Missing BR-CO-4 validation**

**Referenced in:** `check.go:30-32` (commented)

BR-CO-4: Each invoice line must be categorized with a VAT category code (BT-151).

Currently not validated. Should check that `invoiceLine.TaxCategoryCode` is not empty for all invoice lines.

---

### 14. **Missing BR-CO-17 validation**

**Referenced in:** `check.go:70-72` (commented)

BR-CO-17: VAT category tax amount (BT-117) must equal VAT category taxable amount (BT-116) × VAT rate (BT-119) ÷ 100, rounded to 2 decimals.

This calculation should be validated in `checkBRO()`.

---

### 15. **Missing BR-CO-18 validation**

**Referenced in:** `check.go:73-74` (commented)

BR-CO-18: An invoice should contain at least one VAT BREAKDOWN (BG-23).

Should validate `len(inv.TradeTaxes) >= 1`.

---

### 16. **Missing BR-CO-19 validation**

**Referenced in:** `check.go:75-77` (commented)

BR-CO-19: If INVOICING PERIOD (BG-14) is used, either start date (BT-73) or end date (BT-74) or both must be filled.

---

### 17. **Missing BR-CO-20 validation**

**Referenced in:** `check.go:78-80` (commented)

BR-CO-20: If INVOICE LINE PERIOD (BG-26) is used, either start date (BT-134) or end date (BT-135) or both must be filled.

---

### 18. **Missing BR-CO-25 validation**

**Referenced in:** `check.go:96-98` (commented)

BR-CO-25: If amount due for payment (BT-115) is positive, either payment due date (BT-9) or payment terms (BT-20) must be present.

---

### 19. **Incomplete category-specific validations**

Many category-specific business rules are documented in comments (BR-AE-*, BR-E-*, BR-G-*, BR-IC-*, BR-IG-*, BR-IP-*, BR-O-*, BR-S-*) but not implemented. These include:

- BR-AE-1 through BR-AE-10 (Reverse charge)
- BR-E-1 through BR-E-10 (Exempt from VAT)
- BR-G-1 through BR-G-10 (Export outside EU)
- BR-IC-1 through BR-IC-12 (Intra-community supply)
- BR-IG-1 through BR-IG-10 (IGIC - Canary Islands)
- BR-IP-1 through BR-IP-10 (IPSI - Ceuta/Melilla)
- BR-O-1 through BR-O-14 (Not subject to VAT)
- BR-S-1 through BR-S-10 (Standard rated)

---

## Test Issues

### 20. **Test has duplicate LineID** 🔴 ACTIVE

**Status:** 🔴 **ACTIVE BUG**
**Severity:** Medium
**Location:** `einvoice_test.go:67-78`

The test example contains duplicate LineID values, which violates BR-21:

```go
InvoiceLines: []einvoice.InvoiceLine{
    {
        LineID: "1",  // Line 67 ❌
        ItemName: "Item name one",
        ...
    },
    {
        LineID: "1",  // Line 78 ❌ DUPLICATE!
        ItemName: "Item name two",
        ...
    },
},
```

**Violation:** BR-21 states "Each invoice line must have a unique invoice line identifier."

**Impact:**
- The test creates an invalid invoice that would fail BR-21 validation
- Sets a bad example for library users
- May pass tests despite producing spec-violating output

**Fix:** Change the second line to `LineID: "2"`.

---

## Code Quality Issues

### 21. **checkBR and checkBRO mixing of responsibilities**

The distinction between `checkBR()` (basic rules BR-1 through BR-65) and `checkBRO()` (cardinality/calculation rules BR-CO-*) is not consistently maintained. For example:

- BR-CO-10 is in `checkBRO()` ✓ (correct)
- But BR-29 and BR-30 (period date logic) are in `checkBR()` when they could be considered BR-CO rules

This is a minor organizational issue but doesn't affect correctness.

---

### 22. **Missing validation: Line item total calculation**

**Location:** `check.go:554-563`

The `checkOther()` function validates:
```
line.Total == line.BilledQuantity × line.NetPrice
```

However, this doesn't account for line-level allowances and charges. According to the spec, the line net amount should be calculated as:

```
Line Net Amount (BT-131) = (BilledQuantity × NetPrice) - Σ(LineAllowances) + Σ(LineCharges)
```

The current check would fail for invoices with line-level allowances/charges even if they're correctly calculated.

---

### 23. **Writer doesn't validate before writing**

The `Write()` function in `writer.go` doesn't call `inv.check()` before writing. This means invalid invoices can be written to XML. Consider adding validation before output, or at least documenting that callers should validate first.

---

### 24. **Missing SchemaType initialization when creating invoices programmatically**

When users create invoices programmatically (as in the test example), they need to manually set `SchemaType = CII`. If they forget, the `Write()` method will return `ErrUnsupportedSchema`. This should be documented or defaulted.

---

## Recommendations

### ✅ Recent Fixes (Completed)
The following bugs have been successfully fixed:
- ✅ `UpdateTotals()` calculation (Bugs #1, #2, #3)
- ✅ Error handling in `getDecimal()` (Bug #9)
- ✅ Incorrect error messages (Bugs #6, #7)
- ✅ Duplicate code removal (Bug #10)
- ✅ Missing field parsing (Bug #11)

### Priority 1 (High - Fix Soon) 🔴
1. **Fix `UpdateApplicableTradeTax()` to include document-level allowances/charges** (Bug #4)
   - Impact: Tax basis amounts are incorrect when document-level allowances/charges exist
   - Required for: BR-45, BR-S-8, BR-AE-8, BR-E-8, etc.

2. **Fix BR-45 tax basis key calculation** (Bug #5)
   - Use composite key: `CategoryCode + "_" + Percent`
   - Impact: Incorrect validation when multiple categories share the same tax rate

### Priority 2 (Medium - Improve Quality)
3. **Fix test duplicate LineID** (Bug #20)
   - Simple fix: Change line 78 from `LineID: "1"` to `LineID: "2"`
   - Impact: Test currently creates spec-violating invoices

4. **Fix decimal formatting consistency** (Bug #8)
   - Decision needed: Use `StringFixed(2)` or document why 12 decimals needed
   - Impact: Minor - may cause parsing issues with some readers

5. **Implement missing BR-CO-* validations** (Issues #12-19)
   - BR-CO-3, BR-CO-4, BR-CO-17, BR-CO-18, BR-CO-19, BR-CO-20, BR-CO-25
   - Impact: Invoices may violate spec without detection

### Priority 3 (Low - Future Enhancements)
6. **Add line-level allowances/charges to total calculation check** (Issue #22)
   - Current check only validates: `line.Total == line.BilledQuantity × line.NetPrice`
   - Should account for line-level allowances/charges

7. **Implement category-specific business rules** (Issue #19)
   - BR-AE-*, BR-E-*, BR-G-*, BR-IC-*, BR-IG-*, BR-IP-*, BR-O-*, BR-S-*
   - Impact: Category-specific violations not detected

8. **Add pre-write validation** (Issue #23)
   - Consider calling `inv.check()` before writing XML
   - Or document that callers should validate first

9. **Improve API ergonomics** (Issue #24)
   - Default `SchemaType = CII` for programmatically created invoices
   - Better documentation of initialization requirements

---

## Additional Minor Issues

### 25. **Typo in profile name string** 🔴 ACTIVE

**Status:** 🔴 **ACTIVE BUG**
**Severity:** Trivial
**Location:** `model.go:49`

The CProfileEN16931 String() method contains a typo:

```go
case CProfileEN16931:
    return "EN 19631"  // ❌ Should be "EN 16931"
```

**Impact:** Low - cosmetic issue in string representation, doesn't affect functionality.

**Fix:** Change to `return "EN 16931"` to match the actual standard number.

---

## Testing Recommendations

1. Add unit tests for `UpdateTotals()` with various scenarios:
   - Multiple invoice lines
   - Document-level allowances and charges
   - Rounding amounts
   - Multiple calls to UpdateTotals()

2. Add integration tests with real ZUGFeRD/Factur-X samples covering:
   - All profile types (Minimum, Basic, EN16931, Extended)
   - All VAT category codes (S, E, AE, K, G, O, etc.)
   - Document-level and line-level allowances/charges

3. Add validation tests for all BR-CO-* rules

4. Add round-trip tests (parse → write → parse → compare)

5. Add BR-21 validation test to catch duplicate LineID values
