package errors

type Kind uint8

const (
	Unknown Kind = iota
	Validation
	NotFound
	Conflict
	DomainRule
	Authorization
	Authentication
	Infrastructure
	Internal
)

var kindNames = map[Kind]string{
	Unknown:        "unknown",
	Validation:     "validation",
	NotFound:       "not_found",
	Conflict:       "conflict",
	DomainRule:     "domain_rule",
	Authorization:  "authorization",
	Authentication: "authentication",
	Infrastructure: "infrastructure",
	Internal:       "internal",
}

func (k Kind) String() string {
	if name, ok := kindNames[k]; ok {
		return name
	}
	return "unknown"
}
