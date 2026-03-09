package analyzer

import (
	"bufio"
	"os"
	"sync"
	"time"
)

// FileCacheKey represents a unique identifier for a file in the cache
type FileCacheKey struct {
	path    string
	modTime time.Time
}

// FileCacheValue holds the cached content and metadata of a file
type FileCacheValue struct {
	content []string
	modTime time.Time
	size    int64
	readAt  time.Time
}

// FileCache is a thread-safe cache for file contents
type FileCache struct {
	cache       map[FileCacheKey]*FileCacheValue
	mu          sync.RWMutex
	maxSize     int64         // Max cache size in bytes
	currentSize int64         // Current cache size in bytes
	ttl         time.Duration // Time-to-live for cache entries
}

// NewFileCache creates a new file cache with specified capacity and TTL
func NewFileCache(maxSizeMB int, ttlMinutes int) *FileCache {
	return &FileCache{
		cache:   make(map[FileCacheKey]*FileCacheValue),
		maxSize: int64(maxSizeMB) * 1024 * 1024,
		ttl:     time.Duration(ttlMinutes) * time.Minute,
	}
}

// GetFileContent retrieves file content from cache or reads it from disk
func (fc *FileCache) GetFileContent(path string) ([]string, int64, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}

	// Check cache first
	key := FileCacheKey{
		path:    path,
		modTime: info.ModTime(),
	}

	fc.mu.RLock()
	if entry, exists := fc.cache[key]; exists {
		fc.mu.RUnlock()
		return entry.content, entry.size, nil
	}
	fc.mu.RUnlock()

	// Read file from disk with buffered I/O
	content, err := readFileWithBufferedIO(path)
	if err != nil {
		return nil, 0, err
	}

	// Create cache entry
	entry := &FileCacheValue{
		content: content,
		modTime: info.ModTime(),
		size:    info.Size(),
		readAt:  time.Now(),
	}

	// Add to cache with eviction logic
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Evict old entries if cache is full
	if fc.currentSize+info.Size() > fc.maxSize {
		fc.evictEntries()
	}

	fc.cache[key] = entry
	fc.currentSize += info.Size()

	return content, info.Size(), nil
}

// readFileWithBufferedIO reads a file using buffered I/O for efficiency
func readFileWithBufferedIO(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines (default is 64KB)
	buf := make([]byte, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, cap(buf))

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// evictEntries removes old entries to make space in the cache
func (fc *FileCache) evictEntries() {
	// Remove entries older than TTL
	now := time.Now()
	for key, entry := range fc.cache {
		if now.Sub(entry.readAt) > fc.ttl {
			delete(fc.cache, key)
			fc.currentSize -= entry.size
		}
	}

	// If still full, remove 20% of entries (oldest first)
	if fc.currentSize > fc.maxSize {
		var entries []struct {
			key   FileCacheKey
			entry *FileCacheValue
		}

		for key, entry := range fc.cache {
			entries = append(entries, struct {
				key   FileCacheKey
				entry *FileCacheValue
			}{key, entry})
		}

		// Sort by read time (oldest first)
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].entry.readAt.After(entries[j].entry.readAt) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Evict 20% of entries
		evictCount := len(entries) / 5
		for i := 0; i < evictCount; i++ {
			delete(fc.cache, entries[i].key)
			fc.currentSize -= entries[i].entry.size
		}
	}
}

// ClearCache clears all entries from the cache
func (fc *FileCache) ClearCache() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.cache = make(map[FileCacheKey]*FileCacheValue)
	fc.currentSize = 0
}

// FileSizeLimitExceeded checks if the file exceeds the specified size limit
func FileSizeLimitExceeded(path string, maxSizeMB int) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false // Treat errors as not exceeding limit
	}

	maxSize := int64(maxSizeMB) * 1024 * 1024
	return info.Size() > maxSize
}
