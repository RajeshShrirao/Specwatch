package strategy

import (
	"context"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// RequiredStrategy checks for required patterns in code
type RequiredStrategy struct {
	BaseStrategy
	name     string
	category string
}

// NewRequiredStrategy creates a new RequiredStrategy
func NewRequiredStrategy() *RequiredStrategy {
	return &RequiredStrategy{
		name:     "required",
		category: "required",
	}
}

// Name returns the strategy name
func (s *RequiredStrategy) Name() string {
	return s.name
}

// Category returns the strategy category
func (s *RequiredStrategy) Category() string {
	return s.category
}

// CanCheck determines if this strategy can handle the given rule
func (s *RequiredStrategy) CanCheck(rule interface{}) bool {
	_, ok := rule.([]spec.RequiredRule)
	return ok
}

// Check performs the required patterns analysis
func (s *RequiredStrategy) Check(ctx context.Context, params CheckParams) []analyzer.Violation {
	var violations []analyzer.Violation

	rules, ok := params.Rule.([]spec.RequiredRule)
	if !ok || len(rules) == 0 {
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

	// Check for required try/catch in async functions
	for _, rule := range rules {
		if rule.Target == "async functions" && rule.Check == "try/catch" {
			violations = append(violations, s.checkTryCatch(params.FilePath, content, rule)...)
		}
	}

	return violations
}

// checkTryCatch checks if async functions have try/catch blocks
func (s *RequiredStrategy) checkTryCatch(path string, content []string, rule spec.RequiredRule) []analyzer.Violation {
	var violations []analyzer.Violation

	for i, line := range content {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Check for async function
		if strings.Contains(line, "async ") && (strings.Contains(line, "function") || strings.Contains(line, "=") || strings.Contains(line, "(")) {
			// Found a potential async function
			// Check the next ~10 lines or 400 characters for "try"
			foundTry := false
			lookAheadLines := 10
			for j := i; j < i+lookAheadLines && j < len(content); j++ {
				if strings.Contains(content[j], "try {") || strings.Contains(content[j], "try{") {
					foundTry = true
					break
				}
			}

			if !foundTry {
				violations = append(violations, analyzer.Violation{
					File:       path,
					Line:       i + 1,
					Rule:       "required.try_catch",
					Severity:   spec.SeverityError,
					Excerpt:    strings.TrimSpace(line),
					Suggestion: rule.Message,
				})
			}
		}
	}

	return violations
}

// Ensure RequiredStrategy implements RuleStrategy
var _ RuleStrategy = (*RequiredStrategy)(nil)
