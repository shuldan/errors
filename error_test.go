package errors

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCode_New(t *testing.T) {
	t.Parallel()
	code := Code("TEST_001")
	msg := "test message"
	err := code.New(msg)

	if err.Code != code {
		t.Errorf("expected code %v, got %v", code, err.Code)
	}
	if err.Message != msg {
		t.Errorf("expected message %v, got %v", msg, err.Message)
	}
	if err.Details == nil {
		t.Error("expected Details map to be initialized")
	}
	if err.Stack == "" {
		t.Error("expected Stack to be set")
	}
	if err.Timestamp.IsZero() {
		t.Error("expected Timestamp to be set")
	}
}

func TestWithPrefix_WithSuffix(t *testing.T) {
	t.Parallel()
	prefix := "ERR"
	codeGen := WithPrefix(prefix)

	code := codeGen("CUSTOM")
	expected := Code("ERR_CUSTOM")
	if code != expected {
		t.Errorf("expected %v, got %v", expected, code)
	}
}

func TestWithPrefix_WithEmptySuffix(t *testing.T) {
	t.Parallel()
	prefix := "ERR"
	codeGen := WithPrefix(prefix)

	code := codeGen("")
	if !strings.HasPrefix(string(code), "ERR_") {
		t.Errorf("expected prefix ERR_, got %v", code)
	}
	if !strings.Contains(string(code), "0001") {
		t.Errorf("expected counter, got %v", code)
	}
}

func TestWithPrefix_WithoutSuffix(t *testing.T) {
	t.Parallel()
	prefix := "TEST"
	codeGen := WithPrefix(prefix)

	code1 := codeGen()
	code2 := codeGen()

	if code1 == code2 {
		t.Error("expected different codes for sequential calls")
	}
	if !strings.HasPrefix(string(code1), "TEST_") {
		t.Errorf("expected prefix TEST_, got %v", code1)
	}
}

func TestError_Error_SimpleMessage(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_001",
		Message: "simple error",
		Details: make(map[string]interface{}),
	}

	result := err.Error()
	expected := "ERR_001: simple error"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestError_Error_WithTemplate(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_002",
		Message: "error with {{.field}}",
		Details: map[string]interface{}{
			"field": "value",
		},
	}

	result := err.Error()
	expected := "ERR_002: error with value"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestError_Error_WithCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("root cause")
	err := &Error{
		Code:    "ERR_003",
		Message: "wrapped error",
		Details: make(map[string]interface{}),
		Cause:   cause,
	}

	result := err.Error()
	if !strings.Contains(result, "ERR_003") {
		t.Errorf("expected code in result, got %q", result)
	}
	if !strings.Contains(result, "wrapped error") {
		t.Errorf("expected message in result, got %q", result)
	}
	if !strings.Contains(result, "caused by") {
		t.Errorf("expected cause indicator, got %q", result)
	}
	if !strings.Contains(result, "root cause") {
		t.Errorf("expected cause message, got %q", result)
	}
}

func TestError_Error_InvalidTemplate(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_004",
		Message: "invalid {{.field",
		Details: make(map[string]interface{}),
	}

	result := err.Error()
	expected := "ERR_004: invalid {{.field"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestError_Error_TemplateExecuteError(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_005",
		Message: "error {{.missing}}",
		Details: make(map[string]interface{}),
	}

	result := err.Error()
	expected := "ERR_005: error <no value>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestError_Error_EmptyMessage(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_006",
		Message: "",
		Details: make(map[string]interface{}),
	}

	result := err.Error()
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestError_FormatSimpleMessage_NoCause(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_007",
		Message: "simple",
		Details: make(map[string]interface{}),
	}

	result := err.formatSimpleMessage()
	expected := "ERR_007: simple"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestError_FormatSimpleMessage_WithCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying")
	err := &Error{
		Code:    "ERR_008",
		Message: "wrapper",
		Details: make(map[string]interface{}),
		Cause:   cause,
	}

	result := err.formatSimpleMessage()
	if !strings.Contains(result, "caused by") {
		t.Errorf("expected cause indicator, got %q", result)
	}
}

func TestError_FormatSimpleMessage_EmptyMessage(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_009",
		Message: "",
		Details: make(map[string]interface{}),
	}

	result := err.formatSimpleMessage()
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestError_WithCause(t *testing.T) {
	t.Parallel()
	original := Error{
		Code:    "ERR_010",
		Message: "original",
		Details: map[string]interface{}{"key": "value"},
	}
	cause := errors.New("new cause")

	result := original.WithCause(cause)

	if !errors.Is(cause, result.Cause) {
		t.Error("expected cause to be set")
	}
	if result.Code != original.Code {
		t.Error("expected code to be preserved")
	}
	if result.Message != original.Message {
		t.Error("expected message to be preserved")
	}
	if len(result.Details) != len(original.Details) {
		t.Error("expected details to be copied")
	}
}

