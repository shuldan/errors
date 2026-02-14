package errors

import (
	"testing"
)

func TestCode_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		code Code
		want string
	}{
		{"non_empty", Code("ERR_001"), "ERR_001"},
		{"empty", Code(""), ""},
		{"unicode", Code("ошибка"), "ошибка"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.code.String(); got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNewCode_Basic(t *testing.T) {
	t.Parallel()
	b := NewCode("TEST")
	if b == nil {
		t.Fatal("expected non-nil CodeBuilder")
	}
	if b.code != Code("TEST") {
		t.Errorf("expected code TEST, got %v", b.code)
	}
}

func TestNewCode_Empty(t *testing.T) {
	t.Parallel()
	b := NewCode("")
	if b.code != Code("") {
		t.Errorf("expected empty code, got %v", b.code)
	}
}

func TestWithPrefix_Basic(t *testing.T) {
	t.Parallel()
	fn := WithPrefix("SVC")
	b := fn("ERR")
	if b.code != Code("SVC_ERR") {
		t.Errorf("expected SVC_ERR, got %v", b.code)
	}
}

func TestWithPrefix_EmptyParts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix string
		suffix string
		want   Code
	}{
		{"empty_prefix", "", "ERR", Code("_ERR")},
		{"empty_name", "SVC", "", Code("SVC_")},
		{"both_empty", "", "", Code("_")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fn := WithPrefix(tt.prefix)
			b := fn(tt.suffix)
			if b.code != tt.want {
				t.Errorf("expected %v, got %v", tt.want, b.code)
			}
		})
	}
}

func TestCodeBuilder_Kind(t *testing.T) {
	t.Parallel()
	b := NewCode("T").Kind(NotFound)
	if b.kind != NotFound {
		t.Errorf("expected NotFound, got %v", b.kind)
	}
}

func TestCodeBuilder_Severity(t *testing.T) {
	t.Parallel()
	b := NewCode("T").Severity(SeverityCritical)
	if b.severity != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %v", b.severity)
	}
}

func TestCodeBuilder_Chaining(t *testing.T) {
	t.Parallel()
	e := NewCode("C").Kind(Validation).Severity(SeverityWarning).New("msg")
	if e.code != Code("C") {
		t.Errorf("expected code C, got %v", e.code)
	}
	if e.kind != Validation {
		t.Errorf("expected Validation, got %v", e.kind)
	}
	if e.severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %v", e.severity)
	}
	if e.message != "msg" {
		t.Errorf("expected msg, got %v", e.message)
	}
}

func TestCodeBuilder_New_EmptyMessage(t *testing.T) {
	t.Parallel()
	e := NewCode("X").New("")
	if e.message != "" {
		t.Errorf("expected empty message, got %v", e.message)
	}
}
