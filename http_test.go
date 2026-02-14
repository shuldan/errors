package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestToHTTPStatus_AllKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		kind   Kind
		expect int
	}{
		{"Unknown", Unknown, http.StatusInternalServerError},
		{"Validation", Validation, http.StatusBadRequest},
		{"NotFound", NotFound, http.StatusNotFound},
		{"Conflict", Conflict, http.StatusConflict},
		{"DomainRule", DomainRule, http.StatusUnprocessableEntity},
		{"Authorization", Authorization, http.StatusForbidden},
		{"Authentication", Authentication, http.StatusUnauthorized},
		{"Infrastructure", Infrastructure, http.StatusServiceUnavailable},
		{"Internal", Internal, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := NewCode("X").Kind(tt.kind).New("m")
			c := e.clone()
			got := ToHTTPStatus(c)
			if got != tt.expect {
				t.Errorf("expected %d, got %d", tt.expect, got)
			}
		})
	}
}

func TestToHTTPStatus_NonErrorType(t *testing.T) {
	t.Parallel()
	got := ToHTTPStatus(errors.New("plain"))
	if got != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", got)
	}
}

func TestToHTTPStatus_UnknownKind(t *testing.T) {
	t.Parallel()
	e := &Error{kind: Kind(255)}
	c := e.clone()
	got := ToHTTPStatus(c)
	if got != http.StatusInternalServerError {
		t.Errorf("expected 500 for unmapped kind, got %d", got)
	}
}

func TestToPublicJSON_WithError(t *testing.T) {
	t.Parallel()
	e := NewCode("PJ").Kind(Validation).New("pub msg")
	c := e.WithDetail("field", "email")
	data := ToPublicJSON(c)
	var pe PublicError
	if err := json.Unmarshal(data, &pe); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if pe.Code != "PJ" {
		t.Errorf("expected PJ, got %v", pe.Code)
	}
	if pe.Message != "pub msg" {
		t.Errorf("expected 'pub msg', got %v", pe.Message)
	}
	if pe.Details["field"] != "email" {
		t.Errorf("expected field=email, got %v", pe.Details["field"])
	}
}

func TestToPublicJSON_NonError(t *testing.T) {
	t.Parallel()
	data := ToPublicJSON(errors.New("plain"))
	var pe PublicError
	if err := json.Unmarshal(data, &pe); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if pe.Code != string(Internal) {
		t.Errorf("expected internal code, got %v", pe.Code)
	}
	if pe.Message != "internal error" {
		t.Errorf("expected 'internal error', got %v", pe.Message)
	}
}

func TestToPublicError_WithError(t *testing.T) {
	t.Parallel()
	e := NewCode("PE").Kind(NotFound).New("not found msg")
	c := e.WithDetail("id", 42)
	pe := ToPublicError(c)
	if pe.Code != "PE" {
		t.Errorf("expected PE, got %v", pe.Code)
	}
	if pe.Message != "not found msg" {
		t.Errorf("expected 'not found msg', got %v", pe.Message)
	}
}

func TestToPublicError_NonError(t *testing.T) {
	t.Parallel()
	pe := ToPublicError(errors.New("something"))
	if pe.Code != string(Internal) {
		t.Errorf("expected internal, got %v", pe.Code)
	}
	if pe.Message != "internal error" {
		t.Errorf("expected 'internal error', got %v", pe.Message)
	}
}
