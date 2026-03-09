package analyzer

import (
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestViolationPool(t *testing.T) {
	// Get a violation from pool
	v := GetViolation()
	if v == nil {
		t.Fatal("Expected non-nil violation")
	}

	// Set some values
	v.File = "test.go"
	v.Line = 10
	v.Rule = "test-rule"
	v.Severity = "error"
	v.Excerpt = "test excerpt"
	v.Suggestion = "test suggestion"

	// Return to pool
	PutViolation(v)

	// Get another violation - should be the same object reset
	v2 := GetViolation()
	if v2 == nil {
		t.Fatal("Expected non-nil violation from pool")
	}

	// Check values are reset
	if v2.File != "" {
		t.Errorf("Expected File to be reset, got %q", v2.File)
	}
}

func TestStringBuilderPool(t *testing.T) {
	// Get a string builder from pool
	sb := GetStringBuilder()
	if sb == nil {
		t.Fatal("Expected non-nil string builder")
	}

	// Write something
	sb.WriteString("test content")

	// Return to pool
	PutStringBuilder(sb)

	// Get another - should be reset
	sb2 := GetStringBuilder()
	if sb2 == nil {
		t.Fatal("Expected non-nil string builder from pool")
	}

	// Should be empty
	if sb2.Len() != 0 {
		t.Errorf("Expected empty string builder, got %d bytes", sb2.Len())
	}
}

func TestBufferPool(t *testing.T) {
	// Get a buffer from pool
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("Expected non-nil buffer")
	}

	// Write something
	buf = append(buf, []byte("test content")...)

	// Return to pool
	PutBuffer(buf)

	// Get another - should be empty
	buf2 := GetBuffer()
	if buf2 == nil {
		t.Fatal("Expected non-nil buffer from pool")
	}

	// Should be empty
	if len(buf2) != 0 {
		t.Errorf("Expected empty buffer, got %d bytes", len(buf2))
	}
}

func TestViolationSlice(t *testing.T) {
	// Create a new violation slice
	vs := NewViolationSlice(10)
	if vs == nil {
		t.Fatal("Expected non-nil violation slice")
	}

	// Append some violations
	vs.Append(Violation{File: "test1.go", Line: 1})
	vs.Append(Violation{File: "test2.go", Line: 2})

	if vs.Len() != 2 {
		t.Errorf("Expected 2 violations, got %d", vs.Len())
	}

	// Reset and reuse
	vs.Reset()
	if vs.Len() != 0 {
		t.Errorf("Expected 0 after reset, got %d", vs.Len())
	}

	// Append more
	vs.Append(Violation{File: "test3.go", Line: 3})
	if vs.Len() != 1 {
		t.Errorf("Expected 1 after append, got %d", vs.Len())
	}
}

func TestViolationSlicePool(t *testing.T) {
	// Get from pool
	vs := GetViolationSlice()
	if vs == nil {
		t.Fatal("Expected non-nil violation slice from pool")
	}

	// Use it
	vs.Append(Violation{File: "test.go"})

	// Return to pool
	PutViolationSlice(vs)

	// Get again - should be reset
	vs2 := GetViolationSlice()
	if vs2.Len() != 0 {
		t.Errorf("Expected 0 from pool, got %d", vs2.Len())
	}
}

func TestMemoryMonitor(t *testing.T) {
	// Create monitor with limit
	mm := NewMemoryMonitor(1024, time.Second, nil)

	if mm == nil {
		t.Fatal("Expected non-nil memory monitor")
	}

	// Test GetCurrentMemory
	memBefore := mm.GetCurrentMemory()
	if memBefore < 0 {
		t.Errorf("Expected non-negative memory, got %d MB", memBefore)
	}

	// Test IsOverLimit when disabled (0 limit)
	mm2 := NewMemoryMonitor(0, time.Second, nil)
	if mm2.IsOverLimit() {
		t.Error("Expected false when disabled")
	}
}

func TestMemoryMonitorNotifications(t *testing.T) {
	var currentMB int64
	var maxMB int64

	notifyCalled := make(chan bool, 1)

	mm := NewMemoryMonitor(100, 50*time.Millisecond, func(cur, mx int64) {
		currentMB = cur
		maxMB = mx
		notifyCalled <- true
	})

	mm.Start()

	// Wait for notification
	select {
	case <-notifyCalled:
		// Good
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected notification within 200ms")
	}

	// Check values
	if currentMB < 0 {
		t.Errorf("Expected non-negative current memory, got %d", currentMB)
	}

	if maxMB != 100 {
		t.Errorf("Expected max memory 100, got %d", maxMB)
	}

	mm.Stop()
}

func TestConcurrentPoolAccess(t *testing.T) {
	var wg sync.WaitGroup

	// Test concurrent access to ViolationPool
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := GetViolation()
			// Simulate some work
			v.File = "test"
			PutViolation(v)
		}()
	}

	wg.Wait()
}

func BenchmarkViolationPool(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		v := GetViolation()
		v.File = "test"
		v.Line = 1
		v.Rule = "test"
		PutViolation(v)
	}
}

func BenchmarkViolationSlice(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		vs := GetViolationSlice()
		for j := 0; j < 10; j++ {
			vs.Append(Violation{File: "test.go", Line: j})
		}
		_ = vs.Slice()
		PutViolationSlice(vs)
	}
}

func BenchmarkStringBuilderPool(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sb := GetStringBuilder()
		sb.WriteString("test content ")
		sb.WriteString("more content")
		_ = sb.String()
		PutStringBuilder(sb)
	}
}

func BenchmarkMemoryMonitor(b *testing.B) {
	mm := NewMemoryMonitor(0, time.Second, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = mm.GetCurrentMemory()
	}
}

// Ensure strings.Builder is used correctly
func TestStringBuilderUsage(t *testing.T) {
	// Test that string builder is properly reused
	sb := GetStringBuilder()

	// Write using strings.Builder
	sb.WriteString("line1\n")
	sb.WriteString("line2\n")
	sb.WriteString("line3\n")

	content := sb.String()
	lines := strings.Split(content, "\n")

	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines, got %d", len(lines))
	}

	// Reset for pool return
	sb.Reset()
	PutStringBuilder(sb)

	// Verify GC works with pooled objects
	runtime.GC()
}
