package common

import (
	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// ViolationBuilder provides a fluent interface for creating Violation objects
type ViolationBuilder struct {
	v analyzer.Violation
}

// NewViolationBuilder creates a new ViolationBuilder
func NewViolationBuilder() *ViolationBuilder {
	return &ViolationBuilder{}
}

// WithFile sets the file path
func (vb *ViolationBuilder) WithFile(path string) *ViolationBuilder {
	vb.v.File = path
	return vb
}

// WithLine sets the line number
func (vb *ViolationBuilder) WithLine(line int) *ViolationBuilder {
	vb.v.Line = line
	return vb
}

// WithRule sets the rule identifier
func (vb *ViolationBuilder) WithRule(rule string) *ViolationBuilder {
	vb.v.Rule = rule
	return vb
}

// WithSeverity sets the severity
func (vb *ViolationBuilder) WithSeverity(sev spec.Severity) *ViolationBuilder {
	vb.v.Severity = sev
	return vb
}

// WithExcerpt sets the code excerpt
func (vb *ViolationBuilder) WithExcerpt(excerpt string) *ViolationBuilder {
	vb.v.Excerpt = excerpt
	return vb
}

// WithSuggestion sets the suggestion message
func (vb *ViolationBuilder) WithSuggestion(suggestion string) *ViolationBuilder {
	vb.v.Suggestion = suggestion
	return vb
}

// Build returns the constructed Violation
func (vb *ViolationBuilder) Build() analyzer.Violation {
	return vb.v
}

// BuildPtr returns a pointer to the constructed Violation
func (vb *ViolationBuilder) BuildPtr() *analyzer.Violation {
	return &vb.v
}

// Reset resets the builder for reuse
func (vb *ViolationBuilder) Reset() {
	vb.v = analyzer.Violation{}
}

// ViolationFromForbidden creates a violation from a forbidden rule
func ViolationFromForbidden(file string, line int, rule spec.ForbiddenRule, excerpt string) analyzer.Violation {
	return analyzer.Violation{
		File:       file,
		Line:       line,
		Rule:       "forbidden",
		Severity:   spec.SeverityError,
		Excerpt:    excerpt,
		Suggestion: rule.Message,
	}
}

// ViolationFromNaming creates a violation from a naming rule
func ViolationFromNaming(file string, rule string, expected string, filename string) analyzer.Violation {
	return analyzer.Violation{
		File:       file,
		Line:       0,
		Rule:       "naming." + rule,
		Severity:   spec.SeverityWarning,
		Excerpt:    filename,
		Suggestion: "Expected: " + expected,
	}
}

// ViolationFromLimit creates a violation from a limit rule
func ViolationFromLimit(file string, rule string, current int, max int) analyzer.Violation {
	return analyzer.Violation{
		File:       file,
		Line:       0,
		Rule:       "limits." + rule,
		Severity:   spec.SeverityError,
		Excerpt:    "Current: " + formatInt(current),
		Suggestion: "Max allowed: " + formatInt(max),
	}
}

// ViolationFromRequired creates a violation from a required rule
func ViolationFromRequired(file string, line int, rule spec.RequiredRule, excerpt string) analyzer.Violation {
	return analyzer.Violation{
		File:       file,
		Line:       line,
		Rule:       "required",
		Severity:   spec.SeverityError,
		Excerpt:    excerpt,
		Suggestion: rule.Message,
	}
}

// ViolationFromArchitecture creates a violation from an architecture rule
func ViolationFromArchitecture(file string, line int, rule spec.ArchitectureRule, excerpt string) analyzer.Violation {
	return analyzer.Violation{
		File:       file,
		Line:       line,
		Rule:       "architecture",
		Severity:   spec.SeverityError,
		Excerpt:    excerpt,
		Suggestion: rule.Description,
	}
}

func formatInt(i int) string {
	return string(rune('0' + i%10))
}
