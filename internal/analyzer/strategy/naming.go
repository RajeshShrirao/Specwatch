package strategy

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// NamingStrategy checks file and identifier naming conventions
type NamingStrategy struct {
	BaseStrategy
	name     string
	category string
}

// NewNamingStrategy creates a new NamingStrategy
func NewNamingStrategy() *NamingStrategy {
	return &NamingStrategy{
		name:     "naming",
		category: "naming",
	}
}

// Name returns the strategy name
func (s *NamingStrategy) Name() string {
	return s.name
}

// Category returns the strategy category
func (s *NamingStrategy) Category() string {
	return s.category
}

// CanCheck determines if this strategy can handle the given rule
func (s *NamingStrategy) CanCheck(rule interface{}) bool {
	_, ok := rule.(spec.NamingRules)
	return ok
}

// Check performs the naming convention analysis
func (s *NamingStrategy) Check(ctx context.Context, params CheckParams) []analyzer.Violation {
	var violations []analyzer.Violation

	rules, ok := params.Rule.(spec.NamingRules)
	if !ok {
		return violations
	}

	filename := filepath.Base(params.FilePath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Check file naming
	if rules.Files != "" {
		if violation := s.checkFileNaming(params.FilePath, nameWithoutExt, rules.Files); violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// checkFileNaming checks if the filename follows the naming convention
func (s *NamingStrategy) checkFileNaming(path, nameWithoutExt, convention string) *analyzer.Violation {
	// Relax for common special files like README.md or spec.md
	lowerName := strings.ToLower(nameWithoutExt)
	if strings.Contains(lowerName, "readme") || lowerName == "spec" || lowerName == "license" {
		return nil
	}

	switch convention {
	case "kebab-case":
		if !isKebabCase(nameWithoutExt) {
			return &analyzer.Violation{
				File:       path,
				Line:       0,
				Rule:       "naming.files",
				Severity:   spec.SeverityWarning,
				Excerpt:    filepath.Base(path),
				Suggestion: "File name should be kebab-case",
			}
		}
	case "camelCase":
		if !isCamelCase(nameWithoutExt) {
			return &analyzer.Violation{
				File:       path,
				Line:       0,
				Rule:       "naming.files",
				Severity:   spec.SeverityWarning,
				Excerpt:    filepath.Base(path),
				Suggestion: "File name should be camelCase",
			}
		}
	case "PascalCase":
		if !isPascalCase(nameWithoutExt) {
			return &analyzer.Violation{
				File:       path,
				Line:       0,
				Rule:       "naming.files",
				Severity:   spec.SeverityWarning,
				Excerpt:    filepath.Base(path),
				Suggestion: "File name should be PascalCase",
			}
		}
	}

	return nil
}

// isKebabCase checks if name is kebab-case (lowercase with hyphens)
func isKebabCase(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i > 0 && r == '-' {
			continue
		}
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// isCamelCase checks if name is camelCase
func isCamelCase(name string) bool {
	if name == "" {
		return false
	}
	// First char must be lowercase
	first := name[0]
	if (first < 'a' || first > 'z') && first != '_' {
		return false
	}
	// Rest can be alphanumeric
	for _, r := range name[1:] {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

// isPascalCase checks if name is PascalCase
func isPascalCase(name string) bool {
	if name == "" {
		return false
	}
	// First char must be uppercase
	first := name[0]
	if (first < 'A' || first > 'Z') && first != '_' {
		return false
	}
	// Rest can be alphanumeric
	for _, r := range name[1:] {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

// Ensure NamingStrategy implements RuleStrategy
var _ RuleStrategy = (*NamingStrategy)(nil)
