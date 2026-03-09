package common

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
)

// BaseChecker provides reusable analysis utilities
type BaseChecker struct {
	fileCache    *analyzer.FileCache
	patternCache *analyzer.PatternCache
	extensions   []string
}

// NewBaseChecker creates a new BaseChecker
func NewBaseChecker(fileCache *analyzer.FileCache, patternCache *analyzer.PatternCache, extensions []string) *BaseChecker {
	return &BaseChecker{
		fileCache:    fileCache,
		patternCache: patternCache,
		extensions:   extensions,
	}
}

// GetFileContent retrieves file content from cache or reads from disk
func (bc *BaseChecker) GetFileContent(path string) ([]string, error) {
	if bc.fileCache == nil {
		return nil, nil
	}
	lines, _, err := bc.fileCache.GetFileContent(path)
	return lines, err
}

// GetOrCompilePattern returns a compiled regex from cache or compiles it
func (bc *BaseChecker) GetOrCompilePattern(pattern string) (*regexp.Regexp, error) {
	if bc.patternCache == nil {
		return regexp.Compile(pattern)
	}
	return bc.patternCache.GetPattern(pattern)
}

// IsSupported checks if the file extension is supported
func (bc *BaseChecker) IsSupported(path string) bool {
	if len(bc.extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range bc.extensions {
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
func (bc *BaseChecker) ShouldSkip(category string, skipCategories []string) bool {
	for _, skip := range skipCategories {
		if strings.EqualFold(skip, category) {
			return true
		}
	}
	return false
}

// FileSizeLimitExceeded checks if the file exceeds the size limit
func (bc *BaseChecker) FileSizeLimitExceeded(path string, maxSizeMB int) bool {
	return analyzer.FileSizeLimitExceeded(path, maxSizeMB)
}

// GetFilename returns the filename without path
func (bc *BaseChecker) GetFilename(path string) string {
	return filepath.Base(path)
}

// GetExtension returns the file extension
func (bc *BaseChecker) GetExtension(path string) string {
	return filepath.Ext(path)
}

// GetFilenameWithoutExt returns filename without extension
func (bc *BaseChecker) GetFilenameWithoutExt(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// PrecompileForbiddenPatterns compiles all forbidden patterns
func (bc *BaseChecker) PrecompileForbiddenPatterns(patterns []string) (map[string]*regexp.Regexp, error) {
	compiled := make(map[string]*regexp.Regexp)
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		re, err := bc.GetOrCompilePattern(pattern)
		if err != nil {
			return nil, err
		}
		compiled[pattern] = re
	}
	return compiled, nil
}

// SetExtensions updates the supported extensions
func (bc *BaseChecker) SetExtensions(extensions []string) {
	bc.extensions = extensions
}

// SetFileCache updates the file cache
func (bc *BaseChecker) SetFileCache(cache *analyzer.FileCache) {
	bc.fileCache = cache
}

// SetPatternCache updates the pattern cache
func (bc *BaseChecker) SetPatternCache(cache *analyzer.PatternCache) {
	bc.patternCache = cache
}
