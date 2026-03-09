package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// CheckForbidden patterns — regex on file content
func CheckForbidden(path string, rules []spec.ForbiddenRule) []Violation {
	var violations []Violation

	if len(rules) == 0 {
		return violations
	}

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, rule := range rules {
			if rule.Pattern != "" {
				if strings.Contains(line, rule.Pattern) {
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

	// Check file naming (kebab-case)
	if rules.Files == "kebab-case" {
		isKebab := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(nameWithoutExt)
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
func CheckLimits(path string, limits spec.LimitRules) []Violation {
	var violations []Violation

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	importCount := 0
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
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
