package main

import "unicode/utf8"

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var out string
	col := 0
	for _, r := range s {
		if r == '\n' {
			out += "\n"
			col = 0
			continue
		}
		if r == ' ' && col >= width {
			out += "\n"
			col = 0
			continue
		}
		out += string(r)
		col++
		if col >= width && r != '\n' {
			out += "\n"
			col = 0
		}
	}
	return out
}

func padRight(s string, n int) string {
	w := utf8.RuneCountInString(s)
	if w >= n {
		return s
	}
	return s + string(make([]rune, n-w))
}

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
