package main

import (
	"strings"
	"unicode/utf8"
)

// underline returns s followed by a newline and an underline made of '-'.
// The underline length equals the number of runes in s, so it aligns with
// Unicode strings as expected (not bytes).
func underline(s string) string {
	return s + "\n" + strings.Repeat("-", utf8.RuneCountInString(s))
}

// wrapTextIndent wraps text at word boundaries (spaces) to a target width n,
// measured in runes (Unicode-safe). Words longer than n are NOT split and will
// occupy their own line. Lines are separated by '\n'. If n < 1, text is
// returned unchanged.
//
// Indentation:
//   - The leading run of spaces/tabs at the start of text is preserved on the
//     FIRST output line.
//   - Each SUBSEQUENT line is prefixed with `indent` spaces.
//
// Whitespace-only input:
//   - If text contains only spaces/tabs, an empty string is returned.
//
// Notes on measurement/performance:
//   - Line and word lengths are measured in runes (utf8), not bytes.
//   - The current line width is tracked incrementally to avoid copying from
//     strings.Builder on each check.
func wrapTextIndent(text string, n int, indent int) string {
	if n < 1 {
		// If width is less than 1, return the input unchanged.
		return text
	}

	// Preserve the initial whitespace prefix (spaces/tabs only) from the input.
	prefix := ""
	for _, r := range text {
		if r == ' ' || r == '\t' {
			prefix += string(r)
		} else {
			break
		}
	}

	// Split into words using Fields (collapses runs of whitespace).
	words := strings.Fields(text)
	if len(words) == 0 {
		// Whitespace-only input => empty string.
		return ""
	}

	var wrapped []string
	var line strings.Builder

	// Start the first line with the preserved prefix.
	line.WriteString(prefix)

	// Track the current line width in runes (Unicode-safe).
	currWidth := utf8.RuneCountInString(prefix)

	for _, w := range words {
		wr := utf8.RuneCountInString(w)

		if currWidth == 0 {
			// First visible token in this line: write the word directly.
			line.WriteString(w)
			currWidth += wr
			continue
		}

		// Check if we can add a space plus the next word within the target width.
		if currWidth+1+wr <= n {
			// Append with a separating space.
			line.WriteString(" ")
			line.WriteString(w)
			currWidth += 1 + wr
		} else {
			// Line is full: commit it, then start a new indented line.
			wrapped = append(wrapped, line.String())
			line.Reset()

			// Add indentation to subsequent lines.
			line.WriteString(strings.Repeat(" ", indent))
			currWidth = indent

			// Place the word on the new line. Long words may exceed n (by design).
			line.WriteString(w)
			currWidth += wr
		}
	}

	// Flush the final line if it contains content.
	if line.Len() > 0 {
		wrapped = append(wrapped, line.String())
	}

	// Join all lines with '\n'.
	return strings.Join(wrapped, "\n")
}

// padRight returns s padded on the right with spaces to ensure a minimum width
// of n runes. If s is already at least n runes wide, it is returned unchanged.
// (Rune-aware: multi-byte Unicode characters count as a single column unit.)
func padRight(s string, n int) string {
	w := utf8.RuneCountInString(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

// padMiddle places left and right into a field of total width runes, inserting
// spaces between them. If left+right is already >= width runes, the two strings
// are simply concatenated without extra spaces. Width is evaluated in runes.
func padMiddle(left, right string, width int) string {
	lw := utf8.RuneCountInString(left)
	rw := utf8.RuneCountInString(right)
	if lw+rw >= width {
		return left + right
	}
	mid := strings.Repeat(" ", width-lw-rw)
	return left + mid + right
}

// hr returns a horizontal rule of n copies of '─'. If n < 1, a single '─' is
// returned. The result length is measured in runes, not bytes.
func hr(n int) string {
	if n < 1 {
		n = 1
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = '─'
	}
	return string(b)
}
