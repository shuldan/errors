package errors

type Severity uint8

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityCritical
)

var severityNames = map[Severity]string{
	SeverityError:    "error",
	SeverityWarning:  "warning",
	SeverityCritical: "critical",
}

func (s Severity) String() string {
	if name, ok := severityNames[s]; ok {
		return name
	}
	return "error"
}
