package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rajeshshrirao/specwatch/internal/spec"
)

type Engine struct {
	Rules          *spec.RuleSet
	SkipCategories []string
	Extensions     []string
}

func NewEngine(rules *spec.RuleSet) *Engine {
	return &Engine{
		Rules:      rules,
		Extensions: []string{".ts", ".tsx", ".js", ".jsx"},
	}
}

func (e *Engine) shouldSkip(category string) bool {
	for _, skip := range e.SkipCategories {
		if strings.EqualFold(skip, category) {
			return true
		}
	}
	return false
}

func (e *Engine) isSupported(path string) bool {
	if len(e.Extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range e.Extensions {
		if !strings.HasPrefix(supported, ".") {
			supported = "." + supported
		}
		if ext == strings.ToLower(supported) {
			return true
		}
	}
	return false
}

func (e *Engine) Analyze(path string) ([]Violation, time.Duration) {
	start := time.Now()
	var violations []Violation

	if !e.isSupported(path) {
		return violations, time.Since(start)
	}

	// Check forbidden patterns
	if !e.shouldSkip("forbidden") && len(e.Rules.Forbidden) > 0 {
		violations = append(violations, CheckForbidden(path, e.Rules.Forbidden)...)
	}

	// Check naming
	if !e.shouldSkip("naming") {
		violations = append(violations, CheckNaming(path, e.Rules.Naming)...)
	}

	// Check limits
	if !e.shouldSkip("limits") && (e.Rules.Limits.MaxFileLines > 0 || e.Rules.Limits.MaxImports > 0) {
		violations = append(violations, CheckLimits(path, e.Rules.Limits)...)
	}

	// Check required try/catch for async functions
	if !e.shouldSkip("required") {
		violations = append(violations, CheckRequiredTryCatch(path)...)
	}

	// Check import boundaries for architecture rules
	if !e.shouldSkip("architecture") && len(e.Rules.Architecture) > 0 {
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

		if !e.isSupported(path) {
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
