package analyzer

import (
	"regexp"
	"testing"
)

func TestPatternCache(t *testing.T) {
	cache := NewPatternCache(10)

	// Test getting a pattern
	re, err := cache.GetPattern(`^\d+$`)
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	if !re.MatchString("123") {
		t.Error("Expected pattern to match '123'")
	}

	// Test cache hit
	re2, err := cache.GetPattern(`^\d+$`)
	if err != nil {
		t.Fatalf("Failed to get cached pattern: %v", err)
	}

	if re != re2 {
		t.Error("Expected same regex instance from cache")
	}

	// Test invalid pattern
	_, err = cache.GetPattern(`[invalid`)
	if err == nil {
		t.Error("Expected error for invalid pattern")
	}
}

func TestPatternCacheEviction(t *testing.T) {
	cache := NewPatternCache(3)

	// Add more patterns than cache size
	if _, err := cache.GetPattern("^a$"); err != nil {
		t.Fatalf("Failed to get pattern: %v", err)
	}
	if _, err := cache.GetPattern("^b$"); err != nil {
		t.Fatalf("Failed to get pattern: %v", err)
	}
	if _, err := cache.GetPattern("^c$"); err != nil {
		t.Fatalf("Failed to get pattern: %v", err)
	}

	// This should trigger eviction
	re, err := cache.GetPattern(`^d$`)
	if err != nil {
		t.Fatalf("Failed to get pattern: %v", err)
	}

	if re == nil {
		t.Error("Expected non-nil regex")
	}
}

func TestPrecompiledNamingPatterns(t *testing.T) {
	tests := []struct {
		name      string
		pattern   *regexp.Regexp
		testCases []struct {
			input    string
			expected bool
		}
	}{
		{
			name:    "kebabCase",
			pattern: GetKebabCasePattern(),
			testCases: []struct {
				input    string
				expected bool
			}{
				{"my-function", true},
				{"my", true},
				{"my-function-name", true},
				{"MyFunction", false},
				{"my_function", false},
				{"", false},
			},
		},
		{
			name:    "camelCase",
			pattern: GetCamelCasePattern(),
			testCases: []struct {
				input    string
				expected bool
			}{
				{"myFunction", true},
				{"my", true},
				{"MyFunction", false},
				{"my_function", false},
				{"", false},
			},
		},
		{
			name:    "pascalCase",
			pattern: GetPascalCasePattern(),
			testCases: []struct {
				input    string
				expected bool
			}{
				{"MyFunction", true},
				{"My", true},
				{"myFunction", false},
				{"my_function", false},
				{"", false},
			},
		},
		{
			name:    "screamingSnake",
			pattern: GetScreamingSnakePattern(),
			testCases: []struct {
				input    string
				expected bool
			}{
				{"MY_CONSTANT", true},
				{"MY", true},
				{"myConstant", false},
				{"my_constant", false},
				{"", false},
			},
		},
		{
			name:    "interface",
			pattern: GetInterfacePattern(),
			testCases: []struct {
				input    string
				expected bool
			}{
				{"IUser", true},
				{"I", false},
				{"User", false},
				{"iUser", false},
				{"", false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tc := range tt.testCases {
				result := tt.pattern.MatchString(tc.input)
				if result != tc.expected {
					t.Errorf("For input %q: expected %v, got %v", tc.input, tc.expected, result)
				}
			}
		})
	}
}

func TestIsValidNaming(t *testing.T) {
	tests := []struct {
		name       string
		convention string
		input      string
		expected   bool
	}{
		{"kebab-case valid", "kebab-case", "my-function", true},
		{"kebab-case invalid", "kebab-case", "MyFunction", false},
		{"camelCase valid", "camelCase", "myFunction", true},
		{"camelCase invalid", "camelCase", "MyFunction", false},
		{"PascalCase valid", "PascalCase", "MyFunction", true},
		{"PascalCase invalid", "PascalCase", "myFunction", false},
		{"SCREAMING_SNAKE valid", "SCREAMING_SNAKE_CASE", "MY_CONSTANT", true},
		{"SCREAMING_SNAKE invalid", "SCREAMING_SNAKE_CASE", "myConstant", false},
		{"I PascalCase valid", "I PascalCase", "IUser", true},
		{"I PascalCase invalid", "I PascalCase", "User", false},
		{"unknown convention", "unknown", "anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidNaming(tt.input, tt.convention)
			if result != tt.expected {
				t.Errorf("IsValidNaming(%q, %q): expected %v, got %v", tt.input, tt.convention, tt.expected, result)
			}
		})
	}
}

func TestPrecompileForbiddenPatterns(t *testing.T) {
	cache := NewPatternCache(10)
	patterns := []string{`^\s*console\.log`, `debugger`, `TODO.*FIXME`}

	compiled, err := PrecompileForbiddenPatterns(patterns, cache)
	if err != nil {
		t.Fatalf("Failed to precompile patterns: %v", err)
	}

	if len(compiled) != len(patterns) {
		t.Errorf("Expected %d compiled patterns, got %d", len(patterns), len(compiled))
	}

	// Verify patterns work
	if !compiled[`^\s*console\.log`].MatchString("console.log") {
		t.Error("Expected pattern to match 'console.log'")
	}

	if !compiled[`^\s*console\.log`].MatchString("  console.log") {
		t.Error("Expected pattern to match '  console.log'")
	}

	if !compiled[`debugger`].MatchString("debugger") {
		t.Error("Expected pattern to match 'debugger'")
	}
}
