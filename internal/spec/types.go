package spec

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type RuleSet struct {
	Stack        StackRules
	Structure    []StructureRule
	Naming       NamingRules
	Forbidden    []ForbiddenRule
	Required     []RequiredRule
	Architecture []ArchitectureRule
	Limits       LimitRules
}

type StackRules struct {
	Language  string
	Framework string
	Styling   string
	Runtime   string
}

type StructureRule struct {
	Type     string // e.g., "components", "api routes"
	Pattern  string // e.g., "src/components/**"
}

type NamingRules struct {
	Components string // e.g., "PascalCase"
	Functions  string // e.g., "camelCase"
	Files      string // e.g., "kebab-case"
	Constants  string // e.g., "SCREAMING_SNAKE_CASE"
	Interfaces string // e.g., "PascalCase prefixed with I"
}

type ForbiddenRule struct {
	Pattern string
	Import  string
	Message string
}

type RequiredRule struct {
	Target  string // e.g., "async functions"
	Check   string // e.g., "try/catch"
	Message string
}

type ArchitectureRule struct {
	ID                string   // e.g., "architecture.no_direct_db"
	Description       string   // e.g., "no direct db calls outside src/lib/db"
	ExemptPathPattern string   // e.g., "src/lib/db" - path pattern that is exempt from this rule
	Patterns          []string // explicit patterns to match (regex, case-insensitive)
}

type LimitRules struct {
	MaxFileLines     int
	MaxFunctionLines int
	MaxImports       int
	MaxProps         int
}
