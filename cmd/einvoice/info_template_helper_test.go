package main

import (
	"strings"
	"testing"
)

func TestUnderline(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"basic", "Hello", "Hello\n-----"},
		{"empty", "", "\n"},
		{"unicode", "Hä😀", "Hä😀\n---"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := underline(tc.in); got != tc.want {
				t.Fatalf("underline(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestWrapTextIndent(t *testing.T) {
	type checkFn func(t *testing.T, out string)

	hasMultipleLines := func() checkFn {
		return func(t *testing.T, out string) {
			if len(strings.Split(out, "\n")) < 2 {
				t.Fatalf("expected multiple lines, got: %q", out)
			}
		}
	}
	allButFirstIndented := func(spaces int) checkFn {
		return func(t *testing.T, out string) {
			lines := strings.Split(out, "\n")
			prefix := strings.Repeat(" ", spaces)
			for i := range lines {
				if i == 0 {
					continue
				}
				if !strings.HasPrefix(lines[i], prefix) {
					t.Fatalf("line %d not indented with %d spaces: %q", i+1, spaces, lines[i])
				}
			}
		}
	}
	preservesLeadingWS := func(ws string) checkFn {
		return func(t *testing.T, out string) {
			if !strings.HasPrefix(out, ws) {
				t.Fatalf("expected leading whitespace %q to be preserved, got: %q", ws, out)
			}
		}
	}
	noSplitSingleLongWord := func(expected string) checkFn {
		return func(t *testing.T, out string) {
			if strings.Contains(out, "\n") {
				t.Fatalf("a single long word must not be split, got: %q", out)
			}
			if out != expected {
				t.Fatalf("expected unchanged long word, got: %q", out)
			}
		}
	}

	tests := []struct {
		name   string
		text   string
		width  int
		indent int
		checks []checkFn
		want   string // optional exact match
	}{
		{
			name:   "wraps_and_indents",
			text:   "This is a simple test of the wrapping function.",
			width:  15,
			indent: 4,
			checks: []checkFn{hasMultipleLines(), allButFirstIndented(4)},
		},
		{
			name:   "width_lt_1_returns_original",
			text:   "Word",
			width:  0,
			indent: 2,
			want:   "Word",
		},
		{
			name:   "single_long_word_not_split",
			text:   "Supercalifragilisticexpialidocious",
			width:  10,
			indent: 2,
			checks: []checkFn{noSplitSingleLongWord("Supercalifragilisticexpialidocious")},
		},
		{
			name:   "preserves_leading_spaces",
			text:   "   indented text test",
			width:  10,
			indent: 2,
			checks: []checkFn{preservesLeadingWS("   ")},
		},
		{
			name:   "only_whitespace_input",
			text:   "    \t   ",
			width:  10,
			indent: 2,
			want:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := wrapTextIndent(tc.text, tc.width, tc.indent)
			if tc.want != "" && got != tc.want {
				t.Fatalf("wrapTextIndent() = %q, want %q", got, tc.want)
			}
			for _, chk := range tc.checks {
				chk(t, got)
			}

			// Sanity check: Only when all words are <= width, the produced
			// lines must remain <= width.
			if tc.width >= 1 && tc.want == "" {
				words := strings.Fields(tc.text)
				maxWordLen := 0
				for _, w := range words {
					if l := len(w); l > maxWordLen {
						maxWordLen = l
					}
				}
				if maxWordLen <= tc.width {
					for i, line := range strings.Split(got, "\n") {
						if len(line) > tc.width {
							t.Fatalf("line %d exceeds width %d: %q (len=%d)", i+1, tc.width, line, len(line))
						}
					}
				}
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"pads_short", "Hi", 5, "Hi   "},
		{"no_truncate_when_wide_enough", "Hello", 3, "Hello"},
		{"exact_width", "Hey", 3, "Hey"},
		{"unicode_wide_runes_counted_as_runes", "Ä", 2, "Ä "}, // rune-aware padding
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := padRight(tc.in, tc.n); got != tc.want {
				t.Fatalf("padRight(%q,%d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
		})
	}
}

func TestPadMiddle(t *testing.T) {
	tests := []struct {
		name        string
		left, right string
		width       int
		want        string
	}{
		{"pads_between", "A", "B", 5, "A   B"},
		{"concatenate_when_over", "Hello", "World", 5, "HelloWorld"},
		{"exact_width_no_padding", "AB", "CD", 4, "ABCD"},
		{"unicode_rune_count", "Ä", "Ω", 3, "Ä Ω"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := padMiddle(tc.left, tc.right, tc.width); got != tc.want {
				t.Fatalf("padMiddle(%q,%q,%d) = %q, want %q", tc.left, tc.right, tc.width, got, tc.want)
			}
		})
	}
}

func TestHr(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"five", 5, "─────"},
		{"zero_defaults_to_one", 0, "─"},
		{"negative_defaults_to_one", -3, "─"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hr(tc.n); got != tc.want {
				t.Fatalf("hr(%d) = %q, want %q", tc.n, got, tc.want)
			}
		})
	}
}