func TestError_WithDetail(t *testing.T) {
	t.Parallel()
	original := &Error{
		Code:    "ERR_011",
		Message: "test",
		Details: map[string]interface{}{"existing": "value"},
	}

	result := original.WithDetail("new", 123)

	if result.Details["new"] != 123 {
		t.Error("expected new detail to be added")
	}
	if result.Details["existing"] != "value" {
		t.Error("expected existing details to be preserved")
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()
	cause := errors.New("wrapped")
	err := &Error{
		Code:    "ERR_012",
		Message: "wrapper",
		Details: make(map[string]interface{}),
		Cause:   cause,
	}

	result := err.Unwrap()
	if !errors.Is(result, cause) {
		t.Errorf("expected %v, got %v", cause, result)
	}
}

func TestError_Unwrap_NoCause(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "ERR_013",
		Message: "no cause",
		Details: make(map[string]interface{}),
	}

	result := err.Unwrap()
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestError_Is_SameCode(t *testing.T) {
	t.Parallel()
	err1 := &Error{Code: "ERR_014"}
	err2 := &Error{Code: "ERR_014"}

	if !err1.Is(err2) {
		t.Error("expected errors with same code to match")
	}
}

func TestError_Is_DifferentCode(t *testing.T) {
	t.Parallel()
	err1 := &Error{Code: "ERR_015"}
	err2 := &Error{Code: "ERR_016"}

	if err1.Is(err2) {
		t.Error("expected errors with different codes not to match")
	}
}

func TestError_Is_NotErrorType(t *testing.T) {
	t.Parallel()
	err1 := &Error{Code: "ERR_017"}
	err2 := errors.New("standard error")

	if err1.Is(err2) {
		t.Error("expected not to match with standard error")
	}
}

func TestGetStack(t *testing.T) {
	t.Parallel()
	stack := getStack()

	if stack == "" {
		t.Error("expected non-empty stack")
	}
	if !strings.Contains(stack, "goroutine") {
		t.Errorf("expected goroutine in stack, got %s", stack)
	}
}

func TestError_Fields(t *testing.T) {
	t.Parallel()
	now := time.Now()
	err := &Error{
		Code:      "TEST",
		Message:   "msg",
		Details:   map[string]interface{}{"k": "v"},
		Cause:     errors.New("cause"),
		Stack:     "stack",
		Timestamp: now,
	}

	if err.Code != "TEST" {
		t.Errorf("expected Code TEST, got %v", err.Code)
	}
	if err.Message != "msg" {
		t.Errorf("expected Message msg, got %v", err.Message)
	}
	if err.Timestamp != now {
		t.Error("expected Timestamp to match")
	}
}

func TestWithPrefix_Concurrent(t *testing.T) {
	t.Parallel()
	codeGen := WithPrefix("CONCURRENT")
	done := make(chan Code, 100)

	for i := 0; i < 100; i++ {
		go func() {
			done <- codeGen()
		}()
	}

	seen := make(map[Code]bool)
	for i := 0; i < 100; i++ {
		code := <-done
		if seen[code] {
			t.Errorf("duplicate code: %v", code)
		}
		seen[code] = true
	}
}

func TestError_WithDetail_Isolation(t *testing.T) {
	t.Parallel()
	original := &Error{
		Code:    "ISO",
		Message: "test",
		Details: map[string]interface{}{"a": 1},
	}

	modified := original.WithDetail("b", 2)
	original.Details["c"] = 3

	if _, exists := modified.Details["c"]; exists {
		t.Error("modification of original should not affect clone")
	}
}

func TestError_Error_WithComplexTemplate(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "COMPLEX",
		Message: "user {{.user}} failed: {{.reason}}",
		Details: map[string]interface{}{
			"user":   "admin",
			"reason": "timeout",
		},
	}

	result := err.Error()
	expected := "COMPLEX: user admin failed: timeout"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestError_Error_Panic_Recovery(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "PANIC_TEST",
		Message: "{{.field}}",
		Details: map[string]interface{}{
			"field": func() {},
		},
	}

	result := err.Error()
	if result == "" {
		t.Error("expected non-empty result after panic recovery")
	}
}

func TestError_WithCause_ChainedCalls(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "CHAIN",
		Message: "msg",
		Details: make(map[string]interface{}),
	}

	cause1 := errors.New("cause1")
	cause2 := errors.New("cause2")

	result := err.WithCause(cause1).WithCause(cause2)

	if !errors.Is(cause2, result.Cause) {
		t.Error("expected latest cause to be set")
	}
}

func TestError_WithDetail_OverwriteValue(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "OVERWRITE",
		Message: "test",
		Details: map[string]interface{}{"key": "old"},
	}

	result := err.WithDetail("key", "new")

	if result.Details["key"] != "new" {
		t.Errorf("expected 'new', got %v", result.Details["key"])
	}
}

func TestError_WithDetail_MultipleDetails(t *testing.T) {
	t.Parallel()
	err := &Error{
		Code:    "MULTI",
		Message: "test",
		Details: map[string]interface{}{"a": 1},
	}

	result := err.WithDetail("b", 2).WithDetail("c", 3)

	if result.Details["a"] != 1 {
		t.Error("expected detail 'a' to be preserved")
	}
	if result.Details["b"] != 2 {
		t.Error("expected detail 'b' to be added")
	}
	if result.Details["c"] != 3 {
		t.Error("expected detail 'c' to be added")
	}
}
