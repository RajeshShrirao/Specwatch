package strategy

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// ArchitectureStrategy checks architectural constraints
type ArchitectureStrategy struct {
	BaseStrategy
	name     string
	category string
}

// NewArchitectureStrategy creates a new ArchitectureStrategy
func NewArchitectureStrategy() *ArchitectureStrategy {
	return &ArchitectureStrategy{
		name:     "architecture",
		category: "architecture",
	}
}

// Name returns the strategy name
func (s *ArchitectureStrategy) Name() string {
	return s.name
}

// Category returns the strategy category
func (s *ArchitectureStrategy) Category() string {
	return s.category
}

// CanCheck determines if this strategy can handle the given rule
func (s *ArchitectureStrategy) CanCheck(rule interface{}) bool {
	_, ok := rule.([]spec.ArchitectureRule)
	return ok
}

// Check performs the architectural constraints analysis
func (s *ArchitectureStrategy) Check(ctx context.Context, params CheckParams) []analyzer.Violation {
	var violations []analyzer.Violation

	rules, ok := params.Rule.([]spec.ArchitectureRule)
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

	// Check each architecture rule
	for _, rule := range rules {
		violations = append(violations, s.checkRule(params.FilePath, content, rule)...)
	}

	return violations
}

// checkRule checks a specific architectural rule
func (s *ArchitectureStrategy) checkRule(path string, content []string, rule spec.ArchitectureRule) []analyzer.Violation {
	var violations []analyzer.Violation

	// Heuristic: "no direct db calls outside src/lib/db"
	// If file is NOT in src/lib/db, check for forbidden imports or patterns
	absPath, _ := filepath.Abs(path)
	isDbFile := strings.Contains(absPath, "src/lib/db")

	if !isDbFile {
		// Check for direct DB-related keywords or imports
		if strings.Contains(rule.Description, "no direct db calls") {
			for lineNum, line := range content {
				if strings.Contains(line, " prisma.") ||
					strings.Contains(line, " mongoose.") ||
					strings.Contains(line, " sequelize.") ||
					strings.Contains(line, "db.") {
					violations = append(violations, analyzer.Violation{
						File:       path,
						Line:       lineNum + 1,
						Rule:       "architecture.no_direct_db",
						Severity:   spec.SeverityError,
						Excerpt:    strings.TrimSpace(line),
						Suggestion: "Direct database calls are only allowed in src/lib/db",
					})
				}
			}
		}
	}

	return violations
}

// Ensure ArchitectureStrategy implements RuleStrategy
var _ RuleStrategy = (*ArchitectureStrategy)(nil)
