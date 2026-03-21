package strategy

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// fileTouchesArchitectureRules checks if a file might touch any of the user's
// architecture rules by extracting keywords from the rule descriptions and
// checking if they appear in the file content. This is more permissive than
// hardcoded keyword lists and adapts to user's custom architecture rules.
func fileTouchesArchitectureRules(content []string, rules []spec.ArchitectureRule) bool {
	if len(rules) == 0 || len(content) == 0 {
		return false
	}

	// Extract keywords from rule descriptions
	keywords := extractRuleKeywords(rules)
	if len(keywords) == 0 {
		// If no keywords can be extracted, be permissive and return true
		// to ensure AI analysis can still run
		return true
	}

	// Check if any keyword appears in the file content
	for _, line := range content {
		lineLower := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lineLower, strings.ToLower(keyword)) {
				return true
			}
		}
	}

	return false
}

// extractRuleKeywords extracts potential keywords/patterns from architecture rule descriptions
func extractRuleKeywords(rules []spec.ArchitectureRule) []string {
	var keywords []string
	seen := make(map[string]bool)

	for _, rule := range rules {
		if rule.Description == "" {
			continue
		}

		// Extract words that look like identifiers or technical terms
		// This regex matches words that could be import names, package names, etc.
		re := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
		matches := re.FindAllString(rule.Description, -1)

		for _, match := range matches {
			// Filter out common words that aren't useful keywords
			lower := strings.ToLower(match)
			if isCommonWord(lower) {
				continue
			}
			if !seen[lower] {
				seen[lower] = true
				keywords = append(keywords, match)
			}
		}
	}

	return keywords
}

// isCommonWord returns true if the word is too common to be a useful keyword
func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"no": true, "not": true, "only": true, "allowed": true,
		"in": true, "outside": true, "the": true, "are": true,
		"be": true, "to": true, "from": true, "must": true,
		"should": true, "can": true, "may": true, "will": true,
		"direct": true, "calls": true, "files": true, "file": true,
		"function": true, "functions": true, "class": true, "classes": true,
		"import": true, "imports": true, "export": true, "exports": true,
		"use": true, "using": true, "used": true, "with": true,
		"without": true, "have": true, "has": true, "and": true,
		"or": true, "if": true, "then": true, "else": true,
		"when": true, "where": true, "how": true, "what": true,
		"which": true, "who": true, "whom": true, "this": true,
		"that": true, "these": true, "those": true, "all": true,
		"any": true, "some": true, "every": true, "each": true,
		"such": true, "like": true, "as": true, "for": true,
		"by": true, "on": true, "at": true, "of": true,
	}
	return commonWords[word]
}

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
		violations = append(violations, s.checkRule(params.FilePath, content, rule, rules)...)
	}

	return violations
}

// checkRule checks a specific architectural rule
// allRules is passed to enable extracting keywords from all rules for matching
func (s *ArchitectureStrategy) checkRule(path string, content []string, rule spec.ArchitectureRule, allRules []spec.ArchitectureRule) []analyzer.Violation {
	var violations []analyzer.Violation

	// Use the new fileTouchesArchitectureRules function to check if the file
	// might touch any architecture rules based on user's custom rules, not hardcoded keywords
	if !fileTouchesArchitectureRules(content, allRules) {
		return violations
	}

	// Determine exempt path pattern from rule (use ExemptPathPattern or default)
	exemptPathPattern := rule.ExemptPathPattern
	if exemptPathPattern == "" {
		// Default to extracting from description if no explicit pattern
		// This maintains backward compatibility
		exemptPathPattern = "src/lib/db"
	}

	// Heuristic: "no direct db calls outside src/lib/db"
	// If file is NOT in exempt path, check for forbidden imports or patterns
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't get absolute path, log and skip this file
		// Treat as not-a-db-file to be safe (fail open for exemptions)
		absPath = path
	}
	isDbFile := strings.Contains(absPath, exemptPathPattern)

	if !isDbFile {
		// Check for direct DB-related keywords or imports
		// Use explicit Patterns from the rule if available, otherwise extract from description
		var patterns []string
		if len(rule.Patterns) > 0 {
			patterns = rule.Patterns
		} else {
			// Fallback to extracting from description (for backward compatibility)
			patterns = extractRuleKeywords([]spec.ArchitectureRule{rule})
		}

		// Precompile regexes for each pattern once before iterating over content
		// This improves performance significantly for large files
		type compiledPattern struct {
			re       *regexp.Regexp
			fallback string // used when regex compilation fails
		}
		compiledPatterns := make([]compiledPattern, len(patterns))
		for i, pattern := range patterns {
			patternLower := strings.ToLower(pattern)
			re, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(patternLower) + `\b`)
			if err != nil {
				// Record the fallback pattern for later use
				compiledPatterns[i] = compiledPattern{re: nil, fallback: patternLower}
			} else {
				compiledPatterns[i] = compiledPattern{re: re, fallback: ""}
			}
		}

		for lineNum, line := range content {
			lineLower := strings.ToLower(line)
			for _, cp := range compiledPatterns {
				// Use word-boundary regex for matching to avoid matching inside identifiers
				// e.g., "db" should not match "debug" or "adb"
				if cp.re != nil {
					// Use precompiled regex
					if cp.re.MatchString(lineLower) {
						violations = append(violations, analyzer.Violation{
							File:       path,
							Line:       lineNum + 1,
							Rule:       "architecture.no_direct_db",
							Severity:   spec.SeverityError,
							Excerpt:    strings.TrimSpace(line),
							Suggestion: fmt.Sprintf("Direct database calls are only allowed in %s", exemptPathPattern),
						})
						break
					}
				} else if cp.fallback != "" {
					// Fall back to simple substring match for patterns that failed compilation
					if strings.Contains(lineLower, cp.fallback) {
						violations = append(violations, analyzer.Violation{
							File:       path,
							Line:       lineNum + 1,
							Rule:       "architecture.no_direct_db",
							Severity:   spec.SeverityError,
							Excerpt:    strings.TrimSpace(line),
							Suggestion: fmt.Sprintf("Direct database calls are only allowed in %s", exemptPathPattern),
						})
						break
					}
				}
			}
		}
	}

	return violations
}

// Ensure ArchitectureStrategy implements RuleStrategy
var _ RuleStrategy = (*ArchitectureStrategy)(nil)
