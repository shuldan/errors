package errors

import (
	"testing"
)

func TestSeverity_String_AllValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"Error", SeverityError, "error"},
		{"Warning", SeverityWarning, "warning"},
		{"Critical", SeverityCritical, "critical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestSeverity_String_Undefined(t *testing.T) {
	t.Parallel()
	s := Severity(100)
	if got := s.String(); got != "error" {
		t.Errorf("expected 'error' for undefined severity, got %v", got)
	}
}
