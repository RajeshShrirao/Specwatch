package analyzer

import (
	"os"
	"testing"
	"time"
)

func TestFileCache(t *testing.T) {
	// Create a temp test file
	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test content
	content := "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Create cache
	cache := NewFileCache(10, 5) // 10MB cache, 5min TTL

	// Test GetFileContent
	lines, size, err := cache.GetFileContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file content: %v", err)
	}

	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines, got %d", len(lines))
	}

	if size == 0 {
		t.Error("Expected non-zero file size")
	}

	// Test cache hit
	lines2, size2, err := cache.GetFileContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file content (cache hit): %v", err)
	}

	if len(lines) != len(lines2) {
		t.Errorf("Cache hit returned different line count: %d vs %d", len(lines), len(lines2))
	}

	if size != size2 {
		t.Errorf("Cache hit returned different size: %d vs %d", size, size2)
	}

	// Test FileSizeLimitExceeded
	if FileSizeLimitExceeded(tmpFile.Name(), 1) {
		t.Error("Expected file to NOT exceed 1MB limit")
	}

	// Test with non-existent file
	_, _, err = cache.GetFileContent("/nonexistent/file.go")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFileCacheEviction(t *testing.T) {
	// Create a temp test file
	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test content
	content := "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Create cache with small size and short TTL
	cache := NewFileCache(1, 0) // 1 byte (!) to force eviction, 0min TTL

	// First read should work
	lines, _, err := cache.GetFileContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file content: %v", err)
	}

	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines, got %d", len(lines))
	}

	// Wait briefly
	time.Sleep(10 * time.Millisecond)

	// Second read should work (cache may have been evicted)
	lines2, _, err := cache.GetFileContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file content after eviction: %v", err)
	}

	if len(lines2) < 3 {
		t.Errorf("Expected at least 3 lines after eviction, got %d", len(lines2))
	}
}

func TestFileSizeLimitExceeded(t *testing.T) {
	// Create a temp test file
	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write small content
	content := "package main\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test with large limit (should pass)
	if FileSizeLimitExceeded(tmpFile.Name(), 100) {
		t.Error("Expected file to NOT exceed 100MB limit")
	}

	// Test with non-existent file
	if FileSizeLimitExceeded("/nonexistent/file.go", 1) {
		t.Error("Expected non-existent file to NOT exceed limit")
	}
}
