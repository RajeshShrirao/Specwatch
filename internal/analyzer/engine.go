package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rajeshshrirao/specwatch/internal/llm"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

type Engine struct {
	Rules          *spec.RuleSet
	SkipCategories []string
	Extensions     []string
	LLMClient      llm.LLMClient
}

func NewEngine(rules *spec.RuleSet) *Engine {
	return &Engine{
		Rules:      rules,
		Extensions: []string{".go", ".ts", ".tsx", ".js", ".jsx"},
	}
}

// SetLLMClient sets the LLM client for AI-powered analysis
func (e *Engine) SetLLMClient(client llm.LLMClient) {
	e.LLMClient = client
}

// HasLLM returns true if an LLM client is configured
func (e *Engine) HasLLM() bool {
	return e.LLMClient != nil
}

// AnalyzeWithAI performs AI-powered analysis using the LLM
func (e *Engine) AnalyzeWithAI(ctx context.Context, filePath, codeContent, ruleDescription string) ([]Violation, error) {
	if e.LLMClient == nil {
		return nil, fmt.Errorf("LLM client not configured")
	}

	prompt := fmt.Sprintf(`You are a code analyzer. Check if the following code violates this architectural rule: %s

Code:
%s

Respond with a list of violations found, or "OK" if no violations.`, ruleDescription, codeContent)

	result, err := e.LLMClient.Generate(ctx, prompt, "")
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse the LLM response - if it contains "OK", no violations
	if strings.Contains(strings.ToLower(result), "ok") || strings.Contains(strings.ToLower(result), "no violations") {
		return nil, nil
	}

	// Parse violations from LLM response
	// This is a simple implementation - could be enhanced
	violations := parseLLMViolations(filePath, result)
	return violations, nil
}

func parseLLMViolations(filePath, response string) []Violation {
	// Simple implementation - parse lines that look like violations
	var violations []Violation
	lines := strings.Split(response, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 && !strings.HasPrefix(line, "OK") && !strings.HasPrefix(line, "No") {
			violations = append(violations, Violation{
				File:    filePath,
				Line:    i + 1,
				Rule:    "ai-analysis",
				Excerpt: line,
			})
		}
	}
	return violations
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
