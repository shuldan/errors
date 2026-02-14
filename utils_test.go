package errors

import (
	"errors"
	"testing"
)

func TestIs_Matching(t *testing.T) {
	t.Parallel()
	base := errors.New("base")
	wrapped := newTestError("C", "m").WithCause(base)
	if !Is(wrapped, base) {
		t.Errorf("expected Is=true for matching cause")
	}
}

func TestIs_NotMatching(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	if Is(e, errors.New("other")) {
		t.Errorf("expected Is=false for non-matching error")
	}
}

func TestAs_Success(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m").clone()
	var target *Error
	if !As(e, &target) {
		t.Errorf("expected As=true for *Error")
	}
	if target.GetCode() != Code("C") {
		t.Errorf("expected code C, got %v", target.GetCode())
	}
}

func TestAs_Fail(t *testing.T) {
	t.Parallel()
	plain := errors.New("plain")
	var target *Error
	if As(plain, &target) {
		t.Errorf("expected As=false for plain error")
	}
}

func TestUnwrap_WithCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	e := newTestError("C", "m").WithCause(cause)
	if Unwrap(e) != cause {
		t.Errorf("expected Unwrap to return cause")
	}
}

func TestUnwrap_NoCause(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	if Unwrap(e) != nil {
		t.Errorf("expected Unwrap to return nil")
	}
}

func TestJoin_Multiple(t *testing.T) {
	t.Parallel()
	e1 := errors.New("a")
	e2 := errors.New("b")
	joined := Join(e1, e2)
	if joined == nil {
		t.Fatal("expected non-nil joined error")
	}
	if !errors.Is(joined, e1) || !errors.Is(joined, e2) {
		t.Errorf("joined error should contain both errors")
	}
}

func TestJoin_Nil(t *testing.T) {
	t.Parallel()
	joined := Join(nil, nil)
	if joined != nil {
		t.Errorf("expected nil for all-nil Join")
	}
}

func TestWrap_Basic(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	tmpl := newTestError("W", "wrap msg")
	wrapped := Wrap(cause, tmpl)
	if wrapped.GetCause() != cause {
		t.Errorf("expected cause to be set")
	}
	if wrapped.GetCode() != Code("W") {
		t.Errorf("expected code W, got %v", wrapped.GetCode())
	}
}

func TestGetCode_WithError(t *testing.T) {
	t.Parallel()
	e := newError(Code("GC"), NotFound, SeverityError, "m").clone()
	if got := GetCode(e); got != Code("GC") {
		t.Errorf("expected GC, got %v", got)
	}
}

func TestGetCode_NonError(t *testing.T) {
	t.Parallel()
	if got := GetCode(errors.New("x")); got != "" {
		t.Errorf("expected empty code, got %v", got)
	}
}

func TestGetKind_WithError(t *testing.T) {
	t.Parallel()
	e := newError(Code("GK"), Conflict, SeverityError, "m").clone()
	if got := GetKind(e); got != Conflict {
		t.Errorf("expected Conflict, got %v", got)
	}
}

func TestGetKind_NonError(t *testing.T) {
	t.Parallel()
	if got := GetKind(errors.New("x")); got != Unknown {
		t.Errorf("expected Unknown, got %v", got)
	}
}

func TestGetSeverity_WithError(t *testing.T) {
	t.Parallel()
	e := newError(Code("GS"), Validation, SeverityCritical, "m").clone()
	if got := GetSeverity(e); got != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %v", got)
	}
}

func TestGetSeverity_NonError(t *testing.T) {
	t.Parallel()
	if got := GetSeverity(errors.New("x")); got != SeverityError {
		t.Errorf("expected SeverityError, got %v", got)
	}
}

func TestIs_CustomErrorCodes(t *testing.T) {
	t.Parallel()
	e1 := newError(Code("SAME"), Validation, SeverityError, "a").clone()
	e2 := newError(Code("SAME"), NotFound, SeverityCritical, "b").clone()
	if !Is(e1, e2) {
		t.Errorf("expected Is=true for same code via stdlib Is")
	}
}

func TestIs_DifferentCustomErrorCodes(t *testing.T) {
	t.Parallel()
	e1 := newError(Code("A"), Validation, SeverityError, "a").clone()
	e2 := newError(Code("B"), Validation, SeverityError, "b").clone()
	if Is(e1, e2) {
		t.Errorf("expected Is=false for different codes")
	}
}
