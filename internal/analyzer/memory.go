package analyzer

import (
	"runtime"
	"strings"
	"sync"
	"time"
)

// ViolationPool is a sync.Pool for reusing Violation objects
var ViolationPool = sync.Pool{
	New: func() interface{} {
		return &Violation{}
	},
}

// GetViolation returns a Violation from the pool
func GetViolation() *Violation {
	return ViolationPool.Get().(*Violation)
}

// PutViolation returns a Violation to the pool
func PutViolation(v *Violation) {
	// Reset the violation before returning to pool
	v.File = ""
	v.Line = 0
	v.Rule = ""
	v.Severity = ""
	v.Excerpt = ""
	v.Suggestion = ""
	ViolationPool.Put(v)
}

// StringBuilderPool is a sync.Pool for reusing strings.Builder objects
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// GetStringBuilder returns a strings.Builder from the pool
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder returns a strings.Builder to the pool
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// BufferPool(sb *strings.Builder is a sync.Pool for reusing byte buffers
var BufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 8192) // 8KB initial capacity
	},
}

// GetBuffer returns a byte slice from the pool
func GetBuffer() []byte {
	return BufferPool.Get().([]byte)
}

// PutBuffer returns a byte slice to the pool
func PutBuffer(buf []byte) {
	// Only return if it has some capacity (avoid returning empty slices)
	if cap(buf) > 0 {
		// Create a new buffer from scratch to avoid pointer issues
		newBuf := make([]byte, 0, cap(buf))
		BufferPool.Put(newBuf)
	}
}

// MemoryMonitor tracks memory usage
type MemoryMonitor struct {
	enabled     bool
	interval    time.Duration
	maxMemoryMB int64
	lastCheck   time.Time
	notifyFunc  func(int64, int64) // current, max
	stopChan    chan struct{}
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(maxMemoryMB int64, interval time.Duration, notifyFunc func(int64, int64)) *MemoryMonitor {
	return &MemoryMonitor{
		enabled:     maxMemoryMB > 0,
		interval:    interval,
		maxMemoryMB: maxMemoryMB,
		notifyFunc:  notifyFunc,
		stopChan:    make(chan struct{}),
	}
}

// Start begins monitoring memory usage
func (m *MemoryMonitor) Start() {
	if !m.enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.check()
			case <-m.stopChan:
				return
			}
		}
	}()
}

// Stop stops the memory monitor
func (m *MemoryMonitor) Stop() {
	if !m.enabled {
		return
	}
	close(m.stopChan)
}

// check checks current memory usage
func (m *MemoryMonitor) check() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentMB := int64(memStats.Alloc / (1024 * 1024))
	m.lastCheck = time.Now()

	if m.notifyFunc != nil {
		m.notifyFunc(currentMB, m.maxMemoryMB)
	}
}

// GetCurrentMemory returns current memory usage in MB
func (m *MemoryMonitor) GetCurrentMemory() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Alloc / (1024 * 1024))
}

// IsOverLimit returns true if memory usage exceeds the limit
func (m *MemoryMonitor) IsOverLimit() bool {
	if !m.enabled {
		return false
	}
	return m.GetCurrentMemory() > m.maxMemoryMB
}

// ForceGC forces garbage collection
func (m *MemoryMonitor) ForceGC() {
	runtime.GC()
}

// ViolationSlice is a pre-allocated slice for violations
type ViolationSlice struct {
	data []Violation
	pos  int
}

// NewViolationSlice creates a new pre-allocated violation slice
func NewViolationSlice(capacity int) *ViolationSlice {
	return &ViolationSlice{
		data: make([]Violation, 0, capacity),
		pos:  0,
	}
}

// Reset clears the slice for reuse
func (vs *ViolationSlice) Reset() {
	vs.data = vs.data[:0]
	vs.pos = 0
}

// Append adds a violation to the slice
func (vs *ViolationSlice) Append(v Violation) {
	vs.data = append(vs.data, v)
	vs.pos++
}

// Slice returns the underlying slice
func (vs *ViolationSlice) Slice() []Violation {
	return vs.data
}

// Len returns the number of violations
func (vs *ViolationSlice) Len() int {
	return vs.pos
}

// ViolationSlicePool is a sync.Pool for ViolationSlice
var ViolationSlicePool = sync.Pool{
	New: func() interface{} {
		return NewViolationSlice(64) // Default capacity of 64
	},
}

// GetViolationSlice returns a ViolationSlice from the pool
func GetViolationSlice() *ViolationSlice {
	return ViolationSlicePool.Get().(*ViolationSlice)
}

// PutViolationSlice returns a ViolationSlice to the pool
func PutViolationSlice(vs *ViolationSlice) {
	vs.Reset()
	ViolationSlicePool.Put(vs)
}
