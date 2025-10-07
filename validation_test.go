package einvoice

import (
	"errors"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name       string
		violations []SemanticError
		want       string
	}{
		{
			name:       "no violations",
			violations: []SemanticError{},
			want:       "validation failed with no violations",
		},
		{
			name: "single violation",
			violations: []SemanticError{
				{Rule: BR1, Text: "Invoice number is required"},
			},
			want: "validation failed: BR-1 - Invoice number is required",
		},
		{
			name: "multiple violations",
			violations: []SemanticError{
				{Rule: BR1, Text: "Invoice number is required"},
				{Rule: BR2, Text: "Invoice date is required"},
				{Rule: BR3, Text: "Currency is required"},
			},
			want: "validation failed with 3 violations (first: BR-1 - Invoice number is required)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{violations: tt.violations}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Violations(t *testing.T) {
	t.Run("returns copy of violations", func(t *testing.T) {
		original := []SemanticError{
			{Rule: BR1, Text: "Test violation"},
		}
		e := &ValidationError{violations: original}

		// Get violations
		violations := e.Violations()

		// Verify content
		if len(violations) != 1 {
			t.Errorf("Violations() returned %d violations, want 1", len(violations))
		}
		if violations[0].Rule.Code != "BR-1" {
			t.Errorf("Violations()[0].Rule.Code = %v, want BR-1", violations[0].Rule.Code)
		}

		// Modify the returned slice - should not affect internal state
		violations[0].Rule = BR2

		// Verify internal state unchanged
		if e.violations[0].Rule.Code != "BR-1" {
			t.Errorf("Internal violations were modified, want BR-1, got %v", e.violations[0].Rule.Code)
		}
	})

	t.Run("returns nil for nil violations", func(t *testing.T) {
		e := &ValidationError{violations: nil}
		violations := e.Violations()
		if violations != nil {
			t.Errorf("Violations() = %v, want nil", violations)
		}
	})

	t.Run("returns empty slice for empty violations", func(t *testing.T) {
		e := &ValidationError{violations: []SemanticError{}}
		violations := e.Violations()
		if violations == nil {
			t.Error("Violations() = nil, want empty slice")
		}
		if len(violations) != 0 {
			t.Errorf("Violations() length = %d, want 0", len(violations))
		}
	})
}

func TestValidationError_Count(t *testing.T) {
	tests := []struct {
		name       string
		violations []SemanticError
		want       int
	}{
		{
			name:       "no violations",
			violations: []SemanticError{},
			want:       0,
		},
		{
			name: "one violation",
			violations: []SemanticError{
				{Rule: BR1, Text: "Test"},
			},
			want: 1,
		},
		{
			name: "multiple violations",
			violations: []SemanticError{
				{Rule: BR1, Text: "Test 1"},
				{Rule: BR2, Text: "Test 2"},
				{Rule: BR3, Text: "Test 3"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{violations: tt.violations}
			if got := e.Count(); got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_HasRule(t *testing.T) {
	violations := []SemanticError{
		{Rule: BR1, Text: "Test 1"},
		{Rule: BRS8, Text: "Test 2"},
		{Rule: BRCO10, Text: "Test 3"},
	}
	e := &ValidationError{violations: violations}

	tests := []struct {
		name string
		rule string
		want bool
	}{
		{
			name: "rule exists - BR-1",
			rule: "BR-1",
			want: true,
		},
		{
			name: "rule exists - BR-S-8",
			rule: "BR-S-8",
			want: true,
		},
		{
			name: "rule exists - BR-CO-10",
			rule: "BR-CO-10",
			want: true,
		},
		{
			name: "rule does not exist",
			rule: "BR-99",
			want: false,
		},
		{
			name: "empty rule",
			rule: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := e.HasRule(tt.rule); got != tt.want {
				t.Errorf("HasRule(%v) = %v, want %v", tt.rule, got, tt.want)
			}
		})
	}
}

func TestValidationError_AsError(t *testing.T) {
	t.Run("can be used with errors.As", func(t *testing.T) {
		originalErr := &ValidationError{
			violations: []SemanticError{
				{Rule: BR1, Text: "Test violation"},
			},
		}

		var err error = originalErr

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Error("errors.As failed to extract ValidationError")
		}

		if valErr.Count() != 1 {
			t.Errorf("Count() = %d, want 1", valErr.Count())
		}

		if !valErr.HasRule("BR-1") {
			t.Error("HasRule(BR-1) = false, want true")
		}
	})
}
