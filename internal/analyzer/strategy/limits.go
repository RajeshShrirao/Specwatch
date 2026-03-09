package strategy

import (
	"context"
	"fmt"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// LimitsStrategy checks file and code limits
type LimitsStrategy struct {
	BaseStrategy
	name     string
	category string
}

// NewLimitsStrategy creates a new LimitsStrategy
func NewLimitsStrategy() *LimitsStrategy {
	return &LimitsStrategy{
		name:     "limits",
		category: "limits",
	}
}

// Name returns the strategy name
func (s *LimitsStrategy) Name() string {
	return s.name
}

// Category returns the strategy category
func (s *LimitsStrategy) Category() string {
	return s.category
}

// CanCheck determines if this strategy can handle the given rule
func (s *LimitsStrategy) CanCheck(rule interface{}) bool {
	_, ok := rule.(spec.LimitRules)
	return ok
}

// Check performs the limits analysis
func (s *LimitsStrategy) Check(ctx context.Context, params CheckParams) []analyzer.Violation {
	var violations []analyzer.Violation

	rules, ok := params.Rule.(spec.LimitRules)
	if !ok {
		return violations
	}

	// Skip if no limits defined
	if rules.MaxFileLines <= 0 && rules.MaxImports <= 0 {
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

	// Count lines and imports
	lineCount := 0
	importCount := 0
	for _, line := range content {
		lineCount++
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") {
			importCount++
		}
	}

	// Check line limit
	if rules.MaxFileLines > 0 && lineCount > rules.MaxFileLines {
		violations = append(violations, analyzer.Violation{
			File:       params.FilePath,
			Line:       lineCount,
			Rule:       "limits.file_lines",
			Severity:   spec.SeverityError,
			Excerpt:    fmt.Sprintf("File has %d lines", lineCount),
			Suggestion: fmt.Sprintf("Max lines allowed: %d", rules.MaxFileLines),
		})
	}

	// Check import limit
	if rules.MaxImports > 0 && importCount > rules.MaxImports {
		violations = append(violations, analyzer.Violation{
			File:       params.FilePath,
			Line:       0,
			Rule:       "limits.imports",
			Severity:   spec.SeverityWarning,
			Excerpt:    fmt.Sprintf("File has %d imports", importCount),
			Suggestion: fmt.Sprintf("Max imports allowed: %d", rules.MaxImports),
		})
	}

	return violations
}

// Ensure LimitsStrategy implements RuleStrategy
var _ RuleStrategy = (*LimitsStrategy)(nil)
