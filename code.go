package errors

import "fmt"

type Code string

func (c Code) String() string {
	return string(c)
}

type D map[string]any

type CodeBuilder struct {
	code     Code
	kind     Kind
	severity Severity
}

func NewCode(code string) *CodeBuilder {
	return &CodeBuilder{code: Code(code)}
}

func WithPrefix(prefix string) func(name string) *CodeBuilder {
	return func(name string) *CodeBuilder {
		return &CodeBuilder{
			code: Code(fmt.Sprintf("%s_%s", prefix, name)),
		}
	}
}

func (b *CodeBuilder) Kind(k Kind) *CodeBuilder {
	b.kind = k
	return b
}

func (b *CodeBuilder) Severity(s Severity) *CodeBuilder {
	b.severity = s
	return b
}

func (b *CodeBuilder) New(msg string) *Error {
	return newError(b.code, b.kind, b.severity, msg)
}
