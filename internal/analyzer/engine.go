package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rajeshshrirao/specwatch/internal/spec"
)

type Engine struct {
	Rules *spec.RuleSet
}

func NewEngine(rules *spec.RuleSet) *Engine {
	return &Engine{Rules: rules}
}

func (e *Engine) Analyze(path string) ([]Violation, time.Duration) {
	start := time.Now()
	var violations []Violation

	// Only analyze supported files (TS/JS only for now)
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
		return violations, time.Since(start)
	}

	// Check forbidden patterns
	if len(e.Rules.Forbidden) > 0 {
		violations = append(violations, CheckForbidden(path, e.Rules.Forbidden)...)
	}

	// Check naming
	violations = append(violations, CheckNaming(path, e.Rules.Naming)...)

	// Check limits
	if e.Rules.Limits.MaxFileLines > 0 || e.Rules.Limits.MaxImports > 0 {
		violations = append(violations, CheckLimits(path, e.Rules.Limits)...)
	}

	// Check required try/catch for async functions
	violations = append(violations, CheckRequiredTryCatch(path)...)

	// Check import boundaries for architecture rules
	if len(e.Rules.Architecture) > 0 {
		violations = append(violations, CheckImportBoundaries(path, e.Rules.Architecture)...)
	}

	return violations, time.Since(start)
}

func (e *Engine) AnalyzeAll(root string) ([]Violation, time.Duration) {
	start := time.Now()
	var allViolations []Violation

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
			return nil
		}

		violations, _ := e.Analyze(path)
		allViolations = append(allViolations, violations...)
		return nil
	})

	if err != nil {
		return allViolations, time.Since(start)
	}

	return allViolations, time.Since(start)
}
