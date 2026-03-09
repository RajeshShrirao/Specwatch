package strategy

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

// RuleStrategy defines the interface for analysis strategies.
// Each strategy handles a specific type of rule (forbidden, naming, limits, etc.)
type RuleStrategy interface {
	// Name returns the unique identifier for this strategy
	Name() string

	// Category returns the category name for skipping purposes
	// (e.g., "forbidden", "naming", "limits", "architecture")
	Category() string

	// Check performs the analysis and returns violations
	Check(ctx context.Context, params CheckParams) []analyzer.Violation

	// CanCheck determines if this strategy can handle the given rule
	CanCheck(rule interface{}) bool
}

// CheckParams contains common parameters for all check operations
type CheckParams struct {
	// FilePath is the path to the file being analyzed
	FilePath string

	// Content is the file content as slice of lines
	Content []string

	// Rule is the rule to check against (type varies by strategy)
	Rule interface{}

	// Cache is the file content cache
	Cache *analyzer.FileCache

	// Compiled contains pre-compiled regex patterns
	Compiled map[string]*regexp.Regexp

	// LLMClient is optional AI analysis client
	LLMClient interface {
		Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	}

	// NamingRules contains naming conventions to check
	NamingRules spec.NamingRules

	// LimitRules contains limits to check
	LimitRules spec.LimitRules

	// ArchitectureRules contains architectural constraints
	ArchitectureRules []spec.ArchitectureRule

	// ForbiddenRules contains forbidden patterns
	ForbiddenRules []spec.ForbiddenRule

	// RequiredRules contains required patterns
	RequiredRules []spec.RequiredRule

	// SkipCategories contains categories to skip
	SkipCategories []string

	// Extensions contains supported file extensions
	Extensions []string

	// MaxFileSizeMB is max file size in MB
	MaxFileSizeMB int
}

// StrategyOption is a functional options for configuring strategies
type StrategyOption func(*CheckParams)

// WithFilePath sets the file path parameter
func WithFilePath(path string) StrategyOption {
	return func(p *CheckParams) {
		p.FilePath = path
	}
}

// WithContent sets the content parameter
func WithContent(content []string) StrategyOption {
	return func(p *CheckParams) {
		p.Content = content
	}
}

// WithCache sets the cache parameter
func WithCache(cache *analyzer.FileCache) StrategyOption {
	return func(p *CheckParams) {
		p.Cache = cache
	}
}

// WithLLMClient sets the LLM client parameter
func WithLLMClient(client interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}) StrategyOption {
	return func(p *CheckParams) {
		p.LLMClient = client
	}
}

// BaseStrategy provides common functionality for all strategies
type BaseStrategy struct{}

// GetFileContent retrieves file content from cache or reads from disk
func (bs *BaseStrategy) GetFileContent(path string, cache *analyzer.FileCache) ([]string, error) {
	if cache == nil {
		return nil, nil
	}
	lines, _, err := cache.GetFileContent(path)
	return lines, err
}

// IsSupported checks if the file extension is supported
func (bs *BaseStrategy) IsSupported(path string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range extensions {
		supported = strings.ToLower(supported)
		if !strings.HasPrefix(supported, ".") {
			supported = "." + supported
		}
		if ext == supported {
			return true
		}
	}
	return false
}

// ShouldSkip checks if the category should be skipped
func (bs *BaseStrategy) ShouldSkip(category string, skipCategories []string) bool {
	for _, skip := range skipCategories {
		if strings.EqualFold(skip, category) {
			return true
		}
	}
	return false
}

// GetFilename returns the filename without path
func (bs *BaseStrategy) GetFilename(path string) string {
	return filepath.Base(path)
}

// GetExtension returns the file extension
func (bs *BaseStrategy) GetExtension(path string) string {
	return filepath.Ext(path)
}

// GetFilenameWithoutExt returns filename without extension
func (bs *BaseStrategy) GetFilenameWithoutExt(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}
