package strategy

import (
	"context"
	"fmt"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// AIStrategy uses LLM for AI-powered analysis
type AIStrategy struct {
	BaseStrategy
	name     string
	category string
}

// NewAIStrategy creates a new AIStrategy
func NewAIStrategy() *AIStrategy {
	return &AIStrategy{
		name:     "ai",
		category: "ai",
	}
}

// Name returns the strategy name
func (s *AIStrategy) Name() string {
	return s.name
}

// Category returns the strategy category
func (s *AIStrategy) Category() string {
	return s.category
}

// CanCheck determines if this strategy can handle the given rule
func (s *AIStrategy) CanCheck(rule interface{}) bool {
	// AI strategy can check any rule when LLM is available
	return s != nil
}

// Check performs AI-powered analysis using LLM
func (s *AIStrategy) Check(ctx context.Context, params CheckParams) []analyzer.Violation {
	var violations []analyzer.Violation

	// Check if LLM client is available
	if params.LLMClient == nil {
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

	// Build rule description from available rules
	ruleDescription := s.buildRuleDescription(params)

	// Skip if no rules to check
	if ruleDescription == "" {
		return violations
	}

	// Call LLM for analysis
	prompt := fmt.Sprintf(`You are a code analyzer. Check if the following code violates this architectural rule: %s

Code:
%s

Respond with a list of violations found, or "OK" if no violations.`, ruleDescription, strings.Join(content, "\n"))

	result, err := params.LLMClient.Generate(ctx, "", prompt)
	if err != nil {
		// Return empty on error - don't block other checks
		return violations
	}

	// Parse the LLM response
	if strings.Contains(strings.ToLower(result), "ok") || strings.Contains(strings.ToLower(result), "no violations") {
		return nil
	}

	// Parse violations from LLM response
	violations = s.parseLLMViolations(params.FilePath, result)

	return violations
}

// buildRuleDescription creates a description from available rules
func (s *AIStrategy) buildRuleDescription(params CheckParams) string {
	var rules []string

	// Add forbidden rules
	for _, rule := range params.ForbiddenRules {
		if rule.Message != "" {
			rules = append(rules, rule.Message)
		}
	}

	// Add architecture rules
	for _, rule := range params.ArchitectureRules {
		if rule.Description != "" {
			rules = append(rules, rule.Description)
		}
	}

	// Add required rules
	for _, rule := range params.RequiredRules {
		if rule.Message != "" {
			rules = append(rules, rule.Message)
		}
	}

	// Add naming rules
	if params.NamingRules.Files != "" {
		rules = append(rules, fmt.Sprintf("Files must be %s", params.NamingRules.Files))
	}

	return strings.Join(rules, "; ")
}

// parseLLMViolations parses violations from LLM response
func (s *AIStrategy) parseLLMViolations(filePath, response string) []analyzer.Violation {
	var violations []analyzer.Violation
	lines := strings.Split(response, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		// Filter out non-violation lines
		if len(line) > 10 && !strings.HasPrefix(line, "OK") && !strings.HasPrefix(line, "No") {
			violations = append(violations, analyzer.Violation{
				File:     filePath,
				Line:     i + 1,
				Rule:     "ai-analysis",
				Severity: spec.SeverityWarning,
				Excerpt:  line,
			})
		}
	}

	return violations
}

// Ensure AIStrategy implements RuleStrategy
var _ RuleStrategy = (*AIStrategy)(nil)
