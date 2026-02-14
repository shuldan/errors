package errors

import (
	"errors"
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As[T error](err error, target *T) bool {
	return errors.As(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}

func Wrap(cause error, template *Error) *Error {
	return template.WithCause(cause)
}

func GetCode(err error) Code {
	var e *Error
	if As(err, &e) {
		return e.code
	}
	return ""
}

func GetKind(err error) Kind {
	var e *Error
	if As(err, &e) {
		return e.kind
	}
	return Unknown
}

func GetSeverity(err error) Severity {
	var e *Error
	if As(err, &e) {
		return e.severity
	}
	return SeverityError
}
