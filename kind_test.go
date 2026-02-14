package errors

import (
	"testing"
)

func TestKind_String_AllValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kind Kind
		want string
	}{
		{"Unknown", Unknown, "unknown"},
		{"Validation", Validation, "validation"},
		{"NotFound", NotFound, "not_found"},
		{"Conflict", Conflict, "conflict"},
		{"DomainRule", DomainRule, "domain_rule"},
		{"Authorization", Authorization, "authorization"},
		{"Authentication", Authentication, "authentication"},
		{"Infrastructure", Infrastructure, "infrastructure"},
		{"Internal", Internal, "internal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestKind_String_Undefined(t *testing.T) {
	t.Parallel()
	k := Kind(200)
	if got := k.String(); got != "unknown" {
		t.Errorf("expected 'unknown' for undefined kind, got %v", got)
	}
}
