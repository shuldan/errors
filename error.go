package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"text/template"
	"time"
)

type Error struct {
	code      Code
	kind      Kind
	severity  Severity
	message   string
	tmpl      *template.Template
	details   map[string]any
	cause     error
	stack     []uintptr
	timestamp time.Time
}

func newError(code Code, kind Kind, severity Severity, msg string) *Error {
	e := &Error{
		code:     code,
		kind:     kind,
		severity: severity,
		message:  msg,
	}

	parsed, err := template.New("").Parse(msg)
	if err == nil {
		e.tmpl = parsed
	}

	return e
}

func (e *Error) Error() string {
	msg := e.renderMessage()
	if msg == "" {
		if e.cause != nil {
			return fmt.Sprintf("%s: %v", e.code, e.cause)
		}
		return string(e.code)
	}

	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.code, msg, e.cause)
	}

	return fmt.Sprintf("%s: %s", e.code, msg)
}

func (e *Error) renderMessage() string {
	if e.tmpl == nil {
		return e.message
	}

	var buf bytes.Buffer
	if err := e.tmpl.Execute(&buf, e.details); err != nil {
		return e.message
	}

	return buf.String()
}

func (e *Error) clone() *Error {
	c := &Error{
		code:      e.code,
		kind:      e.kind,
		severity:  e.severity,
		message:   e.message,
		tmpl:      e.tmpl,
		details:   make(map[string]any, len(e.details)+1),
		cause:     e.cause,
		timestamp: time.Now(),
	}

	for k, v := range e.details {
		c.details[k] = v
	}

	if len(e.stack) == 0 {
		c.stack = captureStack()
	} else {
		c.stack = e.stack
	}

	return c
}

func (e *Error) WithCause(err error) *Error {
	c := e.clone()
	c.cause = err
	return c
}

func (e *Error) WithDetail(key string, value any) *Error {
	c := e.clone()
	c.details[key] = value
	return c
}

func (e *Error) WithDetails(d D) *Error {
	c := e.clone()
	for k, v := range d {
		c.details[k] = v
	}
	return c
}

func (e *Error) WithSeverity(s Severity) *Error {
	c := e.clone()
	c.severity = s
	return c
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Is(target error) bool {
	var targetErr *Error
	if errors.As(target, &targetErr) {
		return e.code == targetErr.code
	}
	return false
}

func (e *Error) GetCode() Code {
	return e.code
}

func (e *Error) GetKind() Kind {
	return e.kind
}

func (e *Error) GetSeverity() Severity {
	return e.severity
}

func (e *Error) GetMessage() string {
	return e.message
}

func (e *Error) GetTimestamp() time.Time {
	return e.timestamp
}

func (e *Error) GetCause() error {
	return e.cause
}

func (e *Error) Detail(key string) (any, bool) {
	if e.details == nil {
		return nil, false
	}
	v, ok := e.details[key]
	return v, ok
}

func (e *Error) Details() map[string]any {
	if e.details == nil {
		return nil
	}
	cp := make(map[string]any, len(e.details))
	for k, v := range e.details {
		cp[k] = v
	}
	return cp
}

func (e *Error) StackTrace() string {
	if len(e.stack) == 0 {
		return ""
	}
	return formatStack(e.stack)
}

func (e *Error) RootCause() error {
	if e.cause == nil {
		return nil
	}

	root := e.cause
	for {
		unwrapped := errors.Unwrap(root)
		if unwrapped == nil {
			return root
		}
		root = unwrapped
	}
}

func (e *Error) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			_, _ = fmt.Fprint(f, e.Error())
			if st := e.StackTrace(); st != "" {
				_, _ = fmt.Fprintf(f, "\n\nStack trace:\n%s", st)
			}
			if e.cause != nil {
				_, _ = fmt.Fprintf(f, "\nCaused by: %+v", e.cause)
			}
			return
		}
		_, _ = fmt.Fprint(f, e.Error())
	case 's':
		_, _ = fmt.Fprint(f, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(f, "%q", e.Error())
	}
}

type jsonError struct {
	Code      Code           `json:"code"`
	Kind      string         `json:"kind"`
	Severity  string         `json:"severity"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	Cause     string         `json:"cause,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e *Error) MarshalJSON() ([]byte, error) {
	je := jsonError{
		Code:      e.code,
		Kind:      e.kind.String(),
		Severity:  e.severity.String(),
		Message:   e.renderMessage(),
		Details:   e.details,
		Timestamp: e.timestamp,
	}

	if e.cause != nil {
		je.Cause = e.cause.Error()
	}

	return json.Marshal(je)
}

func captureStack() []uintptr {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(4, pcs)
	return pcs[:n]
}

func formatStack(pcs []uintptr) string {
	var buf bytes.Buffer
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&buf, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return buf.String()
}
