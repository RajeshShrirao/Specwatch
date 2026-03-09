package strategy

import (
	"context"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// ForbiddenStrategy checks for forbidden patterns in code
type ForbiddenStrategy struct {
	BaseStrategy
	name     string
	category string
}

// NewForbiddenStrategy creates a new ForbiddenStrategy
func NewForbiddenStrategy() *ForbiddenStrategy {
	return &ForbiddenStrategy{
		name:     "forbidden",
		category: "forbidden",
	}
}

// Name returns the strategy name
func (s *ForbiddenStrategy) Name() string {
	return s.name
}

// Category returns the strategy category
func (s *ForbiddenStrategy) Category() string {
	return s.category
}

// CanCheck determines if this strategy can handle the given rule
func (s *ForbiddenStrategy) CanCheck(rule interface{}) bool {
	_, ok := rule.([]spec.ForbiddenRule)
	return ok
}

// Check performs the forbidden pattern analysis
func (s *ForbiddenStrategy) Check(ctx context.Context, params CheckParams) []analyzer.Violation {
	var violations []analyzer.Violation

	rules, ok := params.Rule.([]spec.ForbiddenRule)
	if !ok {
		return violations
	}

	if len(rules) == 0 {
		return violations
	}

	// Get content if not provided
	content := params.Content
	if len(content) == 0 && params.Cache != nil {
		var err error
		content, err = s.GetFileContent(params.FilePath, params.Cache)
		if err != nil {
			return violations
		}
	}

	lineNum := 0
	for _, line := range content {
		lineNum++

		for _, rule := range rules {
			// Check pattern match
			if rule.Pattern != "" {
				if s.matchPattern(line, rule.Pattern, params.Compiled) {
					violations = append(violations, analyzer.Violation{
						File:       params.FilePath,
						Line:       lineNum,
						Rule:       "forbidden.pattern",
						Severity:   spec.SeverityError,
						Excerpt:    strings.TrimSpace(line),
						Suggestion: rule.Message,
					})
				}
			}

			// Check import match
			if rule.Import != "" {
				if s.matchImport(line, rule.Import) {
					violations = append(violations, analyzer.Violation{
						File:       params.FilePath,
						Line:       lineNum,
						Rule:       "forbidden.import",
						Severity:   spec.SeverityError,
						Excerpt:    strings.TrimSpace(line),
						Suggestion: rule.Message,
					})
				}
			}
		}
	}

	return violations
}

// matchPattern checks if line matches the pattern using pre-compiled regex
func (s *ForbiddenStrategy) matchPattern(line, pattern string, compiled map[string]*regexp.Regexp) bool {
	// Use pre-compiled pattern if available
	if re, ok := compiled[pattern]; ok {
		return re.MatchString(line)
	}
	// Fall back to simple string contains
	return strings.Contains(line, pattern)
}

// matchImport checks if line contains forbidden import
func (s *ForbiddenStrategy) matchImport(line, importPath string) bool {
	return (strings.HasPrefix(strings.TrimSpace(line), "import") ||
		strings.Contains(line, "require(") ||
		strings.Contains(line, "from")) &&
		strings.Contains(line, importPath)
}

// Ensure ForbiddenStrategy implements RuleStrategy
var _ RuleStrategy = (*ForbiddenStrategy)(nil)
