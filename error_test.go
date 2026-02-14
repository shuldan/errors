package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func newTestError(code, msg string) *Error {
	return newError(Code(code), Validation, SeverityError, msg)
}

func TestError_Error_AllBranches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			"msg_and_cause",
			newTestError("C", "m").clone().WithCause(errors.New("root")),
			"C: m: root",
		},
		{
			"msg_no_cause",
			newTestError("C", "hello"),
			"C: hello",
		},
		{
			"no_msg_with_cause",
			newError(Code("C"), Validation, SeverityError, "").
				WithCause(errors.New("root")),
			"C: root",
		},
		{
			"no_msg_no_cause",
			newError(Code("C"), Validation, SeverityError, ""),
			"C",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestError_RenderMessage_WithTemplate(t *testing.T) {
	t.Parallel()
	e := newError(Code("C"), Validation, SeverityError, "hello {{.Name}}")
	c := e.WithDetail("Name", "world")
	got := c.Error()
	if !strings.Contains(got, "hello world") {
		t.Errorf("expected rendered template, got %q", got)
	}
}

func TestError_RenderMessage_TemplateSuccess_MissingKey(t *testing.T) {
	t.Parallel()
	e := newError(Code("C"), Validation, SeverityError, "val={{.Missing}}")
	got := e.renderMessage()
	if got != "val=<no value>" {
		t.Errorf("expected 'val=<no value>', got %q", got)
	}
}

func TestError_RenderMessage_TemplateExecuteError(t *testing.T) {
	t.Parallel()
	e := newError(Code("C"), Validation, SeverityError, "{{call .Fn}}")
	got := e.renderMessage()
	if got != "{{call .Fn}}" {
		t.Errorf("expected fallback to raw message, got %q", got)
	}
}

func TestError_RenderMessage_InvalidTemplateSyntax(t *testing.T) {
	t.Parallel()
	e := newError(Code("C"), Validation, SeverityError, "{{bad")
	if e.tmpl != nil {
		t.Errorf("expected nil tmpl for invalid syntax")
	}
	got := e.renderMessage()
	if got != "{{bad" {
		t.Errorf("expected raw message, got %q", got)
	}
}

func TestError_Clone_StackBehavior(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	c1 := e.clone()
	if len(c1.stack) == 0 {
		t.Error("expected stack to be captured on first clone")
	}
	c2 := c1.clone()
	if len(c2.stack) == 0 {
		t.Error("expected stack to be preserved on re-clone")
	}
}

func TestError_WithCause(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	cause := errors.New("cause")
	c := e.WithCause(cause)
	if c.cause != cause {
		t.Errorf("expected cause to be set")
	}
	if e.cause != nil {
		t.Errorf("original must not be modified")
	}
}

func TestError_WithDetail(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	c := e.WithDetail("k", "v")
	v, ok := c.Detail("k")
	if !ok || v != "v" {
		t.Errorf("expected detail k=v, got %v %v", v, ok)
	}
	_, ok2 := e.Detail("k")
	if ok2 {
		t.Errorf("original must not have detail")
	}
}

func TestError_WithDetails(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	c := e.WithDetails(D{"a": 1, "b": 2})
	if v, _ := c.Detail("a"); v != 1 {
		t.Errorf("expected a=1, got %v", v)
	}
	if v, _ := c.Detail("b"); v != 2 {
		t.Errorf("expected b=2, got %v", v)
	}
}

func TestError_WithSeverity(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	c := e.WithSeverity(SeverityCritical)
	if c.severity != SeverityCritical {
		t.Errorf("expected SeverityCritical")
	}
	if e.severity != SeverityError {
		t.Errorf("original must not be modified")
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()
	cause := errors.New("inner")
	e := newTestError("C", "m").WithCause(cause)
	if e.Unwrap() != cause {
		t.Errorf("Unwrap should return cause")
	}
	e2 := newTestError("C", "m")
	if e2.Unwrap() != nil {
		t.Errorf("Unwrap should return nil when no cause")
	}
}

func TestError_Is_SameCode(t *testing.T) {
	t.Parallel()
	e1 := newError(Code("X"), Validation, SeverityError, "a")
	e2 := newError(Code("X"), NotFound, SeverityCritical, "b")
	if !e1.Is(e2) {
		t.Errorf("expected Is=true for same code")
	}
}

func TestError_Is_DifferentCode(t *testing.T) {
	t.Parallel()
	e1 := newError(Code("X"), Validation, SeverityError, "a")
	e2 := newError(Code("Y"), Validation, SeverityError, "b")
	if e1.Is(e2) {
		t.Errorf("expected Is=false for different code")
	}
}

func TestError_Is_NonError(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	if e.Is(errors.New("plain")) {
		t.Errorf("expected Is=false for non-Error target")
	}
}

func TestError_Getters(t *testing.T) {
	t.Parallel()
	e := newError(Code("G"), NotFound, SeverityWarning, "gm")
	c := e.WithCause(errors.New("gc"))
	if c.GetCode() != Code("G") {
		t.Errorf("GetCode mismatch")
	}
	if c.GetKind() != NotFound {
		t.Errorf("GetKind mismatch")
	}
	if c.GetSeverity() != SeverityWarning {
		t.Errorf("GetSeverity mismatch")
	}
	if c.GetMessage() != "gm" {
		t.Errorf("GetMessage mismatch")
	}
	if c.GetCause() == nil {
		t.Errorf("GetCause should not be nil")
	}
	if c.GetTimestamp().IsZero() {
		t.Errorf("GetTimestamp should not be zero")
	}
}

func TestError_Detail_NilMap(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	v, ok := e.Detail("any")
	if ok || v != nil {
		t.Errorf("expected nil/false for nil details map")
	}
}

func TestError_Details_NilMap(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	if e.Details() != nil {
		t.Errorf("expected nil for nil details map")
	}
}

func TestError_Details_Copy(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m").WithDetail("k", "v")
	d := e.Details()
	d["k"] = "modified"
	v, _ := e.Detail("k")
	if v == "modified" {
		t.Errorf("Details should return a copy")
	}
}

func TestError_StackTrace_Empty(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	if got := e.StackTrace(); got != "" {
		t.Errorf("expected empty stack trace, got %q", got)
	}
}

func TestError_StackTrace_NonEmpty(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m").WithCause(nil)
	if got := e.StackTrace(); got == "" {
		t.Errorf("expected non-empty stack trace after clone")
	}
}

func TestError_RootCause_NoCause(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	if e.RootCause() != nil {
		t.Errorf("expected nil root cause")
	}
}

func TestError_RootCause_ChainedCauses(t *testing.T) {
	t.Parallel()
	root := errors.New("root")
	mid := fmt.Errorf("mid: %w", root)
	e := newTestError("C", "m").WithCause(mid)
	if e.RootCause() != root {
		t.Errorf("expected root cause to be the deepest error")
	}
}

func TestError_RootCause_SingleCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("single")
	e := newTestError("C", "m").WithCause(cause)
	if e.RootCause() != cause {
		t.Errorf("expected root cause to be the single cause")
	}
}

func TestError_Format_VerbV_Verbose(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m").WithCause(errors.New("cause"))
	got := fmt.Sprintf("%+v", e)
	if !strings.Contains(got, "C: m: cause") {
		t.Errorf("expected error string in verbose, got %q", got)
	}
	if !strings.Contains(got, "Stack trace:") {
		t.Errorf("expected stack trace in verbose, got %q", got)
	}
	if !strings.Contains(got, "Caused by:") {
		t.Errorf("expected caused by in verbose, got %q", got)
	}
}

func TestError_Format_VerbV_NoStack(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	got := fmt.Sprintf("%+v", e)
	if strings.Contains(got, "Stack trace:") {
		t.Errorf("expected no stack trace, got %q", got)
	}
}

func TestError_Format_VerbV_Simple(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	got := fmt.Sprintf("%v", e)
	if got != "C: m" {
		t.Errorf("expected 'C: m', got %q", got)
	}
}

func TestError_Format_VerbS(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	got := fmt.Sprintf("%s", e)
	if got != "C: m" {
		t.Errorf("expected 'C: m', got %q", got)
	}
}

func TestError_Format_VerbQ(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m")
	got := fmt.Sprintf("%q", e)
	if got != `"C: m"` {
		t.Errorf("expected quoted, got %q", got)
	}
}

func TestError_Format_VerbV_NoCause_Verbose(t *testing.T) {
	t.Parallel()
	e := newTestError("C", "m").WithDetail("k", "v")
	got := fmt.Sprintf("%+v", e)
	if strings.Contains(got, "Caused by:") {
		t.Errorf("expected no Caused by, got %q", got)
	}
}

func TestError_MarshalJSON_WithCause(t *testing.T) {
	t.Parallel()
	e := newError(Code("J"), NotFound, SeverityWarning, "jmsg").
		WithCause(errors.New("jcause"))
	data, err := e.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	var je jsonError
	if err := json.Unmarshal(data, &je); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if je.Code != Code("J") {
		t.Errorf("expected code J, got %v", je.Code)
	}
	if je.Cause != "jcause" {
		t.Errorf("expected cause jcause, got %v", je.Cause)
	}
}

func TestError_MarshalJSON_NoCause(t *testing.T) {
	t.Parallel()
	e := newError(Code("J"), NotFound, SeverityWarning, "jmsg").
		WithDetail("x", 1)
	data, err := e.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	var je jsonError
	if err := json.Unmarshal(data, &je); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if je.Cause != "" {
		t.Errorf("expected empty cause, got %v", je.Cause)
	}
	if je.Kind != "not_found" {
		t.Errorf("expected not_found kind, got %v", je.Kind)
	}
	if je.Severity != "warning" {
		t.Errorf("expected warning severity, got %v", je.Severity)
	}
}
