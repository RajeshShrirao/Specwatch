package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// CheckRequiredTryCatch uses heuristics to check for try/catch in async functions
func CheckRequiredTryCatch(path string) []Violation {
	var violations []Violation

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	content := ""
	lines := []string{}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lines = append(lines, line)
		content += line + "\n"
	}

	for i, line := range lines {
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
			for j := i; j < i+lookAheadLines && j < len(lines); j++ {
				if strings.Contains(lines[j], "try {") || strings.Contains(lines[j], "try{") {
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
func CheckImportBoundaries(path string, rules []spec.ArchitectureRule) []Violation {
	var violations []Violation

	// Heuristic: "no direct db calls outside src/lib/db"
	// If file is NOT in src/lib/db, check for forbidden imports or patterns
	absPath, _ := filepath.Abs(path)
	isDbFile := strings.Contains(absPath, "src/lib/db")

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

		if !isDbFile {
			// Check for direct DB-related keywords or imports if not in DB layer
			for _, rule := range rules {
				if strings.Contains(rule.Description, "no direct db calls") {
					if strings.Contains(line, " prisma.") || strings.Contains(line, " mongoose.") || strings.Contains(line, " sequelize.") {
						violations = append(violations, Violation{
							File:       path,
							Line:       lineNum,
							Rule:       "architecture.no_direct_db",
							Severity:   spec.SeverityError,
							Excerpt:    strings.TrimSpace(line),
							Suggestion: "Direct database calls are only allowed in src/lib/db",
						})
					}
				}
			}
		}
	}

	return violations
}
