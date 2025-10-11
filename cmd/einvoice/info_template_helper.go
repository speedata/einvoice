package main

import (
	"strings"
	"unicode/utf8"
)

// underline returns the input string s with a new line and an underline
// consisting of '-' characters of the same length as s.
func underline(s string) string {
	return s + "\n" + strings.Repeat("-", utf8.RuneCountInString(s))
}

// wrapTextIndent wraps the input text to the specified width n. It breaks lines at
// word boundaries (spaces) where possible. If a single word exceeds the width n,
// it will be placed on its own line without breaking it. Lines are separated by
// newline characters. If n is less than 1, the original text is returned.
// indentation is added to each line except the first.
func wrapTextIndent(text string, n int, indent int) string {
	if n < 1 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var wrappedLines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() == 0 {
			// Start a new line with the current word
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= n {
			// Append the word to the current line
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			// Current line is full, save it and start a new line
			wrappedLines = append(wrappedLines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(strings.Repeat(" ", indent))
			currentLine.WriteString(word)
		}
	}
	// Append any remaining text in the current line
	if currentLine.Len() > 0 {
		wrappedLines = append(wrappedLines, currentLine.String())
	}

	return strings.Join(wrappedLines, "\n")
}

// padRight pads the string s with spaces on the right to ensure it is at least
// n characters wide. If s is already n or more characters wide, it is returned
// unchanged.
func padRight(s string, n int) string {
	w := utf8.RuneCountInString(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func padMiddle(left, right string, width int) string {
	lw := utf8.RuneCountInString(left)
	rw := utf8.RuneCountInString(right)
	if lw+rw >= width {
		return left + right
	}
	mid := strings.Repeat(" ", width-lw-rw)
	return left + mid + right
}

// hr returns a horizontal rule (line) of n characters wide using the '─'
// character.
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
