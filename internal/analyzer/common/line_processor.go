package common

import (
	"regexp"
	"strings"
)

// LineProcessor provides utilities for processing lines of code
type LineProcessor struct{}

// NewLineProcessor creates a new LineProcessor
func NewLineProcessor() *LineProcessor {
	return &LineProcessor{}
}

// ProcessLines iterates over lines and applies a function to each
func (lp *LineProcessor) ProcessLines(content []string, fn func(line string, lineNum int) bool) []string {
	var matched []string
	for i, line := range content {
		if fn(line, i+1) {
			matched = append(matched, line)
		}
	}
	return matched
}

// FindPatternMatches finds lines that match any of the given patterns
func (lp *LineProcessor) FindPatternMatches(lines []string, patterns map[string]*regexp.Regexp) []string {
	var matches []string
	for _, line := range lines {
		for _, re := range patterns {
			if re.MatchString(line) {
				matches = append(matches, line)
				break
			}
		}
	}
	return matches
}

// ExtractImports extracts import statements from lines
func (lp *LineProcessor) ExtractImports(lines []string) []string {
	var imports []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") ||
			strings.HasPrefix(trimmed, "import (") ||
			strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
			imports = append(imports, trimmed)
		}
	}
	return imports
}

// CountLinesOfCode counts non-empty, non-comment lines
func (lp *LineProcessor) CountLinesOfCode(lines []string) int {
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment-only lines
		if trimmed == "" ||
			strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "#") {
			continue
		}
		count++
	}
	return count
}

// CountImportStatements counts import statements
func (lp *LineProcessor) CountImportStatements(lines []string) int {
	count := 0
	inBlockImport := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle import blocks
		if strings.HasPrefix(trimmed, "import (") {
			inBlockImport = true
			continue
		}
		if inBlockImport && trimmed == ")" {
			inBlockImport = false
			continue
		}
		if inBlockImport && trimmed != "" {
			count++
			continue
		}

		// Handle single line imports
		if strings.HasPrefix(trimmed, "import ") && !strings.HasPrefix(trimmed, "import (") {
			count++
		}
	}
	return count
}

// ContainsForbidden checks if lines contain any forbidden pattern
func (lp *LineProcessor) ContainsForbidden(lines []string, patterns map[string]*regexp.Regexp) (bool, string) {
	for _, line := range lines {
		for pattern, re := range patterns {
			if re.MatchString(line) {
				return true, pattern
			}
		}
	}
	return false, ""
}

// IsComment checks if a line is a comment
func (lp *LineProcessor) IsComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") ||
		strings.HasPrefix(trimmed, "#")
}

// IsEmpty checks if a line is empty or whitespace-only
func (lp *LineProcessor) IsEmpty(line string) bool {
	return strings.TrimSpace(line) == ""
}

// HasAnyPrefix checks if line has any of the given prefixes
func (lp *LineProcessor) HasAnyPrefix(line string, prefixes ...string) bool {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// HasAnySuffix checks if line has any of the given suffixes
func (lp *LineProcessor) HasAnySuffix(line string, suffixes ...string) bool {
	trimmed := strings.TrimSpace(line)
	for _, suffix := range suffixes {
		if strings.HasSuffix(trimmed, suffix) {
			return true
		}
	}
	return false
}
