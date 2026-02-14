package errors

import (
	"encoding/json"
	"net/http"
)

var httpStatusMap = map[Kind]int{
	Unknown:        http.StatusInternalServerError,
	Validation:     http.StatusBadRequest,
	NotFound:       http.StatusNotFound,
	Conflict:       http.StatusConflict,
	DomainRule:     http.StatusUnprocessableEntity,
	Authorization:  http.StatusForbidden,
	Authentication: http.StatusUnauthorized,
	Infrastructure: http.StatusServiceUnavailable,
	Internal:       http.StatusInternalServerError,
}

func ToHTTPStatus(err error) int {
	var e *Error
	if As(err, &e) {
		if status, ok := httpStatusMap[e.kind]; ok {
			return status
		}
	}
	return http.StatusInternalServerError
}

type PublicError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func ToPublicJSON(err error) []byte {
	var e *Error
	if As(err, &e) {
		pe := PublicError{
			Code:    string(e.code),
			Message: e.renderMessage(),
			Details: e.Details(),
		}
		data, _ := json.Marshal(pe)
		return data
	}

	pe := PublicError{
		Code:    string(Internal),
		Message: "internal error",
	}
	data, _ := json.Marshal(pe)
	return data
}

func ToPublicError(err error) PublicError {
	var e *Error
	if As(err, &e) {
		return PublicError{
			Code:    string(e.code),
			Message: e.renderMessage(),
			Details: e.Details(),
		}
	}
	return PublicError{
		Code:    string(Internal),
		Message: "internal error",
	}
}
