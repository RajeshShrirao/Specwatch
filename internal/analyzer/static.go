package analyzer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// CheckForbidden patterns — regex on file content
func CheckForbidden(path string, rules []spec.ForbiddenRule, cache *FileCache, compiled map[string]*regexp.Regexp) []Violation {
	var violations []Violation

	if len(rules) == 0 {
		return violations
	}

	content, _, err := cache.GetFileContent(path)
	if err != nil {
		return nil
	}

	lineNum := 0
	for _, line := range content {
		lineNum++

		for _, rule := range rules {
			if rule.Pattern != "" {
				// Use pre-compiled pattern if available, otherwise fall back to strings.Contains
				if re, ok := compiled[rule.Pattern]; ok {
					if re.MatchString(line) {
						violations = append(violations, Violation{
							File:       path,
							Line:       lineNum,
							Rule:       "forbidden.pattern",
							Severity:   spec.SeverityError,
							Excerpt:    strings.TrimSpace(line),
							Suggestion: rule.Message,
						})
					}
				} else if strings.Contains(line, rule.Pattern) {
					violations = append(violations, Violation{
						File:       path,
						Line:       lineNum,
						Rule:       "forbidden.pattern",
						Severity:   spec.SeverityError,
						Excerpt:    strings.TrimSpace(line),
						Suggestion: rule.Message,
					})
				}
			}

			if rule.Import != "" {
				// Simple heuristic for imports
				if (strings.HasPrefix(line, "import") || strings.Contains(line, "require(")) && strings.Contains(line, rule.Import) {
					violations = append(violations, Violation{
						File:       path,
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

// CheckNaming checks filename against expected convention
func CheckNaming(path string, rules spec.NamingRules) []Violation {
	var violations []Violation
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Check file naming using pre-compiled patterns
	if rules.Files == "kebab-case" {
		isKebab := GetKebabCasePattern().MatchString(nameWithoutExt)
		// Relax for common special files like README.md or spec.md
		if !isKebab && !strings.Contains(strings.ToLower(nameWithoutExt), "readme") && nameWithoutExt != "spec" {
			violations = append(violations, Violation{
				File:       path,
				Line:       0,
				Rule:       "naming.files",
				Severity:   spec.SeverityWarning,
				Excerpt:    filename,
				Suggestion: "File name should be kebab-case",
			})
		}
	}

	return violations
}

// CheckLimits: line count, import count
func CheckLimits(path string, limits spec.LimitRules, cache *FileCache) []Violation {
	var violations []Violation

	content, _, err := cache.GetFileContent(path)
	if err != nil {
		return nil
	}

	lineCount := 0
	importCount := 0
	for _, line := range content {
		lineCount++
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			importCount++
		}
	}

	if limits.MaxFileLines > 0 && lineCount > limits.MaxFileLines {
		violations = append(violations, Violation{
			File:       path,
			Line:       lineCount,
			Rule:       "limits.file_lines",
			Severity:   spec.SeverityError,
			Excerpt:    "File too long",
			Suggestion: "Max lines allowed: " + fmt.Sprintf("%d", limits.MaxFileLines),
		})
	}

	if limits.MaxImports > 0 && importCount > limits.MaxImports {
		violations = append(violations, Violation{
			File:       path,
			Line:       0,
			Rule:       "limits.imports",
			Severity:   spec.SeverityWarning,
			Excerpt:    "Too many imports",
			Suggestion: "Max imports allowed: " + fmt.Sprintf("%d", limits.MaxImports),
		})
	}

	return violations
}
