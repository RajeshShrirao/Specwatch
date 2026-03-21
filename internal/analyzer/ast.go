package analyzer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// extractArchitectureKeywords extracts keywords from architecture rules for matching
func extractArchitectureKeywords(rules []spec.ArchitectureRule) []string {
	var keywords []string
	seen := make(map[string]bool)

	for _, rule := range rules {
		if rule.Description == "" {
			continue
		}

		// Extract words that look like identifiers or technical terms
		re := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
		matches := re.FindAllString(rule.Description, -1)

		for _, match := range matches {
			// Filter out common words that aren't useful keywords
			lower := strings.ToLower(match)
			if isCommonArchitectureWord(lower) {
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

// isCommonArchitectureWord returns true if the word is too common to be a useful keyword
func isCommonArchitectureWord(word string) bool {
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

// CheckRequiredTryCatch uses heuristics to check for try/catch in async functions
func CheckRequiredTryCatch(path string, cache *FileCache) []Violation {
	var violations []Violation

	content, _, err := cache.GetFileContent(path)
	if err != nil {
		return nil
	}

	for i, line := range content {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

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
				violations = append(violations, Violation{
					File:       path,
					Line:       i + 1,
					Rule:       "required.try_catch",
					Severity:   spec.SeverityError,
					Excerpt:    strings.TrimSpace(line),
					Suggestion: "Async functions should be wrapped in try/catch blocks",
				})
			}
		}
	}

	return violations
}

// CheckImportBoundaries checks if imports violate architectural rules
func CheckImportBoundaries(path string, rules []spec.ArchitectureRule, cache *FileCache) []Violation {
	var violations []Violation

	if len(rules) == 0 {
		return violations
	}

	content, _, err := cache.GetFileContent(path)
	if err != nil {
		return nil
	}

	// Determine exempt path pattern from rules (use first rule with ExemptPathPattern set)
	// Default to "src/lib/db" if not specified
	exemptPathPattern := "src/lib/db"
	for _, rule := range rules {
		if rule.ExemptPathPattern != "" {
			exemptPathPattern = rule.ExemptPathPattern
			break
		}
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

	for lineNum, line := range content {
		if !isDbFile {
			// Check for direct DB-related keywords or imports if not in DB layer
			// Extract keywords per-rule to avoid false positives from other rules
			for _, rule := range rules {
				// Check if this rule is about "no direct db calls" or has explicit patterns
				isDbRule := strings.Contains(strings.ToLower(rule.Description), "no direct db calls") ||
					rule.ID == "architecture.no_direct_db"

				if isDbRule {
					// Extract keywords for this specific rule only
					ruleKeywords := extractArchitectureKeywords([]spec.ArchitectureRule{rule})
					for _, keyword := range ruleKeywords {
						if strings.Contains(strings.ToLower(line), strings.ToLower(keyword)) {
							violations = append(violations, Violation{
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
	}

	return violations
}
