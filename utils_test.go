package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestIs_BothNil(t *testing.T) {
	t.Parallel()
	result := Is(nil, nil)
	if result {
		t.Error("expected false when both errors are nil")
	}
}

func TestIs_FirstNil(t *testing.T) {
	t.Parallel()
	target := errors.New("target")
	result := Is(nil, target)
	if result {
		t.Error("expected false when first error is nil")
	}
}

func TestIs_SecondNil(t *testing.T) {
	t.Parallel()
	err := errors.New("error")
	result := Is(err, nil)
	if result {
		t.Error("expected false when target is nil")
	}
}

func TestIs_Matching(t *testing.T) {
	t.Parallel()
	target := errors.New("sentinel")
	err := fmt.Errorf("wrapped: %w", target)

	result := Is(err, target)
	if !result {
		t.Error("expected true for matching errors")
	}
}

func TestIs_NotMatching(t *testing.T) {
	t.Parallel()
	err := errors.New("error1")
	target := errors.New("error2")

	result := Is(err, target)
	if result {
		t.Error("expected false for non-matching errors")
	}
}

func TestIs_WithCustomError(t *testing.T) {
	t.Parallel()
	baseErr := Code("ERR_001").New("base")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	result := Is(wrappedErr, baseErr)
	if !result {
		t.Error("expected true for wrapped custom error")
	}
}

func TestAs_NilError(t *testing.T) {
	t.Parallel()
	var target *Error

	result := As(nil, &target)
	if result {
		t.Error("expected false when error is nil")
	}
}

func TestAs_Matching(t *testing.T) {
	t.Parallel()
	err := &Error{Code: "TEST", Message: "test"}
	var target *Error

	result := As(err, &target)
	if !result {
		t.Error("expected true for matching type")
	}
	if target == nil {
		t.Error("expected target to be set")
	}
	if target.Code != "TEST" {
		t.Errorf("expected code TEST, got %v", target.Code)
	}
}

func TestAs_WrappedMatching(t *testing.T) {
	t.Parallel()
	baseErr := &Error{Code: "WRAPPED", Message: "msg"}
	wrapped := fmt.Errorf("wrapper: %w", baseErr)
	var target *Error

	result := As(wrapped, &target)
	if !result {
		t.Error("expected true for wrapped matching type")
	}
	if target.Code != "WRAPPED" {
		t.Errorf("expected code WRAPPED, got %v", target.Code)
	}
}

func TestAs_NotMatching(t *testing.T) {
	t.Parallel()
	err := errors.New("standard error")
	var target *Error

	result := As(err, &target)
	if result {
		t.Error("expected false for non-matching type")
	}
}

func TestUnwrap_Wrapped(t *testing.T) {
	t.Parallel()
	base := errors.New("base")
	wrapped := fmt.Errorf("wrapped: %w", base)

	result := Unwrap(wrapped)
	if !errors.Is(result, base) {
		t.Errorf("expected %v, got %v", base, result)
	}
}

func TestUnwrap_NotWrapped(t *testing.T) {
	t.Parallel()
	err := errors.New("not wrapped")

	result := Unwrap(err)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestUnwrap_Nil(t *testing.T) {
	t.Parallel()
	result := Unwrap(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestUnwrap_CustomError(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	err := &Error{
		Code:    "ERR",
		Message: "msg",
		Cause:   cause,
		Details: make(map[string]interface{}),
	}

	result := Unwrap(err)
	if !errors.Is(result, cause) {
		t.Errorf("expected %v, got %v", cause, result)
	}
}

func TestJoin_NoErrors(t *testing.T) {
	t.Parallel()
	result := Join()
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestJoin_SingleError(t *testing.T) {
	t.Parallel()
	err := errors.New("single")
	result := Join(err)

	if result == nil {
		t.Error("expected non-nil result")
	}
	if result != nil && result.Error() != "single" {
		t.Errorf("expected 'single', got %v", result.Error())
	}
}

func TestJoin_MultipleErrors(t *testing.T) {
	t.Parallel()
	err1 := errors.New("error1")
	err2 := errors.New("error2")
	result := Join(err1, err2)

	if result == nil {
		t.Error("expected non-nil result")
	}
	if result != nil {
		str := result.Error()
		if str == "" {
			t.Error("expected non-empty error message")
		}
	}
}

func TestJoin_WithNils(t *testing.T) {
	t.Parallel()
	err := errors.New("real error")
	result := Join(nil, err, nil)

	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestGetErrorCode_CustomError(t *testing.T) {
	t.Parallel()
	err := &Error{Code: "CUSTOM_ERR"}

	code := GetErrorCode(err)
	if code != "CUSTOM_ERR" {
		t.Errorf("expected CUSTOM_ERR, got %v", code)
	}
}

func TestGetErrorCode_WrappedCustomError(t *testing.T) {
	t.Parallel()
	baseErr := &Error{Code: "BASE_ERR"}
	wrapped := fmt.Errorf("wrapped: %w", baseErr)

	code := GetErrorCode(wrapped)
	if code != "BASE_ERR" {
		t.Errorf("expected BASE_ERR, got %v", code)
	}
}

func TestGetErrorCode_StandardError(t *testing.T) {
	t.Parallel()
	err := errors.New("standard")

	code := GetErrorCode(err)
	if code != "" {
		t.Errorf("expected empty code, got %v", code)
	}
}

func TestGetErrorCode_Nil(t *testing.T) {
	t.Parallel()
	code := GetErrorCode(nil)
	if code != "" {
		t.Errorf("expected empty code, got %v", code)
	}
}
