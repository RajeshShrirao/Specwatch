package analyzer

import (
	"regexp"
	"sync"
)

// PatternCache is a thread-safe cache for compiled regex patterns
type PatternCache struct {
	cache   map[string]*regexp.Regexp
	mu      sync.RWMutex
	maxSize int
}

// Common regex patterns that are pre-compiled
var (
	// kebabCasePattern matches kebab-case filenames
	kebabCasePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	// camelCasePattern matches camelCase identifiers
	camelCasePattern = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)

	// pascalCasePattern matches PascalCase identifiers
	pascalCasePattern = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

	// screamingSnakePattern matches SCREAMING_SNAKE_CASE
	screamingSnakePattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`)

	// interfacePattern matches interface names prefixed with I
	interfacePattern = regexp.MustCompile(`^I[A-Z][a-zA-Z0-9]*$`)
)

// NewPatternCache creates a new pattern cache with specified max size
func NewPatternCache(maxSize int) *PatternCache {
	if maxSize <= 0 {
		maxSize = 100 // Default to 100 patterns
	}
	return &PatternCache{
		cache:   make(map[string]*regexp.Regexp),
		maxSize: maxSize,
	}
}

// GetPattern returns a compiled regex from cache or compiles and caches it
func (pc *PatternCache) GetPattern(pattern string) (*regexp.Regexp, error) {
	// Check if already compiled
	pc.mu.RLock()
	if re, exists := pc.cache[pattern]; exists {
		pc.mu.RUnlock()
		return re, nil
	}
	pc.mu.RUnlock()

	// Compile the pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// Add to cache with eviction if full
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Double-check after acquiring write lock
	if re, exists := pc.cache[pattern]; exists {
		return re, nil
	}

	// Evict if cache is full
	if len(pc.cache) >= pc.maxSize {
		// Simple eviction: remove first entry (FIFO)
		for k := range pc.cache {
			delete(pc.cache, k)
			break
		}
	}

	pc.cache[pattern] = re
	return re, nil
}

// GetKebabCasePattern returns the pre-compiled kebab-case pattern
func GetKebabCasePattern() *regexp.Regexp {
	return kebabCasePattern
}

// GetCamelCasePattern returns the pre-compiled camelCase pattern
func GetCamelCasePattern() *regexp.Regexp {
	return camelCasePattern
}

// GetPascalCasePattern returns the pre-compiled PascalCase pattern
func GetPascalCasePattern() *regexp.Regexp {
	return pascalCasePattern
}

// GetScreamingSnakePattern returns the pre-compiled SCREAMING_SNAKE_CASE pattern
func GetScreamingSnakePattern() *regexp.Regexp {
	return screamingSnakePattern
}

// GetInterfacePattern returns the pre-compiled interface pattern
func GetInterfacePattern() *regexp.Regexp {
	return interfacePattern
}

// IsValidNaming checks if a name matches the specified convention using pre-compiled patterns
func IsValidNaming(name, convention string) bool {
	switch convention {
	case "kebab-case":
		return kebabCasePattern.MatchString(name)
	case "camelCase":
		return camelCasePattern.MatchString(name)
	case "PascalCase":
		return pascalCasePattern.MatchString(name)
	case "SCREAMING_SNAKE_CASE":
		return screamingSnakePattern.MatchString(name)
	case "I PascalCase":
		// Must start with I and be PascalCase
		return len(name) > 1 && name[0] == 'I' && pascalCasePattern.MatchString(name[1:])
	default:
		return true
	}
}

// PrecompileForbiddenPatterns compiles all forbidden patterns from rules
func PrecompileForbiddenPatterns(patterns []string, cache *PatternCache) (map[string]*regexp.Regexp, error) {
	compiled := make(map[string]*regexp.Regexp)

	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		re, err := cache.GetPattern(pattern)
		if err != nil {
			return nil, err
		}
		compiled[pattern] = re
	}

	return compiled, nil
}
